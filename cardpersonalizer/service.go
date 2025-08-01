package cardpersonalizer

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/moov-io/ftdc-from-tap-to-auth/cardpersonalizer/card"
	"github.com/moov-io/ftdc-from-tap-to-auth/cardpersonalizer/models"
	"github.com/moov-io/ftdc-from-tap-to-auth/terminal"
	"github.com/moov-io/iso8583/encoding"
	"golang.org/x/exp/slog"
)

type CardJob struct {
	Card   models.CardRequest
	Result chan CardResult
	State  models.JobState
}

type CardResult struct {
	Card models.CardResponse
	Err  error
}

type Service struct {
	cardReader *terminal.CardReader
	jobQueue   chan CardJob
	state      *JobState
	logger     *slog.Logger
	done       chan struct{}
	mu         sync.RWMutex
}

// JobState represents the current state of job processing
type JobState struct {
	jobQueue        []CardJob
	jobInProgress   bool
	cardPresentChan <-chan error
	cardRemovedChan <-chan error
}

func NewService(logger *slog.Logger, readerName string) (*Service, error) {
	cardReader, err := terminal.NewCardReader()
	if err != nil {
		return nil, fmt.Errorf("creating card reader: %w", err)
	}
	for _, reader := range cardReader.Readers {
		if strings.Contains(reader, readerName) {
			cardReader.SelectedReader = reader
			logger.Info("selected card reader", slog.String("reader", reader))
			break
		}
	}
	if cardReader.SelectedReader == "" {
		return nil, fmt.Errorf("no card reader found with name: %s", readerName)
	}

	return &Service{
		cardReader: cardReader,
		jobQueue:   make(chan CardJob, 100),
		state: &JobState{
			jobQueue: make([]CardJob, 0),
		},
		logger: logger,
		done:   make(chan struct{}),
	}, nil
}

func (s *Service) GetJobs() []models.CardJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.state == nil {
		return []models.CardJob{}
	}

	jobs := make([]models.CardJob, len(s.state.jobQueue))
	for i, job := range s.state.jobQueue {
		jobs[i] = models.CardJob{
			Name:  job.Card.Name,
			State: job.State,
		}
	}
	return jobs
}

func (s *Service) Run(ctx context.Context) {
	s.logger.Info("starting card personalizer service")
	defer s.logger.Info("stopped card personalizer service")

	for {
		select {
		case job, ok := <-s.jobQueue:
			if !ok {
				s.logger.Info("job queue closed, stopping service")
				return
			}
			s.dequeueJob(job)

		case err := <-s.state.cardPresentChan:
			s.handleCardPresent(err)

		case err := <-s.state.cardRemovedChan:
			s.handleCardRemoved(err)

		case <-ctx.Done():
			s.cardReader.Close()
			s.logger.Info("context cancelled, stopping service")
			return
		}
	}
}

func (s *Service) handleCardPresent(err error) {
	if !s.state.jobInProgress {
		return
	}

	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			s.logger.Info("card wait timeout...", slog.String("job", s.state.jobQueue[0].Card.Name))
			s.state.cardPresentChan = s.cardReader.WaitForCardAsync(time.Second * 5)
			return
		}

		s.mu.Lock()
		s.logger.Error("waiting for card", slog.String("error", err.Error()))
		s.state.jobQueue[0].Result <- CardResult{Err: err}
		s.state.jobQueue[0].State = models.JobStateFailed
		s.mu.Unlock()
		// Wait for card to be removed before processing next job
		s.state.cardRemovedChan = s.cardReader.WaitForCardRemoveAsync(time.Second * 5)
		return
	}

	s.mu.Lock()
	s.state.cardPresentChan = nil
	s.state.jobQueue[0].State = models.JobStateProcess
	s.mu.Unlock()
	s.logger.Info("processing job", slog.String("job", s.state.jobQueue[0].Card.Name), slog.String("state", string(s.state.jobQueue[0].State)))

	cardResponse, err := s.PersonalizeCard(s.state.jobQueue[0].Card)
	if err != nil {
		s.logger.Error("processing job", slog.String("job", s.state.jobQueue[0].Card.Name), slog.String("error", err.Error()))
		s.mu.Lock()
		s.state.jobQueue[0].Result <- CardResult{Err: err}
		s.state.jobQueue[0].State = models.JobStateFailed
		s.mu.Unlock()
		// Wait for card to be removed before processing next job
		s.state.cardRemovedChan = s.cardReader.WaitForCardRemoveAsync(time.Second * 5)
		return
	}

	s.mu.Lock()
	s.state.jobQueue[0].Result <- CardResult{Card: cardResponse}
	s.state.jobQueue[0].State = models.JobStateDone
	s.mu.Unlock()
	s.logger.Info("completed job", slog.String("job", s.state.jobQueue[0].Card.Name), slog.String("state", string(s.state.jobQueue[0].State)))

	// Wait for card to be removed before processing next job
	s.state.cardRemovedChan = s.cardReader.WaitForCardRemoveAsync(time.Second * 5)
}

