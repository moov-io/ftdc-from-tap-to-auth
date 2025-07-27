package paycard

import (
	"reflect"
	"testing"
)

// TestCalaculateSFI tests the CalculateSFI function
func TestCalculateSFI(t *testing.T) {
	tests := []struct {
		input    byte
		expected byte
	}{
		{0x01, 0x0C},
		{0x02, 0x14},
		{0x03, 0x1C},
		{0x04, 0x24},
		{0x05, 0x2C},
	}

	for _, test := range tests {
		result := CalculateSFI(test.input)
		if result != test.expected {
			t.Errorf("CalculateSFI(%X) = %X; want %X", test.input, result, test.expected)
		}
	}
}

// TestGenerateReadCommands tests the GenerateReadCommands function
// It checks if the generated commands match the expected output for various AFL inputs.
// The AFL (Application File Locator) is a byte array that contains information about the files to be read.
// The function generates READ RECORD commands based on the AFL input.
// Each command is a byte array that follows the EMV standard for reading records from a card.
func TestGenerateReadCommands(t *testing.T) {
	tests := []struct {
		input    []byte
		expected [][]byte
	}{
		{
			input: []byte{0x08, 0x01, 0x01, 0x00},
			expected: [][]byte{
				{0x00, 0xB2, 0x01, 0x0C, 0x00},
			},
		},
		{
			input: []byte{0x10, 0x02, 0x02, 0x01},
			expected: [][]byte{
				{0x00, 0xB2, 0x02, 0x14, 0x00},
			},
		},
		{
			input: []byte{0x20, 0x01, 0x02, 0x00},
			expected: [][]byte{
				{0x00, 0xB2, 0x01, 0x24, 0x00},
				{0x00, 0xB2, 0x02, 0x24, 0x00},
			},
		},
	}

	for _, test := range tests {
		result := GenerateReadCommands(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("GenerateReadCommands(% X) = % X; want % X", test.input, result, test.expected)
		}
	}
}
