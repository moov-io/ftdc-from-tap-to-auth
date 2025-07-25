package models

import (
	"math/rand"
	"strconv"
	"time"
)

func GenerateCardNumber(bin string) string {
	return completeDigits(bin, 16)
}

func completeDigits(bin string, l int) string {
	rand.Seed(time.Now().Unix())

	randomNumberLength := l - (len(bin) + 1)

	for i := 0; i < randomNumberLength; i++ {
		rand := rand.Intn(10)
		bin += strconv.Itoa(rand)
	}

	checkDigit := generateCheckDigit(bin)
	return bin + checkDigit
}

// Luhn algorithm to generate the check digit
func generateCheckDigit(cnumber string) string {
	sum := 0
	for i := 0; i < len(cnumber); i++ {
		digit, _ := strconv.Atoi((string(cnumber[i])))

		if i%2 == 0 {
			digit = digit * 2

			if digit > 9 {
				digit = (digit / 10) + (digit % 10)
			}

		}

		sum += digit

	}

	mod := sum % 10

	if mod == 0 {
		return strconv.Itoa(0)
	}

	return strconv.Itoa(10 - mod)
}

// Checks if the card number has valid check digit
func CheckLuhn(cnumber string) bool {
	if len(cnumber) < 12 || len(cnumber) > 19 {
		return false
	}
	checkDigit := cnumber[len(cnumber)-1:]
	return checkDigit == generateCheckDigit(cnumber[:len(cnumber)-1])
}
