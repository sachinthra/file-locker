package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sachinthra/file-locker/backend/internal/auth"
	"github.com/sachinthra/file-locker/backend/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	jwtService *auth.JWTService
	redisCache *storage.RedisCache
}

func NewAuthHandler(jwtService *auth.JWTService, redisCache *storage.RedisCache) *AuthHandler {
	return &AuthHandler{
		jwtService: jwtService,
		redisCache: redisCache,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type AuthResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
	Email  string `json:"email,omitempty"`
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Username and password required")
		return
	}

	// Get user from Redis
	userData, err := h.redisCache.GetUser(r.Context(), req.Username)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(userData.PasswordHash), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(userData.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Save session in Redis (24 hour expiry)
	if err := h.redisCache.SaveSession(r.Context(), token, userData.UserID, 24*time.Hour); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		Token:  token,
		UserID: userData.UserID,
		Email:  userData.Email,
	})
}

func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Username and password required")
		return
	}

	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	// Check if user already exists
	exists, err := h.redisCache.UserExists(r.Context(), req.Username)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to check user existence")
		return
	}
	if exists {
		respondError(w, http.StatusConflict, "Username already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Generate user ID
	userID := uuid.New().String()

	// Save user to Redis (format: "userID:hashedPassword:email")
	if err := h.redisCache.SaveUser(r.Context(), req.Username, userID, string(hashedPassword), req.Email, 0); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	log.Printf("Token %v", token)

	// Save session
	if err := h.redisCache.SaveSession(r.Context(), token, userID, 24*time.Hour); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	respondJSON(w, http.StatusCreated, AuthResponse{
		Token:  token,
		UserID: userID,
		Email:  req.Email,
	})
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondError(w, http.StatusUnauthorized, "Authorization header required")
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		respondError(w, http.StatusUnauthorized, "Invalid authorization format")
		return
	}

	token := parts[1]

	// Validate token to get userID
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Delete session from Redis
	if err := h.redisCache.DeleteSession(r.Context(), token); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete session")
		return
	}

	// Keep claims for potential logging
	_ = claims

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}
