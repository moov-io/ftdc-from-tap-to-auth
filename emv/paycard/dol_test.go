package paycard

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestParsePDOL(t *testing.T) {
	// PDOL from Ramp visa card
	pdolHex := "9F66049F02069F03069F1A0295055F2A029A039C019F3704"
	data, err := hex.DecodeString(pdolHex)
	if err != nil {
		fmt.Println("Error decoding PDOL hex string:", err)
		return
	}

	_, err = ParseDOL(data)
	if err != nil {
		fmt.Println("Error parsing PDOL:", err)
		return
	}

	// Print the parsed Tag and Length pairs
	//	PrettyPrintDOL(result)
}

func TestPrasePDOL2(t *testing.T) {
	// PDOL from Chase United MileagePlus card
	pdolHex := "9F66049F02069F03069F1A0295055F2A029A039C019F3704"
	data, err := hex.DecodeString(pdolHex)
	if err != nil {
		fmt.Println("Error decoding PDOL hex string:", err)
		return
	}

	_, err = ParseDOL(data)
	if err != nil {
		fmt.Println("Error parsing PDOL:", err)
		return
	}

	// Print the parsed Tag and Length pairs
	// PrettyPrintDOL(result)
}
