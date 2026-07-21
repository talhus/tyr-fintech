package services_test

import (
	"context"
	"testing"

	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/internal/services"
)

type mockCardRepository struct {
	createFunc              func(ctx context.Context, card *models.Card) error
	getByUserIDFunc         func(ctx context.Context, userID string) ([]models.Card, error)
	getCardDetailsFunc      func(ctx context.Context, cardID, userID string) (*models.Card, error)
	updateStatusFunc        func(ctx context.Context, cardID, userID string, status models.CardStatus) error
	getCardTransactionsFunc func(ctx context.Context, cardID, userID string) ([]models.Transaction, error)
	processPaymentFunc      func(ctx context.Context, transactionID, cardID, cvv string, expiryMonth, expiryYear int, amount int64, merchantName string) (*models.CardPaymentResult, error)
}

func (m *mockCardRepository) Create(ctx context.Context, card *models.Card) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, card)
	}
	return nil
}

func (m *mockCardRepository) GetByUserID(ctx context.Context, userID string) ([]models.Card, error) {
	if m.getByUserIDFunc != nil {
		return m.getByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockCardRepository) GetCardDetails(ctx context.Context, cardID, userID string) (*models.Card, error) {
	if m.getCardDetailsFunc != nil {
		return m.getCardDetailsFunc(ctx, cardID, userID)
	}
	return nil, nil
}

func (m *mockCardRepository) UpdateStatus(ctx context.Context, cardID, userID string, status models.CardStatus) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, cardID, userID, status)
	}
	return nil
}

func (m *mockCardRepository) GetCardTransactions(ctx context.Context, cardID, userID string) ([]models.Transaction, error) {
	if m.getCardTransactionsFunc != nil {
		return m.getCardTransactionsFunc(ctx, cardID, userID)
	}
	return nil, nil
}

func (m *mockCardRepository) ProcessPayment(ctx context.Context, transactionID, cardID, cvv string, expiryMonth, expiryYear int, amount int64, merchantName string) (*models.CardPaymentResult, error) {
	if m.processPaymentFunc != nil {
		return m.processPaymentFunc(ctx, transactionID, cardID, cvv, expiryMonth, expiryYear, amount, merchantName)
	}
	return &models.CardPaymentResult{
		TransactionID: transactionID,
	}, nil
}

// tests
func TestCardService_CreateCard(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		walletID    string
		limitAmount int64
		mockCardErr error
		wantErr     bool
	}{
		{
			name:        "Success-Valid Wallet Ownership",
			userID:      "user-123",
			walletID:    "wallet-abc",
			limitAmount: 50000,
			mockCardErr: nil,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cardRepo := &mockCardRepository{
				createFunc: func(ctx context.Context, card *models.Card) error {
					return tt.mockCardErr
				},
			}
			service := services.NewCardService(cardRepo, nil)

			card, err := service.CreateCard(context.Background(), tt.userID, tt.walletID, tt.limitAmount)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if card == nil {
					t.Error("CreateCard() expected card to be non-nil on success")
				} else {
					if card.UserID != tt.userID {
						t.Errorf("CreateCard() UserID = %v, want %v", card.UserID, tt.userID)
					}
					if card.WalletID != tt.walletID {
						t.Errorf("CreateCard() WalletID = %v, want %v", card.WalletID, tt.walletID)
					}
					if card.LimitAmount != tt.limitAmount {
						t.Errorf("CreateCard() LimitAmount = %v, want %v", card.LimitAmount, tt.limitAmount)
					}
				}
			}
		})
	}
}
