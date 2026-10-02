package models

import "time"

type CheckoutSessionStatus string

const (
	CheckoutStatusPending CheckoutSessionStatus = "PENDING"
	CheckoutStatusPaid    CheckoutSessionStatus = "PAID"
	CheckoutStatusExpired CheckoutSessionStatus = "EXPIRED"
	CheckoutStatusFailed  CheckoutSessionStatus = "FAILED"
)

type CheckoutSession struct {
	ID            string                `db:"id" json:"id"`
	MerchantID    string                `db:"merchant_id" json:"merchant_id"`
	OrderID       string                `db:"order_id" json:"order_id"`
	Amount        int64                 `db:"amount" json:"amount"`
	Currency      string                `db:"currency" json:"currency"`
	CallbackURL   string                `db:"callback_url" json:"callback_url"`
	WebHookURL    string                `db:"webhook_url" json:"webhook_url"`
	Status        CheckoutSessionStatus `db:"status" json:"status"`
	PayerUserID   *string               `db:"payer_user_id" json:"payer_user_id"`
	TransactionID *string               `db:"transaction_id" json:"transaction_id"`
	ExpiresAt     time.Time             `db:"expires_at" json:"expires_at"`
	CreatedAt     time.Time             `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time             `db:"updated_at" json:"updated_at"`
}
