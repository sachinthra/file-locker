package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sachinthra/file-locker/backend/internal/storage"
)

// Mock Redis for testing
type mockRedisCache struct {
	sessions map[string]string
	counters map[string]int64
}

func (m *mockRedisCache) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.sessions[key]
	return exists, nil
}

func (m *mockRedisCache) Incr(ctx context.Context, key string) (int64, error) {
	m.counters[key]++
	return m.counters[key], nil
}

func (m *mockRedisCache) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	m.sessions[key] = value
	return nil
}

func (m *mockRedisCache) SaveSession(ctx context.Context, token, userID string, expiration time.Duration) error {
	m.sessions["session:"+userID] = token
	return nil
}

func (m *mockRedisCache) GetSession(ctx context.Context, token string) (string, error) {
	if userID, exists := m.sessions[token]; exists {
		return userID, nil
	}
	return "", nil
}

func TestNewAuthMiddleware(t *testing.T) {
	jwtService := NewJWTService("test-secret", 3600)
	redis, _ := storage.NewRedisCache("localhost:6379", "", 0)

	middleware := NewAuthMiddleware(jwtService, redis, nil)

	if middleware == nil {
		t.Fatal("Expected middleware to be created")
	}

	if middleware.jwtService == nil {
		t.Error("Expected jwtService to be set")
	}

	if middleware.redisCache == nil {
		t.Error("Expected redisCache to be set")
	}
}

func TestRequireAuth_NoAuthHeader(t *testing.T) {
	jwtService := NewJWTService("test-secret", 3600)

	// Create actual AuthMiddleware with type assertion workaround
	redis, _ := storage.NewRedisCache("localhost:6379", "", 0)
	middleware := NewAuthMiddleware(jwtService, redis, nil)

	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestRequireAuth_InvalidFormat(t *testing.T) {
	jwtService := NewJWTService("test-secret", 3600)
	redis, _ := storage.NewRedisCache("localhost:6379", "", 0)
	middleware := NewAuthMiddleware(jwtService, redis, nil)

	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "InvalidToken"},
		{"Wrong prefix", "Basic token123"},
		{"Multiple spaces", "Bearer  token  with  spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tt.header)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("Expected status 401 for %s, got %d", tt.name, rec.Code)
			}
		})
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	jwtService := NewJWTService("test-secret", 3600)
	redis, _ := storage.NewRedisCache("localhost:6379", "", 0)
	middleware := NewAuthMiddleware(jwtService, redis, nil)

	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestRequireAuth_ContextUserID(t *testing.T) {
	// This test would require a working Redis instance
	// Skipping for unit tests
	t.Skip("Requires Redis instance")
}

func TestRateLimitMiddleware_NoAuth(t *testing.T) {
	jwtService := NewJWTService("test-secret", 3600)
	redis, _ := storage.NewRedisCache("localhost:6379", "", 0)
	middleware := NewAuthMiddleware(jwtService, redis, nil)

	rateLimiter := middleware.RateLimitMiddleware(5, 1*time.Minute)
	handler := rateLimiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/endpoint", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 without auth, got %d", rec.Code)
	}
}

func TestRateLimitMiddleware_WithUserID(t *testing.T) {
	// This test would require proper context setup
	t.Skip("Requires full authentication setup")
}

func TestAuthMiddleware_Integration(t *testing.T) {
	// This test would require Redis
	t.Skip("Requires Redis instance for integration testing")
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	// Create service with 1 second expiry
	jwtService := NewJWTService("test-secret", 1)
	redis, _ := storage.NewRedisCache("localhost:6379", "", 0)
	middleware := NewAuthMiddleware(jwtService, redis, nil)

	// Generate token
	token, err := jwtService.GenerateToken("user123")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait for expiration
	time.Sleep(2 * time.Second)

	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for expired token, got %d", rec.Code)
	}
}

func TestRateLimitMiddleware_WindowCalculation(t *testing.T) {
	// Test that window calculation works correctly
	window := 1 * time.Minute
	currentTime := time.Now().Unix()
	currentWindow := currentTime / int64(window.Seconds())

	if currentWindow <= 0 {
		t.Error("Expected positive window value")
	}

	// Different times in same window should have same window number
	time1 := time.Now()
	time.Sleep(100 * time.Millisecond)
	time2 := time.Now()

	window1 := time1.Unix() / int64(window.Seconds())
	window2 := time2.Unix() / int64(window.Seconds())

	if window1 != window2 {
		t.Error("Expected same window for times within same minute")
	}
}

func TestAuthMiddleware_ChainedMiddleware(t *testing.T) {
	jwtService := NewJWTService("test-secret", 3600)
	redis, _ := storage.NewRedisCache("localhost:6379", "", 0)
	middleware := NewAuthMiddleware(jwtService, redis, nil)

	// Create a chain of middleware
	rateLimiter := middleware.RateLimitMiddleware(5, 1*time.Minute)
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	})

	// Chain: RequireAuth -> RateLimit -> Handler
	chain := middleware.RequireAuth(rateLimiter(finalHandler))

	// Test without auth
	req := httptest.NewRequest("GET", "/api/resource", nil)
	rec := httptest.NewRecorder()

	chain.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}
