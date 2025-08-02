package terminal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/moov-io/bertlv"
	"github.com/moov-io/ftdc-from-tap-to-auth/acquirer/client"
	"github.com/moov-io/ftdc-from-tap-to-auth/acquirer/models"
	"github.com/moov-io/ftdc-from-tap-to-auth/printer"
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

	err = t.printReceipt(payment, tags)
	if err != nil {
		return fmt.Errorf("printing receipt: %w", err)
	}

	return nil
}

func (t *Terminal) printReceipt(payment models.Payment, tags []bertlv.TLV) error {
	if t.config.PrinterURL == "" {
		fmt.Println("No printer configured, skipping receipt printing.")
		return nil
	}

	var cardholder string

	cardholderTag, ok := bertlv.FindFirstTag(tags, cardHolderNameTag)
	if !ok {
		cardholder = ""
	} else {
		cardholder = string(cardholderTag.Value)
	}

	receipt := printer.Receipt{
		PaymentID:          payment.ID,
		ProcessingDateTime: payment.CreatedAt,
		PAN:                fmt.Sprintf("%s****%s", payment.Card.First6, payment.Card.Last4),
		Cardholder:         cardholder,
		Amount:             payment.Amount,
		AuthorizationCode:  payment.AuthorizationCode,
		ResponseCode:       payment.ResponseCode,
	}

	receiptJSON, err := json.Marshal(receipt)
	if err != nil {
		return fmt.Errorf("marshalling receipt: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/receipts", t.config.PrinterURL), bytes.NewReader(receiptJSON))
	if err != nil {
		return fmt.Errorf("creating request to printer: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sending receipt to printer: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("printer returned error: %s; status code: %d", responseBody, resp.StatusCode)
	}

	var printJob printer.PrintJob
	err = json.NewDecoder(resp.Body).Decode(&printJob)
	if err != nil {
		return fmt.Errorf("decoding print job response: %w", err)
	}

	fmt.Println("Receipt printed successfully!")
	fmt.Printf("Print Job: Number in Queue=%d, Waiting Time=%d seconds\n",
		printJob.NumberInQueue,
		printJob.WaitingTime,
	)

	return nil
}
