package printer_test

import (
	"testing"

	"github.com/moov-io/ftdc-from-tap-to-auth/printer"
	"github.com/stretchr/testify/require"
)

func TestBitmapLogo(t *testing.T) {
	logo, err := printer.NewLogoBitmap()
	require.NoError(t, err)
	require.NotNil(t, logo)
}
