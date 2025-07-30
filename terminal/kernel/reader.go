package kernel

import (
	"errors"
	"fmt"
)

type RawCardReader interface {
	SendAPDU(command []byte) ([]byte, error)
}

type CardReader interface {
	SendAPDU(command APDUCommand) (APDUResponse, error)
}

// Adapter that wraps an existing raw card reader
type CardReaderAdapter struct {
	rawReader RawCardReader
}

func NewCardReaderAdapter(rawReader RawCardReader) *CardReaderAdapter {
	return &CardReaderAdapter{
		rawReader: rawReader,
	}
}

func (adapter *CardReaderAdapter) SendAPDU(command APDUCommand) (APDUResponse, error) {
	// Convert APDUCommand to []byte
	rawCommand, err := encodeAPDUCommand(command)
	if err != nil {
		return APDUResponse{}, fmt.Errorf("failed to encode APDU command: %w", err)
	}

	rawResponse, err := adapter.rawReader.SendAPDU(rawCommand)
	if err != nil {
		return APDUResponse{}, fmt.Errorf("card communication failed: %w", err)
	}

	response, err := decodeAPDUResponse(rawResponse)
	if err != nil {
		return APDUResponse{}, fmt.Errorf("failed to decode APDU response: %w", err)
	}

	return response, nil
}

func encodeAPDUCommand(cmd APDUCommand) ([]byte, error) {
	// Basic APDU structure: CLA INS P1 P2 [Lc Data] [Le]
	result := []byte{cmd.CLA, cmd.INS, cmd.P1, cmd.P2}

	// Add data if present
	if len(cmd.Data) > 0 {
		if len(cmd.Data) > 255 {
			return nil, errors.New("data too long for short APDU")
		}
		result = append(result, byte(len(cmd.Data))) // Lc
		result = append(result, cmd.Data...)         // Data
	}

	// Add Le if present
	if cmd.Le != nil {
		result = append(result, *cmd.Le)
	}

	return result, nil
}

// Helper function to decode []byte to APDUResponse
func decodeAPDUResponse(raw []byte) (APDUResponse, error) {
	if len(raw) < 2 {
		return APDUResponse{}, errors.New("response too short, must contain at least SW1 SW2")
	}

	// Last 2 bytes are always SW1 SW2
	dataLen := len(raw) - 2

	response := APDUResponse{
		SW1: raw[dataLen],
		SW2: raw[dataLen+1],
	}

	// Everything before SW1 SW2 is data
	if dataLen > 0 {
		response.Data = make([]byte, dataLen)
		copy(response.Data, raw[:dataLen])
	}

	return response, nil
}

func ptrByte(b byte) *byte {
	return &b
}
