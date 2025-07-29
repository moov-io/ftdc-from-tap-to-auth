package main

import (
	"fmt"
	"os"

	tm "github.com/moov-io/ftdc-from-tap-to-auth/terminal"
)

func main() {
	// show current directory
	err := runTerminal()
	if err != nil {
		fmt.Println("Error running terminal:", err)
		os.Exit(1)
	}
}

func runTerminal() error {
	// we should read the config and flags and pass them to the terminal
	cfg, err := tm.NewConfigFromFile("configs/terminal.yaml")
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	terminal, err := tm.NewTerminal(cfg)
	if err != nil {
		return fmt.Errorf("creating terminal: %w", err)
	}

	err = terminal.Run()
	if err != nil {
		return fmt.Errorf("running terminal: %w", err)
	}

	return nil

}
