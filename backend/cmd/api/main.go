package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iamtbay/tyr-fintech/config"
	"github.com/iamtbay/tyr-fintech/internal/db"
	"github.com/iamtbay/tyr-fintech/internal/handlers"
	"github.com/iamtbay/tyr-fintech/internal/middleware"
	"github.com/iamtbay/tyr-fintech/internal/notifications"
	"github.com/iamtbay/tyr-fintech/internal/queue"
	"github.com/iamtbay/tyr-fintech/internal/repos"
	"github.com/iamtbay/tyr-fintech/internal/services"
	"github.com/iamtbay/tyr-fintech/internal/worker"
	"github.com/iamtbay/tyr-fintech/pkg/encryption"
)

func main() {
	cfg := config.New()
	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to the database", err)
	}
	defer pool.Close()

	//connect to redis
	redisClient, err := db.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v \n", err)
	} else {
		defer redisClient.Close()
	}

	//rabbit mq
	rabbitmqClient, err := queue.NewRabbitMQ(cfg.RabbitmqURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to RabbitMQ: %v \n", err)
	} else {
		defer rabbitmqClient.Close()
	}

	//start worker
	go worker.StartWebhookWorker()

	// Initialize notifications
	hub := notifications.NewHub()
	notificationService := notifications.NewNotificationService(hub)

	//start rabbitmq consumer if connection succeeded
	if rabbitmqClient != nil {
		// Initialize durable RabbitMQ webhook dispatcher
		worker.InitRabbitMQDispatcher(rabbitmqClient)

		// Start durable merchant webhook consumer with exponential backoff & DLQ
		if err := worker.StartMerchantWebhookConsumer(rabbitmqClient.Conn()); err != nil {
			log.Printf("Warning: Failed to start RabbitMQ merchant webhook consumer: %v\n", err)
		}

		eventConsumer := worker.NewEventConsumer(rabbitmqClient.Conn(), hub)
		if err := eventConsumer.StartConsuming(); err != nil {
			log.Printf("Warning: Failed to start RabbitMQ consumer %v \n", err)
		}
	}

	//encryptor
	encryptor, err := encryption.NewAESEncryptor(cfg.CardEncryptionKey)
	if err != nil {
		log.Fatalf("Failed to create encryptor: %v\n", err)
	}
	
	// Initialize repos
	userRepo := repos.NewUserRepository(pool.DB)
	walletRepo := repos.NewWalletRepository(pool.DB)
	transactionRepo := repos.NewTransactionRepository(pool.DB)
	cardRepo := repos.NewCardRepository(pool.DB, encryptor)
	merchantRepo := repos.NewMerchantRepository(pool.DB)
	chargeRepo := repos.NewChargeRepository(pool.DB)
	checkoutRepo := repos.NewCheckoutRepository(pool.DB)

	// Seed demo user for recruiters / instant testing (skip in production)
	if cfg.Env != "production" {
		db.SeedDemoUser(pool.DB, encryptor)
		db.SeedGatewayEntities(pool.DB)
	} else {
		if err := db.EnsureGatewayTables(context.Background(), pool.DB); err != nil {
			log.Printf("Warning: Failed to ensure gateway tables: %v\n", err)
		}
	}

	//initialize exchange rate sv warpped w redis cache
	rawExchangeService := services.NewMockExchangeService()
	var exchangeService services.ExchangeRateProvider = rawExchangeService
	if redisClient != nil {
		//wrap mock exchange sv with 10-minute redis caching
		exchangeService = services.NewCachedExchangeService(rawExchangeService, redisClient.Client, 10*time.Minute)
	}

	// Initialize services
	userService := services.NewUserService(userRepo)
	walletService := services.NewWalletService(walletRepo)

	//transaction service & card service
	transactionService := services.NewTransactionService(transactionRepo, exchangeService, walletRepo, notificationService)
	cardService := services.NewCardService(cardRepo, notificationService)
	chargeService := services.NewChargeService(chargeRepo)
	checkoutService := services.NewCheckoutService(checkoutRepo, merchantRepo, cfg.FrontendURL)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	walletHandler := handlers.NewWalletHandler(walletService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	cardHandler := handlers.NewCardHandler(cardService)
	notificationHandler := handlers.NewNotificationHandler(hub)
	chargeHandler := handlers.NewChargeHandler(chargeService)
	checkoutHandler := handlers.NewCheckoutHandler(checkoutService)

	// Setup Gin router
	r := gin.Default()

	// CORS middleware with whitelist (configurable via ALLOWED_ORIGINS and localhost in dev)
	r.Use(middleware.CORSMiddleware(cfg.Env, cfg.AllowedOrigins))

	//setup redis rate limiter
	var redisLimiter *middleware.RedisRateLimiter
	if redisClient != nil {
		redisLimiter = middleware.NewRedisRateLimiter(redisClient.Client)
	} else {
		redisLimiter = middleware.NewRedisRateLimiter(nil)
	}

	//register routes
	handlers.RegisterRoutes(r, userHandler, walletHandler, transactionHandler, cardHandler, notificationHandler, chargeHandler, checkoutHandler, merchantRepo, redisLimiter)

	// Start Gin HTTP server
	log.Printf("Starting Gin server on %v:%v", cfg.APIHost, cfg.APIPort)
	if err := r.Run(fmt.Sprintf(":%v", cfg.APIPort)); err != nil {
		log.Fatalf("Failed to run Gin server: %v", err)
	}
}
