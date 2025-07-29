package terminal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ReaderIndex   int    `yaml:"reader_index"`   // Index of reader to use, -1 for interactive selection
	MerchantID    string `yaml:"merchant_id"`    // ID of merchant to create payment for
	AcquirerURL   string `yaml:"acquirer_url"`   // URL of the acquirer service
	DefaultAmount int64  `yaml:"default_amount"` // Default amount for payments
}

func NewConfigFromFile(filePath string) (*Config, error) {
	config := &Config{}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}
	return config, nil
}
