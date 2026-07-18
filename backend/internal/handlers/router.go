package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/iamtbay/tyr-fintech/internal/middleware"
)

func RegisterRoutes(r *gin.Engine, userHandler *UserHandler, walletHandler *WalletHandler, txHandler *TransactionHandler, cardHandler *CardHandler) {
	// Public routes
	r.POST("/api/v1/auth/register", userHandler.Register)
	r.POST("/api/v1/auth/login", userHandler.Login)

	// Protected routes group
	authorized := r.Group("/api/v1")
	authorized.Use(middleware.AuthRequired())
	{
		//wallets
		authorized.GET("/wallets", walletHandler.GetWallets)
		authorized.POST("/wallets", walletHandler.Create)
		authorized.GET("/wallets/verify/:walletID", walletHandler.VerifyWallet)
		authorized.DELETE("/wallets/:walletID", walletHandler.DeleteWallet)
		//transactions
		authorized.POST("/transfer", txHandler.Transfer)
		authorized.GET("/transactions/:walletID", txHandler.GetHistory)
		authorized.GET("/transactions/:walletID/export", txHandler.ExportHistory)
		authorized.GET("/exchange-rate", txHandler.GetExchangeRate)
		//cards
		authorized.POST("/cards", cardHandler.CreateCard)
		authorized.GET("/cards", cardHandler.GetCards)
		authorized.GET("/cards/:cardID/details", cardHandler.GetCardDetails)
		authorized.GET("/cards/:cardID/transactions", cardHandler.GetCardTransactions)
		authorized.DELETE("/cards/:cardID", cardHandler.CloseCard)
		authorized.POST("/cards/:cardID/freeze", cardHandler.FreezeCard)
		authorized.POST("/cards/:cardID/unfreeze", cardHandler.ActivateCard)
		authorized.POST("/cards/:cardID/process-payment", cardHandler.ProcessPayment)
		//auth
		authorized.POST("/logout", userHandler.Logout)
	}
}
