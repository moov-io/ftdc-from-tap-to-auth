package kernel

import "fmt"

type APDUCommand struct {
	CLA  byte
	INS  byte
	P1   byte
	P2   byte
	Data []byte
	Le   *byte // nil means no Le, 0 means Le=0
}

type APDUResponse struct {
	Data []byte
	SW1  byte
	SW2  byte
}

type APDUError struct {
	SW uint16 // Status Word (SW1 SW2)
}

func (e *APDUError) Error() string {
	swInfo := GetStatusWordInfo(e.SW)

	return fmt.Sprintf("APDU Error: %04X - %s: %s", e.SW, swInfo.Name, swInfo.Description)
}

func (r APDUResponse) IsSuccess() bool {
	return r.SW1 == 0x90 && r.SW2 == 0x00
}

func (r APDUResponse) StatusWord() uint16 {
	return uint16(r.SW1)<<8 | uint16(r.SW2)
}

func (r APDUResponse) Error() error {
	return &APDUError{SW: r.StatusWord()}
}
