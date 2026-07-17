package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/internal/services"
)

type mockWalletRepository struct {
	funcCreate        func(ctx context.Context, wallet *models.Wallet) error
	funcGetByUserID   func(ctx context.Context, userID string) ([]*models.Wallet, error)
	funcGetWalletByID func(ctx context.Context, walletID int64) (*models.WalletResponse, error)
	funcDelete        func(ctx context.Context, userID, walletID string) error
	funcVerifyWallet  func(ctx context.Context, walletID int64) (*dto.WalletLookUpResult, error)
}

func (m *mockWalletRepository) Create(ctx context.Context, wallet *models.Wallet) error {
	if m.funcCreate != nil {
		return m.funcCreate(ctx, wallet)
	}
	return nil
}
func (m *mockWalletRepository) GetWalletByID(ctx context.Context, walletID int64) (*models.WalletResponse, error) {
	if m.funcGetWalletByID != nil {
		return m.funcGetWalletByID(ctx, walletID)
	}
	return &models.WalletResponse{Balance: 1000, Currency: models.CurrencyUSD}, nil
}
func (m *mockWalletRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Wallet, error) {
	if m.funcGetByUserID != nil {
		return m.funcGetByUserID(ctx, userID)
	}
	return nil, nil
}
func (m *mockWalletRepository) Delete(ctx context.Context, userID, walletID string) error {
	if m.funcDelete != nil {
		return m.funcDelete(ctx, userID, walletID)
	}
	return nil
}
func (m *mockWalletRepository) VerifyWallet(ctx context.Context, walletID int64) (*dto.WalletLookUpResult, error) {
	if m.funcVerifyWallet != nil {
		return m.funcVerifyWallet(ctx, walletID)
	}
	return nil, nil
}

// TESTS
func TestWalletService_DeleteWallet(t *testing.T) {
	tests := []struct {
		name          string
		inputUserID   string
		inputWalletID string
		mockBalance   int64
		mockCreateErr error
		wantErr       bool
	}{
		{
			name:          "success",
			inputUserID:   "1",
			inputWalletID: "1",
			mockBalance:   0,
			mockCreateErr: nil,
			wantErr:       false,
		},
		{
			name:          "wallet not found",
			inputUserID:   "1",
			inputWalletID: "1",
			mockCreateErr: errors.New("wallet not found"),
			wantErr:       true,
		},
		{
			name:          "cannot delete with balance",
			inputUserID:   "1",
			inputWalletID: "1",
			mockBalance:   10,
			mockCreateErr: errors.New("Cannot delete a wallet with a non-zero balance. Please empty or transfer funds first."),
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockWalletRepository{
				funcDelete: func(ctx context.Context, userID, walletID string) error {
					return tt.mockCreateErr
				},
			}
			service := services.NewWalletService(mockRepo)
			err := service.DeleteWallet(context.Background(), tt.inputUserID, tt.inputWalletID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteWallet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}
