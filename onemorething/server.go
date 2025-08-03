package onemorething

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"

	"github.com/moov-io/iso8583"
	connection "github.com/moov-io/iso8583-connection"
)

type server struct {
	logger  *slog.Logger
	addr    string
	wg      sync.WaitGroup
	closed  chan bool
	ln      net.Listener
	mu      sync.Mutex
	counter int
}

func NewServer(logger *slog.Logger, addr string) *server {
	if addr == "" {
		addr = ":"
	}

	return &server{
		logger: logger,
		closed: make(chan bool),
		addr:   addr,
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
	ParticipantName string `iso8583:"25"`
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
		slog.Int("num", s.counter),
	)

	message.MTI("0110")

	err = conn.Reply(message)
	if err != nil {
		s.logger.Error(
			"Error to sending reply",
			slog.String("error", err.Error()),
		)
	}

	if request.ParticipantName == "" {
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

	// TODO: send to printer service
}
