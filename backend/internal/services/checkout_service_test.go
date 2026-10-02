package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/internal/services"
	"github.com/iamtbay/tyr-fintech/internal/worker"
)

type mockCheckoutRepository struct {
	sessions map[string]*models.CheckoutSession
}

func newMockCheckoutRepository() *mockCheckoutRepository {
	return &mockCheckoutRepository{
		sessions: make(map[string]*models.CheckoutSession),
	}
}

func (m *mockCheckoutRepository) CreateSession(ctx context.Context, session *models.CheckoutSession) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *mockCheckoutRepository) GetSessionByID(ctx context.Context, sessionID string) (*models.CheckoutSession, error) {
	s, ok := m.sessions[sessionID]
	if !ok {
		return nil, errors.New("checkout session not found")
	}
	return s, nil
}

func (m *mockCheckoutRepository) GetSessionDetails(ctx context.Context, sessionID string) (*dto.CheckoutSessionDetailsResponse, error) {
	s, ok := m.sessions[sessionID]
	if !ok {
		return nil, errors.New("checkout session not found")
	}
	return &dto.CheckoutSessionDetailsResponse{
		SessionID:    s.ID,
		MerchantName: "Foodeli Inc.",
		OrderID:      s.OrderID,
		Amount:       s.Amount,
		Currency:     s.Currency,
		Status:       string(s.Status),
		ExpiresAt:    s.ExpiresAt,
	}, nil
}

func (m *mockCheckoutRepository) ExecuteWalletPayment(ctx context.Context, sessionID, userID, walletID string, merchant *models.Merchant) (string, error) {
	s, ok := m.sessions[sessionID]
	if !ok {
		return "", errors.New("checkout session not found")
	}
	if s.Status != models.CheckoutStatusPending {
		return "", errors.New("session already paid or invalid")
	}

	if walletID == "empty-wallet-id" {
		return "", errors.New("insufficient wallet balance")
	}

	trxID := "trx_test_wallet_987654"
	s.Status = models.CheckoutStatusPaid
	s.TransactionID = &trxID
	s.PayerUserID = &userID
	return trxID, nil
}

type mockMerchantRepository struct {
	merchant *models.Merchant
}

func (m *mockMerchantRepository) GetMerchantByAPIKeyHash(ctx context.Context, keyHash string) (*models.Merchant, error) {
	return m.merchant, nil
}

func (m *mockMerchantRepository) GetMerchantByID(ctx context.Context, id string) (*models.Merchant, error) {
	return m.merchant, nil
}

func TestCheckoutService_FullFlow(t *testing.T) {
	checkoutRepo := newMockCheckoutRepository()
	merchant := &models.Merchant{
		ID:                       "merchant-foodeli-uuid",
		Name:                     "Foodeli Inc.",
		SettlementWalletID:       "wallet-foodeli-try",
		FeePercentageBasisPoints: 200, // 2.00%
		Status:                   models.MerchantStatusActive,
	}
	merchantRepo := &mockMerchantRepository{merchant: merchant}

	frontendURL := "https://checkout.tyrfintech.com"
	checkoutService := services.NewCheckoutService(checkoutRepo, merchantRepo, frontendURL)

	ctx := context.Background()

	// 1. Merchant creates checkout session
	createReq := &dto.CreateCheckoutSessionRequest{
		OrderID:     "foodeli_order_9988",
		Amount:      15000, // 150.00 TRY
		Currency:    "TRY",
		CallbackURL: "https://foodeli.com/checkout/callback",
		WebhookURL:  "https://api.foodeli.com/webhooks/tyrfintech",
	}

	createResp, err := checkoutService.CreateSession(ctx, merchant, createReq)
	if err != nil {
		t.Fatalf("failed to create checkout session: %v", err)
	}

	if !strings.HasPrefix(createResp.SessionID, "cs_live_") {
		t.Fatalf("expected session ID to start with cs_live_, got %s", createResp.SessionID)
	}
	if !strings.Contains(createResp.CheckoutURL, frontendURL) {
		t.Fatalf("expected checkout URL to use frontendURL, got %s", createResp.CheckoutURL)
	}

	// 2. User loads checkout page UI (gets session details)
	details, err := checkoutService.GetSessionDetails(ctx, createResp.SessionID)
	if err != nil {
		t.Fatalf("failed to get session details: %v", err)
	}
	if details.Amount != 15000 || details.MerchantName != "Foodeli Inc." {
		t.Fatalf("unexpected details: %+v", details)
	}

	// 3. User pays from their TyrFintech wallet
	payReq := &dto.PayCheckoutSessionRequest{
		WalletID: "user-funded-wallet-uuid",
	}

	payResp, err := checkoutService.PaySession(ctx, createResp.SessionID, "user-alice-uuid", payReq)
	if err != nil {
		t.Fatalf("payment failed: %v", err)
	}

	if !payResp.Success {
		t.Fatal("expected payment success true")
	}
	if payResp.Status != "SUCCEEDED" {
		t.Fatalf("expected status SUCCEEDED, got %s", payResp.Status)
	}

	// 4. Verify Redirect URL was constructed with callback parameters
	if !strings.Contains(payResp.RedirectURL, "status=SUCCEEDED") ||
		!strings.Contains(payResp.RedirectURL, "order_id=foodeli_order_9988") ||
		!strings.Contains(payResp.RedirectURL, payResp.TransactionID) {
		t.Fatalf("malformed redirect URL: %s", payResp.RedirectURL)
	}

	// 5. Verify Webhook Event was dispatched to worker queue
	select {
	case event := <-worker.MerchantWebHookQueue:
		if event.Event != "checkout.session.paid" {
			t.Fatalf("expected event checkout.session.paid, got %s", event.Event)
		}
		if event.OrderID != "foodeli_order_9988" {
			t.Fatalf("expected order ID foodeli_order_9988, got %s", event.OrderID)
		}
		if event.WebhookURL != "https://api.foodeli.com/webhooks/tyrfintech" {
			t.Fatalf("expected webhook URL https://api.foodeli.com/webhooks/tyrfintech, got %s", event.WebhookURL)
		}
	default:
		t.Fatal("expected webhook event to be enqueued in MerchantWebHookQueue")
	}
}

func TestCheckoutService_PaySession_InsufficientBalance(t *testing.T) {
	checkoutRepo := newMockCheckoutRepository()
	merchant := &models.Merchant{
		ID:                       "merchant-foodeli-uuid",
		Name:                     "Foodeli Inc.",
		SettlementWalletID:       "wallet-foodeli-try",
		FeePercentageBasisPoints: 200,
		Status:                   models.MerchantStatusActive,
	}
	merchantRepo := &mockMerchantRepository{merchant: merchant}
	checkoutService := services.NewCheckoutService(checkoutRepo, merchantRepo, "http://localhost:3000")

	ctx := context.Background()

	createResp, _ := checkoutService.CreateSession(ctx, merchant, &dto.CreateCheckoutSessionRequest{
		OrderID:     "order_fail_1",
		Amount:      50000,
		Currency:    "TRY",
		CallbackURL: "https://foodeli.com/cb",
		WebhookURL:  "https://foodeli.com/wh",
	})

	_, err := checkoutService.PaySession(ctx, createResp.SessionID, "user-bob-uuid", &dto.PayCheckoutSessionRequest{
		WalletID: "empty-wallet-id",
	})

	if err == nil || !strings.Contains(err.Error(), "insufficient wallet balance") {
		t.Fatalf("expected insufficient balance error, got %v", err)
	}
}
