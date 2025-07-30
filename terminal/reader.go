package terminal

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ebfe/scard"
	"github.com/kr/pretty"
	"github.com/moov-io/bertlv"
	"github.com/moov-io/ftdc-from-tap-to-auth/terminal/paycard"
)

type CardReader struct {
	ctx            *scard.Context
	Readers        []string
	SelectedReader string
	Card           *scard.Card
}

func NewCardReader() (*CardReader, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return nil, fmt.Errorf("failed to establish context: %w", err)
	}

	if isValid, err := ctx.IsValid(); !isValid || err != nil {
		return nil, fmt.Errorf("context is not valid: %w", err)
	}

	readers, err := ctx.ListReaders()
	if err != nil {
		return nil, fmt.Errorf("failed to list card readers: %w", err)
	}

	return &CardReader{
		ctx:     ctx,
		Readers: readers,
	}, nil
}

func (c *CardReader) SendAPDU(cmd []byte) ([]byte, error) {
	// Ensure the card is connected
	if c.Card == nil {
		return nil, fmt.Errorf("no card connected")
	}

	response, err := c.Card.Transmit(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to send APDU command: %w", err)
	}

	return response, nil
}

func (c *CardReader) Close() error {
	if c.Card != nil {
		if err := c.Card.Disconnect(scard.LeaveCard); err != nil {
			return fmt.Errorf("failed to disconnect card: %w", err)
		}
		c.Card = nil
	}
	if c.ctx != nil {
		return c.ctx.Release()
	}
	return nil
}

func (c *CardReader) DisplayReaders() {
	fmt.Printf("Available readers:\n")
	for i, reader := range c.Readers {
		fmt.Printf("  [%d] %s\n", i, reader)
	}
}

func (c *CardReader) SelectReader() (string, error) {
	if len(c.Readers) < 1 {
		return "", fmt.Errorf("no readers found")
	} else if len(c.Readers) == 1 {
		c.SelectedReader = c.Readers[0]
		return c.SelectedReader, nil
	}

	for {
		fmt.Print("\nSelect reader (0-", len(c.Readers)-1, "): ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		// Check if user wants to quit
		if input == "q" || input == "quit" || input == "exit" {
			fmt.Println("Exiting...")
			os.Exit(0)
		}

		// Parse the selection
		index, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Invalid input. Please enter a number between 0 and %d, or 'q' to quit\n", len(c.Readers)-1)
			continue
		}

		if index < 0 || index >= len(c.Readers) {
			fmt.Printf("Invalid reader index. Please enter a number between 0 and %d\n", len(c.Readers)-1)
			continue
		}

		c.SelectedReader = c.Readers[index]
		return c.SelectedReader, nil
	}
}

func (c *CardReader) ConnectToCard() error {
	card, err := c.ctx.Connect(c.SelectedReader, scard.ShareExclusive, scard.ProtocolAny)
	if err != nil {
		return fmt.Errorf("failed to connect to card: %w", err)
	}
	c.Card = card
	return nil
}

func (c *CardReader) WaitForCardAsync(timeout time.Duration) <-chan error {
	resultChan := make(chan error, 1)

	go func() {
		defer close(resultChan)

		readerStates := []scard.ReaderState{
			{
				Reader:       c.SelectedReader,
				CurrentState: scard.StateEmpty,
			},
		}

		fmt.Println("Waiting for card...")
		err := c.ctx.GetStatusChange(readerStates, timeout)
		if err != nil {
			resultChan <- fmt.Errorf("failed to wait for card: %v", err)
			return
		}

		// Check if the card is present
		if readerStates[0].EventState&scard.StatePresent != 0 {
			fmt.Println("Card presented.")
			resultChan <- nil
			return
		}

		resultChan <- fmt.Errorf("timeout waiting for card")
	}()

	return resultChan
}

