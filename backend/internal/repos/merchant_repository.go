package repos

import (
	"context"
	"errors"

	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MerchantRepository interface {
	GetMerchantByAPIKeyHash(ctx context.Context, keyHash string) (*models.Merchant, error)
	GetMerchantByID(ctx context.Context, id string) (*models.Merchant, error)
}

type merchantRepository struct {
	db *pgxpool.Pool
}

func NewMerchantRepository(db *pgxpool.Pool) MerchantRepository {
	return &merchantRepository{db: db}
}

// /methods

// GET MERCHANT API KEY
func (r *merchantRepository) GetMerchantByAPIKeyHash(ctx context.Context, keyHash string) (*models.Merchant, error) {
	query := `SELECT m.id, m.name, m.email,m.settlement_wallet_id,m.fee_percentage_basis_points, m.status,m.created_at FROM merchants m JOIN merchant_api_keys k ON m.id=k.merchant_id WHERE k.key_hash=$1 AND k.status='ACTIVE' AND m.status='ACTIVE' LIMIT 1`

	var merchant models.Merchant
	err := r.db.QueryRow(ctx, query, keyHash).Scan(&merchant.ID, &merchant.Name, &merchant.Email, &merchant.SettlementWalletID, &merchant.FeePercentageBasisPoints, &merchant.Status, &merchant.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("merchant not found or inactive")
		}
		return nil, err
	}

	return &merchant, nil
}

// GET MERCHANT BY ID
func (r *merchantRepository) GetMerchantByID(ctx context.Context, id string) (*models.Merchant, error) {
	query := `SELECT id,name,email,settlement_wallet_id,fee_percentage_basis_points,status,created_at FROM merchants WHERE id=$1`

	var merchant models.Merchant

	err := r.db.QueryRow(ctx, query, id).Scan(&merchant.ID, &merchant.Name, &merchant.Email, &merchant.SettlementWalletID, &merchant.FeePercentageBasisPoints, &merchant.Status, &merchant.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("merchant not found")
		}
		return nil, err
	}

	return &merchant, nil
}
