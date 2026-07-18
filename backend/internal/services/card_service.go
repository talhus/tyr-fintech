package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/pkg/utils"
)

type CardRepository interface {
	Create(ctx context.Context, card *models.Card) error
	GetByUserID(ctx context.Context, userID string) ([]models.Card, error)
	GetCardDetails(ctx context.Context, cardID, userID string) (*models.Card, error)
	UpdateStatus(ctx context.Context, cardID, userID string, status models.CardStatus) error
	GetCardTransactions(ctx context.Context, cardID, userID string) ([]models.Transaction, error)
	ProcessPayment(ctx context.Context, transactionID, cardID, cvv string, expiryMonth, expiryYear int, amount int64, merchantName string) error
}

type CardService struct {
	repo CardRepository
}

func NewCardService(cardRepo CardRepository) *CardService {
	return &CardService{repo: cardRepo}
}

// METHODS

// create card
func (s *CardService) CreateCard(ctx context.Context, userID, walletID string, limitAmount int64) (*models.Card, error) {
	//create card details
	card := &models.Card{
		ID:          uuid.New().String(),
		UserID:      userID,
		WalletID:    walletID,
		LimitAmount: limitAmount,
		CVV:         utils.GenerateCVV(),
		CardNumber:  utils.GenerateCardNumber(),
		ExpiryMonth: int(time.Now().Month()),
		ExpiryYear:  time.Now().Year() + 5,
		Status:      models.CardStatusActive,
	}

	err := s.repo.Create(ctx, card)
	if err != nil {
		return nil, err
	}
	return card, nil

}

// GetCardsByUserID
func (s *CardService) GetCardsByUserID(ctx context.Context, userID string) ([]models.Card, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// GetCardDetails
func (s *CardService) GetCardDetails(ctx context.Context, cardID, userID string) (*models.Card, error) {
	return s.repo.GetCardDetails(ctx, cardID, userID)
}

// UpdateCardStatus
func (s *CardService) UpdateCardStatus(ctx context.Context, cardID, userID string, status models.CardStatus) error {
	return s.repo.UpdateStatus(ctx, cardID, userID, status)
}

// GetCardTransactions
func (s *CardService) GetCardTransactions(ctx context.Context, cardID, userID string) ([]models.Transaction, error) {
	return s.repo.GetCardTransactions(ctx, cardID, userID)
}

// ProcessPayment
func (s *CardService) ProcessPayment(ctx context.Context, cardID, cvv string, expiryMonth, expiryYear int, amount int64, merchantName string) (string, error) {
	transactionID := uuid.New().String()
	if merchantName == "" {
		merchantName = "Online Merchant"
	}
	return transactionID, s.repo.ProcessPayment(ctx, transactionID, cardID, cvv, expiryMonth, expiryYear, amount, merchantName)
}
