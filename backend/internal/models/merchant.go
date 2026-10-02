package models

import "time"

type MerchantStatus string

const (
	MerchantStatusActive    MerchantStatus = "ACTIVE"
	MerchantStatusSuspended MerchantStatus = "SUSPENDED"
)

type APIKeyEnvironment string

const (
	EnvLive APIKeyEnvironment = "live"
	EnvTest APIKeyEnvironment = "test"
)

type APIKeyStatus string

const (
	APIKeyStatusActive  APIKeyStatus = "ACTIVE"
	APIKeyStatusRevoked APIKeyStatus = "REVOKED"
)

// Merchant represents a registered B2B partner (e.g., Foodeli)
type Merchant struct {
	ID                       string         `db:"id" json:"id"`
	Name                     string         `db:"name" json:"name"`
	Email                    string         `db:"email" json:"email"`
	SettlementWalletID       string         `db:"settlement_wallet_id" json:"settlement_wallet_id"`
	FeePercentageBasisPoints int            `db:"fee_percentage_basis_points" json:"fee_percentage_basis_points"` // 200 = 2.00%
	Status                   MerchantStatus `db:"status" json:"status"`
	CreatedAt                time.Time      `db:"created_at" json:"created_at"`
}

// MerchantAPIKey represents hashed API credentials for a merchant
type MerchantAPIKey struct {
	ID          string            `db:"id" json:"id"`
	MerchantID  string            `db:"merchant_id" json:"merchant_id"`
	KeyHash     string            `db:"key_hash" json:"-"`
	Prefix      string            `db:"prefix" json:"prefix"`
	MaskedKey   string            `db:"masked_key" json:"masked_key"`
	Environment APIKeyEnvironment `db:"environment" json:"environment"`
	Status      APIKeyStatus      `db:"status" json:"status"`
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
}
