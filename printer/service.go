package printer

import (
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Service struct {
	logger  *slog.Logger
	printer Printer

	printingDelay time.Duration
	mu            sync.Mutex // to protect the following fields

	queue []Receipt // queue of receipts to be printed

	wg sync.WaitGroup // to wait for all print jobs to finish

	printSignal chan bool // signal to start printing
	done        chan bool // signal to stop printing
}

func NewService(logger *slog.Logger, printer Printer) *Service {
	return &Service{
		logger:        logger,
		printer:       printer,
		printingDelay: 5 * time.Second, // default printing delay
		printSignal:   make(chan bool), // buffered channel to signal printing
		done:          make(chan bool), // channel to signal stopping the service
	}
}

func (s *Service) PrintReceipt(receipt Receipt) (*PrintJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.queue = append(s.queue, receipt)
	jobs := len(s.queue)

	waitingTime := (jobs - 1) * int(s.printingDelay.Seconds())

	s.logger.Info(
		"Queuing receipt for printing",
		slog.String("payment_id", receipt.PaymentID),
		slog.Int("number_in_queue", jobs),
		slog.Int("waiting_time_seconds", waitingTime),
	)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		select {
		case <-s.done:
			s.logger.Info(
				"Printing service is stopping, dropping receipt",
				slog.String("payment_id", receipt.PaymentID),
				slog.Int("number_in_queue", jobs),
				slog.Int("waiting_time_seconds", waitingTime),
			)
		case s.printSignal <- true:
		}
	}()

	return &PrintJob{
		NumberInQueue: jobs,
		WaitingTime:   waitingTime,
	}, nil
}

func (s *Service) Start() {
	s.logger.Info("Starting printing service...")

	s.wg.Add(1)
	// Start a goroutine to handle printing
	go func() {
		defer s.wg.Done()

		for {
			select {
			case <-s.done:
				s.mu.Lock()
				jobs := len(s.queue)
				s.mu.Unlock()

				s.logger.Info(
					"Printing service is stopping",
					slog.Int("jobs_dropped", jobs),
				)
				return
			case <-s.printSignal:
				s.mu.Lock()
				if len(s.queue) == 0 {
					s.mu.Unlock()
					continue
				}

				receipt := s.queue[0]
				s.queue = s.queue[1:] // remove the first receipt from the queue
				s.mu.Unlock()

				// Print the receipt
				if err := s.printReceipt(receipt); err != nil {
					s.logger.Error(
						"Failed to print receipt",
						slog.String("payment_id", receipt.PaymentID),
						slog.String("error", err.Error()),
					)
					continue
				}

				s.logger.Info("Printed receipt successfully", slog.String("payment_id", receipt.PaymentID))

				// Simulate printing delay
				time.Sleep(s.printingDelay)
			}
		}
	}()

	s.logger.Info("Printing service started")
}

func (s *Service) Stop() {
	s.logger.Info("Stopping printing service...")

	// Signal the goroutine to stop
	close(s.done)

	// Wait for the goroutine to finish
	s.wg.Wait()

	s.logger.Info("Printing service stopped")
}

func (s *Service) printReceipt(receipt Receipt) error {
	var err error

	// printing only name
	if receipt.Short {
		name := receipt.Cardholder
		if name == "" {
			return nil
		}

		// 3 digits, hopefully, space and name - max 32 chars
		if len(name) > 28 {
			name = name[:28]
		}

		err = s.printer.PrintLine(name)
		if err != nil {
			return fmt.Errorf("printing line for %s: %w", name, err)
		}
		err = s.printer.Feed(1)
		if err != nil {
			return fmt.Errorf("error feeding paper: %w", err)
		}

		return nil
	}

	logo, err := NewLogoBitmap()
	if err != nil {
		s.logger.Error("error creating logo bitmap", slog.String("error", err.Error()))

		err = s.printer.PrintCentered("*** Fintech DevCon 2025 ***")
		if err != nil {
			return fmt.Errorf("error printing title: %w", err)
		}
	} else {
		err = s.printer.PrintBitmapImage(logo)
		if err != nil {
			fmt.Printf("error printing logo bitmap: %v\n", err)
		}
	}

	err = s.printer.Feed(1)
	if err != nil {
		return fmt.Errorf("error feeding paper: %w", err)
	}
	p := message.NewPrinter(language.English)
	dollars := float64(receipt.Amount) / 100
	amount := p.Sprintf("%.2f", dollars)

	lines := []string{
		fmt.Sprintf("Date, time: %s", receipt.ProcessingDateTime.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("PAN: %s", receipt.PAN),
		fmt.Sprintf("Cardholder: %s", receipt.Cardholder),
		fmt.Sprintf("Amount: %s", amount),
		fmt.Sprintf("Auth Code: %s", receipt.AuthorizationCode),
		fmt.Sprintf("Resp Code: %s", receipt.ResponseCode),
	}

	err = s.printer.PrintLines(lines)
	if err != nil {
		return fmt.Errorf("error printing lines: %w", err)
	}

	err = s.printer.Feed(1)
	if err != nil {
		return fmt.Errorf("error feeding paper: %w", err)
	}

	phrase := motivationalPhrases[rand.Intn(len(motivationalPhrases))]

	err = s.printer.PrintCentered(fmt.Sprintf("* %s *", phrase))
	if err != nil {
		return fmt.Errorf("error printing title: %w", err)
	}

	err = s.printer.Feed(3)
	if err != nil {
		return fmt.Errorf("error feeding paper: %w", err)
	}

	err = s.printer.Cut()
	if err != nil {
		return fmt.Errorf("error cutting paper: %w", err)
	}

	return nil
}
