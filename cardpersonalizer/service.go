package cardpersonalizer

import (
	"fmt"
	"strings"

	"github.com/alovak/cardflow-playground/cardpersonalizer/card"
	"github.com/alovak/cardflow-playground/cardpersonalizer/models"
	"github.com/google/uuid"
	"github.com/moov-io/iso8583/encoding"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s Service) PersonalizeCard(cardReq models.CardRequest) (models.CardResponse, error) {
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
