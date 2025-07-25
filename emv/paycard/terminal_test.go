package paycard

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	// 9F66 04 Terminal Transaction Qualifier
	t9F66         = []byte{0x9F, 0x66, 0x04}
	t9F66Response = []byte{0xB6, 0x00, 0xC0, 0x00}
	// 9F02 06 Authorized Amount
	t9F02         = []byte{0x9F, 0x02, 0x06}
	t9F02Response = []byte{0x00, 0x00, 0x00, 0x00, 0x12, 0x34} // 1234
	// 9F03 06 Secondary Amount
	t9F03         = []byte{0x9F, 0x03, 0x06}
	t9F03Response = []byte{0x00, 0x00, 0x00, 0x00, 0x56, 0x78} // 5678
	// 9F1A 02 Terminal Country Code
	t9F1A         = []byte{0x9F, 0x1A, 0x02}
	t9F1AResponse = []byte{0x08, 0x40} // 0840
	// 95 05 Terminal Verification Results
	t95         = []byte{0x95, 0x05}
	t95Response = []byte{0x00, 0x00, 0x00, 0x00, 0x00} // 0000000000
	// 5F2A 02 Transaction Currency Code
	t5F2A         = []byte{0x5F, 0x2A, 0x02}
	t5F2AResponse = []byte{0x08, 0x40} // 0840
	// 9A 03 Transaction Date
	t9A         = []byte{0x9A, 0x03}
	t9AResponse = []byte{0x21, 0x01, 0x01} // 210101
	// 9C 01 Transaction Type
	t9C         = []byte{0x9C, 0x01}
	t9CResponse = []byte{0x00} // 00
	// 9F37 04 Unpredictable Number
	t9F37         = []byte{0x9F, 0x37, 0x04}
	t9F37Response = []byte{0xA1, 0xB2, 0xC3, 0xD4} // A1B2C3D4
)

func TestBuildPDOLData_TerminalTransactionQualifier(t *testing.T) {
	data := t9F66
	pdol, _ := ParseDOL(data)

	terminal, _ := NewTerminal()
	s := Transaction{}

	result := terminal.BuildPDOLData(&s, pdol)
	if len(result) != 4 {
		t.Errorf("Expected 4 bytes, got %d", len(result))
	}
	require.Equal(t, t9F66Response, result)
}

func TestBuildPDOLData_AuthorizedAmount(t *testing.T) {
	data := t9F02
	pdol, _ := ParseDOL(data)
	terminal, _ := NewTerminal()
	s := Transaction{
		AuthorizedAmount: "1234",
	}
	result := terminal.BuildPDOLData(&s, pdol)
	require.Len(t, result, 6)
	expected, _ := hex.DecodeString("000000001234")
	require.Equal(t, expected, result)
}

func TestBuildPDOLData_SecondaryAmount(t *testing.T) {
	data := t9F03
	pdol, _ := ParseDOL(data)
	terminal, _ := NewTerminal()
	s := Transaction{
		SecondaryAmount: "5678",
	}
	result := terminal.BuildPDOLData(&s, pdol)
	require.Len(t, result, 6)
	expected, _ := hex.DecodeString("000000005678")
	require.Equal(t, expected, result)
}

func TestBuildPDOLData_TerminalCountryCode(t *testing.T) {
	data := []byte{0x9F, 0x1A, 0x02}
	pdol, _ := ParseDOL(data)
	terminal, _ := NewTerminal(WithCountryCode("0840"))
	s := Transaction{}
	result := terminal.BuildPDOLData(&s, pdol)
	require.Len(t, result, 2)
	expected, _ := hex.DecodeString(DefaultCountryCode)
	require.Equal(t, expected, result)
}

func TestBuildPDOLData_TerminalVerificationResults(t *testing.T) {
	// Terminal Verification Results
	// 95 05
	data := []byte{0x95, 0x05}
	pdol, _ := ParseDOL(data)

	terminal, _ := NewTerminal()
	s := Transaction{}
	result := terminal.BuildPDOLData(&s, pdol)
	require.Len(t, result, 5)
	expected, _ := hex.DecodeString(DefaultVerificationResults)
	require.Equal(t, expected, result)
}

func TestBuildPDOLData_TransactionCurrencyCode(t *testing.T) {
	// Transaction Currency Code
	// 5F2A 02
	data := []byte{0x5F, 0x2A, 0x02}
	pdol, _ := ParseDOL(data)

	terminal, _ := NewTerminal()
	s := Transaction{}
	result := terminal.BuildPDOLData(&s, pdol)
	require.Len(t, result, 2)
	expected, _ := hex.DecodeString("0840")
	require.Equal(t, expected, result)
}

func TestBuildPDOLData_TerminalTransactionDate(t *testing.T) {
	// Transaction Date
	// 9A 03
	data := []byte{0x9A, 0x03}
	pdol, _ := ParseDOL(data)

	terminal, _ := NewTerminal()
	s := Transaction{}
	result := terminal.BuildPDOLData(&s, pdol)
	require.Len(t, result, 3)
	expected, _ := hex.DecodeString("210101")
	require.Equal(t, expected, result)
}

func TestBuildPDOLData_TransactionType(t *testing.T) {
	// Transaction Type
	// 9C 01
	data := []byte{0x9C, 0x01}
	pdol, _ := ParseDOL(data)

	terminal, _ := NewTerminal()
	s := Transaction{
		TransactionType: "00",
	}
	result := terminal.BuildPDOLData(&s, pdol)
	require.Len(t, result, 1)
	expected, _ := hex.DecodeString("00")
	require.Equal(t, expected, result)
}

