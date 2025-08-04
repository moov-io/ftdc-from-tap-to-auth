package printer_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/moov-io/ftdc-from-tap-to-auth/log"
	"github.com/moov-io/ftdc-from-tap-to-auth/printer"
	"github.com/stretchr/testify/require"
)

func TestPrinter(t *testing.T) {
	if os.Getenv("PRINTER_TEST") == "" {
		t.Skip("Set PRINTER_TEST=1 to run this test")
	}

	logger := log.New()
	p, err := printer.NewThermalPrinter()
	require.NoError(t, err)
	defer p.Close()

	service := printer.NewService(logger, p)
	service.Start()
	defer service.Stop()

	pJob, err := service.PrintReceipt(printer.Receipt{
		ProcessingDateTime: time.Now(),
		PAN:                "424242424242",
		Cardholder:         "John Doe",
		Amount:             100,
		AuthorizationCode:  "123456",
		ResponseCode:       "14",
		Short:              false,
	})
	require.NoError(t, err)
	require.NotNil(t, pJob)

	for i := range 4 {
		pJob, err := service.PrintReceipt(printer.Receipt{
			Cardholder: fmt.Sprintf("%d - %s", i, "John Doe"),
			Short:      true,
		})
		require.NoError(t, err)
		require.NotNil(t, pJob)
	}

	time.Sleep(30 * time.Second)
}
