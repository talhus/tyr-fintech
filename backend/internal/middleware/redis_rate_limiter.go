package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	rdb *redis.Client
}

func NewRedisRateLimiter(rdb *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{rdb: rdb}
}

func (r *RedisRateLimiter) Limit(keyPrefix string, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if r == nil || r.rdb == nil {
			c.Next()
			return
		}
		//get client ip
		clientIP := c.ClientIP()

		//create unique key
		redisKey := fmt.Sprintf("rate_limit:%s:%s", keyPrefix, clientIP)

		//ctx
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second*2)
		defer cancel()

		//incr ip
		count, err := r.rdb.Incr(ctx, redisKey).Result()
		if err != nil {
			log.Printf("Warning: redis rate limiter error: %v \n", err)
			c.Next()
			return
		}

		if count == 1 {
			r.rdb.Expire(ctx, redisKey, window)
		}

		if count > limit {
			//ttl returns remaing time of a key that has a timeout
			ttl, _ := r.rdb.TTL(ctx, redisKey).Result()
			retrySeconds := int(ttl.Seconds())
			if retrySeconds <= 0 {
				retrySeconds = int(window.Seconds())
			}
			// set http header
			c.Header("Retry-After", fmt.Sprintf("%d", retrySeconds))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Too many requests, please try again later",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
