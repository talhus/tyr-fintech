package repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

// Transfer
func (r *TransactionRepository) Transfer(ctx context.Context, req *dto.TransferRequest, convertedAmount int64) error {
	if req.FromWalletNumber == req.ToWalletNumber {
		return errors.New("cannot transfer to the same wallet")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	resIdemp, err := tx.Exec(ctx, `INSERT INTO idempotency_keys(key) VALUES($1) ON CONFLICT (key) DO NOTHING`, req.TransactionID)
	if err != nil {
		return err
	}
	if resIdemp.RowsAffected() == 0 {
		return errors.New("This transaction has already been processed.")
	}

	firstNumber := req.FromWalletNumber
	secondNumber := req.ToWalletNumber
	if req.FromWalletNumber > req.ToWalletNumber {
		firstNumber = req.ToWalletNumber
		secondNumber = req.FromWalletNumber
	}

	var firstID, firstCurrency string
	err = tx.QueryRow(ctx, `SELECT id, currency FROM wallets WHERE wallet_number=$1 AND deleted_at IS NULL FOR UPDATE`, firstNumber).Scan(&firstID, &firstCurrency)
	if err != nil {
		return errors.New("Source or destination wallet could not be found.")
	}

	var secondID, secondCurrency string
	err = tx.QueryRow(ctx, `SELECT id, currency FROM wallets WHERE wallet_number=$1 AND deleted_at IS NULL FOR UPDATE`, secondNumber).Scan(&secondID, &secondCurrency)
	if err != nil {
		return errors.New("Source or destination wallet could not be found.")
	}

	var fromID, toID string
	if firstNumber == req.FromWalletNumber {
		fromID = firstID
		toID = secondID
	} else {
		fromID = secondID
		toID = firstID
	}

	res, err := tx.Exec(ctx, `UPDATE wallets SET balance=balance-$1 WHERE id=$2 AND balance>=$1`, req.Amount, fromID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("Insufficient funds in the source wallet to complete this transfer.")
	}

	_, err = tx.Exec(ctx, `UPDATE wallets SET balance=balance+$1 WHERE id=$2`, convertedAmount, toID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `INSERT INTO transactions(id,from_wallet_id,to_wallet_id,amount,converted_amount,status) VALUES($1,$2,$3,$4,$5,$6)`, req.TransactionID, fromID, toID, req.Amount, convertedAmount, models.StatusCompleted)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetTransactionsByWalletID
// GetTransactions handles both single wallet and all wallets filtering with optional date range bounds
func (r *TransactionRepository) GetTransactions(ctx context.Context, walletID string, userID string, startDate string, endDate string) ([]*models.Transaction, error) {
	var query string
	var args []interface{}

	if walletID == "all" || walletID == "" {
		query = `
			SELECT 
			t.id,
			COALESCE(t.from_wallet_id, '00000000-0000-0000-0000-000000000000'),
			COALESCE(t.to_wallet_id, '00000000-0000-0000-0000-000000000000'),
			COALESCE(w_from.wallet_number, 0) as from_wallet_number,
			COALESCE(w_to.wallet_number, 0) as to_wallet_number,
			t.amount,
			COALESCE(t.converted_amount, t.amount) as converted_amount,
			t.status,
			t.created_at,
			COALESCE(t.card_id, '00000000-0000-0000-0000-000000000000') as card_id,
			COALESCE(t.merchant_name, '') as merchant_name
			FROM transactions t
			LEFT JOIN wallets w_from ON t.from_wallet_id = w_from.id
			LEFT JOIN wallets w_to ON t.to_wallet_id = w_to.id
			WHERE (w_from.user_id = $1 OR w_to.user_id = $1) AND t.status = 'COMPLETED'
		`
		args = append(args, userID)
	} else {
		query = `
			SELECT 
			t.id,
			COALESCE(t.from_wallet_id, '00000000-0000-0000-0000-000000000000'),
			COALESCE(t.to_wallet_id, '00000000-0000-0000-0000-000000000000'),
			COALESCE(w_from.wallet_number, 0) as from_wallet_number,
			COALESCE(w_to.wallet_number, 0) as to_wallet_number,
			t.amount,
			COALESCE(t.converted_amount, t.amount) as converted_amount,
			t.status,
			t.created_at,
			COALESCE(t.card_id, '00000000-0000-0000-0000-000000000000') as card_id,
			COALESCE(t.merchant_name, '') as merchant_name
			FROM transactions t
			LEFT JOIN wallets w_from ON t.from_wallet_id = w_from.id
			LEFT JOIN wallets w_to ON t.to_wallet_id = w_to.id
			WHERE (t.from_wallet_id = $1 OR t.to_wallet_id = $1) AND t.status = 'COMPLETED'
		`
		args = append(args, walletID)
	}

	if startDate != "" {
		args = append(args, startDate)
		query += fmt.Sprintf(" AND t.created_at >= $%d::timestamp", len(args))
	}
	if endDate != "" {
		args = append(args, endDate)
		query += fmt.Sprintf(" AND t.created_at <= ($%d::date + INTERVAL '1 day')", len(args))
	}

	query += " ORDER BY t.created_at DESC;"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []*models.Transaction{}
	for rows.Next() {
		var t models.Transaction
		var cardIDVal, merchantVal string
		err := rows.Scan(&t.ID, &t.FromWalletID, &t.ToWalletID, &t.FromWalletNumber, &t.ToWalletNumber, &t.Amount, &t.ConvertedAmount, &t.Status, &t.CreatedAt, &cardIDVal, &merchantVal)
		if err != nil {
			return nil, err
		}
		if cardIDVal != "00000000-0000-0000-0000-000000000000" && cardIDVal != "" {
			t.CardID = &cardIDVal
		}
		if merchantVal != "" {
			t.MerchantName = &merchantVal
		}
		transactions = append(transactions, &t)
	}

	return transactions, nil
}

func (r *TransactionRepository) GetTransactionsByWalletID(ctx context.Context, walletID string) ([]*models.Transaction, error) {
	return r.GetTransactions(ctx, walletID, "", "", "")
}
