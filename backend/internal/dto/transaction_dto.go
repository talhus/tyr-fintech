package dto

type TransferRequest struct {
	FromWalletNumber int64  `json:"from_wallet_number"`
	ToWalletNumber   int64  `json:"to_wallet_number"`
	Amount           int64  `json:"amount"`
	TransactionID    string `json:"transaction_id"`
}

type TransactionWebhookEvent struct {
	TransactionID    string `json:"transaction_id"`
	FromWalletNumber int64  `json:"from_wallet_number"`
	ToWalletNumber   int64  `json:"to_wallet_number"`
	Amount           int64  `json:"amount"`
	Status           string `json:"status"`
}
