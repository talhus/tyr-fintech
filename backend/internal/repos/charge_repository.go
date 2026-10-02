package repos

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChargeRepository interface {
	GetIdempotency(ctx context.Context, merchantID, key string) (*models.ChargeIdempotencyRecord, error)
	SaveIdempotency(ctx context.Context, record *models.ChargeIdempotencyRecord) error
	SaveFailedCharge(ctx context.Context, charge *models.Charge) error
	ProcessSettlement(ctx context.Context, charge *models.Charge, entries []*models.LedgerEntry, settlementWalletID string) error
	GetSystemAccountID(ctx context.Context, accountType models.AccountType) (string, error)
}

type chargeRepository struct {
	db *pgxpool.Pool
}

func NewChargeRepository(db *pgxpool.Pool) ChargeRepository {
	return &chargeRepository{db: db}
}

// GetIdempotency finds cached response by merchant ID and idempotency key.
func (r *chargeRepository) GetIdempotency(ctx context.Context, merchantID, key string) (*models.ChargeIdempotencyRecord, error) {
	query := `
		SELECT idempotency_key, merchant_id, request_hash, response_status, response_body, created_at
		FROM charge_idempotency
		WHERE merchant_id = $1 AND idempotency_key = $2
		LIMIT 1;
	`

	var record models.ChargeIdempotencyRecord
	err := r.db.QueryRow(ctx, query, merchantID, key).Scan(
		&record.IdempotencyKey,
		&record.MerchantID,
		&record.RequestHash,
		&record.ResponseStatus,
		&record.ResponseBody,
		&record.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Cache miss is not an error
		}
		return nil, err
	}

	return &record, nil
}

// SaveIdempotency stores or updates the cached HTTP response for a request.
func (r *chargeRepository) SaveIdempotency(ctx context.Context, record *models.ChargeIdempotencyRecord) error {
	query := `
		INSERT INTO charge_idempotency (idempotency_key, merchant_id, request_hash, response_status, response_body)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (merchant_id, idempotency_key) 
		DO UPDATE SET response_status = EXCLUDED.response_status, response_body = EXCLUDED.response_body;
	`
	_, err := r.db.Exec(ctx, query, record.IdempotencyKey, record.MerchantID, record.RequestHash, record.ResponseStatus, record.ResponseBody)
	return err
}

// SaveFailedCharge records an unfulfilled payment attempt for audit logs without modifying ledger balances.
func (r *chargeRepository) SaveFailedCharge(ctx context.Context, charge *models.Charge) error {
	query := `
		INSERT INTO charges (id, merchant_id, amount, currency, fee_amount, net_amount, status, failure_reason, card_last4, card_holder_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
	`
	_, err := r.db.Exec(ctx, query,
		charge.ID,
		charge.MerchantID,
		charge.Amount,
		charge.Currency,
		charge.FeeAmount,
		charge.NetAmount,
		models.ChargeStatusFailed,
		charge.FailureReason,
		charge.CardLast4,
		charge.CardHolderName,
	)
	return err
}

// GetSystemAccountID retrieves the UUID of a system internal clearing/revenue account.
func (r *chargeRepository) GetSystemAccountID(ctx context.Context, accountType models.AccountType) (string, error) {
	var id string
	query := `SELECT id FROM system_accounts WHERE account_type = $1 LIMIT 1;`
	err := r.db.QueryRow(ctx, query, string(accountType)).Scan(&id)
	return id, err
}

// ProcessSettlement executes the charge creation, double-entry ledger entries, and wallet balance updates atomically.
func (r *chargeRepository) ProcessSettlement(ctx context.Context, charge *models.Charge, entries []*models.LedgerEntry, settlementWalletID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Insert Succeeded Charge
	chargeQuery := `
		INSERT INTO charges (id, merchant_id, amount, currency, fee_amount, net_amount, status, failure_reason, card_last4, card_holder_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
	`
	_, err = tx.Exec(ctx, chargeQuery,
		charge.ID,
		charge.MerchantID,
		charge.Amount,
		charge.Currency,
		charge.FeeAmount,
		charge.NetAmount,
		models.ChargeStatusSucceeded,
		"",
		charge.CardLast4,
		charge.CardHolderName,
	)
	if err != nil {
		return err
	}

	// 2. Insert Balanced Double-Entry Ledger Records
	ledgerQuery := `
		INSERT INTO ledger_entries (id, charge_id, account_type, account_id, entry_type, amount, currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`
	for _, entry := range entries {
		if entry.ID == "" {
			entry.ID = uuid.New().String()
		}
		_, err = tx.Exec(ctx, ledgerQuery,
			entry.ID,
			charge.ID,
			string(entry.AccountType),
			entry.AccountID,
			string(entry.EntryType),
			entry.Amount,
			entry.Currency,
		)
		if err != nil {
			return err
		}
	}

	// 3. Update Balances Atomically
	// 3a. Credit Merchant's Settlement Wallet with Net Amount
	walletUpdate := `UPDATE wallets SET balance = balance + $1 WHERE id = $2;`
	resWallet, err := tx.Exec(ctx, walletUpdate, charge.NetAmount, settlementWalletID)
	if err != nil {
		return err
	}
	if resWallet.RowsAffected() == 0 {
		return errors.New("merchant settlement wallet not found")
	}

	// 3b. Update System Accounts
	// Acquiring clearing increases by gross amount
	_, err = tx.Exec(ctx, `UPDATE system_accounts SET balance = balance + $1 WHERE account_type = 'ACQUIRING_CLEARING';`, charge.Amount)
	if err != nil {
		return err
	}

	// Platform fee revenue increases by fee amount
	if charge.FeeAmount > 0 {
		_, err = tx.Exec(ctx, `UPDATE system_accounts SET balance = balance + $1 WHERE account_type = 'PLATFORM_FEE_REVENUE';`, charge.FeeAmount)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
