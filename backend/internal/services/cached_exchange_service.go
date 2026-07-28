package services

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/iamtbay/tyr-fintech/internal/models"
	"github.com/redis/go-redis/v9"
)

type ExchangeRateProvider interface {
	GetRate(ctx context.Context, from, to models.WalletCurrency) (float64, error)
}

type CachedExchangeService struct {
	delegate ExchangeRateProvider
	rdb      *redis.Client
	ttl      time.Duration
}

func NewCachedExchangeService(delegate ExchangeRateProvider, rdb *redis.Client, ttl time.Duration) *CachedExchangeService {
	if ttl == 0 {
		ttl = 10 * time.Minute
	}
	return &CachedExchangeService{
		delegate: delegate,
		rdb:      rdb,
		ttl:      ttl,
	}
}

// GET RATE
func (s *CachedExchangeService) GetRate(ctx context.Context, from, to models.WalletCurrency) (float64, error) {
	if from == to {
		return 1.0, nil
	}

	//construct key
	cacheKey := fmt.Sprintf("exchange_rate:%s:%s", from, to)

	//1 read exchange rate from redis
	cachedVal, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		rate, parseErr := strconv.ParseFloat(cachedVal, 64)
		if parseErr == nil {
			fmt.Printf("Exchange rate cache hit: %s -> %s = %f\n", from, to, rate)
			return rate, nil
		}
	}

	//2 cache miss fetch rate from source
	rate, err := s.delegate.GetRate(ctx, from, to)
	if err != nil {
		return 0, fmt.Errorf("failed to get exchange rate from source: %w", err)
	}

	//3 store rate in redis with ttl
	rateStr := strconv.FormatFloat(rate, 'f', -1, 64)
	setErr := s.rdb.Set(ctx, cacheKey, rateStr, s.ttl).Err()
	if err != nil {
		// log warning if redis caching fails, but do not block the response
		log.Printf("Warning: Failed to cache rate in Redis %v \n", setErr)
	}
	return rate, nil

}
