package models

import "time"

type CardStatus string

const (
	CardStatusActive CardStatus = "active"
	CardStatusClosed CardStatus = "closed"
	CardStatusFrozen CardStatus = "frozen"
)

type Card struct {
	ID          string
	UserID      string
	WalletID    string
	CardNumber  string
	CVV         string
	ExpiryMonth int
	ExpiryYear  int
	LimitAmount int
	SpentAmount int
	Status      CardStatus
	CreatedAt   time.Time
}

//easyjson:json
type CardCreateRequest struct {
	WalletID string `json:"wallet_id"`
	Limit    int    `json:"limit"`
}
