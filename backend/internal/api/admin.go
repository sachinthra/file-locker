package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sachinthra/file-locker/backend/internal/storage"
)

type AdminHandler struct {
	pg         *storage.PostgresStore
	minioStore *storage.MinIOStorage
}

func NewAdminHandler(pg *storage.PostgresStore, minioStore *storage.MinIOStorage) *AdminHandler {
	return &AdminHandler{
		pg:         pg,
		minioStore: minioStore,
	}
}

// Stats represents system statistics
type Stats struct {
	TotalUsers        int   `json:"total_users"`
	TotalFiles        int   `json:"total_files"`
	TotalStorageBytes int64 `json:"total_storage_bytes"`
	ActiveUsers24h    int   `json:"active_users_24h"`
}

// UserInfo represents user information for admin panel
type UserInfo struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	FileCount    int    `json:"file_count"`
	TotalStorage int64  `json:"total_storage"`
	CreatedAt    string `json:"created_at"`
}

// HandleGetStats returns system statistics
func (h *AdminHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get total users
	var totalUsers int
	err := h.pg.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		log.Printf("[admin] Failed to get total users: %v", err)
		http.Error(w, `{"error":"Failed to get statistics"}`, http.StatusInternalServerError)
		return
	}

	// Get total files
	var totalFiles int
	err = h.pg.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM files").Scan(&totalFiles)
	if err != nil {
		log.Printf("[admin] Failed to get total files: %v", err)
		http.Error(w, `{"error":"Failed to get statistics"}`, http.StatusInternalServerError)
		return
	}

	// Get total storage
	var totalStorage sql.NullInt64
	err = h.pg.DB().QueryRowContext(ctx, "SELECT SUM(size) FROM files").Scan(&totalStorage)
	if err != nil {
		log.Printf("[admin] Failed to get total storage: %v", err)
		http.Error(w, `{"error":"Failed to get statistics"}`, http.StatusInternalServerError)
		return
	}

	// Get active users in last 24 hours (based on file uploads or downloads)
	var activeUsers int
	query := `
		SELECT COUNT(DISTINCT user_id) 
		FROM files 
		WHERE created_at > NOW() - INTERVAL '24 hours'
	`
	err = h.pg.DB().QueryRowContext(ctx, query).Scan(&activeUsers)
	if err != nil {
		log.Printf("[admin] Failed to get active users: %v", err)
		http.Error(w, `{"error":"Failed to get statistics"}`, http.StatusInternalServerError)
		return
	}

	stats := Stats{
		TotalUsers:        totalUsers,
		TotalFiles:        totalFiles,
		TotalStorageBytes: totalStorage.Int64,
		ActiveUsers24h:    activeUsers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleGetUsers returns list of all users with their statistics
func (h *AdminHandler) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	query := `
		SELECT 
			u.id,
			u.username,
			u.email,
			u.role,
			u.created_at,
			COALESCE(COUNT(f.id), 0) as file_count,
			COALESCE(SUM(f.size), 0) as total_storage
		FROM users u
		LEFT JOIN files f ON u.id = f.user_id
		GROUP BY u.id, u.username, u.email, u.role, u.created_at
		ORDER BY u.created_at DESC
	`

	rows, err := h.pg.DB().QueryContext(ctx, query)
	if err != nil {
		log.Printf("[admin] Failed to get users: %v", err)
		http.Error(w, `{"error":"Failed to get users"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []UserInfo
	for rows.Next() {
		var user UserInfo
		var createdAt sql.NullTime
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
			&createdAt,
			&user.FileCount,
			&user.TotalStorage,
		)
		if err != nil {
			log.Printf("[admin] Failed to scan user: %v", err)
			continue
		}
		if createdAt.Valid {
			user.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		}
		users = append(users, user)
	}

	if users == nil {
		users = []UserInfo{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
	})
}

// HandleDeleteUser deletes a user and all their files
func (h *AdminHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := chi.URLParam(r, "id")

	if userID == "" {
		http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
		return
	}

	// Get the requesting admin's user ID
	adminUserID := r.Context().Value("userID")
	if adminUserID == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Prevent admin from deleting themselves
	if adminUserID.(string) == userID {
		http.Error(w, `{"error":"Cannot delete your own account"}`, http.StatusBadRequest)
		return
	}

	// Check if user exists
	user, err := h.pg.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("[admin] User not found: %v", err)
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	// Get all files for this user
	files, err := h.pg.ListUserFiles(ctx, userID)
	if err != nil {
		log.Printf("[admin] Failed to list user files: %v", err)
		http.Error(w, `{"error":"Failed to delete user files"}`, http.StatusInternalServerError)
		return
	}

	// Delete all files from MinIO
	deletedCount := 0
	for _, file := range files {
		err := h.minioStore.DeleteFile(ctx, file.MinIOPath)
		if err != nil {
			log.Printf("[admin] Failed to delete file from MinIO: %s, error: %v", file.MinIOPath, err)
			// Continue deleting other files even if one fails
		} else {
			deletedCount++
		}
	}

	log.Printf("[admin] Deleted %d/%d files from MinIO for user %s", deletedCount, len(files), userID)

	// Delete user from database (CASCADE will delete files table entries)
	query := "DELETE FROM users WHERE id = $1"
	_, err = h.pg.DB().ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("[admin] Failed to delete user from database: %v", err)
		http.Error(w, `{"error":"Failed to delete user"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("[admin] Successfully deleted user %s (%s) with %d files", user.Username, userID, len(files))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       fmt.Sprintf("User %s deleted successfully", user.Username),
		"files_deleted": len(files),
	})
}
