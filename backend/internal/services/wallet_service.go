package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
)

type WalletRepository interface {
	Create(ctx context.Context, wallet *models.Wallet) error
	GetByUserID(ctx context.Context, userID string) ([]*models.Wallet, error)
	GetWalletByID(ctx context.Context, walletID int64) (*models.WalletResponse, error)
	Delete(ctx context.Context, userID, walletID string) error
	VerifyWallet(ctx context.Context, walletID int64) (*dto.WalletLookUpResult, error)
}

type WalletService struct {
	walletRepo WalletRepository
}

func NewWalletService(walletRepo WalletRepository) *WalletService {
	return &WalletService{walletRepo: walletRepo}
}

// METHODS
func (s *WalletService) GetByUserID(ctx context.Context, userID string) ([]*models.Wallet, error) {
	return s.walletRepo.GetByUserID(ctx, userID)
}

// CREATE WALLET
func (s *WalletService) CreateWallet(ctx context.Context, req *dto.CreateWallet) error {
	if req.Currency != "TRY" && req.Currency != "USD" && req.Currency != "EUR" {
		return errors.New("Unsupported currency type. Supported currencies are TRY, USD, and EUR.")
	}
	err := s.walletRepo.Create(ctx, &models.Wallet{
		ID:       uuid.New().String(),
		UserID:   req.UserID,
		Currency: req.Currency,
	})
	if err != nil {
		return err
	}
	return nil
}

// DELETE WALLET
func (s *WalletService) DeleteWallet(ctx context.Context, userID string, walletID string) error {
	return s.walletRepo.Delete(ctx, userID, walletID)
}

func (s *WalletService) VerifyWallet(ctx context.Context, walletID int64) (*dto.WalletLookUpResult, error) {
	return s.walletRepo.VerifyWallet(ctx, walletID)
}
