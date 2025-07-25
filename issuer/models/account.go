package models

import (
	"errors"
	"sync"
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
