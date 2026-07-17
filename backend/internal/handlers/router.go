package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/iamtbay/tyr-fintech/internal/middleware"
)

func RegisterRoutes(r *gin.Engine, userHandler *UserHandler, walletHandler *WalletHandler, txHandler *TransactionHandler) {
	// Public routes
	r.POST("/api/v1/auth/register", userHandler.Register)
	r.POST("/api/v1/auth/login", userHandler.Login)

	// Protected routes group
	authorized := r.Group("/api/v1")
	authorized.Use(middleware.AuthRequired())
	{
		authorized.GET("/wallets", walletHandler.GetWallets)
		authorized.POST("/wallets", walletHandler.Create)
		authorized.GET("/wallets/verify/:walletID", walletHandler.VerifyWallet)
		authorized.DELETE("/wallets/:walletID", walletHandler.DeleteWallet)
		authorized.POST("/transfer", txHandler.Transfer)
		authorized.GET("/transactions/:walletID", txHandler.GetHistory)
		authorized.GET("/transactions/:walletID/export", txHandler.ExportHistory)
		authorized.GET("/exchange-rate", txHandler.GetExchangeRate)
		authorized.POST("/logout", userHandler.Logout)
	}
}
