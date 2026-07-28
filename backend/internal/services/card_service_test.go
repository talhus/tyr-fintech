package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/internal/notifications"
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
		UserID:        "user-123",
		UserEmail:     "user@test.com",
		UserName:      "Test User",
		MerchantName:  merchantName,
		Amount:        amount,
	}, nil
}

type mockNotificationService struct {
	notifyUserFunc func(event *notifications.NotificationEvent)
}

func (m *mockNotificationService) NotifyUser(event *notifications.NotificationEvent) {
	if m.notifyUserFunc != nil {
		m.notifyUserFunc(event)
	}
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
		{
			name:        "Failed Create Card",
			userID:      "user-123",
			walletID:    "wallet-abc",
			limitAmount: 50000,
			mockCardErr: errors.New("db error"),
			wantErr:     true,
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

func TestCardService_GetCardsByUserID(t *testing.T) {
	mockRepo := &mockCardRepository{
		getByUserIDFunc: func(ctx context.Context, userID string) ([]models.Card, error) {
			if userID == "user-123" {
				return []models.Card{{ID: "card-1", UserID: userID}}, nil
			}
			return nil, errors.New("user not found")
		},
	}

	service := services.NewCardService(mockRepo, nil)

	cards, err := service.GetCardsByUserID(context.Background(), "user-123")
	if err != nil {
		t.Errorf("GetCardsByUserID() unexpected error = %v", err)
	}
	if len(cards) != 1 {
		t.Errorf("GetCardsByUserID() got %d cards, want 1", len(cards))
	}
}

func TestCardService_GetCardDetails(t *testing.T) {
	mockRepo := &mockCardRepository{
		getCardDetailsFunc: func(ctx context.Context, cardID, userID string) (*models.Card, error) {
			if cardID == "card-1" && userID == "user-123" {
				return &models.Card{ID: cardID, UserID: userID, CardNumber: "1234567812345678"}, nil
			}
			return nil, errors.New("card not found")
		},
	}

	service := services.NewCardService(mockRepo, nil)

	card, err := service.GetCardDetails(context.Background(), "card-1", "user-123")
	if err != nil {
		t.Errorf("GetCardDetails() unexpected error = %v", err)
	}
	if card == nil || card.CardNumber != "1234567812345678" {
		t.Errorf("GetCardDetails() invalid card details returned")
	}
}

func TestCardService_UpdateCardStatus(t *testing.T) {
	mockRepo := &mockCardRepository{
		updateStatusFunc: func(ctx context.Context, cardID, userID string, status models.CardStatus) error {
			if status == models.CardStatusFrozen {
				return nil
			}
			return errors.New("update status error")
		},
	}

	service := services.NewCardService(mockRepo, nil)

	err := service.UpdateCardStatus(context.Background(), "card-1", "user-123", models.CardStatusFrozen)
	if err != nil {
		t.Errorf("UpdateCardStatus() unexpected error = %v", err)
	}
}

func TestCardService_ProcessPayment(t *testing.T) {
	var notifiedEvent *notifications.NotificationEvent

	mockNotif := &mockNotificationService{
		notifyUserFunc: func(event *notifications.NotificationEvent) {
			notifiedEvent = event
		},
	}

	mockRepo := &mockCardRepository{
		processPaymentFunc: func(ctx context.Context, transactionID, cardID, cvv string, expiryMonth, expiryYear int, amount int64, merchantName string) (*models.CardPaymentResult, error) {
			if amount > 100000 {
				return nil, errors.New("limit exceeded")
			}
			return &models.CardPaymentResult{
				TransactionID: transactionID,
				UserID:        "user-123",
				UserEmail:     "john@example.com",
				UserName:      "John Doe",
				MerchantName:  merchantName,
				Amount:        amount,
			}, nil
		},
	}

	service := services.NewCardService(mockRepo, mockNotif)

	// Test Payment Success
	txID, err := service.ProcessPayment(context.Background(), "card-1", "123", 12, 2028, 1100, "Supermarket")
	if err != nil {
		t.Fatalf("ProcessPayment() unexpected error = %v", err)
	}
	if txID == "" {
		t.Error("ProcessPayment() expected non-empty transaction ID")
	}

	// Verify notification payload formatting (1100 cents -> "11.00")
	if notifiedEvent == nil {
		t.Fatal("ProcessPayment() expected notification to be dispatched")
	}
	if notifiedEvent.UserID != "user-123" {
		t.Errorf("Notification UserID got %v, want user-123", notifiedEvent.UserID)
	}
	expectedMsg := "Payment of 11.00 to Supermarket was processed successfully."
	if notifiedEvent.Message != expectedMsg {
		t.Errorf("Notification Message got %q, want %q", notifiedEvent.Message, expectedMsg)
	}

	// Test Payment Failure
	_, errFail := service.ProcessPayment(context.Background(), "card-1", "123", 12, 2028, 500000, "Expensive Shop")
	if errFail == nil {
		t.Error("ProcessPayment() expected error for amount exceeding limit")
	}
}
