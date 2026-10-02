package dto

import "time"

type CreateCheckoutSessionRequest struct {
	OrderID     string `json:"order_id" binding:"required"`
	Amount      int64  `json:"amount" binding:"required,gt=0"`
	Currency    string `json:"currency" binding:"required,len=3"`
	CallbackURL string `json:"callback_url" binding:"required,url"`
	WebhookURL  string `json:"webhook_url" binding:"required,url"`
}

type CreateCheckoutSessionResponse struct {
	SessionID   string    `json:"session_id"`
	CheckoutURL string    `json:"checkout_url"`
	Status      string    `json:"status"` // "PENDING"
	ExpiresAt   time.Time `json:"expires_at"`
}

type CheckoutSessionDetailsResponse struct {
	SessionID    string    `json:"session_id"`
	MerchantName string    `json:"merchant_name"` // "Foodeli Inc."
	OrderID      string    `json:"order_id"`
	Amount       int64     `json:"amount"`
	Currency     string    `json:"currency"`
	Status       string    `json:"status"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type PayCheckoutSessionRequest struct {
	WalletID string `json:"wallet_id" binding:"required,uuid"`
}

type PayCheckoutSessionResponse struct {
	Success       bool   `json:"success"`
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"` // "SUCCEEDED"
	RedirectURL   string `json:"redirect_url"`
}

type MerchantWebhookEvent struct {
	Event         string `json:"event"`
	TransactionID string `json:"transaction_id"`
	SessionID     string `json:"session_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
	WebhookURL    string `json:"webhook_url"`
	Timestamp     int64  `json:"timestamp"`
}
