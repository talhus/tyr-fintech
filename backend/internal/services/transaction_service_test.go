package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/internal/services"
	"github.com/iamtbay/tyr-fintech/internal/worker"
)

type mockTransactionRepository struct {
	transferFunc                  func(ctx context.Context, tx *dto.TransferRequest, convertedAmount int64) error
	getTransactionsByWalletIDFunc func(ctx context.Context, walletID string) ([]*models.Transaction, error)
}

func (m *mockTransactionRepository) Transfer(ctx context.Context, tx *dto.TransferRequest, convertedAmount int64) error {
	return m.transferFunc(ctx, tx, convertedAmount)
}

func (m *mockTransactionRepository) GetTransactionsByWalletID(ctx context.Context, walletID string) ([]*models.Transaction, error) {
	return m.getTransactionsByWalletIDFunc(ctx, walletID)
}

type customMockExchangeService struct {
	getRateFunc func(ctx context.Context, from, to models.WalletCurrency) (float64, error)
}

func (e *customMockExchangeService) GetRate(ctx context.Context, from, to models.WalletCurrency) (float64, error) {
	return e.getRateFunc(ctx, from, to)
}

func TestTransactionService_Transfer(t *testing.T) {
	tests := []struct {
		name                    string
		inputFromWalletNum      int64
		inputToWalletNum        int64
		inputAmount             int64
		inputTransactionID      string
		fromWalletCurrency      models.WalletCurrency
		toWalletCurrency        models.WalletCurrency
		exchangeRate            float64
		exchangeRateErr         error
		walletErr               error
		mockCreateErr           error
		expectedConvertedAmount int64
		wantErr                 bool
	}{
		{
			name:                    "success cross-currency TRY to USD",
			inputFromWalletNum:      1000000001,
			inputToWalletNum:        1000000002,
			inputTransactionID:      "tx-1",
			inputAmount:             100,
			fromWalletCurrency:      models.CurrencyTRY,
			toWalletCurrency:        models.CurrencyUSD,
			exchangeRate:            0.021,
			expectedConvertedAmount: 2,
			wantErr:                 false,
		},
		{
			name:                    "success same currency TRY to TRY",
			inputFromWalletNum:      1000000001,
			inputToWalletNum:        1000000003,
			inputTransactionID:      "tx-2",
			inputAmount:             100,
			fromWalletCurrency:      models.CurrencyTRY,
			toWalletCurrency:        models.CurrencyTRY,
			exchangeRate:            1.0,
			expectedConvertedAmount: 100,
			wantErr:                 false,
		},
		{
			name:               "empty idempotency key",
			inputFromWalletNum: 1000000001,
			inputToWalletNum:   1000000002,
			inputTransactionID: "",
			inputAmount:        100,
			wantErr:            true,
		},
		{
			name:               "wallet fetch error",
			inputFromWalletNum: 1000000001,
			inputToWalletNum:   1000000002,
			inputTransactionID: "tx-3",
			inputAmount:        100,
			walletErr:          errors.New("wallet not found"),
			wantErr:            true,
		},
		{
			name:               "exchange rate service error",
			inputFromWalletNum: 1000000001,
			inputToWalletNum:   1000000002,
			inputTransactionID: "tx-4",
			inputAmount:        100,
			fromWalletCurrency: models.CurrencyTRY,
			toWalletCurrency:   models.CurrencyUSD,
			exchangeRateErr:    errors.New("network timeout"),
			wantErr:            true,
		},
		{
			name:                    "database transaction error",
			inputFromWalletNum:      1000000001,
			inputToWalletNum:        1000000002,
			inputTransactionID:      "tx-5",
			inputAmount:             100,
			fromWalletCurrency:      models.CurrencyTRY,
			toWalletCurrency:        models.CurrencyUSD,
			exchangeRate:            0.021,
			mockCreateErr:           errors.New("database lock failure"),
			expectedConvertedAmount: 2,
			wantErr:                 true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockTransactionRepository{
				transferFunc: func(ctx context.Context, tx *dto.TransferRequest, convertedAmount int64) error {
					if tt.expectedConvertedAmount != 0 && convertedAmount != tt.expectedConvertedAmount {
						t.Errorf("expected converted amount %d, got %d", tt.expectedConvertedAmount, convertedAmount)
					}
					return tt.mockCreateErr
				},
			}

			walletRepo := &mockWalletRepository{
				funcGetWalletByID: func(ctx context.Context, walletID int64) (*models.WalletResponse, error) {
					if tt.walletErr != nil {
						return nil, tt.walletErr
					}
					currency := tt.fromWalletCurrency
					if walletID == tt.inputToWalletNum {
						currency = tt.toWalletCurrency
					}
					return &models.WalletResponse{
						Currency:     currency,
						Balance:      10000,
					}, nil
				},
			}

			exchangeService := &customMockExchangeService{
				getRateFunc: func(ctx context.Context, from, to models.WalletCurrency) (float64, error) {
					if tt.exchangeRateErr != nil {
						return 0, tt.exchangeRateErr
					}
					return tt.exchangeRate, nil
				},
			}

			service := services.NewTransactionService(mockRepo, exchangeService, walletRepo, nil)
			err := service.Transfer(context.Background(), &dto.TransferRequest{
				TransactionID:    tt.inputTransactionID,
				FromWalletNumber: tt.inputFromWalletNum,
				ToWalletNumber:   tt.inputToWalletNum,
				Amount:           tt.inputAmount,
			})

			if !tt.wantErr && err == nil && tt.name != "empty idempotency key" {
				select {
				case event := <-worker.WebHookQueue:
					if event.TransactionID != tt.inputTransactionID {
						t.Errorf("Expected transaction id %s, but got %s", tt.inputTransactionID, event.TransactionID)
					}
					if event.Amount != tt.inputAmount {
						t.Errorf("Expected amount %d, but got %d", tt.inputAmount, event.Amount)
					}
				default:
					t.Errorf("After successfully process, no webhook event sent.")
				}
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Transfer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransactionService_GetHistory(t *testing.T) {
}