func (c *CardReader) WaitForCardRemoveAsync(timeout time.Duration) <-chan error {
	resultChan := make(chan error, 1)

	go func() {
		defer close(resultChan)

		readerStates := []scard.ReaderState{
			{
				Reader:       c.SelectedReader,
				CurrentState: scard.StatePresent,
			},
		}

		fmt.Println("Waiting for card remove...")
		err := c.ctx.GetStatusChange(readerStates, timeout)
		if err != nil {
			resultChan <- fmt.Errorf("failed waiting for card remove: %v", err)
			return
		}

		// Check if the card is removed
		if readerStates[0].EventState&scard.StateEmpty != 0 {
			fmt.Println("Card removed.")
			resultChan <- nil
			return
		}

		resultChan <- fmt.Errorf("timeout waiting for card remove")
	}()

	return resultChan
}

func (c *CardReader) WaitForCard(timeout time.Duration) error {
	return <-c.WaitForCardAsync(timeout)
}

func (c *CardReader) WaitForCardRemove(timeout time.Duration) error {
	return <-c.WaitForCardRemoveAsync(timeout)
}

func (c *CardReader) DirectApplicationSelection(emvCard *paycard.EmvCard) error {
	fmt.Printf("Trying direct application selection...\n")

	// List of common AIDs to try
	knownAIDs := []struct {
		aid         []byte
		label       string
		kernelID    paycard.KernelID
		description string
	}{
		// {[]byte{0xA0, 0x00, 0x00, 0x01, 0x51, 0x00, 0x00}, "Visa", paycard.Kernel3ID, "Visa"},
		// {[]byte{0xA0, 0x00, 0x00, 0x00, 0x03, 0x10, 0x10, 0x10}, "Visa", paycard.Kernel3ID, "Visa"},
		// {[]byte{0xA0, 0x00, 0x00, 0x00, 0x03, 0x10, 0x56}, "Visa", paycard.Kernel3ID, "Visa"},
		// {[]byte{0xA0, 0x00, 0x00, 0x00, 0x03, 0x10, 0x4D}, "Visa", paycard.Kernel3ID, "Visa"},
		{[]byte{0xA0, 0x00, 0x00, 0x00, 0x02, 0x03, 0x04, 0x05}, "FTDC", paycard.Kernel1ID, "Fintech Devcon"},
		// {[]byte{0xD2, 0x76, 0x00, 0x00, 0x85, 0x30, 0x4A, 0x43, 0x4F, 0x90, 0x00, 0x01}, "Visa", paycard.Kernel3ID, "Visa"},
	}

	for _, aidInfo := range knownAIDs {
		fmt.Printf("Trying AID: %X (%s)\n", aidInfo.aid, aidInfo.description)

		// Build SELECT command
		selectCommand := []byte{0x00, 0xA4, 0x04, 0x00, byte(len(aidInfo.aid))}
		selectCommand = append(selectCommand, aidInfo.aid...)

		// Send SELECT command
		response, err := c.Card.Transmit(selectCommand)
		if err != nil {
			fmt.Printf("SELECT failed for %s: %v\n", aidInfo.description, err)
			continue
		}

		// Check status word
		if len(response) < 2 {
			continue
		}

		sw1 := response[len(response)-2]
		sw2 := response[len(response)-1]

		if sw1 == 0x90 && sw2 == 0x00 {
			fmt.Printf("✅ Found application: %s (AID: %X)\n", aidInfo.description, aidInfo.aid)

			emvCard.Applications = append(emvCard.Applications, paycard.Application{
				AID:      aidInfo.aid,
				Label:    aidInfo.label,
				Priority: 0x01,
			})

			return nil
		}

		fmt.Printf("SELECT failed for %s with status %02X%02X\n", aidInfo.description, sw1, sw2)
	}

	return fmt.Errorf("no supported applications found via direct selection")
}

