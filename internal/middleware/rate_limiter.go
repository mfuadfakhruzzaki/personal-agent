package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter represents a simple rate limiter
type RateLimiter struct {
	visitors        map[string]*visitor
	mutex           sync.RWMutex
	rate            int           // requests per second
	burst           int           // burst capacity
	cleanupInterval time.Duration // cleanup interval
}

type visitor struct {
	limiter  *TokenBucket
	lastSeen time.Time
}

type TokenBucket struct {
	tokens    int
	capacity  int
	rate      int
	lastRefill time.Time
	mutex     sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate, burst int, cleanup time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors:        make(map[string]*visitor),
		rate:            rate,
		burst:           burst,
		cleanupInterval: cleanup,
	}

	// Start cleanup goroutine
	go rl.cleanupRoutine()

	return rl
}

// Middleware returns gin middleware for rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		if !rl.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests",
				"code":    http.StatusTooManyRequests,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allow checks if request is allowed for given IP
func (rl *RateLimiter) allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			limiter:  newTokenBucket(rl.rate, rl.burst),
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = v
	}

	v.lastSeen = time.Now()
	return v.limiter.allow()
}

// cleanupRoutine removes old visitors
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanupVisitors()
	}
}

func (rl *RateLimiter) cleanupVisitors() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	cutoff := time.Now().Add(-rl.cleanupInterval)
	for ip, v := range rl.visitors {
		if v.lastSeen.Before(cutoff) {
			delete(rl.visitors, ip)
		}
	}
}

// newTokenBucket creates a new token bucket
func newTokenBucket(rate, capacity int) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		rate:       rate,
		lastRefill: time.Now(),
	}
}

// allow checks if a token is available
func (tb *TokenBucket) allow() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	
	// Add tokens based on elapsed time
	tokensToAdd := int(elapsed.Seconds()) * tb.rate
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	// Check if token is available
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}
