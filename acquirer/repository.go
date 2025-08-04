package acquirer

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/moov-io/ftdc-from-tap-to-auth/acquirer/models"
)

var ErrNotFound = fmt.Errorf("not found")

type persistedData struct {
	Merchants map[string]*models.Merchant `json:"merchants"`
	Payments  map[string]*models.Payment  `json:"payments"`
}

type Repository struct {
	mu sync.RWMutex

	merchants map[string]*models.Merchant
	payments  map[string]*models.Payment
}

func NewRepository() *Repository {
	return &Repository{
		merchants: make(map[string]*models.Merchant),
		payments:  make(map[string]*models.Payment),
	}
}

func (r *Repository) CreateMerchant(merchant *models.Merchant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.merchants[merchant.ID] = merchant

	return nil
}

func (r *Repository) GetMerchant(merchantID string) (*models.Merchant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	merchant, ok := r.merchants[merchantID]
	if !ok {
		return nil, ErrNotFound
	}

	return merchant, nil
}

func (r *Repository) CreatePayment(payment *models.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.payments[payment.ID] = payment

	return nil
}

func (r *Repository) GetPayment(merchantID, paymentID string) (*models.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	payment, ok := r.payments[paymentID]
	if !ok {
		return nil, ErrNotFound
	}

	if payment.MerchantID != merchantID {
		return nil, ErrNotFound
	}

	return payment, nil
}

func (r *Repository) GetPayments(merchantID string) ([]*models.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var payments []*models.Payment
	for _, payment := range r.payments {
		if payment.MerchantID == merchantID {
			payments = append(payments, payment)
		}
	}

	return payments, nil
}

const filename = "db/acquirer_data.json"

func (r *Repository) SaveToFile() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data := persistedData{
		Merchants: r.merchants,
		Payments:  r.payments,
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

	// Initialize maps if they're nil
	if persisted.Merchants != nil {
		r.merchants = persisted.Merchants
	} else {
		r.merchants = make(map[string]*models.Merchant)
	}

	if persisted.Payments != nil {
		r.payments = persisted.Payments
	} else {
		r.payments = make(map[string]*models.Payment)
	}

	return nil
}
