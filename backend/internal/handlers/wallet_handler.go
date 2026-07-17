package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/pkg/response"
	"github.com/iamtbay/tyr-fintech/pkg/utils"
)

type WalletService interface {
	CreateWallet(ctx context.Context, req *dto.CreateWallet) error
	GetByUserID(ctx context.Context, userID string) ([]*models.Wallet, error)
	DeleteWallet(ctx context.Context, userID, walletID string) error
	VerifyWallet(ctx context.Context, walletID int64) (*dto.WalletLookUpResult, error)
}

type WalletHandler struct {
	walletService WalletService
}

func NewWalletHandler(walletService WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

// CREATE WALLET
func (h *WalletHandler) Create(c *gin.Context) {
	var req dto.CreateWallet
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.walletService.CreateWallet(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Wallet created successfully"})
}

// GET WALLETS
func (h *WalletHandler) GetWallets(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	wallets, err := h.walletService.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"wallets": wallets})
}

// VERIFY WALLET
func (h *WalletHandler) VerifyWallet(c *gin.Context) {
	walletIDStr := c.Param("walletID")
	if walletIDStr == "" {
		response.Error(c, http.StatusBadRequest, "Wallet ID is required")
		return
	}
	walletID, err := strconv.ParseInt(walletIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid Wallet ID")
		return
	}
	wallet, err := h.walletService.VerifyWallet(c.Request.Context(), walletID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	wallet.OwnerName = utils.MaskString(wallet.OwnerName)
	response.Success(c, http.StatusOK, wallet)
}

// DELETE WALLET
func (h *WalletHandler) DeleteWallet(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	walletId := c.Param("walletID")
	if walletId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}
	if err := h.walletService.DeleteWallet(c.Request.Context(), userID, walletId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Wallet deleted successfully"})
}
