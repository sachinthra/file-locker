package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/sachinthra/file-locker/backend/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	pgStore *storage.PostgresStore
}

func NewUserHandler(pgStore *storage.PostgresStore) *UserHandler {
	return &UserHandler{
		pgStore: pgStore,
	}
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePasswordResponse represents password change response
type ChangePasswordResponse struct {
	Message string `json:"message"`
}

// HandleChangePassword changes user's password
func (h *UserHandler) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse request
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.CurrentPassword == "" {
		respondError(w, http.StatusBadRequest, "Current password is required")
		return
	}
	if req.NewPassword == "" {
		respondError(w, http.StatusBadRequest, "New password is required")
		return
	}
	if len(req.NewPassword) < 6 {
		respondError(w, http.StatusBadRequest, "New password must be at least 6 characters")
		return
	}
	if req.CurrentPassword == req.NewPassword {
		respondError(w, http.StatusBadRequest, "New password must be different from current password")
		return
	}

	log.Printf("[DEBUG] Password change requested for user: %s", userID)

	// Get user from database
	user, err := h.pgStore.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get user: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		log.Printf("[DEBUG] Current password verification failed for user: %s", userID)
		respondError(w, http.StatusUnauthorized, "Current password is incorrect")
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] Failed to hash new password: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	// Update password in database
	if err := h.pgStore.UpdateUserPassword(r.Context(), userID, string(hashedPassword)); err != nil {
		log.Printf("[ERROR] Failed to update password in database: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	log.Printf("[INFO] Password successfully changed for user: %s", userID)

	// Return success response
	respondJSON(w, http.StatusOK, ChangePasswordResponse{
		Message: "Password changed successfully",
	})
}
