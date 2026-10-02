package repos

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CheckoutRepository interface {
	CreateSession(ctx context.Context, session *models.CheckoutSession) error
	GetSessionByID(ctx context.Context, sessionID string) (*models.CheckoutSession, error)
	GetSessionDetails(ctx context.Context, sessionID string) (*dto.CheckoutSessionDetailsResponse, error)
	ExecuteWalletPayment(ctx context.Context, sessionID, userID, walletID string, merchant *models.Merchant) (string, error)
}

type checkoutRepository struct {
	db *pgxpool.Pool
}

func NewCheckoutRepository(db *pgxpool.Pool) CheckoutRepository {
	return &checkoutRepository{db: db}
}

func (r *checkoutRepository) CreateSession(ctx context.Context, session *models.CheckoutSession) error {
	query := `
		INSERT INTO checkout_sessions (id, merchant_id, order_id, amount, currency, callback_url, webhook_url, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`
	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.MerchantID,
		session.OrderID,
		session.Amount,
		session.Currency,
		session.CallbackURL,
		session.WebHookURL,
		session.Status,
		session.ExpiresAt,
	)
	return err
}

func (r *checkoutRepository) GetSessionByID(ctx context.Context, sessionID string) (*models.CheckoutSession, error) {
	query := `
		SELECT id, merchant_id, order_id, amount, currency, callback_url, webhook_url, status, payer_user_id, transaction_id, expires_at, created_at, updated_at
		FROM checkout_sessions
		WHERE id = $1;
	`
	var s models.CheckoutSession
	err := r.db.QueryRow(ctx, query, sessionID).Scan(
		&s.ID,
		&s.MerchantID,
		&s.OrderID,
		&s.Amount,
		&s.Currency,
		&s.CallbackURL,
		&s.WebHookURL,
		&s.Status,
		&s.PayerUserID,
		&s.TransactionID,
		&s.ExpiresAt,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("checkout session not found")
		}
		return nil, err
	}
	return &s, nil
}

func (r *checkoutRepository) GetSessionDetails(ctx context.Context, sessionID string) (*dto.CheckoutSessionDetailsResponse, error) {
	query := `
		SELECT cs.id, m.name, cs.order_id, cs.amount, cs.currency, cs.status, cs.expires_at
		FROM checkout_sessions cs
		JOIN merchants m ON cs.merchant_id = m.id
		WHERE cs.id = $1;
	`
	var d dto.CheckoutSessionDetailsResponse
	err := r.db.QueryRow(ctx, query, sessionID).Scan(
		&d.SessionID,
		&d.MerchantName,
		&d.OrderID,
		&d.Amount,
		&d.Currency,
		&d.Status,
		&d.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("checkout session not found")
		}
		return nil, err
	}
	return &d, nil
}

func (r *checkoutRepository) ExecuteWalletPayment(ctx context.Context, sessionID, userID, walletID string, merchant *models.Merchant) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	// 1. Lock and validate checkout session
	var session models.CheckoutSession
	sessionQuery := `
		SELECT id, amount, currency, status, expires_at
		FROM checkout_sessions
		WHERE id = $1 FOR UPDATE;
	`
	err = tx.QueryRow(ctx, sessionQuery, sessionID).Scan(
		&session.ID,
		&session.Amount,
		&session.Currency,
		&session.Status,
		&session.ExpiresAt,
	)
	if err != nil {
		return "", errors.New("checkout session not found")
	}

	if session.Status != models.CheckoutStatusPending {
		return "", fmt.Errorf("checkout session is not payable (current status: %s)", session.Status)
	}

	if time.Now().After(session.ExpiresAt) {
		_, _ = tx.Exec(ctx, `UPDATE checkout_sessions SET status = 'EXPIRED' WHERE id = $1;`, sessionID)
		return "", errors.New("checkout session has expired")
	}

	// 2. Lock and validate user wallet
	var walletBalance int64
	var walletCurrency string
	walletQuery := `SELECT balance, currency FROM wallets WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL FOR UPDATE;`
	err = tx.QueryRow(ctx, walletQuery, walletID, userID).Scan(&walletBalance, &walletCurrency)
	if err != nil {
		return "", errors.New("user wallet not found or does not belong to user")
	}

	if walletCurrency != session.Currency {
		return "", fmt.Errorf("currency mismatch: wallet is %s, required %s", walletCurrency, session.Currency)
	}

	if walletBalance < session.Amount {
		return "", errors.New("insufficient wallet balance")
	}

	// 3. Financial calculations in minor units (int64)
	feeBasisPoints := int64(merchant.FeePercentageBasisPoints)
	feeAmount := (session.Amount * feeBasisPoints) / 10000
	netAmount := session.Amount - feeAmount
	txUUID := uuid.New()
	transactionID := txUUID.String()

	// 4. Debit User's Wallet
	_, err = tx.Exec(ctx, `UPDATE wallets SET balance = balance - $1 WHERE id = $2;`, session.Amount, walletID)
	if err != nil {
		return "", err
	}

	// 5. Credit Merchant's Settlement Wallet with netAmount
	_, err = tx.Exec(ctx, `UPDATE wallets SET balance = balance + $1 WHERE id = $2;`, netAmount, merchant.SettlementWalletID)
	if err != nil {
		return "", err
	}

	// 6. Credit Platform Fee Account
	if feeAmount > 0 {
		_, err = tx.Exec(ctx, `UPDATE system_accounts SET balance = balance + $1 WHERE account_type = 'PLATFORM_FEE_REVENUE';`, feeAmount)
		if err != nil {
			return "", err
		}
	}

	// 7. Insert Into Transactions (for user activity log)
	txLogQuery := `
		INSERT INTO transactions (id, from_wallet_id, to_wallet_id, amount, status, merchant_name)
		VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err = tx.Exec(ctx, txLogQuery, txUUID, walletID, merchant.SettlementWalletID, session.Amount, models.StatusCompleted, merchant.Name)
	if err != nil {
		return "", err
	}

	// 8. Update Checkout Session status to PAID
	updateSessionQuery := `
		UPDATE checkout_sessions
		SET status = 'PAID', payer_user_id = $1, transaction_id = $2, updated_at = NOW()
		WHERE id = $3;
	`
	_, err = tx.Exec(ctx, updateSessionQuery, userID, transactionID, sessionID)
	if err != nil {
		return "", err
	}

	// 10. Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return transactionID, nil
}
