package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/iamtbay/tyr-fintech/pkg/encryption"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// SeedDemoUser ensures a demo recruiter user with funded wallets and virtual card exists in PostgreSQL.
func SeedDemoUser(pool *pgxpool.Pool, encryptor encryption.Encryptor) {
	ctx := context.Background()

	demoEmail := "demo@tyr.com"
	demoPassword := "demo123456"

	var userID string
	err := pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, demoEmail).Scan(&userID)
	if err != nil {
		// User does not exist, create demo user
		userID = uuid.New().String()
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(demoPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Warning: Failed to hash demo password: %v\n", err)
			return
		}

		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`,
			userID, "Recruiter Demo User", demoEmail, string(hashedPassword))
		if err != nil {
			log.Printf("Warning: Failed to seed demo user: %v\n", err)
			return
		}
	}

	// Ensure pre-funded wallets exist (TRY, USD, EUR)
	currencies := []struct {
		code    string
		balance int64
	}{
		{"TRY", 2500000}, // 25,000.00 TRY
		{"USD", 150000},  // 1,500.00 USD
		{"EUR", 100000},  // 1,000.00 EUR
	}

	var tryWalletID string
	for _, c := range currencies {
		var wID string
		err := pool.QueryRow(ctx, `SELECT id FROM wallets WHERE user_id = $1 AND currency = $2 AND deleted_at IS NULL`, userID, c.code).Scan(&wID)
		if err != nil {
			// Wallet missing, create it
			wID = uuid.New().String()
			_, err = pool.Exec(ctx, `INSERT INTO wallets (id, user_id, currency, balance) VALUES ($1, $2, $3, $4)`,
				wID, userID, c.code, c.balance)
			if err != nil {
				log.Printf("Warning: Failed to seed demo wallet (%s): %v\n", c.code, err)
			}
		}
		if c.code == "TRY" {
			tryWalletID = wID
		}
	}

	// Ensure virtual card exists for TRY wallet
	if tryWalletID != "" && encryptor != nil {
		var cardCount int
		_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM cards WHERE wallet_id = $1 AND status != 'CLOSED'`, tryWalletID).Scan(&cardCount)
		if cardCount == 0 {
			encNumber, err1 := encryptor.Encrypt("4111111111111111")
			encCVV, err2 := encryptor.Encrypt("123")
			if err1 == nil && err2 == nil {
				cardID := uuid.New().String()
				_, err = pool.Exec(ctx, `INSERT INTO cards (id, user_id, wallet_id, card_number, cvv, expiry_month, expiry_year, limit_amount, spent_amount, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
					cardID, userID, tryWalletID, encNumber, encCVV, 12, 2028, 500000, 0, "ACTIVE")
				if err != nil {
					log.Printf("Warning: Failed to seed demo card: %v\n", err)
				}
			}
		}
	}

	fmt.Println("Demo user successfully verified/seeded (demo@tyr.com / demo123456)")
}

const (
	DefaultFoodeliLiveAPIKey = "tyr_live_foodeli_secret_key_12345"
	DefaultFoodeliTestAPIKey = "tyr_test_foodeli_secret_key_12345"
)

// EnsureGatewayTables ensures that the payment gateway and ledger tables exist.
func EnsureGatewayTables(ctx context.Context, pool *pgxpool.Pool) error {
	ddl := `
	CREATE TABLE IF NOT EXISTS merchants (
	    id UUID PRIMARY KEY,
	    name VARCHAR(255) NOT NULL,
	    email VARCHAR(255) NOT NULL UNIQUE,
	    settlement_wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE RESTRICT,
	    fee_percentage_basis_points INT NOT NULL DEFAULT 200,
	    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'SUSPENDED')),
	    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS merchant_api_keys (
	    id UUID PRIMARY KEY,
	    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
	    key_hash VARCHAR(64) NOT NULL UNIQUE,
	    prefix VARCHAR(20) NOT NULL,
	    masked_key VARCHAR(50) NOT NULL,
	    environment VARCHAR(20) NOT NULL DEFAULT 'live' CHECK (environment IN ('live', 'test')),
	    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'REVOKED')),
	    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS system_accounts (
	    id UUID PRIMARY KEY,
	    account_type VARCHAR(50) NOT NULL UNIQUE CHECK (account_type IN ('ACQUIRING_CLEARING', 'PLATFORM_FEE_REVENUE')),
	    currency VARCHAR(3) NOT NULL,
	    balance BIGINT NOT NULL DEFAULT 0,
	    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS charges (
	    id VARCHAR(64) PRIMARY KEY,
	    merchant_id UUID NOT NULL REFERENCES merchants(id),
	    amount BIGINT NOT NULL CHECK (amount > 0),
	    currency VARCHAR(3) NOT NULL,
	    fee_amount BIGINT NOT NULL DEFAULT 0,
	    net_amount BIGINT NOT NULL DEFAULT 0,
	    status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'SUCCEEDED', 'FAILED')),
	    failure_reason VARCHAR(255) DEFAULT '',
	    card_last4 VARCHAR(4) NOT NULL,
	    card_holder_name VARCHAR(255) NOT NULL,
	    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS ledger_entries (
	    id UUID PRIMARY KEY,
	    charge_id VARCHAR(64) NOT NULL REFERENCES charges(id) ON DELETE CASCADE,
	    account_type VARCHAR(50) NOT NULL CHECK (account_type IN ('ACQUIRING_CLEARING', 'MERCHANT_PAYABLE', 'PLATFORM_FEE_REVENUE')),
	    account_id UUID NOT NULL,
	    entry_type VARCHAR(10) NOT NULL CHECK (entry_type IN ('DEBIT', 'CREDIT')),
	    amount BIGINT NOT NULL CHECK (amount > 0),
	    currency VARCHAR(3) NOT NULL,
	    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS charge_idempotency (
	    idempotency_key VARCHAR(255) NOT NULL,
	    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
	    request_hash VARCHAR(64) NOT NULL,
	    response_status INT NOT NULL,
	    response_body TEXT NOT NULL,
	    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	    PRIMARY KEY (merchant_id, idempotency_key)
	);

	CREATE TABLE IF NOT EXISTS checkout_sessions (
	    id VARCHAR(64) PRIMARY KEY,
	    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
	    order_id VARCHAR(255) NOT NULL,
	    amount BIGINT NOT NULL CHECK (amount > 0),
	    currency VARCHAR(3) NOT NULL,
	    callback_url TEXT NOT NULL,
	    webhook_url TEXT NOT NULL,
	    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PAID', 'EXPIRED', 'FAILED')),
	    payer_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
	    transaction_id VARCHAR(64) DEFAULT NULL,
	    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
	    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_charges_merchant_id ON charges(merchant_id);
	CREATE INDEX IF NOT EXISTS idx_ledger_entries_charge_id ON ledger_entries(charge_id);
	CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_hash ON merchant_api_keys(key_hash);
	CREATE INDEX IF NOT EXISTS idx_checkout_sessions_merchant ON checkout_sessions(merchant_id);
	CREATE INDEX IF NOT EXISTS idx_checkout_sessions_order ON checkout_sessions(merchant_id, order_id);
	`
	_, err := pool.Exec(ctx, ddl)
	return err
}

// SeedGatewayEntities ensures system ledger accounts, Foodeli merchant, and API keys are seeded.
func SeedGatewayEntities(pool *pgxpool.Pool) {
	ctx := context.Background()

	if err := EnsureGatewayTables(ctx, pool); err != nil {
		log.Printf("Warning: Failed to ensure gateway tables exist: %v\n", err)
	}

	// 1. Seed System Accounts (ACQUIRING_CLEARING & PLATFORM_FEE_REVENUE)
	systemAccounts := []string{"ACQUIRING_CLEARING", "PLATFORM_FEE_REVENUE"}
	for _, accType := range systemAccounts {
		var existingID string
		err := pool.QueryRow(ctx, `SELECT id FROM system_accounts WHERE account_type = $1`, accType).Scan(&existingID)
		if err != nil {
			id := uuid.New().String()
			_, err = pool.Exec(ctx, `INSERT INTO system_accounts (id, account_type, currency, balance) VALUES ($1, $2, $3, $4)`,
				id, accType, "TRY", 0)
			if err != nil {
				log.Printf("Warning: Failed to seed system account (%s): %v\n", accType, err)
			}
		}
	}

	// 2. Seed Foodeli Merchant User & Settlement Wallet
	foodeliEmail := "merchant@foodeli.com"
	var foodeliUserID string
	err := pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, foodeliEmail).Scan(&foodeliUserID)
	if err != nil {
		foodeliUserID = uuid.New().String()
		hash, _ := bcrypt.GenerateFromPassword([]byte("foodeli_secure_pass"), bcrypt.DefaultCost)
		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`,
			foodeliUserID, "Foodeli Merchant", foodeliEmail, string(hash))
		if err != nil {
			log.Printf("Warning: Failed to seed foodeli user: %v\n", err)
			return
		}
	}

	var settlementWalletID string
	err = pool.QueryRow(ctx, `SELECT id FROM wallets WHERE user_id = $1 AND currency = 'TRY' AND deleted_at IS NULL`, foodeliUserID).Scan(&settlementWalletID)
	if err != nil {
		settlementWalletID = uuid.New().String()
		_, err = pool.Exec(ctx, `INSERT INTO wallets (id, user_id, currency, balance) VALUES ($1, $2, $3, $4)`,
			settlementWalletID, foodeliUserID, "TRY", 0)
		if err != nil {
			log.Printf("Warning: Failed to seed Foodeli settlement wallet: %v\n", err)
			return
		}
	}

	// 3. Seed Merchant Record
	var merchantID string
	err = pool.QueryRow(ctx, `SELECT id FROM merchants WHERE email = $1`, foodeliEmail).Scan(&merchantID)
	if err != nil {
		merchantID = uuid.New().String()
		_, err = pool.Exec(ctx, `INSERT INTO merchants (id, name, email, settlement_wallet_id, fee_percentage_basis_points, status) VALUES ($1, $2, $3, $4, $5, $6)`,
			merchantID, "Foodeli Inc.", foodeliEmail, settlementWalletID, 200, "ACTIVE")
		if err != nil {
			log.Printf("Warning: Failed to seed Foodeli merchant: %v\n", err)
			return
		}
	}

	// 4. Seed API Keys (Live and Test)
	liveKey := DefaultFoodeliLiveAPIKey
	if envLive := os.Getenv("FOODELI_LIVE_API_KEY"); envLive != "" {
		liveKey = envLive
	}
	testKey := DefaultFoodeliTestAPIKey
	if envTest := os.Getenv("FOODELI_TEST_API_KEY"); envTest != "" {
		testKey = envTest
	}

	keys := []struct {
		rawKey string
		prefix string
		env    string
	}{
		{liveKey, "tyr_live_", "live"},
		{testKey, "tyr_test_", "test"},
	}

	for _, k := range keys {
		hasher := sha256.New()
		hasher.Write([]byte(k.rawKey))
		keyHash := hex.EncodeToString(hasher.Sum(nil))

		var existingKeyID string
		err := pool.QueryRow(ctx, `SELECT id FROM merchant_api_keys WHERE key_hash = $1`, keyHash).Scan(&existingKeyID)
		if err != nil {
			keyID := uuid.New().String()
			maskedKey := fmt.Sprintf("%s••••%s", k.prefix, k.rawKey[len(k.rawKey)-4:])
			_, err = pool.Exec(ctx, `INSERT INTO merchant_api_keys (id, merchant_id, key_hash, prefix, masked_key, environment, status) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				keyID, merchantID, keyHash, k.prefix, maskedKey, k.env, "ACTIVE")
			if err != nil {
				log.Printf("Warning: Failed to seed API key (%s): %v\n", k.env, err)
			}
		}
	}

	fmt.Println("Gateway entities successfully verified/seeded (Foodeli API Key: tyr_live_foodeli_secret_key_12345)")
}

