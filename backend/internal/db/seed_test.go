package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/iamtbay/tyr-fintech/internal/db"
	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/repos"
	"github.com/iamtbay/tyr-fintech/internal/services"
	"github.com/iamtbay/tyr-fintech/pkg/encryption"
)

func TestSeedDemoUserAndLogin(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://admin:secretpassword@localhost:55432/fintech?sslmode=disable"
	}

	pool, err := db.Connect(dsn)
	if err != nil {
		t.Skipf("Skipping integration test: database connection failed (%v)", err)
		return
	}
	defer pool.Close()

	encryptor, err := encryption.NewAESEncryptor("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	// 1. Seed demo user
	db.SeedDemoUser(pool.DB, encryptor)

	// 2. Verify demo user exists in database
	userRepo := repos.NewUserRepository(pool.DB)
	userService := services.NewUserService(userRepo)

	ctx := context.Background()
	user, err := userRepo.GetByEmail(ctx, "demo@tyr.com")
	if err != nil {
		t.Fatalf("Demo user was not found in DB: %v", err)
	}

	if user.Email != "demo@tyr.com" {
		t.Fatalf("Expected email demo@tyr.com, got %s", user.Email)
	}

	// 3. Test UserService.Login with demo password
	loginResp, err := userService.Login(ctx, &dto.LoginUserRequest{
		Email:    "demo@tyr.com",
		Password: "demo123456",
	})
	if err != nil {
		t.Fatalf("Login for demo user failed: %v", err)
	}
	if loginResp.User.Email != "demo@tyr.com" {
		t.Fatalf("Expected logged in user email demo@tyr.com, got %s", loginResp.User.Email)
	}

	// Check wallets
	walletRepo := repos.NewWalletRepository(pool.DB)
	wallets, err := walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to query demo wallets: %v", err)
	}
	if len(wallets) < 3 {
		t.Fatalf("Expected at least 3 pre-funded wallets (TRY, USD, EUR), got %d", len(wallets))
	}

	// Check card
	cardRepo := repos.NewCardRepository(pool.DB, encryptor)
	cards, err := cardRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to query demo cards: %v", err)
	}
	if len(cards) < 1 {
		t.Fatalf("Expected at least 1 virtual card, got %d", len(cards))
	}

	if cards[0].CardNumber != "4111111111111111" {
		t.Fatalf("Expected decrypted card number 4111111111111111, got %s", cards[0].CardNumber)
	}
	if cards[0].CVV != "123" {
		t.Fatalf("Expected decrypted CVV 123, got %s", cards[0].CVV)
	}

	t.Logf("Demo User Seed Verification PASSED: User ID=%s, Wallets=%d, Card PAN=%s, CVV=%s",
		user.ID, len(wallets), cards[0].CardNumber, cards[0].CVV)
}