func (s *Service) handleCardRemoved(err error) {
	if !s.state.jobInProgress {
		return
	}

	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			s.logger.Info("card remove wait timeout...", slog.String("job", s.state.jobQueue[0].Card.Name))
			s.state.cardRemovedChan = s.cardReader.WaitForCardRemoveAsync(time.Second * 5)
			return
		}
	}

	s.logger.Info("card removed", slog.String("job", s.state.jobQueue[0].Card.Name))
	// Remove the completed job from queue
	if len(s.state.jobQueue) > 0 {
		s.mu.Lock()
		s.state.jobQueue = s.state.jobQueue[1:]
		s.mu.Unlock()
	}

	s.startNextJob()
}

func (s *Service) startNextJob() {
	// Reset state
	s.mu.Lock()
	s.state.jobInProgress = false
	s.state.cardPresentChan = nil
	s.state.cardRemovedChan = nil
	s.mu.Unlock()

	// Start processing next job if available
	if len(s.state.jobQueue) > 0 {
		s.mu.Lock()
		s.state.jobInProgress = true
		s.mu.Unlock()
		s.logger.Info("starting next job", slog.String("job", s.state.jobQueue[0].Card.Name), slog.String("state", string(s.state.jobQueue[0].State)))
		s.state.cardPresentChan = s.cardReader.WaitForCardAsync(time.Second * 5)
	}
}

func (s *Service) dequeueJob(job CardJob) {
	// Set initial state to queue
	s.mu.Lock()
	job.State = models.JobStateQueue
	s.state.jobQueue = append(s.state.jobQueue, job)
	s.mu.Unlock()
	s.logger.Info("added job to queue", slog.String("job", job.Card.Name), slog.String("state", string(job.State)), slog.Int("queue_length", len(s.state.jobQueue)))

	// Start processing if no job is currently active
	if !s.state.jobInProgress {
		s.startNextJob()
	}
}

func (s *Service) EnqueueCardRequest(card models.CardRequest) <-chan CardResult {
	resultChan := make(chan CardResult, 1)
	s.jobQueue <- CardJob{
		Card:   card,
		Result: resultChan,
	}
	s.logger.Info("enqueuing card job", slog.String("job", card.Name))
	return resultChan
}

func (s *Service) PersonalizeCard(cardReq models.CardRequest) (models.CardResponse, error) {
	if err := cardReq.Validate(); err != nil {
		return models.CardResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	// Generate unique request ID for workspace isolation
	requestID := uuid.NewString()

	var expiry string
	if strings.Contains(cardReq.ExpiryDate, "/") {
		// TODO: update to YYMMDD , day is 31, even for Feb
		expiry = fmt.Sprintf("%s%s", cardReq.ExpiryDate[3:], cardReq.ExpiryDate[:2])
	} else {
		expiry = fmt.Sprintf("%s%s", cardReq.ExpiryDate[2:], cardReq.ExpiryDate[:2])
	}

	// Create operations for updating EMVStaticData.java
	operations := []card.Operation{
		{
			LineNumber:  73,
			Tags:        []string{"5A"},
			Value:       cardReq.PAN,
			Encoding:    encoding.BCD,
			Description: "PAN",
		},
		{
			LineNumber:  75,
			Tags:        []string{"5F", "24"},
			Value:       expiry,
			Encoding:    encoding.BCD,
			Description: "Expiry Date",
		},
		{
			LineNumber:  76,
			Tags:        []string{"5F", "20"},
			Value:       cardReq.Name,
			Encoding:    encoding.ASCII,
			Description: "Cardholder Name",
		},
		{
			LineNumber:  103,
			Tags:        []string{"00"},
			Value:       cardReq.PIN,
			Encoding:    encoding.BCD,
			Description: "PIN",
		},
	}

	if err := card.UpdateEMVStaticData(operations, requestID); err != nil {
		return models.CardResponse{}, fmt.Errorf("updating EMV data: %w", err)
	}

	if err := card.FlashCardWithBinary(requestID); err != nil {
		return models.CardResponse{}, fmt.Errorf("flashing card: %w", err)
	}

	return models.CardResponse{
		CardHolder: cardReq.Name,
		PAN:        cardReq.PAN,
		ExpiryDate: cardReq.ExpiryDate,
		PIN:        cardReq.PIN,
	}, nil
}
