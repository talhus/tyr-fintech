package services

import (
	"context"
	"fmt"

	"github.com/iamtbay/tyr-fintech/internal/models"
)

type CurrencyPair struct {
	From models.WalletCurrency
	To   models.WalletCurrency
}

// MockExchangeService struct
type MockExchangeService struct {
	rates map[CurrencyPair]float64
}

func NewMockExchangeService() *MockExchangeService {
	return &MockExchangeService{
		rates: map[CurrencyPair]float64{
			{models.CurrencyTRY, models.CurrencyUSD}: 0.021,
			{models.CurrencyUSD, models.CurrencyTRY}: 47.06,
			{models.CurrencyUSD, models.CurrencyEUR}: 0.93,
			{models.CurrencyEUR, models.CurrencyUSD}: 1.07,
			{models.CurrencyEUR, models.CurrencyTRY}: 54.46,
			{models.CurrencyTRY, models.CurrencyEUR}: 0.019,
		},
	}
}

func (s *MockExchangeService) GetRate(ctx context.Context, from, to models.WalletCurrency) (float64, error) {
	pair := CurrencyPair{From: from, To: to}

	if from == to {
		return 1, nil
	} else if rate, ok := s.rates[pair]; ok {
		return rate, nil
	} else {
		return 0, fmt.Errorf("exchange rate not found for %s/%s", from, to)
	}
}
