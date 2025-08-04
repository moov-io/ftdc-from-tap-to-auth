package issuer

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/moov-io/ftdc-from-tap-to-auth/issuer/models"
)

var ErrNotFound = fmt.Errorf("not found")

type persistedData struct {
	Cards        []*models.Card        `json:"cards"`
	Accounts     []*models.Account     `json:"accounts"`
	Transactions []*models.Transaction `json:"transactions"`
}

type Repository struct {
	Cards        []*models.Card
	Accounts     []*models.Account
	Transactions []*models.Transaction

	mu sync.RWMutex
}

func NewRepository() *Repository {
	return &Repository{
		Cards:        make([]*models.Card, 0),
		Accounts:     make([]*models.Account, 0),
		Transactions: make([]*models.Transaction, 0),
	}
}

func (r *Repository) CreateAccount(account *models.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Accounts = append(r.Accounts, account)

	return nil
}

func (r *Repository) GetAccounts() ([]models.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	accounts := make([]models.Account, 0)
	for _, account := range r.Accounts {
		accounts = append(accounts, models.Account{
			ID:               account.ID,
			OwnerName:        account.OwnerName,
			AvailableBalance: account.AvailableBalance,
			HoldBalance:      account.HoldBalance,
			Currency:         account.Currency,
		})
	}
	return accounts, nil
}

func (r *Repository) GetAccount(accountID string) (*models.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, account := range r.Accounts {
		if account.ID == accountID {
			return account, nil
		}
	}

	return nil, ErrNotFound
}

func (r *Repository) CreateCard(card *models.Card) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Cards = append(r.Cards, card)

	return nil
}

func (r *Repository) FindCardForAuthorization(card models.Card) (*models.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.Cards {
		match := c.Number == card.Number

		if match {
			return c, nil
		}
	}

	return nil, ErrNotFound
}

func (r *Repository) CreateTransaction(transaction *models.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Transactions = append(r.Transactions, transaction)

	return nil
}

// ListTransactions returns all transactions for a given account ID.
func (r *Repository) ListTransactions(accountID string) ([]*models.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	transactions := make([]*models.Transaction, 0)

	for _, transaction := range r.Transactions {
		if transaction.AccountID == accountID {
			transactions = append(transactions, transaction)
		}
	}

	return transactions, nil
}

const filename = "db/issuer_data.json"

func (r *Repository) SaveToFile() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data := persistedData{
		Cards:        r.Cards,
		Accounts:     r.Accounts,
		Transactions: r.Transactions,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}

func (r *Repository) LoadFromFile() error {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's okay
		}
		return err
	}

	var persisted persistedData
	if err := json.Unmarshal(data, &persisted); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.Cards = persisted.Cards
	r.Accounts = persisted.Accounts
	r.Transactions = persisted.Transactions

	return nil
}
