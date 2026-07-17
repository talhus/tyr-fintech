package services

import (
	"context"
	"errors"
	"math"

	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/internal/worker"
)

// ExhangeService interface
type ExhangeService interface {
	GetRate(ctx context.Context, from, to models.WalletCurrency) (float64, error)
}

type TransactionRepository interface {
	Transfer(ctx context.Context, req *dto.TransferRequest, convertedAmount int64) error
	GetTransactionsByWalletID(ctx context.Context, walletID string) ([]*models.Transaction, error)
}

type TransactionService struct {
	repo            TransactionRepository
	exchangeService ExhangeService
	walletRepo      WalletRepository
}

func NewTransactionService(repo TransactionRepository, exchangeService ExhangeService, walletRepo WalletRepository) *TransactionService {
	return &TransactionService{repo: repo, exchangeService: exchangeService, walletRepo: walletRepo}
}

// Transfer
func (s *TransactionService) Transfer(ctx context.Context, req *dto.TransferRequest) error {
	if req.TransactionID == "" {
		return errors.New("Transaction request missing required idempotency key.")
	}
	sender, err := s.walletRepo.GetWalletByID(ctx, req.FromWalletNumber)
	if err != nil {
		return err
	}
	receiver, err := s.walletRepo.GetWalletByID(ctx, req.ToWalletNumber)
	if err != nil {
		return err
	}

	//convert
	rate, err := s.exchangeService.GetRate(ctx, sender.Currency, receiver.Currency)
	if err != nil {
		return err
	}
	convertedAmount := int64(math.Round(float64(req.Amount) * rate))

	//transfer
	err = s.repo.Transfer(ctx, req, convertedAmount)
	if err != nil {
		return err
	}
	worker.WebHookQueue <- &dto.TransactionWebhookEvent{
		TransactionID:    req.TransactionID,
		FromWalletNumber: req.FromWalletNumber,
		ToWalletNumber:   req.ToWalletNumber,
		Amount:           req.Amount,
		Status:           "COMPLETED",
	}
	return nil
}

// GetHistory
func (s *TransactionService) GetHistory(ctx context.Context, walletID string) ([]*models.Transaction, error) {
	return s.repo.GetTransactionsByWalletID(ctx, walletID)
}

// GetExchangeRate
func (s *TransactionService) GetExchangeRate(ctx context.Context, from, to models.WalletCurrency) (float64, error) {
	return s.exchangeService.GetRate(ctx, from, to)
}