func TestBuildPDOLData_UnpredictableNumber(t *testing.T) {
	// Unpredictable Number
	// 9F37 04
	data := []byte{0x9F, 0x37, 0x04}
	pdol, _ := ParseDOL(data)

	terminal, _ := NewTerminal()
	s := Transaction{}
	result := terminal.BuildPDOLData(&s, pdol)
	require.Len(t, result, 4)
	expected, _ := hex.DecodeString(DefaultUnpredictableNumber)
	require.Equal(t, expected, result)
}

func TestBuildPDOLData_manual(t *testing.T) {
	// 9F66 04 9F02 06 9F03 06 9F1A 02 95 05 5F2A 02 9A 03 9C 01 9F37 04
	data := t9F66
	data = append(data, t9F02...)
	data = append(data, t9F03...)
	data = append(data, t9F1A...)
	data = append(data, t95...)
	data = append(data, t5F2A...)
	data = append(data, t9A...)
	data = append(data, t9C...)
	data = append(data, t9F37...)
	//	t.Errorf("PDOL: %s", hex.EncodeToString(data))
	// expected responses built from the above PDOL
	expected := t9F66Response                     // B600C000
	expected = append(expected, t9F02Response...) // 1234
	expected = append(expected, t9F03Response...) // 5678
	expected = append(expected, t9F1AResponse...) // 0840
	expected = append(expected, t95Response...)   // 0000000000
	expected = append(expected, t5F2AResponse...) // 0840
	expected = append(expected, t9AResponse...)   // 210101
	expected = append(expected, t9CResponse...)   // 00
	expected = append(expected, t9F37Response...) // A1B2C3D4

	pdol, _ := ParseDOL(data)
	length := 0
	for _, d := range pdol {
		length += d.Length
	}

	terminal, _ := NewTerminal()

	s := Transaction{
		AuthorizedAmount: "1234",
		SecondaryAmount:  "5678",
		TransactionType:  "00",
	}
	result := terminal.BuildPDOLData(&s, pdol)
	require.Len(t, result, length)
	require.Equal(t, expected, result)
}

func TestBuildPDOLData_Ramp(t *testing.T) {
	// PDOL from Ramp visa card
	pdolHex := "9f66049f02069f03069f1a0295055f2a029a039c019f3704"
	data, _ := hex.DecodeString(pdolHex)
	pdol, _ := ParseDOL(data)
	length := 0
	for _, d := range pdol {
		length += d.Length
	}
	terminal, _ := NewTerminal()

	s := Transaction{
		AuthorizedAmount: "1234",
		SecondaryAmount:  "5678",
		TransactionType:  "00",
	}

	result := terminal.BuildPDOLData(&s, pdol)
	require.Len(t, result, length)
}

func TestBuildPDOLData_Chase(t *testing.T) {
	// PDOL from Chase United MileagePlus card
	pdolHex := "9F66049F02069F03069F1A0295055F2A029A039C019F3704"
	data, _ := hex.DecodeString(pdolHex)
	pdol, _ := ParseDOL(data)

	length := 0
	for _, d := range pdol {
		length += d.Length
	}
	terminal, _ := NewTerminal()
	s := Transaction{
		AuthorizedAmount: "1234",
		SecondaryAmount:  "5678",
		TransactionType:  "00",
	}

	result := terminal.BuildPDOLData(&s, pdol)
	processing := GetProcessingOptions(result)
	expected := "80a80000238321b600c00000000000123400000000567808400000000000084021010100a1b2c3d400"
	if expected != hex.EncodeToString(processing) {
		t.Errorf("Expected %s, got %s", expected, hex.EncodeToString(processing))
	}
	if len(result) != length {
		t.Errorf("Expected %d bytes, got %d", length, len(result))
	}
}

func TestBuildPDOLData_CapitalOne(t *testing.T) {
	// PDOL from Capital One card
	pdolHex := "9F66049F02069F37049F1A02"
	data, _ := hex.DecodeString(pdolHex)
	pdol, _ := ParseDOL(data)
	length := 0
	for _, d := range pdol {
		length += d.Length
	}
	terminal, _ := NewTerminal()
	s := Transaction{
		AuthorizedAmount: "1234",
		SecondaryAmount:  "5678",
		TransactionType:  "00",
	}

	result := terminal.BuildPDOLData(&s, pdol)
	processing := GetProcessingOptions(result)
	expected := "80a80000128310b600c000000000001234a1b2c3d4084000"
	if expected != hex.EncodeToString(processing) {
		t.Errorf("Expected %s, got %s", expected, hex.EncodeToString(processing))
	}
	if len(result) != length {
		t.Errorf("Expected %d bytes, got %d", length, len(result))
	}
}

// 80 a8 0000 10       b600c000 000000001234 a1b2c3d4 0840
// 80 a8 0000 10       b600c000 000000001234 a1b2c3d4 0840 00
// 80 a8 0000 11 83 10 b600c000 000000001234 a1b2c3d4 0840 00
// 80 a8 0000 12 83 10 b600c000 000000001234 a1b2c3d4 0840 00
// 80 a8 0000 11 83 10 b600c000 000000001234 a1b2c3d4 0840
// 80 A8 0000 12 83 10 B620C000 000000001000 823DDE7A 0124 00
// 80 A8 0000 02 83 00

// CARDHOLDER/VISA
