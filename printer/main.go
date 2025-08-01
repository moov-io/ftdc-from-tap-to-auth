package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

func main() {
	printer, err := NewThermalPrinter()
	if err != nil {
		log.Fatalf("Failed to initialize printer: %v", err)
	}
	defer printer.Close()

	fmt.Println("Printer connected successfully!")

	err = printer.PrintCentered("*** Fintech DevCon 2025 ***")
	if err != nil {
		log.Printf("Error printing title: %v", err)
	}

	err = printer.Feed(1)
	if err != nil {
		log.Printf("Error feeding paper: %v", err)
	}

	lines := []string{
		fmt.Sprintf("Date, time: %s", time.Now().Format("2006-01-02 15:04:05")),
		fmt.Sprintf("PAN: %s", "1234 5678 9012 3456"),
		fmt.Sprintf("Cardholder: %s", "John Doe"),
		fmt.Sprintf("Amount: %s", "$100.00"),
		fmt.Sprintf("Auth Code: %s", "123456"),
		fmt.Sprintf("Resp Code: %s", "00"),
	}

	err = printer.PrintLines(lines)
	if err != nil {
		log.Printf("Error printing lines: %v", err)
	}

	err = printer.Feed(1)
	if err != nil {
		log.Printf("Error feeding paper: %v", err)
	}

	phrase := motivationalPhrases[rand.Intn(len(motivationalPhrases))]

	err = printer.PrintCentered(fmt.Sprintf("* %s *", phrase))
	if err != nil {
		log.Printf("Error printing title: %v", err)
	}

	err = printer.Feed(2)
	if err != nil {
		log.Printf("Error feeding paper: %v", err)
	}

	err = printer.Cut()
	if err != nil {
		log.Printf("Error cutting paper: %v", err)
	}

	fmt.Println("Done!")
}
