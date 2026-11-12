package terminal_test

import (
	"testing"

	"github.com/moov-io/ftdc-from-tap-to-auth/terminal"
)

func TestPCSC(t *testing.T) {
	_, err := terminal.NewCardReader()
	if err != nil {
		t.Skipf("PC/SC service or reader not available: %v", err)
	}
}
