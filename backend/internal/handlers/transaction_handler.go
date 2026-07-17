package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/pkg/export"
	"github.com/iamtbay/tyr-fintech/pkg/response"
)

type TransactionService interface {
	Transfer(ctx context.Context, req *dto.TransferRequest) error
	GetHistory(ctx context.Context, walletID string) ([]*models.Transaction, error)
	GetExchangeRate(ctx context.Context, from, to models.WalletCurrency) (float64, error)
}

type TransactionHandler struct {
	transactionService TransactionService
}

func NewTransactionHandler(transactionService TransactionService) *TransactionHandler {
	return &TransactionHandler{transactionService: transactionService}
}

// HANDLER FUNCTIONS
func (h *TransactionHandler) Transfer(c *gin.Context) {
	var req dto.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve idempotency key from header if present
	if idempKey := c.GetHeader("X-Idempotency-Key"); idempKey != "" {
		req.TransactionID = idempKey
	}

	if err := h.transactionService.Transfer(c.Request.Context(), &req); err != nil {
		if err.Error() == "idempotency key is required" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "transaction already processed" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "insufficient balance" || err.Error() == "wallet not found" {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "currency mismatch: cannot transfer between different currencies" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer completed successfully", "transaction_id": req.TransactionID})
}

func (h *TransactionHandler) GetHistory(c *gin.Context) {
	walletID := c.Param("walletID")
	if walletID == "" {
		response.Error(c, http.StatusBadRequest, "wallet id can't be empty")
		return
	}

	txns, err := h.transactionService.GetHistory(c.Request.Context(), walletID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, txns)
}

func (h *TransactionHandler) ExportHistory(c *gin.Context) {
	walletID := c.Param("walletID")
	if walletID == "" {
		response.Error(c, http.StatusBadRequest, "wallet id can't be empty")
		return
	}

	//
	format := c.DefaultQuery("format", "csv")

	//
	txns, err := h.transactionService.GetHistory(c.Request.Context(), walletID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	switch format {
	case "pdf":
		pdfBytes, err := export.TransactionsToPDF(walletID, txns)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "PDF couldn't produce: "+err.Error())
			return
		}

		c.Header("Content-Disposition", "attachment; filename=statement.pdf")
		c.Data(http.StatusOK, "application/pdf", pdfBytes)
	case "csv":
		csvBytes, err := export.TransactionsToCSV(txns)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "CSV couldn't produce: "+err.Error())
			return
		}

		c.Header("Content-Disposition", "attachment; filename=statement.csv")
		c.Data(http.StatusOK, "text/csv", csvBytes)
	default:
		response.Error(c, http.StatusBadRequest, "unsupported format")
		return
	}

}

func (h *TransactionHandler) GetExchangeRate(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		response.Error(c, http.StatusBadRequest, "from and to currencies are required")
		return
	}
	rate, err := h.transactionService.GetExchangeRate(c.Request.Context(), models.WalletCurrency(from), models.WalletCurrency(to))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, gin.H{"rate": rate})
}
