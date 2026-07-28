package dto

import "github.com/iamtbay/tyr-fintech/internal/models"

type CreateWallet struct {
	UserID   string                `json:"user_id"`
	Currency models.WalletCurrency `json:"currency" binding:"required,oneof=TRY USD EUR"`
	Balance  int64                 `json:"balance"`
}

type WalletLookUpResult struct {
	OwnerName string                `json:"owner_name"`
	Currency  models.WalletCurrency `json:"currency"`
}
