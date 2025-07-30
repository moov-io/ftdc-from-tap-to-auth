package terminal

import (
	"fmt"
	"time"

	"github.com/moov-io/bertlv"
	"github.com/moov-io/ftdc-from-tap-to-auth/acquirer/client"
	"github.com/moov-io/ftdc-from-tap-to-auth/acquirer/models"
	"github.com/moov-io/ftdc-from-tap-to-auth/terminal/kernel"
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
	if t.config.Kernel != "universal" {
		return t.RunFTDCKernel()
	}

	return t.RunUniversalKernel()
}

func (t *Terminal) RunFTDCKernel() error {
	fmt.Println("📱FTDC Terminal is running...")

	amount, err := t.promptForAmount()
	if err != nil {
		return fmt.Errorf("prompting for amount: %w", err)
	}

	cardReader, err := t.cardReaderWithCard()
	if err != nil {
		return fmt.Errorf("reading card: %w", err)
	}

	k := kernel.NewFTDCKernel(kernel.NewCardReaderAdapter(cardReader))

	err = k.Process()
	if err != nil {
		return fmt.Errorf("processing kernel: %w", err)
	}

	fmt.Println("*********************************************")
	fmt.Println("EMV Tags read from card")
	bertlv.PrettyPrint(k.TagsDB)
	fmt.Println("*********************************************")

	// Send payment request to the acquirer
	err = t.createPayment(amount, k.TagsDB)
	if err != nil {
		return fmt.Errorf("creating payment: %w", err)
	}

	return nil
}

func (t *Terminal) promptForAmount() (int64, error) {
	var amount int64
	if t.config.DefaultAmount != 0 {
		amount = t.config.DefaultAmount
		fmt.Printf("Using default amount: %d cents\n", t.config.DefaultAmount)
	} else {
		fmt.Println("Please enter amount (in cents, e.g., 100 for $1.00):")

		_, err := fmt.Scanf("%d", &amount)
		if err != nil {
			return 0, fmt.Errorf("failed to read amount: %w", err)
		}

		if amount <= 0 {
			return 0, fmt.Errorf("amount must be greater than 0")
		}
	}

	return amount, nil
}

func (t *Terminal) cardReaderWithCard() (*CardReader, error) {
	cardReader, err := NewCardReader()
	if err != nil {
		return nil, fmt.Errorf("creating card reader: %w", err)
	}

	if t.config.ReaderIndex < 0 {
		cardReader.DisplayReaders()
		_, err = cardReader.SelectReader()
		if err != nil {
			return nil, fmt.Errorf("selecting card reader: %w", err)
		}
	} else if t.config.ReaderIndex >= 0 && t.config.ReaderIndex < len(cardReader.Readers) {
		cardReader.SelectedReader = cardReader.Readers[t.config.ReaderIndex]
	}

	fmt.Println("Using NFC reader:", cardReader.SelectedReader)

	timeout := time.Second * 60

	err = cardReader.WaitForCard(timeout)
	if err != nil {
		return nil, fmt.Errorf("waiting for card: %w", err)
	}

	err = cardReader.ConnectToCard()
	if err != nil {
		return nil, fmt.Errorf("connecting to card: %w", err)
	}

	fmt.Println("Card connected successfully!")

	return cardReader, nil
}

const (
	panTag            = "5A"
	expDateTag        = "5F24"
	cardHolderNameTag = "5F20"
	appIDTag          = "84"
	appLabelTag       = "50"
)

func (t *Terminal) createPayment(amount int64, tags []bertlv.TLV) error {
	fmt.Println("Sending payment request to acquirer...")

	paymentTags := bertlv.CopyTags(tags, []string{
		panTag,
		expDateTag,
		cardHolderNameTag,
		appIDTag,
		appLabelTag,
	}...)

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
