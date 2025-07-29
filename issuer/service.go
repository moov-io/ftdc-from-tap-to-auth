package issuer

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	cardpersonalizer "github.com/moov-io/ftdc-from-tap-to-auth/cardpersonalizer/client"
	cpm "github.com/moov-io/ftdc-from-tap-to-auth/cardpersonalizer/models"
	"github.com/moov-io/ftdc-from-tap-to-auth/issuer/models"
	"golang.org/x/exp/slog"
)

type Service struct {
	logger           *slog.Logger
	repo             *Repository
	cardpersonalizer *cardpersonalizer.Client
}

func NewService(logger *slog.Logger, repo *Repository, cardpersonalizer *cardpersonalizer.Client) *Service {
	return &Service{
		logger:           logger,
		repo:             repo,
		cardpersonalizer: cardpersonalizer,
	}
}

func (i *Service) CreateAccount(req models.CreateAccount) (*models.Account, error) {
	account := &models.Account{
		ID:               uuid.New().String(),
		OwnerName:        req.OwnerName,
		AvailableBalance: req.Balance,
		Currency:         req.Currency,
	}

	err := i.repo.CreateAccount(account)
	if err != nil {
		return nil, fmt.Errorf("creating account: %w", err)
	}

	return account, nil
}

func (i *Service) GetAccounts() ([]models.Account, error) {
	accounts, err := i.repo.GetAccounts()
	if err != nil {
		return nil, fmt.Errorf("getting accounts: %w", err)
	}

	return accounts, nil
}

func (i *Service) GetAccount(accountID string) (*models.Account, error) {
	account, err := i.repo.GetAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("finding account: %w", err)
	}

	return account, nil
}

func (i *Service) IssueCard(accountID string, cardRequest models.CardRequest, shouldPersonalize bool) (*models.Card, error) {
	account, err := i.repo.GetAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("finding account: %w", err)
	}

	card := &models.Card{
		ID:                    uuid.New().String(),
		AccountID:             accountID,
		CardHolderName:        account.OwnerName,
		CardVerificationValue: cardRequest.CardVerificationValue,
		ExpirationDate:        cardRequest.ExpiryDate,
	}

	// TODO: hardcode number so in emulator mode, we can test without cardpersonalizer
	card.Number = models.GenerateCardNumber("7")

	if shouldPersonalize {
		cr := cpm.CardRequest{
			Name:       account.OwnerName,
			ExpiryDate: cardRequest.ExpiryDate,
			PAN:        card.Number,
			PIN:        cardRequest.PIN,
		}

		_, err := i.cardpersonalizer.PersonalizeCard(cr)
		if err != nil {
			return nil, fmt.Errorf("personalizing card: %w", err)
		}
	}

	err = i.repo.CreateCard(card)
	if err != nil {
		return nil, fmt.Errorf("creating card: %w", err)
	}

	return card, nil
}

// ListTransactions returns a list of transactions for the given account ID.
func (i *Service) ListTransactions(accountID string) ([]*models.Transaction, error) {
	transactions, err := i.repo.ListTransactions(accountID)
	if err != nil {
		return nil, fmt.Errorf("listing transactions: %w", err)
	}

	return transactions, nil
}

func (i *Service) AuthorizeRequest(req models.AuthorizationRequest) (models.AuthorizationResponse, error) {
	// we can do whatever we want here with EMVPayload
	// req.EMVPayload => ...
	i.logger.Info(
		"authorizing request",
		slog.Int64("amount", req.Amount),
		slog.String("currency", req.Currency),
		slog.String("merchant", req.Merchant.Name),
		slog.String("pan", req.Card.Number),
		slog.String("expiration date", req.Card.ExpirationDate),
		slog.String("card holder name", req.Card.CardHolderName),
	)

	card, err := i.repo.FindCardForAuthorization(req.Card)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return models.AuthorizationResponse{
				ApprovalCode: models.ApprovalCodeInvalidCard,
			}, nil
		}

		return models.AuthorizationResponse{}, fmt.Errorf("finding card: %w", err)
	}

	account, err := i.repo.GetAccount(card.AccountID)
	if err != nil {
		return models.AuthorizationResponse{}, fmt.Errorf("finding account: %w", err)
	}

	transaction := &models.Transaction{
		ID:        uuid.New().String(),
		AccountID: card.AccountID,
		CardID:    card.ID,
		Amount:    req.Amount,
		Currency:  req.Currency,
		Merchant:  req.Merchant,
	}

	err = i.repo.CreateTransaction(transaction)
	if err != nil {
		return models.AuthorizationResponse{}, fmt.Errorf("creating transaction: %w", err)
	}

	// hold the funds on the account
	err = account.Hold(req.Amount)
	if err != nil {
		// handle insufficient funds
		if !errors.Is(err, models.ErrInsufficientFunds) {
			return models.AuthorizationResponse{}, fmt.Errorf("holding funds: %w", err)
		}

		return models.AuthorizationResponse{
			ApprovalCode: models.ApprovalCodeInsufficientFunds,
		}, nil
	}

	transaction.ApprovalCode = models.ApprovalCodeApproved
	transaction.AuthorizationCode = generateAuthorizationCode()
	transaction.Status = models.TransactionStatusAuthorized

	return models.AuthorizationResponse{
		AuthorizationCode: transaction.AuthorizationCode,
		ApprovalCode:      transaction.ApprovalCode,
	}, nil
}

func generateAuthorizationCode() string {
	return generateRandomNumber(6)
}

func generateRandomNumber(length int) string {
	rand.Seed(time.Now().UnixNano())

	// Generate a 6-digit random number
	randomDigits := make([]int, length)
	for i := 0; i < len(randomDigits); i++ {
		randomDigits[i] = rand.Intn(10)
	}

	var number string
	for _, digit := range randomDigits {
		number += fmt.Sprintf("%d", digit)
	}

	return number
}
