package dto

import (
	"github.com/iamtbay/tyr-fintech/internal/models"
)

type CardCreateRequest struct {
	WalletID    string `json:"wallet_id" binding:"required"`
	LimitAmount int64  `json:"limit_amount" binding:"required,gt=0"`
}

type CardResponse struct {
	ID          string            `json:"id"`
	WalletID    string            `json:"wallet_id"`
	CardNumber  string            `json:"card_number"`
	ExpiryMonth int               `json:"expiry_month"`
	ExpiryYear  int               `json:"expiry_year"`
	LimitAmount int64             `json:"limit_amount"`
	SpentAmount int64             `json:"spent_amount"`
	Status      models.CardStatus `json:"status"`
	Currency    string            `json:"currency"`
}

type CardDetailResponse struct {
	ID          string            `json:"id"`
	WalletID    string            `json:"wallet_id"`
	CardNumber  string            `json:"card_number"`
	CVV         string            `json:"cvv"`
	ExpiryMonth int               `json:"expiry_month"`
	ExpiryYear  int               `json:"expiry_year"`
	LimitAmount int64             `json:"limit_amount"`
	SpentAmount int64             `json:"spent_amount"`
	Status      models.CardStatus `json:"status"`
	Currency    string            `json:"currency"`
}

// PAYMENT
type CardPaymentRequest struct {
	CardNumber   string `json:"card_number"`
	CVV          string `json:"cvv" binding:"required,len=3"`
	ExpiryMonth  int    `json:"expiry_month" binding:"required,min=1,max=12"`
	ExpiryYear   int    `json:"expiry_year" binding:"required,gt=0"`
	Amount       int64  `json:"amount" binding:"required,gt=0"`
	MerchantName string `json:"merchant_name"`
}

type CardPaymentResponse struct {
	TransactionID string            `json:"transaction_id"`
	AmountCharged int64             `json:"amount_charged"`
	Message       string            `json:"message"`
	Status        CardPaymentStatus `json:"status"`
}
type CardPaymentStatus string

const (
	CardPaymentStatusApproved CardPaymentStatus = "approved"
	CardPaymentStatusDeclined CardPaymentStatus = "declined"
)
