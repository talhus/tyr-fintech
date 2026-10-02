package models

import "time"

type AccountType string

const (
	AccountTypeAcquiringClearing AccountType = "ACQUIRING_CLEARING"
	AccountTypeMerchantPayable   AccountType = "MERCHANT_PAYABLE"
	AccountTypePlatformFeeRevenue AccountType = "PLATFORM_FEE_REVENUE"
)

type EntryType string

const (
	EntryTypeDebit  EntryType = "DEBIT"
	EntryTypeCredit EntryType = "CREDIT"
)

// SystemAccount represents an internal balance ledger account (e.g. Card settlement clearing, Gateway fee revenue)
type SystemAccount struct {
	ID          string      `db:"id" json:"id"`
	AccountType AccountType `db:"account_type" json:"account_type"`
	Currency    string      `db:"currency" json:"currency"`
	Balance     int64       `db:"balance" json:"balance"`
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
}

// LedgerEntry represents an atomic double-entry bookkeeping record
type LedgerEntry struct {
	ID          string      `db:"id" json:"id"`
	ChargeID    string      `db:"charge_id" json:"charge_id"`
	AccountType AccountType `db:"account_type" json:"account_type"`
	AccountID   string      `db:"account_id" json:"account_id"`
	EntryType   EntryType   `db:"entry_type" json:"entry_type"`
	Amount      int64       `db:"amount" json:"amount"` // strictly in minor currency units
	Currency    string      `db:"currency" json:"currency"`
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
}

// ChargeIdempotencyRecord represents cached gateway idempotency responses
type ChargeIdempotencyRecord struct {
	IdempotencyKey string    `db:"idempotency_key" json:"idempotency_key"`
	MerchantID     string    `db:"merchant_id" json:"merchant_id"`
	RequestHash    string    `db:"request_hash" json:"request_hash"`
	ResponseStatus int       `db:"response_status" json:"response_status"`
	ResponseBody   string    `db:"response_body" json:"response_body"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}
