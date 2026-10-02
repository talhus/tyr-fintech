package services

import (
	"strconv"
	"strings"
	"time"

	"github.com/iamtbay/tyr-fintech/pkg/utils"
)

type CardValidationResult struct {
	IsValid       bool
	FailureReason string
}

func ValidateAndSimulateCard(cardNumber, expireMonth, expireYear, cvv string) CardValidationResult {
	cleanCardNumber := strings.ReplaceAll(cardNumber, " ", "")
	cleanCardNumber = strings.ReplaceAll(cleanCardNumber, "-", "")

	if len(cleanCardNumber) < 13 || len(cleanCardNumber) > 19 {
		return CardValidationResult{
			IsValid:       false,
			FailureReason: "Invalid card number length",
		}
	}

	if len(cvv) < 3 || len(cvv) > 4 {
		return CardValidationResult{IsValid: false, FailureReason: "Invalid CVV"}
	}
	// 1. Deterministic Test Card Rule
	switch {
	case strings.HasSuffix(cleanCardNumber, "5001"):
		return CardValidationResult{IsValid: false, FailureReason: "Insufficient funds"}
	case strings.HasSuffix(cleanCardNumber, "5002"):
		return CardValidationResult{IsValid: false, FailureReason: "Card expired"}
	case strings.HasSuffix(cleanCardNumber, "5003"):
		return CardValidationResult{IsValid: false, FailureReason: "Invalid CVV"}
	case strings.HasSuffix(cleanCardNumber, "0000"):
		return CardValidationResult{IsValid: true}
	}

	//2 validate expiration date
	month, err := strconv.Atoi(expireMonth)
	if err != nil || month < 1 || month > 12 {
		return CardValidationResult{IsValid: false, FailureReason: "Invalid expiration month"}
	}
	year, err := strconv.Atoi(expireYear)
	if err != nil {
		return CardValidationResult{IsValid: false, FailureReason: "Invalid expiration year"}
	}
	if year < 100 {
		year += 2000
	}
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())
	if year < currentYear || (year == currentYear && month < currentMonth) {
		return CardValidationResult{IsValid: false, FailureReason: "Card expired"}
	}

	//3 Check Luhn algorithm
	if !utils.ValidateLuhn(cleanCardNumber) {
		return CardValidationResult{IsValid: false, FailureReason: "Invalid card number"}
	}
	return CardValidationResult{
		IsValid: true,
	}
}
