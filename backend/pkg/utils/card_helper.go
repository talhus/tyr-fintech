package utils

import (
	"crypto/rand"
	"math/big"
	"strconv"
)

// GENERATE CARD NUMBER
func GenerateCardNumber() string {
	baseNum := "411111"
	max := big.NewInt(900000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return ""
	}
	randomNumber := n.Int64() + 100000000
	prefix := baseNum + strconv.FormatInt(randomNumber, 10)
	checkDigit := calculateLuhnCheckDigit(prefix)
	return prefix + checkDigit
}

// Luhn algorithm implementation
// Source: https://en.wikipedia.org/wiki/Luhn_algorithm
func calculateLuhnCheckDigit(prefix string) string {
	sum := 0
	alternate := true
	for i := len(prefix) - 1; i >= 0; i-- {
		num := int(prefix[i] - '0')
		if alternate {
			num *= 2
		}
		if num > 9 {
			num -= 9
		}
		sum += num
		alternate = !alternate
	}
	checkDigit := (10 - (sum % 10)) % 10
	return strconv.Itoa(checkDigit)
}

// GENERATE CVV
func GenerateCVV() string {
	max := big.NewInt(900)
	n, _ := rand.Int(rand.Reader, max)
	cvv := n.Int64() + 100
	return strconv.FormatInt(cvv, 10)
}
