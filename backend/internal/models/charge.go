package models

import "time"

type ChargeStatus string

const (
	ChargeStatusPending   ChargeStatus = "PENDING"
	ChargeStatusSucceeded ChargeStatus = "SUCCEEDED"
	ChargeStatusFailed    ChargeStatus = "FAILED"
)

// Charge represents an incoming payment charge processed through the gateway
type Charge struct {
	ID             string       `db:"id" json:"id"` // trx_...
	MerchantID     string       `db:"merchant_id" json:"merchant_id"`
	Amount         int64        `db:"amount" json:"amount"` // in minor currency units (e.g. 10000 = 100.00 TRY)
	Currency       string       `db:"currency" json:"currency"`
	FeeAmount      int64        `db:"fee_amount" json:"fee_amount"`
	NetAmount      int64        `db:"net_amount" json:"net_amount"`
	Status         ChargeStatus `db:"status" json:"status"`
	FailureReason  string       `db:"failure_reason" json:"failure_reason"`
	CardLast4      string       `db:"card_last4" json:"card_last4"`
	CardHolderName string       `db:"card_holder_name" json:"card_holder_name"`
	CreatedAt      time.Time    `db:"created_at" json:"created_at"`
}
