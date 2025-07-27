package models

import (
	"regexp"

	"github.com/moov-io/ftdc-from-tap-to-auth/cardpersonalizer/card"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Card struct {
	ID                    string `json:"id"`
	AccountID             string `json:"account_id"`
	CardHolderName        string `json:"card_holder_name"`
	Number                string `json:"pan"`
	ExpirationDate        string `json:"expiry"`
	CardVerificationValue string `json:"cvv"`
}

type CardRequest struct {
	ExpiryDate            string `json:"expiry"`
	CardVerificationValue string `json:"cvv"`
	PIN                   string `json:"pin"`
}

// Validate validates the CardRequest struct
func (c CardRequest) Validate() error {
	// Convert iterator to slice for validation
	var brandNames []string
	for brand := range card.CardNameToIndex {
		brandNames = append(brandNames, brand)
	}

	return validation.ValidateStruct(&c,
		validation.Field(&c.ExpiryDate, validation.Required, validation.Match(regexp.MustCompile(`^(0[1-9]|1[0-2])([0-9]{2})$`)).Error("must be in MMYY format")),
		validation.Field(&c.PIN, validation.Required, validation.Length(4, 4), validation.Match(regexp.MustCompile(`^[0-9]*$`))),
	)
}
