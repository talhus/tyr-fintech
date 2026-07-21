package models

import "time"

type CardStatus string

const (
	CardStatusActive CardStatus = "ACTIVE"
	CardStatusClosed CardStatus = "CLOSED"
	CardStatusFrozen CardStatus = "FROZEN"
)

type Card struct {
	ID          string
	UserID      string
	WalletID    string
	CardNumber  string
	CVV         string
	ExpiryMonth int
	ExpiryYear  int
	LimitAmount int64
	SpentAmount int64
	Status      CardStatus
	Currency    string
	CreatedAt   time.Time
}

type CardPaymentResult struct {
	TransactionID string `db:"transaction_id" json:"transaction_id"`
	UserID        string `db:"user_id" json:"user_id"`
	UserEmail     string `db:"email" json:"email"`
	UserName      string `db:"name" json:"name"`
	MerchantName  string `db:"merchant_name" json:"merchant_name"`
	Amount        int64  `db:"amount" json:"amount"`
}
