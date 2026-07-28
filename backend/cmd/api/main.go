package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/iamtbay/tyr-fintech/config"
	"github.com/iamtbay/tyr-fintech/internal/db"
	"github.com/iamtbay/tyr-fintech/internal/handlers"
	"github.com/iamtbay/tyr-fintech/internal/notifications"
	"github.com/iamtbay/tyr-fintech/internal/repos"
	"github.com/iamtbay/tyr-fintech/internal/services"
	"github.com/iamtbay/tyr-fintech/internal/worker"
)

func main() {
	cfg := config.New()
	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to the database", err)
	}
	defer pool.Close()

	//start worker
	go worker.StartWebhookWorker()

	// Initialize notifications
	hub := notifications.NewHub()
	notificationService := notifications.NewNotificationService(hub)

	// Initialize repos
	userRepo := repos.NewUserRepository(pool.DB)
	walletRepo := repos.NewWalletRepository(pool.DB)
	transactionRepo := repos.NewTransactionRepository(pool.DB)
	cardRepo := repos.NewCardRepository(pool.DB)

	// Initialize services
	userService := services.NewUserService(userRepo)
	walletService := services.NewWalletService(walletRepo)
	//mock exchange
	exchangeService := services.NewMockExchangeService()
	//transaction service & card service
	transactionService := services.NewTransactionService(transactionRepo, exchangeService, walletRepo, notificationService)
	cardService := services.NewCardService(cardRepo, notificationService)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	walletHandler := handlers.NewWalletHandler(walletService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	cardHandler := handlers.NewCardHandler(cardService)
	notificationHandler := handlers.NewNotificationHandler(hub)

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("http://%v", cfg.APIHost))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Idempotency-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	handlers.RegisterRoutes(r, userHandler, walletHandler, transactionHandler, cardHandler, notificationHandler)

	// Start Gin HTTP server
	log.Printf("Starting Gin server on %v:%v", cfg.APIHost, cfg.APIPort)
	if err := r.Run(fmt.Sprintf(":%v", cfg.APIPort)); err != nil {
		log.Fatalf("Failed to run Gin server: %v", err)
	}
}
