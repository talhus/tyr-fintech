package services_test

import (
	"testing"

	"github.com/iamtbay/tyr-fintech/internal/services"
)

func TestValidateAndSimulateCard(t *testing.T) {
	tests := []struct {
		name          string
		cardNumber    string
		expireMonth   string
		expireYear    string
		cvv           string
		expectValid   bool
		failureReason string
	}{
		{
			name:        "Deterministic Success - Ends in 0000",
			cardNumber:  "4111111111110000",
			expireMonth: "12",
			expireYear:  "28",
			cvv:         "123",
			expectValid: true,
		},
		{
			name:          "Deterministic Failure - Ends in 5001 Insufficient Funds",
			cardNumber:    "4111111111115001",
			expireMonth:   "12",
			expireYear:    "28",
			cvv:           "123",
			expectValid:   false,
			failureReason: "Insufficient funds",
		},
		{
			name:          "Deterministic Failure - Ends in 5002 Card Expired",
			cardNumber:    "4111111111115002",
			expireMonth:   "12",
			expireYear:    "28",
			cvv:           "123",
			expectValid:   false,
			failureReason: "Card expired",
		},
		{
			name:          "Deterministic Failure - Ends in 5003 Invalid CVV",
			cardNumber:    "4111111111115003",
			expireMonth:   "12",
			expireYear:    "28",
			cvv:           "123",
			expectValid:   false,
			failureReason: "Invalid CVV",
		},
		{
			name:        "Real Luhn Valid Card (4111111111111111)",
			cardNumber:  "4111111111111111",
			expireMonth: "10",
			expireYear:  "2030",
			cvv:         "999",
			expectValid: true,
		},
		{
			name:          "Invalid Luhn Checksum",
			cardNumber:    "4111111111111112", // Corrupted last digit
			expireMonth:   "10",
			expireYear:    "2030",
			cvv:         "999",
			expectValid:   false,
			failureReason: "Invalid card number",
		},
		{
			name:          "Expired Card Date",
			cardNumber:    "4111111111111111",
			expireMonth:   "01",
			expireYear:    "2020",
			cvv:           "123",
			expectValid:   false,
			failureReason: "Card expired",
		},
		{
			name:          "Invalid CVV Length",
			cardNumber:    "4111111111111111",
			expireMonth:   "12",
			expireYear:    "2028",
			cvv:           "12",
			expectValid:   false,
			failureReason: "Invalid CVV",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := services.ValidateAndSimulateCard(tt.cardNumber, tt.expireMonth, tt.expireYear, tt.cvv)
			if res.IsValid != tt.expectValid {
				t.Fatalf("expected valid %v, got %v (reason: %s)", tt.expectValid, res.IsValid, res.FailureReason)
			}
			if !tt.expectValid && res.FailureReason != tt.failureReason {
				t.Fatalf("expected failure reason %q, got %q", tt.failureReason, res.FailureReason)
			}
		})
	}
}
