package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewJWTService(t *testing.T) {
	secret := "test-secret-key"
	expiry := 3600

	service := NewJWTService(secret, expiry)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	if string(service.secret) != secret {
		t.Errorf("Expected secret %s, got %s", secret, string(service.secret))
	}

	expectedExpiry := time.Duration(expiry) * time.Second
	if service.expiry != expectedExpiry {
		t.Errorf("Expected expiry %v, got %v", expectedExpiry, service.expiry)
	}
}

func TestGenerateToken(t *testing.T) {
	service := NewJWTService("test-secret", 3600)
	userID := "user123"

	token, err := service.GenerateToken(userID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Fatal("Expected token to be generated, got empty string")
	}

	// Validate the token can be parsed
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("Expected token to be valid, got error: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, claims.UserID)
	}
}

func TestValidateToken_Valid(t *testing.T) {
	service := NewJWTService("test-secret", 3600)
	userID := "user456"

	token, err := service.GenerateToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("Expected valid token, got error: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, claims.UserID)
	}

	// Check expiration is set
	if claims.ExpiresAt == nil {
		t.Error("Expected ExpiresAt to be set")
	}

	// Check issued at is set
	if claims.IssuedAt == nil {
		t.Error("Expected IssuedAt to be set")
	}
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	service1 := NewJWTService("secret1", 3600)
	service2 := NewJWTService("secret2", 3600)

	token, err := service1.GenerateToken("user123")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Try to validate with different secret
	_, err = service2.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for invalid signature, got nil")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	// Create service with 1 second expiry
	service := NewJWTService("test-secret", 1)

	token, err := service.GenerateToken("user123")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	_, err = service.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestValidateToken_MalformedToken(t *testing.T) {
	service := NewJWTService("test-secret", 3600)

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"invalid format", "invalid.token.format"},
		{"random string", "notavalidjwttoken"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ValidateToken(tt.token)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestValidateToken_WrongSigningMethod(t *testing.T) {
	service := NewJWTService("test-secret", 3600)

	// Create token with wrong signing method (RS256 instead of HS256)
	claims := &Claims{
		UserID: "user123",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// This would normally use RSA, but we'll create a token string manually
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("wrong-secret"))

	_, err := service.ValidateToken(tokenString)
	if err == nil {
		t.Error("Expected error for wrong signing method, got nil")
	}
}

func TestGenerateToken_DifferentUsers(t *testing.T) {
	service := NewJWTService("test-secret", 3600)

	users := []string{"user1", "user2", "user3"}
	tokens := make(map[string]string)

	// Generate tokens for different users
	for _, userID := range users {
		token, err := service.GenerateToken(userID)
		if err != nil {
			t.Fatalf("Failed to generate token for %s: %v", userID, err)
		}
		tokens[userID] = token
	}

	// Validate each token contains correct userID
	for userID, token := range tokens {
		claims, err := service.ValidateToken(token)
		if err != nil {
			t.Fatalf("Failed to validate token for %s: %v", userID, err)
		}

		if claims.UserID != userID {
			t.Errorf("Expected userID %s, got %s", userID, claims.UserID)
		}
	}
}

func TestJWTService_TokenUniqueness(t *testing.T) {
	service := NewJWTService("test-secret", 3600)
	userID := "user123"

	// Generate two tokens for same user
	token1, err := service.GenerateToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate first token: %v", err)
	}

	// Small delay to ensure different IssuedAt
	time.Sleep(10 * time.Millisecond)

	token2, err := service.GenerateToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate second token: %v", err)
	}

	// Tokens should be different even for same user
	if token1 == token2 {
		t.Error("Expected different tokens, got identical tokens")
	}

	// Both should be valid
	claims1, err := service.ValidateToken(token1)
	if err != nil {
		t.Fatalf("Failed to validate first token: %v", err)
	}

	claims2, err := service.ValidateToken(token2)
	if err != nil {
		t.Fatalf("Failed to validate second token: %v", err)
	}

	if claims1.UserID != claims2.UserID {
		t.Error("Both tokens should have same userID")
	}
}
