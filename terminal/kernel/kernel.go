package kernel

import (
	"fmt"

	"github.com/moov-io/bertlv"
)

// FTDCKernel is the main kernel for processing payment cards
var ftdcApplicationID = []byte{0xA0, 0x00, 0x00, 0x00, 0x02, 0x03, 0x04, 0x05}

type FTDCKernel struct {
	reader CardReader
	TagsDB []bertlv.TLV
}

func NewFTDCKernel(reader CardReader) *FTDCKernel {
	return &FTDCKernel{
		reader: reader,
	}
}

func (kt *FTDCKernel) Process() error {
	err := kt.SelectApplication(ftdcApplicationID)
	if err != nil {
		return fmt.Errorf("selecting default application: %w", err)
	}

	err = kt.readRecords()
	if err != nil {
		return fmt.Errorf("reading records: %w", err)
	}

	return nil
}

// SelectApplication selects a specific application by its AID
// This is the core EMV command that "opens" a payment application on the card
// For FTDC cards, we don't parse the FCI response as it's not ... ready
func (k *FTDCKernel) SelectApplication(aid []byte) error {
	// Create SELECT command for the specific AID
	selectCmd := NewSelectCommand(aid)

	// Send the command to the card
	resp, err := k.reader.SendAPDU(selectCmd)
	if err != nil {
		return fmt.Errorf("failed to send SELECT command: %w", err)
	}

	// Check if the application was successfully selected
	if !resp.IsSuccess() {
		return fmt.Errorf("failed to select application %X: %w", aid, resp.Error())
	}

	// parse response
	fciTemplate, err := bertlv.Decode(resp.Data)
	if err != nil {
		return fmt.Errorf("parsing FCI response: %w", err)
	}

	appID, found := bertlv.FindFirstTag(fciTemplate, "84")
	if found {
		k.TagsDB = append(k.TagsDB, appID)
	}

	appLabel, found := bertlv.FindFirstTag(fciTemplate, "50")
	if found {
		k.TagsDB = append(k.TagsDB, appLabel)
	}

	fmt.Printf("✅ Application %X - %s selected successfully\n", appID.Value, appLabel.Value)
	fmt.Printf("✅ FCI response received for selected application\n")
	bertlv.PrettyPrint(fciTemplate)

	return nil
}

// readRecords reads the cardholder data records (PAN, name, expiration date)
func (k *FTDCKernel) readRecords() error {

	// Create READ RECORD command
	readCmd := NewReadRecordCommand(1, 1)

	// Send command to card
	resp, err := k.reader.SendAPDU(readCmd)
	if err != nil {
		return fmt.Errorf("failed to send READ RECORD command: %w", err)
	}

	// Check if command was successful
	if !resp.IsSuccess() {
		return fmt.Errorf("READ RECORD command failed: %w", resp.Error())
	}

	// Parse the TLV response
	tlvs, err := bertlv.Decode(resp.Data)
	if err != nil {
		return fmt.Errorf("failed to decode READ RECORD response: %w", err)
	}
	fmt.Printf("✅ READ RECORD successful\n")
	bertlv.PrettyPrint(tlvs)

	responseMessageTemplate, found := bertlv.FindFirstTag(tlvs, "70")
	if !found {
		return fmt.Errorf("response message template (70) not found in READ RECORD response")
	}

	// Store the parsed TLVs in the kernel's database
	k.TagsDB = append(k.TagsDB, responseMessageTemplate.TLVs...)

	// Process the response based on template type
	return nil
}
