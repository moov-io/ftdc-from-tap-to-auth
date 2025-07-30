package terminal

type Config struct {
	ReaderIndex   int    `yaml:"reader_index"`   // Index of reader to use, -1 for interactive selection
	MerchantID    string `yaml:"merchant_id"`    // ID of merchant to create payment for
	AcquirerURL   string `yaml:"acquirer_url"`   // URL of the acquirer service
	DefaultAmount int64  `yaml:"default_amount"` // Default amount for payments
}

func DefaultConfig() *Config {
	return &Config{
		ReaderIndex:   -1,                      // Use interactive selection by default
		MerchantID:    "",                      // No default merchant ID
		AcquirerURL:   "http://localhost:8080", // Default URL for acquirer service
		DefaultAmount: 100,                     // Default amount of 1.00 in minor units (e.g., cents)
	}
}
