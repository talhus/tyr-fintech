package repos

import (
	"context"

	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (id,name,email,password_hash) VALUES ($1,$2,$3,$4)`
	_, err := r.db.Exec(ctx, query, user.ID, user.Name, user.Email, user.PasswordHash)
	if err != nil {
		return err
	}
	return nil
}

// GET BY ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `SELECT id,name,email,password_hash,created_at FROM users WHERE id = $1`
	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GET BY EMAIL
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id,name,email,password_hash,created_at FROM users WHERE email = $1`
	user := &models.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}
