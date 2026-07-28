package services_test

import (
	"context"
	"testing"

	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/iamtbay/tyr-fintech/internal/services"
)

func TestMockExchangeService_GetRate(t *testing.T) {
	svc := services.NewMockExchangeService()

	tests := []struct {
		name     string
		from     models.WalletCurrency
		to       models.WalletCurrency
		wantRate float64
		wantErr  bool
	}{
		{
			name:     "Same Currency TRY to TRY",
			from:     models.CurrencyTRY,
			to:       models.CurrencyTRY,
			wantRate: 1.0,
			wantErr:  false,
		},
		{
			name:     "USD to TRY",
			from:     models.CurrencyUSD,
			to:       models.CurrencyTRY,
			wantRate: 47.06,
			wantErr:  false,
		},
		{
			name:     "TRY to USD",
			from:     models.CurrencyTRY,
			to:       models.CurrencyUSD,
			wantRate: 0.021,
			wantErr:  false,
		},
		{
			name:     "EUR to TRY",
			from:     models.CurrencyEUR,
			to:       models.CurrencyTRY,
			wantRate: 54.46,
			wantErr:  false,
		},
		{
			name:     "TRY to EUR",
			from:     models.CurrencyTRY,
			to:       models.CurrencyEUR,
			wantRate: 0.019,
			wantErr:  false,
		},
		{
			name:     "Unsupported currency pair",
			from:     models.CurrencyUSD,
			to:       "GBP",
			wantRate: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate, err := svc.GetRate(context.Background(), tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if rate != tt.wantRate {
					t.Errorf("GetRate() got rate = %v, want %v", rate, tt.wantRate)
				}
			}
		})
	}
}
