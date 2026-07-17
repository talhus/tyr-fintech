package dto

import "github.com/iamtbay/tyr-fintech/internal/models"

type CreateWallet struct {
	UserID   string                `json:"user_id" validate:"required"`
	Currency models.WalletCurrency `json:"currency" validate:"required"`
	Balance  int64                 `json:"balance"`
}

type WalletLookUpResult struct {
	OwnerName string
	Currency  models.WalletCurrency
}