func (c *CardReader) SelectPPSE(emvCard *paycard.EmvCard) (bool, error) {
	// // APDU command to select the PPSE (Proximity Payment System Environment)
	fmt.Println("=> 💳 SELECT FILE 2PAY.SYS.DDF01 to get the PPSE directory ...")

	// Construct the select command for a contactless card using the PPSE
	response, err := c.Card.Transmit(paycard.SelectPPSE.Bytes())
	if err != nil {
		// Default to PSE if PPSE is not found on the card.
		// TODO: Implement PSE logic here. PSE is for dipping the chip
		emvCard.Contactless = false
		return false, fmt.Errorf("failed to trasmit select cmd: %v", err)
	}

	sw1 := response[len(response)-2]
	sw2 := response[len(response)-1]

	if sw1 != 0x90 || sw2 != 0x00 {
		fmt.Printf("PPSE not found (status %02X%02X), trying direct application selection\n", sw1, sw2)
		return false, nil
	} else {
		fmt.Printf("Select 2PAY.SYS.DDF01 Raw Response: %X\n", response)

		// Parse the FCI Template to get the AID
		err = emvCard.Parse2Pay(response)
		if err != nil {
			return false, fmt.Errorf("Failed to parse 2PAY.SYS.DDF01: %w", err)
		}

		// TODO: Make logging optional for every step of parsing.
		bertlv.PrettyPrint(emvCard.FileControldInformation)

		return true, nil
	}
}

func (c *CardReader) SelectAID(emvCard *paycard.EmvCard) (string, error) {
	/**
	Select the Appropriate AID:
	- Based on terminal configurations for the supported processor, the reader will then select the AID that corresponds to a supported payment application.
	- Once an AID is selected, the reader sends another APDU SELECT command with the chosen AID, effectively “entering” that application on the card.
	- This step initializes the specific application, allowing further commands (such as Get Processing Options or Read Record) to proceed with transaction data exchange.
	**/

	fmt.Printf("=> 💳 Selecting AID %s...\n", emvCard.Applications[0].Label)

	// todo: Create a parser of the Application.AID for the application types that Moov supports
	// use the application that moov supports for the AID selection

	// Select the first application on the card (this is a simplified example)
	cmd := paycard.SelectAID(emvCard.Applications[0].AID)

	response, err := c.Card.Transmit(cmd.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to send APDU: %w", err)
	}
	// Check the response status word If the AID is not found or cannot be selected, the card may return an error status like 6A82 (File not found).
	fmt.Printf("Select AID Raw Response: %X\n", response)
	err = emvCard.ParseAIDResponse(response)
	if err != nil {
		return "", fmt.Errorf("failed to parse AID response: %w", err)
	}

	fmt.Printf("AIDResponse: %# v\n", pretty.Formatter(emvCard.AIDResponse))

	return fmt.Sprintf("%X", emvCard.Applications[0].AID), nil
}

