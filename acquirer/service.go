package acquirer

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/moov-io/bertlv"
	"github.com/moov-io/ftdc-from-tap-to-auth/acquirer/models"
)

type Service struct {
	repo          *Repository
	iso8583Client ISO8583Client
}

type ISO8583Client interface {
	AuthorizePayment(payment *models.Payment, card models.CreatePayment, merchant models.Merchant) (models.AuthorizationResponse, error)
}

func NewService(repo *Repository, iso8583Client ISO8583Client) *Service {
	return &Service{
		repo:          repo,
		iso8583Client: iso8583Client,
	}
}

func (a *Service) CreateMerchant(create models.CreateMerchant) (*models.Merchant, error) {
	merchant := &models.Merchant{
		ID:         uuid.New().String(),
		Name:       create.Name,
		MCC:        create.MCC,
		PostalCode: create.PostalCode,
		WebSite:    create.WebSite,
	}

	err := a.repo.CreateMerchant(merchant)
	if err != nil {
		return nil, fmt.Errorf("creating merchant: %w", err)
	}

	return merchant, nil
}

type card struct {
	PAN            string `bertlv:"5A"`
	ExpirationDate string `bertlv:"5F24"` // YYMMDD format
	CardholderName string `bertlv:"5F20"`
}

func (a *Service) CreatePayment(merchantID string, create models.CreatePayment) (*models.Payment, error) {
	payment := &models.Payment{
		ID:         uuid.New().String(),
		MerchantID: merchantID,
		Amount:     create.Amount,
		Currency:   create.Currency,
		Status:     models.PaymentStatusPending,
		CreatedAt:  time.Now(),
	}

	// if we have emv payload, we will use it to extract card details
	if len(create.EMVPayload) != 0 {
		emvTags, err := bertlv.Decode(create.EMVPayload)
		if err != nil {
			return nil, fmt.Errorf("decoding EMV payload: %w", err)
		}

		c := &card{}

		err = bertlv.Unmarshal(emvTags, c)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling EMV tags: %w", err)
		}

		payment.Card = models.SafeCard{
			First6:         c.PAN[:6],
			Last4:          c.PAN[len(c.PAN)-4:],
			ExpirationDate: fmt.Sprintf("%s%s", c.ExpirationDate[2:4], c.ExpirationDate[:2]), // MMYY format
		}
	} else {
		// then it's e-commerce payment
		payment.Card = models.SafeCard{
			First6:         create.Card.Number[:6],
			Last4:          create.Card.Number[len(create.Card.Number)-4:],
			ExpirationDate: create.Card.ExpirationDate,
		}
	}

	err := a.repo.CreatePayment(payment)
	if err != nil {
		return nil, fmt.Errorf("creating payment: %w", err)
	}

	merchant, err := a.repo.GetMerchant(merchantID)
	if err != nil {
		return nil, fmt.Errorf("getting merchant: %w", err)
	}

	response, err := a.iso8583Client.AuthorizePayment(payment, create, *merchant)
	if err != nil {
		payment.Status = models.PaymentStatusError
		// update payment details
		return nil, fmt.Errorf("authorizing payment: %w", err)
	}

	payment.AuthorizationCode = response.AuthorizationCode

	if response.ApprovalCode == "00" {
		payment.Status = models.PaymentStatusAuthorized
	} else {
		payment.Status = models.PaymentStatusDeclined
	}

	return payment, nil
}

func (a *Service) GetPayment(merchantID, paymentID string) (*models.Payment, error) {
	payment, err := a.repo.GetPayment(merchantID, paymentID)
	if err != nil {
		return nil, fmt.Errorf("getting payment: %w", err)
	}

	return payment, nil
}
