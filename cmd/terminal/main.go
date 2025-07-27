package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/moov-io/ftdc-from-tap-to-auth/acquirer/client"
	"github.com/moov-io/ftdc-from-tap-to-auth/acquirer/models"
	tm "github.com/moov-io/ftdc-from-tap-to-auth/terminal"
	"github.com/moov-io/ftdc-from-tap-to-auth/terminal/paycard"
)

func main() {

	readerIndex := flag.Int("reader", -1, "Index of reader to use (default: interactive selection)")
	cvv := flag.String("cvv", "123", "CVV code of the card (default: 123)")
	amount := flag.Int64("amount", 0, "Amount to authorize (default: 10)")
	merchantID := flag.String("merchant", "", "ID of merchant to create payment for")

	// Parse the command-line arguments
	flag.Parse()

	if *amount <= 0 {
		fmt.Println("Amount must be greater than 0")
		os.Exit(1)
	}
	if *merchantID == "" {
		fmt.Println("Merchant ID must be provided")
		os.Exit(1)
	}

	// create a new terminal
	terminal, err := paycard.NewTerminal(
		paycard.WithCountryCode("0840"),  // USA
		paycard.WithCurrencyCode("0840"), // USD
	)
	if err != nil {
		fmt.Println("Failed to create terminal:", err)
		os.Exit(1)
	}

	// create a new session for this card transaction
	session := paycard.Transaction{
		AuthorizedAmount: "1234",
		SecondaryAmount:  "5678", // TODO(adam): does this cause issues?
		TransactionType:  "00",
	}

	// TODO: This Main should all be inside of a NewTerminal() function that returns a terminal struct

	// Establish context with PC/SC reader
	cardReader, err := tm.NewCardReader()
	if err != nil {
		fmt.Println("Failed to create card reader:", err)
		os.Exit(1)
	}

	if readerIndex == nil || *readerIndex < 0 {
		cardReader.DisplayReaders()
		_, err = cardReader.SelectReader()
		if err != nil {
			fmt.Println("selecting reader:", err)
			os.Exit(1)
		}
	} else if *readerIndex >= 0 && *readerIndex < len(cardReader.Readers) {
		cardReader.SelectedReader = cardReader.Readers[*readerIndex]
	}
	fmt.Println("Using reader:", cardReader.SelectedReader)

	timeout := time.Second * 60

	_ = cardReader.WaitForCard(timeout)

	err = cardReader.ConnectToCard()
	if err != nil {
		fmt.Println("Failed to connect to card:", err)
		os.Exit(1)
	}

	// We have a emvCard to start parsing
	emvCard := paycard.NewEmvCard(true)

	ppseSelected, err := cardReader.SelectPPSE(emvCard)
	if ppseSelected {
		aid, err := cardReader.SelectAID(emvCard)
		if err != nil {
			fmt.Println("Failed to select AID:", err)
			os.Exit(1)
		}
		session.AID = aid

		err = cardReader.ProcessPDOL(emvCard, terminal, &session)
		if err != nil {
			fmt.Println("Failed to process PDOL:", err)
			os.Exit(1)
		}
	} else {
		err = cardReader.DirectApplicationSelection(emvCard)
		if err != nil {
			fmt.Println("Failed to direct application selection:", err)
			os.Exit(1)
		}
	}

	// Print the GPO response
	// fmt.Printf("GPO Response:\n%s\n", emvCard.GPOResponse.String())

	if emvCard.GPOResponse.AFL != nil {
		err = cardReader.ProcessAFL(emvCard)
		if err != nil {
			fmt.Println("Failed to process AFL:", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("No AFL found, trying READ RECORD command 00:B2:01:0C...")

		card, err := cardReader.ReadRecord(emvCard)
		if err != nil {
			fmt.Println("Failed to read record:", err)
			os.Exit(1)
		}
		card.CardVerificationValue = *cvv
		err = createPayment(*merchantID, *amount, card)
		if err != nil {
			fmt.Println("Failed to create payment:", err)
			os.Exit(1)
		}
	}
}

func createPayment(merchantID string, amount int64, card models.Card) error {

	merchant := client.New("http://127.0.0.1:8080")

	fmt.Printf("\ncard: %s\n", card)
	_, err := merchant.CreatePayment(merchantID, models.CreatePayment{
		Amount:   amount,
		Currency: "USD",
		Card:     card,
	})
	if err != nil {
		return fmt.Errorf("creating payment: %w", err)
	}

	return nil
}
