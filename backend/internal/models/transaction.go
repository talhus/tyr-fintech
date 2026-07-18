package models

import "time"

type TransactionStatus string

const (
	StatusPending   TransactionStatus = "PENDING"
	StatusCompleted TransactionStatus = "COMPLETED"
	StatusFailed    TransactionStatus = "FAILED"
)

type Transaction struct {
	ID               string            `db:"id" json:"id"`
	FromWalletID     string            `db:"from_wallet_id" json:"from_wallet_id"`
	ToWalletID       string            `db:"to_wallet_id" json:"to_wallet_id"`
	FromWalletNumber int64             `db:"from_wallet_number" json:"from_wallet_number"`
	ToWalletNumber   int64             `db:"to_wallet_number" json:"to_wallet_number"`
	Amount           int64             `db:"amount" json:"amount"`
	ConvertedAmount  int64             `db:"converted_amount" json:"converted_amount"`
	Status           TransactionStatus `db:"status" json:"status" validate:"oneof=PENDING COMPLETED FAILED"`
	CardID           *string           `db:"card_id" json:"card_id,omitempty"`
	MerchantName     *string           `db:"merchant_name" json:"merchant_name,omitempty"`
	CreatedAt        time.Time         `db:"created_at" json:"created_at"`
}