func (c *CardReader) ProcessPDOL(emvCard *paycard.EmvCard, terminal *paycard.Terminal, session *paycard.Transaction) error {
	// If a PDOL was included in the FCI (9F38 tag), you must structure your GPO command to match this requirement.
	// If the PDOL is empty, you can send a GPO command with an empty data field.

	// Send the GPO command with an empty PDOL
	fmt.Println("=> 💳 Sending Get Processing Options (GPO) command...")
	// The APDU for the GET PROCESSING OPTIONS command typically looks like:

	// Create the PDOL data for the GPO command
	var command []byte
	if len(emvCard.AIDResponse.PDOL) > 0 {
		// Card explicitly requested PDOL data
		fmt.Println("Card requested PDOL data:")
		paycard.PrettyPrintDOL(emvCard.AIDResponse.PDOL)

		pdolRequest := terminal.BuildPDOLData(session, emvCard.AIDResponse.PDOL)
		command = paycard.GetProcessingOptions(pdolRequest)
	} else {
		// Try minimal GPO first
		fmt.Println("No PDOL requested, trying minimal GPO...")
		command = paycard.GetProcessingOptions(nil)
	}

	fmt.Printf("GPO Command: %X\n", command)
	response, err := c.Card.Transmit(command)
	if err != nil {
		return fmt.Errorf("Failed to send GET PROCESSING OPTIONS APDU: %w", err)
	}
	// Check the response status word
	fmt.Printf("GPO Raw Response: %X\n", response)

	// Check response status
	// Analyze any status words (e.g., 6985, 6A81, 6A82) returned from the card for clues about why the command failed.
	// todo: not sure why this logic does not detect the 6985 status word?
	if len(response) >= 2 {
		sw1 := response[len(response)-2]
		sw2 := response[len(response)-1]

		if sw1 == 0x67 && sw2 == 0x00 { // Wrong length
			// Card needs data but didn't specify via PDOL
			// Could try card-specific defaults here if needed
			fmt.Println("Card rejected minimal GPO, needs specific data")
		} else {
			fmt.Printf("GPO Raw Response: %X\n", response)
			err = emvCard.ParseGPOResponse(response)
			if err != nil {
				return fmt.Errorf("Failed to parse GPO response: %w", err)
			}
		}
	}
	return nil
}

func (c *CardReader) ProcessAFL(emvCard *paycard.EmvCard) error {
	fmt.Printf("AFL: %X\n", emvCard.GPOResponse.AFL)
	// Parse the AFL (Application File Locator) to get the SFI and record numbers
	// The AFL contains the SFI and record numbers for the data files to be read.
	//commands
	// Generate the READ RECORD commands based on the AFL
	commands := paycard.GenerateReadCommands(emvCard.GPOResponse.AFL)
	fmt.Printf("Commands: %X\n", commands)
	for _, cmd := range commands {
		response, err := c.Card.Transmit(cmd)
		if err != nil {
			// pretty prent the READ RECORD command
			fmt.Printf("Failed to send READ RECORD command: %X\n", cmd)
		}
		if len(response) > 0 {
			// check if response is not 6A83 (Record Not Found)
			if response[len(response)-2] != 0x6A || response[len(response)-1] != 0x83 {

				// parse the SFI into the emvCard
				emvCard.ParseSFI(response)
				ShowBerTLV(response)
			}
		}
	}
	return nil
}

func (c *CardReader) ReadRecord(card *paycard.EmvCard) error {
	// Send READ RECORD command: 00 B2 01 0C
	// B2 = READ RECORD instruction
	// 01 = Record number
	// 0C = Reference control parameter (SFI = 0, mode = 4)
	readCommand := []byte{0x00, 0xB2, 0x01, 0x0C}

	response, err := c.Card.Transmit(readCommand)
	if err != nil {
		return fmt.Errorf("sending READ RECORD command: %X, error: %w", readCommand, err)
	}

	fmt.Printf("READ RECORD Response: %X\n", response)

	// Check if response has status word
	if len(response) >= 2 {
		sw1 := response[len(response)-2]
		sw2 := response[len(response)-1]

		if sw1 == 0x90 && sw2 == 0x00 {
			fmt.Println("✅ READ RECORD command successful")

			tlvs, err := bertlv.Decode(response)
			if err != nil {
				return fmt.Errorf("Failed to decode READ RECORD response: %w", err)
			}

			bertlv.PrettyPrint(tlvs)

			// find response message template 70
			responseMessageTemplate, ok := bertlv.FindFirstTag(tlvs, "70")
			if !ok {
				return fmt.Errorf("Failed to find response message template (70) in READ RECORD response")
			}

			for _, tlv := range responseMessageTemplate.TLVs {
				card.TagsDB = append(card.TagsDB, tlv)
			}

			return nil
		} else {
			return fmt.Errorf("READ RECORD command failed with status %02X%02X", sw1, sw2)
		}
	}

	return fmt.Errorf("READ RECORD command response does not contain status word")
}
