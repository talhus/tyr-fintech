package export

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/iamtbay/tyr-fintech/internal/models"
)

func TransactionsToCSV(transactions []*models.Transaction) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	headers := []string{
		"Transaction ID", "Sender Wallet", "Receiver Wallet", "Description", "Amount", "Converted Amount", "Status", "Date",
	}
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	for _, tx := range transactions {
		description := "Transfer"
		if tx.MerchantName != nil && *tx.MerchantName != "" {
			description = fmt.Sprintf("Card: %s", *tx.MerchantName)
		}

		sender := fmt.Sprintf("%d", tx.FromWalletNumber)
		if tx.FromWalletNumber == 0 {
			sender = "N/A"
		}

		receiver := fmt.Sprintf("%d", tx.ToWalletNumber)
		if tx.ToWalletNumber == 0 {
			receiver = "N/A"
		}

		row := []string{
			tx.ID,
			sender,
			receiver,
			description,
			fmt.Sprintf("%.2f", float64(tx.Amount)/100.0),
			fmt.Sprintf("%.2f", float64(tx.ConvertedAmount)/100.0),
			string(tx.Status),
			tx.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
