package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type clientTracker struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter throttles requests per client IP address.
type IPRateLimiter struct {
	mu      sync.RWMutex
	clients map[string]*clientTracker
	r       rate.Limit
	b       int
}

// NewIPRateLimiter creates a new IPRateLimiter with background cleanup to prevent memory leaks.
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		clients: make(map[string]*clientTracker),
		r:       r,
		b:       b,
	}

	// Periodically purge inactive clients from memory
	go limiter.cleanup(5 * time.Minute)

	return limiter
}

func (i *IPRateLimiter) getClient(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	client, exists := i.clients[ip]
	if !exists {
		limiter := rate.NewLimiter(i.r, i.b)
		i.clients[ip] = &clientTracker{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	client.lastSeen = time.Now()
	return client.limiter
}

func (i *IPRateLimiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		i.mu.Lock()
		for ip, client := range i.clients {
			if time.Since(client.lastSeen) > 3*interval {
				delete(i.clients, ip)
			}
		}
		i.mu.Unlock()
	}
}

// RateLimit returns a Gin middleware that enforces the specified rate limit per client IP.
func RateLimit(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		clientLimiter := limiter.getClient(ip)

		if !clientLimiter.Allow() {
			c.Header("Retry-After", "1")

			// Return HTMX error partial if requested via HTMX
			if c.GetHeader("HX-Request") == "true" {
				c.Header("Content-Type", "text/html; charset=utf-8")
				c.String(http.StatusTooManyRequests, `<div class="bg-red-50 border border-red-200 rounded-lg p-4 text-center"><p class="text-red-700 font-medium">⚠️ Too many requests. Please slow down and try again.</p></div>`)
			} else {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error":   true,
					"message": "too many requests, please slow down",
				})
			}
			c.Abort()
			return
		}

		c.Next()
	}
}
