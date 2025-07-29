package terminal

import (
	"fmt"
	"os"
	"time"

	"github.com/moov-io/bertlv"
	"github.com/moov-io/ftdc-from-tap-to-auth/acquirer/client"
	"github.com/moov-io/ftdc-from-tap-to-auth/acquirer/models"
	"github.com/moov-io/ftdc-from-tap-to-auth/terminal/paycard"
)

type Terminal struct {
	config *Config
}

func NewTerminal(cfg *Config) (*Terminal, error) {
	return &Terminal{
		config: cfg,
	}, nil
}

func (t *Terminal) Run() error {
	fmt.Println("FTDC Terminal is running...")

	var amount int64
	if t.config.DefaultAmount != 0 {
		amount = t.config.DefaultAmount
		fmt.Printf("Using default amount: %d cents\n", t.config.DefaultAmount)
	} else {
		fmt.Println("Please enter amount (in cents, e.g., 100 for $1.00):")

		_, err := fmt.Scanf("%d", &amount)
		if err != nil {
			return fmt.Errorf("failed to read amount: %w", err)
		}

		if amount <= 0 {
			return fmt.Errorf("amount must be greater than 0")
		}
	}

	err := t.run(amount)
	if err != nil {
		return fmt.Errorf("running terminal: %w", err)
	}

	// Implementation of terminal run logic
	// This should include reading card data, processing payments, etc.
	return nil
}

func (t *Terminal) run(amount int64) error {
	// create a new terminal
	terminal, err := paycard.NewTerminal(
		paycard.WithCountryCode("0840"),  // USA
		paycard.WithCurrencyCode("0840"), // USD
	)
	if err != nil {
		return fmt.Errorf("creating terminal: %w", err)
	}

	// create a new session for this card transaction
	session := paycard.Transaction{
		AuthorizedAmount: fmt.Sprintf("%012d", amount),
		SecondaryAmount:  "000000000000",
		TransactionType:  "00",
	}

	// Establish context with PC/SC reader
	cardReader, err := NewCardReader()
	if err != nil {
		fmt.Println("Failed to create card reader:", err)
		os.Exit(1)
	}

	if t.config.ReaderIndex < 0 {
		cardReader.DisplayReaders()
		_, err = cardReader.SelectReader()
		if err != nil {
			fmt.Println("selecting reader:", err)
			os.Exit(1)
		}
	} else if t.config.ReaderIndex >= 0 && t.config.ReaderIndex < len(cardReader.Readers) {
		cardReader.SelectedReader = cardReader.Readers[t.config.ReaderIndex]
	}

	fmt.Println("Using NFC reader:", cardReader.SelectedReader)

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

	if emvCard.GPOResponse.AFL != nil {
		err = cardReader.ProcessAFL(emvCard)
		if err != nil {
			fmt.Println("Failed to process AFL:", err)
			os.Exit(1)
		}
	}

	fmt.Println("Tags from card:")

	bertlv.PrettyPrint(emvCard.TagsDB)

	err = t.createPayment(amount, emvCard.TagsDB)
	if err != nil {
		return fmt.Errorf("creating payment: %w", err)
	}

	return nil

}

const (
	panTag            = "5A"
	expDateTag        = "5F24"
	cardHolderNameTag = "5F20"
)

func (t *Terminal) createPayment(amount int64, tags []bertlv.TLV) error {
	fmt.Println("Sending payment request to acquirer...")

	paymentTags := bertlv.CopyTags(tags, []string{panTag, expDateTag, cardHolderNameTag}...)

	emvPayload, err := bertlv.Encode(paymentTags)
	if err != nil {
		return fmt.Errorf("encoding EMV payload: %w", err)
	}

	merchant := client.New(t.config.AcquirerURL)
	payment, err := merchant.CreatePayment(
		t.config.MerchantID,
		models.CreatePayment{
			Amount:     amount,
			Currency:   "USD",
			EMVPayload: emvPayload,
		},
	)
	if err != nil {
		return fmt.Errorf("creating payment: %w", err)
	}

	fmt.Printf("Payment created successfully: ID=%s, Status=%s, Authorization Code=%s\n",
		payment.ID,
		payment.Status,
		payment.AuthorizationCode,
	)

	return nil
}
