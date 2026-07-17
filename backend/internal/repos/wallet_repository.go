package repos

import (
	"context"
	"errors"
	"net/http"

	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/pkg/apperrors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepository struct {
	pool *pgxpool.Pool
}

func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{pool: pool}
}

// CREATE
func (r *WalletRepository) Create(ctx context.Context, wallet *models.Wallet) error {
	query := `INSERT INTO wallets (id,user_id,currency) VALUES ($1,$2,$3)`

	_, err := r.pool.Exec(ctx, query, wallet.ID, wallet.UserID, wallet.Currency)
	if err != nil {
		return err
	}
	return nil
}

// GET WALLET BY ID

func (r *WalletRepository) GetWalletByID(ctx context.Context, walletID int64) (*models.WalletResponse, error) {
	query := `SELECT balance,currency FROM wallets WHERE wallet_number = $1 AND deleted_at IS NULL`
	row := r.pool.QueryRow(ctx, query, walletID)
	var walletResponse models.WalletResponse
	err := row.Scan(&walletResponse.Balance, &walletResponse.Currency)
	if err != nil {
		return nil, err
	}
	return &walletResponse, nil
}

// GET WALLETS
func (r *WalletRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Wallet, error) {
	query := `SELECT id, user_id, currency, balance, created_at, COALESCE(wallet_number, 0) AS wallet_number FROM wallets WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []*models.Wallet
	for rows.Next() {
		w := &models.Wallet{}
		err := rows.Scan(&w.ID, &w.UserID, &w.Currency, &w.Balance, &w.CreatedAt, &w.WalletNumber)
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, w)
	}
	return wallets, nil
}

// DELETE WALLET
func (r *WalletRepository) Delete(ctx context.Context, userID, walletID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var balance float64
	err = tx.QueryRow(ctx, `SELECT balance FROM wallets WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL FOR UPDATE`, walletID, userID).Scan(&balance)
	if err != nil {
		return err
	}

	if balance > 0 {
		return errors.New("Cannot delete a wallet with a non-zero balance. Please empty or transfer funds first.")
	}

	_, err = tx.Exec(ctx, `UPDATE wallets SET deleted_at = NOW() WHERE user_id = $1 AND id = $2`, userID, walletID)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

// VERIFY WALLET
func (r *WalletRepository) VerifyWallet(ctx context.Context, walletID int64) (*dto.WalletLookUpResult, error) {
	query := `
	SELECT u.name, w.currency
	FROM wallets w
	JOIN users u ON w.user_id=u.id
	WHERE w.wallet_number = $1 AND w.deleted_at IS NULL
	`
	var walletResult dto.WalletLookUpResult
	err := r.pool.QueryRow(ctx, query, walletID).Scan(&walletResult.OwnerName, &walletResult.Currency)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.New(http.StatusNotFound, "Wallet Not Found")
		}
		return nil, apperrors.New(http.StatusInternalServerError, err.Error())
	}
	return &walletResult, nil
}
