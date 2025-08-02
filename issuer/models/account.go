package models

import (
	"errors"
	"sync"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

type CreateAccount struct {
	OwnerName string `json:"owner"`
	Balance   int64  `json:"balance"`
	Currency  string `json:"currency"`
}

type Account struct {
	ID               string `json:"id"`
	OwnerName        string `json:"owner"`
	AvailableBalance int64  `json:"balance"`
	HoldBalance      int64  `json:"hold_balance"`
	Currency         string `json:"currency"`

	mu sync.Mutex
}

func (a *Account) Hold(amount int64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.AvailableBalance < amount {
		return ErrInsufficientFunds
	}

	a.AvailableBalance -= amount
	a.HoldBalance += amount

	return nil
}

// Validate validates the CreateAccount struct
func (c CreateAccount) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.OwnerName, validation.Required, validation.Length(1, 26), validation.By(isASCII)),
		validation.Field(&c.Balance,
			validation.Required,
			validation.Min(int64(1)).Error("must be greater than 0"),
			validation.Max(int64(100000001)).Error("must be at most 100000000"),
		),
		validation.Field(&c.Currency, validation.Required, validation.Length(3, 3), validation.In("USD").Error("must be USD")),
	)
}

func isASCII(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil // skip validation if not a string
	}
	for _, c := range str {
		if c > 127 {
			return errors.New("must contain only ASCII characters")
		}
	}
	return nil
}
