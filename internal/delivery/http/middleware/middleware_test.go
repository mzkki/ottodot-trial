package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRequestID(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		reqID := c.GetString("request_id")
		if reqID == "" {
			t.Error("expected non-empty request_id in context")
		}
		c.String(http.StatusOK, reqID)
	})

	// Case 1: No incoming X-Request-ID header
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header in response")
	}

	// Case 2: Incoming X-Request-ID header preserved
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Request-ID", "custom-id-123")
	r.ServeHTTP(w2, req2)

	if w2.Header().Get("X-Request-ID") != "custom-id-123" {
		t.Errorf("expected custom-id-123, got %s", w2.Header().Get("X-Request-ID"))
	}
}

func TestRecoveryLog(t *testing.T) {
	logger := zap.NewNop()
	r := gin.New()
	r.Use(RecoveryLog(logger))
	r.GET("/panic", func(c *gin.Context) {
		panic("something went terribly wrong")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestAccessLog(t *testing.T) {
	logger := zap.NewNop()
	r := gin.New()
	r.Use(AccessLog(logger))
	r.GET("/ok", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.GET("/bad", func(c *gin.Context) {
		c.Status(http.StatusBadRequest)
	})
	r.GET("/server-err", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})

	for _, path := range []string{"/ok", "/bad", "/server-err"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", path, nil)
		r.ServeHTTP(w, req)
	}
}

func TestSecurityHeaders(t *testing.T) {
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/secure", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/secure", nil)
	r.ServeHTTP(w, req)

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected X-Content-Type-Options: nosniff, got %s", w.Header().Get("X-Content-Type-Options"))
	}
	if w.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Errorf("expected X-Frame-Options: SAMEORIGIN, got %s", w.Header().Get("X-Frame-Options"))
	}
	if w.Header().Get("X-XSS-Protection") != "1; mode=block" {
		t.Errorf("expected X-XSS-Protection: 1; mode=block, got %s", w.Header().Get("X-XSS-Protection"))
	}
	if w.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Errorf("expected Referrer-Policy: strict-origin-when-cross-origin, got %s", w.Header().Get("Referrer-Policy"))
	}
}

func TestRateLimit(t *testing.T) {
	// Create limiter: 1 req/sec with burst of 2
	limiter := NewIPRateLimiter(1, 2)

	r := gin.New()
	r.Use(RateLimit(limiter))
	r.GET("/api/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// First 2 requests should succeed (burst = 2)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/test", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected request %d to succeed with 200, got %d", i+1, w.Code)
		}
	}

	// 3rd request should be throttled (429 Too Many Requests)
	wJSON := httptest.NewRecorder()
	reqJSON := httptest.NewRequest("GET", "/api/test", nil)
	r.ServeHTTP(wJSON, reqJSON)

	if wJSON.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests, got %d", wJSON.Code)
	}
	if wJSON.Header().Get("Retry-After") != "1" {
		t.Errorf("expected Retry-After: 1, got %s", wJSON.Header().Get("Retry-After"))
	}

	// 4th request with HTMX header should return 429 with HTMX partial
	wHTMX := httptest.NewRecorder()
	reqHTMX := httptest.NewRequest("GET", "/api/test", nil)
	reqHTMX.Header.Set("HX-Request", "true")
	r.ServeHTTP(wHTMX, reqHTMX)

	if wHTMX.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for HTMX request, got %d", wHTMX.Code)
	}

	// Test client tracking and cleanup
	limiter.mu.Lock()
	if len(limiter.clients) == 0 {
		t.Error("expected at least 1 tracked client IP")
	}
	limiter.mu.Unlock()
}


