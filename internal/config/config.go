package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func NewFromFile[T any](filePath string, config *T) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	err = yaml.Unmarshal(content, config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return nil
}
