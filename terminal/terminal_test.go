package terminal_test

import (
	"testing"

	"github.com/moov-io/ftdc-from-tap-to-auth/terminal"
	"github.com/stretchr/testify/require"
)

func TestPCSC(t *testing.T) {
	_, err := terminal.NewCardReader()
	require.NoError(t, err, "creating card reader")
}
