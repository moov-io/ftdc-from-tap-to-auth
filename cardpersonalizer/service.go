package cardpersonalizer

import (
	"context"
	"fmt"
	"strings"
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
}

// JobState represents the current state of job processing
type JobState struct {
	jobQueue        []CardJob
	currentJob      *CardJob
	cardPresentChan <-chan error
	cardRemovedChan <-chan error
}

func NewService(logger *slog.Logger, readerName string) (*Service, error) {
	cardReader, err := terminal.NewCardReader()
	if err != nil {
		return nil, fmt.Errorf("creating card reader: %w", err)
	}
	for _, reader := range cardReader.Readers {
		if reader == readerName {
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

func (s *Service) GetJobs() []string {
	// Return a copy of the job queue to avoid race conditions
	jobs := make([]string, len(s.state.jobQueue))
	for i, job := range s.state.jobQueue {
		jobs[i] = job.Card.Name
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
			s.dequeueJob(job, s.state)

		case err := <-s.state.cardPresentChan:
			s.handleCardPresent(err, s.state)

		case err := <-s.state.cardRemovedChan:
			s.handleCardRemoved(err, s.state)

		case <-ctx.Done():
			s.logger.Info("context cancelled, stopping service")
			return
		}
	}
}

func (s *Service) dequeueJob(job CardJob, state *JobState) {
	state.jobQueue = append(state.jobQueue, job)
	s.logger.Info("added job to queue", slog.String("job", job.Card.Name), slog.Int("queue_length", len(state.jobQueue)))

	// Start processing if no job is currently active
	if state.currentJob == nil {
		s.startNextJob(state)
	}
}

func (s *Service) handleCardPresent(err error, state *JobState) {
	if state.currentJob == nil {
		return
	}

	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			s.logger.Info("card wait timeout...", slog.String("job", state.currentJob.Card.Name))
			state.cardPresentChan = s.cardReader.WaitForCardAsync(time.Second * 5)
			return
		}

		s.logger.Error("waiting for card", slog.String("error", err.Error()))
		state.currentJob.Result <- CardResult{Err: err}
		s.startNextJob(state)
		return
	}

	state.cardPresentChan = nil
	s.logger.Info("processing job", slog.String("job", state.currentJob.Card.Name))

	cardResponse, err := s.PersonalizeCard(state.currentJob.Card)
	if err != nil {
		s.logger.Error("processing job", slog.String("job", state.currentJob.Card.Name), slog.String("error", err.Error()))
		state.currentJob.Result <- CardResult{Err: err}
		s.startNextJob(state)
		return
	}

	state.currentJob.Result <- CardResult{Card: cardResponse}
	s.logger.Info("completed job", slog.String("job", state.currentJob.Card.Name))

	// Wait for card to be removed before processing next job
	state.cardRemovedChan = s.cardReader.WaitForCardRemoveAsync(time.Second * 5)
}

func (s *Service) handleCardRemoved(err error, state *JobState) {
	if state.currentJob == nil {
		return
	}

	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			s.logger.Info("card remove wait timeout...", slog.String("job", state.currentJob.Card.Name))
			state.cardRemovedChan = s.cardReader.WaitForCardRemoveAsync(time.Second * 5)
			return
		}
	}

	// Remove the completed job from queue
	if len(state.jobQueue) > 0 {
		state.jobQueue = state.jobQueue[1:]
	}

	s.logger.Info("card removed", slog.String("job", state.currentJob.Card.Name))
	s.startNextJob(state)
}

func (s *Service) startNextJob(state *JobState) {
	// Reset state
	state.currentJob = nil
	state.cardPresentChan = nil
	state.cardRemovedChan = nil

	// Start processing next job if available
	if len(state.jobQueue) > 0 {
		state.currentJob = &state.jobQueue[0]
		s.logger.Info("starting next job", slog.String("job", state.currentJob.Card.Name))
		state.cardPresentChan = s.cardReader.WaitForCardAsync(time.Second * 5)
	}
}

func (s *Service) SubmitCardRequest(card models.CardRequest) <-chan CardResult {
	resultChan := make(chan CardResult, 1)
	s.jobQueue <- CardJob{
		Card:   card,
		Result: resultChan,
	}
	s.logger.Info("submitting card job", slog.String("job", card.Name))
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
