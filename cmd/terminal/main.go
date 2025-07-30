package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/moov-io/ftdc-from-tap-to-auth/internal/config"
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
	// kernel flag
	kernel := flag.String("kernel", "", "Kernel to use for the terminal (e.g., 'universal' or 'ftdc')")
	flag.Parse()

	// we should read the config and flags and pass them to the terminal
	cfg := &tm.Config{}

	err := config.NewFromFile("configs/terminal.yaml", cfg)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if *kernel != "" {
		cfg.Kernel = *kernel
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
