package models

import (
	"regexp"

	"github.com/moov-io/ftdc-from-tap-to-auth/cardpersonalizer/card"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// CardRequest represents the request body for forging a card
type CardRequest struct {
	Name       string `json:"name"`
	PAN        string `json:"pan"`
	ExpiryDate string `json:"expiry"`
	PIN        string `json:"pin"`
}

// Validate validates the CardRequest struct
func (c CardRequest) Validate() error {
	// Convert iterator to slice for validation
	var brandNames []string
	for brand := range card.CardNameToIndex {
		brandNames = append(brandNames, brand)
	}

	return validation.ValidateStruct(&c,
		validation.Field(&c.Name, validation.Required, validation.Length(1, 50)),
		validation.Field(&c.PAN, validation.Length(13, 19), validation.Match(regexp.MustCompile(`^[0-9]*$`))),
		validation.Field(&c.ExpiryDate, validation.Required, validation.Match(regexp.MustCompile(`^(0[1-9]|1[0-2])([0-9]{2})$`)).Error("must be in MMYY format")),
		validation.Field(&c.PIN, validation.Required, validation.Length(4, 4), validation.Match(regexp.MustCompile(`^[0-9]*$`))),
	)
}

// CardResponse represents the response when a card is personalized
type CardResponse struct {
	CardHolder string `json:"name"`
	PAN        string `json:"pan"`
	ExpiryDate string `json:"expiry"`
	PIN        string `json:"pin"`
}
