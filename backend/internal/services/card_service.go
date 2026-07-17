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
	UpdateStatus(ctx context.Context, cardID, userID string, status models.CardStatus) error
	ProcessPayment(ctx context.Context, cardNumber, cvv string, expiryMonth, expiryYear int, amount int64) error
}

type CardService struct {
	repo CardRepository
}

func NewCardService(cardRepo CardRepository) *CardService {
	return &CardService{repo: cardRepo}
}

// METHODS

// create card
func (s *CardService) CreateCard(ctx context.Context, userID, walletID string, limitAmount int) (*models.Card, error) {
	//create card details
	card := &models.Card{
		ID:          uuid.NewString(),
		UserID:      userID,
		WalletID:    walletID,
		LimitAmount: limitAmount,
		CVV:         utils.GenerateCVV(),
		CardNumber:  utils.GenerateCardNumber(),
		ExpiryMonth: int(time.Now().Month()) + 1,
		ExpiryYear:  time.Now().Year() + 5,
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

// UpdateCardStatus
func (s *CardService) UpdateCardStatus(ctx context.Context, cardID, userID string, status models.CardStatus) error {
	return s.repo.UpdateStatus(ctx, cardID, userID, status)
}

// ProcessPayment
func (s *CardService) ProcessPayment(ctx context.Context, cardNumber, cvv string, expiryMonth, expiryYear int, amount int64) error {
	return s.repo.ProcessPayment(ctx, cardNumber, cvv, expiryMonth, expiryYear, amount)
}
