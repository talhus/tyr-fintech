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
