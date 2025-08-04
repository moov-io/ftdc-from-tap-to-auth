package onemorething

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/moov-io/ftdc-from-tap-to-auth/printer"
	"github.com/moov-io/iso8583"
	connection "github.com/moov-io/iso8583-connection"
)

type server struct {
	logger     *slog.Logger
	addr       string
	wg         sync.WaitGroup
	closed     chan bool
	ln         net.Listener
	mu         sync.Mutex
	counter    int
	printerURL string
}

func NewServer(logger *slog.Logger, addr, printerURL string) *server {
	if addr == "" {
		addr = ":"
	}

	return &server{
		logger:     logger,
		closed:     make(chan bool),
		addr:       addr,
		printerURL: printerURL,
	}
}

func (s *server) Start() error {
	s.logger.Info("Starting ISO 8583 server...")

	var err error

	s.ln, err = net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("creating listener: %w", err)
	}

	s.logger.Info(
		"Listening...",
		slog.String("addr", s.ln.Addr().String()),
	)

	s.wg.Add(1)
	go func() {
		for {
			conn, err := s.ln.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					break
				}

				s.logger.Error(
					"Error accepting connection",
					slog.String("error", err.Error()),
				)
				break
			}

			s.wg.Add(1)
			go s.handleConn(conn)
		}
		s.logger.Info("Stopped listening")
		s.wg.Done()
	}()

	return nil
}

func (s *server) Stop() error {
	s.logger.Info("Stopping ISO 8583 server...")

	close(s.closed)

	s.ln.Close()

	s.wg.Wait()

	s.logger.Info("ISO 8583 server stopped")

	return nil
}

func (s *server) handleConn(conn net.Conn) {
	defer s.wg.Done()

	s.logger.Info(
		"Accepted new connection",
		slog.String("string", conn.RemoteAddr().String()),
	)

	isoConn, err := connection.NewFrom(
		conn,
		spec,
		func(r io.Reader) (int, error) {
			header := make([]byte, 2)
			_, err := r.Read(header)
			if err != nil {
				// if errors.Is(err, os.ErrDeadlineExceeded) {
				// 	return 0, err
				// }
				// s.logger.Error(
				// 	"Failed to read length header",
				// 	slog.String("error", err.Error()),
				// )
				return 0, fmt.Errorf("reading length header: %w", err)
			}

			length := int(header[0])<<8 | int(header[1])
			return length, nil
		},
		func(w io.Writer, length int) (int, error) {
			header := make([]byte, 2)

			header[0] = byte(length >> 8)
			header[1] = byte(length)

			n, err := w.Write(header)
			if err != nil {
				s.logger.Error(
					"Failed to write length header",
					slog.String("error", err.Error()),
				)
				return 0, fmt.Errorf("writing length header: %w", err)
			}

			return n, nil
		},
		connection.InboundMessageHandler(s.handleRequest),
	)
	if err != nil {
		s.logger.Error(
			"Error creating ISO 8583 connection",
			slog.String("error", err.Error()),
		)
	}

	select {
	case <-s.closed:
	case <-isoConn.Done():
		s.logger.Info(
			"Connection closed",
			slog.String("string", conn.RemoteAddr().String()),
		)
	}

	conn.Close()
}

type requestData struct {
	MTI             string `iso8583:"0"`
	PAN             string `iso8583:"2"`
	Amount          int64  `iso8583:"4"`
	ParticipantName string `iso8583:"25"`
}

type responseData struct {
	MTI               string `iso8583:"0"`
	ApprovalCode      string `iso8583:"5"`
	AuthorizationCode string `iso8583:"6"`
}

func (s *server) handleRequest(conn *connection.Connection, message *iso8583.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++

	request := &requestData{}

	err := message.Unmarshal(request)
	if err != nil {
		s.logger.Error(
			"Error to handle incoming request",
			slog.String("error", err.Error()),
		)

		return
	}

	s.logger.Info(
		"Received request",
		slog.String("name", request.ParticipantName),
		slog.String("pan", request.PAN),
		slog.Int64("amount", request.Amount),
		slog.Int("position", s.counter),
	)

	response := &responseData{
		MTI:               "0110",
		ApprovalCode:      "00",
		AuthorizationCode: fmt.Sprintf("%06d", s.counter),
	}

	err = message.Marshal(response)
	if err != nil {
		s.logger.Error(
			"Error to marshal response",
			slog.String("error", err.Error()),
		)
		return
	}

	err = conn.Reply(message)
	if err != nil {
		s.logger.Error(
			"Error to sending reply",
			slog.String("error", err.Error()),
		)
	}

	if request.ParticipantName == "" {
		s.logger.Warn("Participant name is empty, skipping printing")
		return
	}

	s.logger.Info(
		"Replied",
		slog.String("name", request.ParticipantName),
	)

	f, err := os.OpenFile("backup.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		s.logger.Error("Failed to open backup file")
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("%d - %s\n", s.counter, request.ParticipantName))

	err = s.printParticipant(request, response)
	if err != nil {
		s.logger.Error(
			"printing participant",
			slog.String("error", err.Error()),
		)
		return
	}
}

func (s *server) printParticipant(request *requestData, response *responseData) error {
	if s.printerURL == "" {
		return fmt.Errorf("printer is not configured")
	}

	receipt := printer.Receipt{
		Cardholder:         request.ParticipantName,
		Short:              false,
		ProcessingDateTime: time.Now(),
		PAN:                request.PAN,
		Amount:             request.Amount,
		AuthorizationCode:  response.AuthorizationCode,
		ResponseCode:       response.ApprovalCode,
	}

	receiptJSON, err := json.Marshal(receipt)
	if err != nil {
		return fmt.Errorf("marshalling receipt: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/receipts", s.printerURL), bytes.NewReader(receiptJSON))
	if err != nil {
		return fmt.Errorf("creating request to printer: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sending receipt to printer: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("printer returned error: %s; status code: %d", responseBody, resp.StatusCode)
	}

	var printJob printer.PrintJob
	err = json.NewDecoder(resp.Body).Decode(&printJob)
	if err != nil {
		return fmt.Errorf("decoding print job response: %w", err)
	}

	s.logger.Info("printer job accepted",
		slog.Int("queue", printJob.NumberInQueue),
		slog.Int("waiting", printJob.WaitingTime),
	)

	return nil
}
