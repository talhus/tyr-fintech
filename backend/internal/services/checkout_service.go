package services

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/internal/repos"
	"github.com/iamtbay/tyr-fintech/internal/worker"
)

type CheckoutService interface {
	CreateSession(ctx context.Context, merchant *models.Merchant, req *dto.CreateCheckoutSessionRequest) (*dto.CreateCheckoutSessionResponse, error)
	GetSessionDetails(ctx context.Context, sessionID string) (*dto.CheckoutSessionDetailsResponse, error)
	PaySession(ctx context.Context, sessionID, userID string, req *dto.PayCheckoutSessionRequest) (*dto.PayCheckoutSessionResponse, error)
}

type checkoutService struct {
	checkoutRepo repos.CheckoutRepository
	merchantRepo repos.MerchantRepository
	frontendURL  string
}

func NewCheckoutService(checkoutRepo repos.CheckoutRepository, merchantRepo repos.MerchantRepository, frontendURL string) CheckoutService {
	cleanURL := strings.TrimRight(frontendURL, "/")
	if cleanURL == "" {
		cleanURL = "http://localhost:3005"
	}
	return &checkoutService{
		checkoutRepo: checkoutRepo,
		merchantRepo: merchantRepo,
		frontendURL:  cleanURL,
	}
}

func (s *checkoutService) CreateSession(ctx context.Context, merchant *models.Merchant, req *dto.CreateCheckoutSessionRequest) (*dto.CreateCheckoutSessionResponse, error) {
	sessionID := fmt.Sprintf("cs_live_%s", strings.ReplaceAll(uuid.New().String(), "-", "")[:16])
	expiresAt := time.Now().Add(30 * time.Minute)

	session := &models.CheckoutSession{
		ID:          sessionID,
		MerchantID:  merchant.ID,
		OrderID:     req.OrderID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		CallbackURL: req.CallbackURL,
		WebHookURL:  req.WebhookURL,
		Status:      models.CheckoutStatusPending,
		ExpiresAt:   expiresAt,
	}

	if err := s.checkoutRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	checkoutURL := fmt.Sprintf("%s/checkout?session_id=%s", s.frontendURL, sessionID)

	return &dto.CreateCheckoutSessionResponse{
		SessionID:   sessionID,
		CheckoutURL: checkoutURL,
		Status:      string(models.CheckoutStatusPending),
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *checkoutService) GetSessionDetails(ctx context.Context, sessionID string) (*dto.CheckoutSessionDetailsResponse, error) {
	return s.checkoutRepo.GetSessionDetails(ctx, sessionID)
}

func (s *checkoutService) PaySession(ctx context.Context, sessionID, userID string, req *dto.PayCheckoutSessionRequest) (*dto.PayCheckoutSessionResponse, error) {
	session, err := s.checkoutRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	merchant, err := s.merchantRepo.GetMerchantByID(ctx, session.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}

	// Atomically execute payment settlement
	transactionID, err := s.checkoutRepo.ExecuteWalletPayment(ctx, sessionID, userID, req.WalletID, merchant)
	if err != nil {
		return nil, err
	}

	// Build return URL with query parameters for Foodeli redirect
	separator := "?"
	if strings.Contains(session.CallbackURL, "?") {
		separator = "&"
	}
	redirectURL := fmt.Sprintf("%s%sstatus=SUCCEEDED&transaction_id=%s&order_id=%s&session_id=%s",
		session.CallbackURL,
		separator,
		url.QueryEscape(transactionID),
		url.QueryEscape(session.OrderID),
		url.QueryEscape(sessionID),
	)

	// Enqueue asynchronous webhook event to Foodeli (Durable RabbitMQ + In-Memory Fallback)
	worker.EnqueueMerchantWebhook(&dto.MerchantWebhookEvent{
		Event:         "checkout.session.paid",
		TransactionID: transactionID,
		SessionID:     sessionID,
		OrderID:       session.OrderID,
		Amount:        session.Amount,
		Currency:      session.Currency,
		Status:        "SUCCEEDED",
		WebhookURL:    session.WebHookURL,
		Timestamp:     time.Now().Unix(),
	})

	return &dto.PayCheckoutSessionResponse{
		Success:       true,
		TransactionID: transactionID,
		Status:        "SUCCEEDED",
		RedirectURL:   redirectURL,
	}, nil
}
