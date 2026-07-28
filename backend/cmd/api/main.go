package main

import (
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
		eventConsumer := worker.NewEventConsumer(rabbitmqClient.Conn(), hub)
		if err := eventConsumer.StartConsuming(); err != nil {
			log.Printf("Warning: Failed to start RabbitMQ consumer %v \n", err)
		}
	}

	// Initialize repos
	userRepo := repos.NewUserRepository(pool.DB)
	walletRepo := repos.NewWalletRepository(pool.DB)
	transactionRepo := repos.NewTransactionRepository(pool.DB)
	cardRepo := repos.NewCardRepository(pool.DB)

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

	//setup redis rate limiter
	redisLimiter := middleware.NewRedisRateLimiter(redisClient.Client)

	//register routes
	handlers.RegisterRoutes(r, userHandler, walletHandler, transactionHandler, cardHandler, notificationHandler, redisLimiter)

	// Start Gin HTTP server
	log.Printf("Starting Gin server on %v:%v", cfg.APIHost, cfg.APIPort)
	if err := r.Run(fmt.Sprintf(":%v", cfg.APIPort)); err != nil {
		log.Fatalf("Failed to run Gin server: %v", err)
	}
}
