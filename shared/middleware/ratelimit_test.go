package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(60, 10) // 60 req/min, burst 10

	key := "test-client"

	// Should allow burst requests
	for i := 0; i < 10; i++ {
		if !limiter.Allow(key) {
			t.Errorf("Request %d should be allowed (within burst)", i+1)
		}
	}

	// 11th request should be denied (burst exhausted)
	if limiter.Allow(key) {
		t.Error("Request should be denied after burst exhausted")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	limiter := NewRateLimiter(60, 2) // 60 req/min = 1 req/sec, burst 2

	key := "test-client"

	// Use up burst
	limiter.Allow(key)
	limiter.Allow(key)

	// Should be denied
	if limiter.Allow(key) {
		t.Error("Request should be denied")
	}

	// Wait for refill (need at least 1 second for 1 token at 60 req/min)
	time.Sleep(1100 * time.Millisecond)

	// Should be allowed again
	if !limiter.Allow(key) {
		t.Error("Request should be allowed after refill")
	}
}

func TestRateLimiter_MultipleClients(t *testing.T) {
	limiter := NewRateLimiter(60, 5) // 60 req/min, burst 5

	client1 := "client-1"
	client2 := "client-2"

	// Exhaust client1's burst
	for i := 0; i < 5; i++ {
		limiter.Allow(client1)
	}

	// client1 should be denied
	if limiter.Allow(client1) {
		t.Error("Client 1 should be denied")
	}

	// client2 should still be allowed
	if !limiter.Allow(client2) {
		t.Error("Client 2 should be allowed")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	limiter := NewRateLimiter(60, 2) // Very restrictive for testing
	middleware := RateLimitMiddleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrappedHandler := middleware(handler)

	// First request should succeed
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("First request status = %d, want %d", w.Code, http.StatusOK)
	}

	// Second request should succeed (burst)
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w = httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Second request status = %d, want %d", w.Code, http.StatusOK)
	}

	// Third request should be rate limited
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w = httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Third request status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimitMiddleware_XForwardedFor(t *testing.T) {
	limiter := NewRateLimiter(60, 1)
	middleware := RateLimitMiddleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	// First request with X-Forwarded-For
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("First request status = %d, want %d", w.Code, http.StatusOK)
	}

	// Second request should be rate limited (same X-Forwarded-For)
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	w = httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Second request status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}

func TestEndpointRateLimiter(t *testing.T) {
	erl := NewEndpointRateLimiter(60, 10) // Default: 60 req/min, burst 10
	erl.AddEndpoint("/api/login", 10, 2)   // Stricter: 10 req/min, burst 2

	// Test default endpoint
	defaultLimiter := erl.GetLimiter("/api/users")
	defaultExpected := erl.defaultLimiter
	if defaultLimiter != defaultExpected {
		t.Error("Should return default limiter for unknown endpoint")
	}

	// Test specific endpoint
	loginLimiter := erl.GetLimiter("/api/login")
	if loginLimiter == defaultExpected {
		t.Error("Should return specific limiter for /api/login")
	}

	// Verify different burst sizes
	key := "test-client"

	// Default endpoint should allow 10 requests
	for i := 0; i < 10; i++ {
		if !defaultLimiter.Allow(key) {
			t.Errorf("Default endpoint request %d should be allowed", i+1)
		}
	}

	// Login endpoint should only allow 2 requests (different key to avoid cross-contamination)
	key2 := "test-client-2"
	for i := 0; i < 2; i++ {
		if !loginLimiter.Allow(key2) {
			t.Errorf("Login endpoint request %d should be allowed", i+1)
		}
	}

	// 3rd request to login should be denied
	if loginLimiter.Allow(key2) {
		t.Error("Login endpoint should deny 3rd request")
	}
}

func TestRateLimiter_Concurrent(t *testing.T) {
	limiter := NewRateLimiter(1000, 100) // High limits for concurrent test

	key := "concurrent-client"
	var wg sync.WaitGroup
	requests := 50

	allowed := 0
	var mu sync.Mutex

	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.Allow(key) {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Should allow all requests within burst
	if allowed != requests {
		t.Errorf("Allowed %d requests, want %d", allowed, requests)
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	limiter := NewRateLimiter(1000, 100)
	key := "benchmark-client"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(key)
	}
}

func BenchmarkRateLimiter_MultipleClients(b *testing.B) {
	limiter := NewRateLimiter(1000, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "client-" + string(rune(i%100))
		limiter.Allow(key)
	}
}
