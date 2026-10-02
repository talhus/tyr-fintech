package dto

type ChargeRequest struct {
	Amount         int64  `json:"amount" binding:"required,gt=0"`
	Currency       string `json:"currency" binding:"required,len=3"`
	CardNumber     string `json:"card_number" binding:"required,min=12,max=19"`
	CardHolderName string `json:"card_holder_name" binding:"required"`
	ExpireMonth    string `json:"expire_month" binding:"required"`
	ExpireYear     string `json:"expire_year" binding:"required"`
	CVV            string `json:"cvv" binding:"required,min=3,max=4"`
}

type ChargeResponse struct {
	Success       bool   `json:"success"`
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	FailureReason string `json:"failure_reason"`
}
