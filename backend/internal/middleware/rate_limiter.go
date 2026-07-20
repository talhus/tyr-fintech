package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter manages rate limiters per client IP for a specific middleware instance.
type IPRateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*client
	rate     rate.Limit
	burst    int
	cleanup  time.Duration
}

// NewIPRateLimiter creates a new rate limiter instance with isolated IP tracking.
func NewIPRateLimiter(requestsPerWindow int, window time.Duration) *IPRateLimiter {
	if requestsPerWindow <= 0 || window <= 0 {
		requestsPerWindow = 60
		window = time.Minute
	}

	limit := rate.Limit(float64(requestsPerWindow) / window.Seconds())
	limiter := &IPRateLimiter{
		clients: make(map[string]*client),
		rate:    limit,
		burst:   requestsPerWindow,
		cleanup: 3 * time.Minute,
	}

	go limiter.cleanupClients()
	return limiter
}

func (i *IPRateLimiter) cleanupClients() {
	for {
		time.Sleep(time.Minute)
		i.mu.Lock()
		for ip, c := range i.clients {
			if time.Since(c.lastSeen) > i.cleanup {
				delete(i.clients, ip)
			}
		}
		i.mu.Unlock()
	}
}

func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	cl, exists := i.clients[ip]
	if !exists {
		limiter := rate.NewLimiter(i.rate, i.burst)
		i.clients[ip] = &client{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}
	cl.lastSeen = time.Now()
	return cl.limiter
}

// RateLimit returns a Gin middleware that limits requests per IP.
func RateLimit(requestsPerWindow int, window time.Duration) gin.HandlerFunc {
	limiter := NewIPRateLimiter(requestsPerWindow, window)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		l := limiter.getLimiter(ip)

		if !l.Allow() {
			retrySeconds := int(window.Seconds())
			if retrySeconds < 1 {
				retrySeconds = 1
			}
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

