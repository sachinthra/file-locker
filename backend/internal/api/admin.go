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
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	pg          *storage.PostgresStore
	minioStore  *storage.MinIOStorage
	redisCache  *storage.RedisCache
	auditLogger *AuditLogger
}

func NewAdminHandler(pg *storage.PostgresStore, minioStore *storage.MinIOStorage, redisCache *storage.RedisCache) *AdminHandler {
	return &AdminHandler{
		pg:          pg,
		minioStore:  minioStore,
		redisCache:  redisCache,
		auditLogger: NewAuditLogger(pg),
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
	IsActive     bool   `json:"is_active"`
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
			u.is_active,
			u.created_at,
			COALESCE(COUNT(f.id), 0) as file_count,
			COALESCE(SUM(f.size), 0) as total_storage
		FROM users u
		LEFT JOIN files f ON u.id = f.user_id
		GROUP BY u.id, u.username, u.email, u.role, u.is_active, u.created_at
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
			&user.IsActive,
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

	// Log audit action
	adminID := r.Context().Value("userID").(string)
	h.auditLogger.LogAdminAction(ctx, adminID, "USER_DELETED", "user", userID, map[string]interface{}{
		"username":      user.Username,
		"files_deleted": len(files),
	}, GetClientIP(r))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       fmt.Sprintf("User %s deleted successfully", user.Username),
		"files_deleted": len(files),
	})
}

// ================================================================
// ADMIN GOVERNANCE FEATURES
// ================================================================

// HandleUpdateUserStatus toggles user account active/suspended status
func (h *AdminHandler) HandleUpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := chi.URLParam(r, "id")
	adminID := r.Context().Value("userID").(string)

	if userID == "" {
		http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
		return
	}

	// Prevent admin from suspending themselves
	if adminID == userID {
		http.Error(w, `{"error":"Cannot modify your own account status"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Get user info before update
	user, err := h.pg.GetUserByID(ctx, userID)
	if err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	// Update user status
	query := "UPDATE users SET is_active = $1, updated_at = NOW() WHERE id = $2"
	_, err = h.pg.DB().ExecContext(ctx, query, req.IsActive, userID)
	if err != nil {
		log.Printf("[admin] Failed to update user status: %v", err)
		http.Error(w, `{"error":"Failed to update user status"}`, http.StatusInternalServerError)
		return
	}

	// If suspending user, revoke all their sessions
	if !req.IsActive {
		h.redisCache.DeleteUserSessions(ctx, userID)
		log.Printf("[admin] Revoked all sessions for suspended user: %s", userID)
	}

	// Log audit action
	action := "USER_SUSPENDED"
	if req.IsActive {
		action = "USER_ACTIVATED"
	}
	h.auditLogger.LogAdminAction(ctx, adminID, action, "user", userID, map[string]interface{}{
		"username": user.Username,
	}, GetClientIP(r))

	log.Printf("[admin] User %s status changed to active=%v by %s", user.Username, req.IsActive, adminID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "User status updated successfully",
		"is_active": req.IsActive,
	})
}

// HandleUpdateUserRole changes user role (admin/user)
func (h *AdminHandler) HandleUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := chi.URLParam(r, "id")
	adminID := r.Context().Value("userID").(string)

	if userID == "" {
		http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
		return
	}

	// Prevent admin from demoting themselves
	if adminID == userID {
		http.Error(w, `{"error":"Cannot change your own role"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate role
	if req.Role != "admin" && req.Role != "user" {
		http.Error(w, `{"error":"Invalid role. Must be 'admin' or 'user'"}`, http.StatusBadRequest)
		return
	}

	// Get user info before update
	user, err := h.pg.GetUserByID(ctx, userID)
	if err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	oldRole := user.Role

	// Update user role
	query := "UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2"
	_, err = h.pg.DB().ExecContext(ctx, query, req.Role, userID)
	if err != nil {
		log.Printf("[admin] Failed to update user role: %v", err)
		http.Error(w, `{"error":"Failed to update user role"}`, http.StatusInternalServerError)
		return
	}

	// Log audit action
	h.auditLogger.LogAdminAction(ctx, adminID, "ROLE_CHANGED", "user", userID, map[string]interface{}{
		"username": user.Username,
		"old_role": oldRole,
		"new_role": req.Role,
	}, GetClientIP(r))

	log.Printf("[admin] User %s role changed from %s to %s by %s", user.Username, oldRole, req.Role, adminID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User role updated successfully",
		"role":    req.Role,
	})
}

// HandleResetUserPassword allows admin to force reset a user's password
func (h *AdminHandler) HandleResetUserPassword(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := chi.URLParam(r, "id")
	adminID := r.Context().Value("userID").(string)

	if userID == "" {
		http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate password strength
	if len(req.NewPassword) < 8 {
		http.Error(w, `{"error":"Password must be at least 8 characters"}`, http.StatusBadRequest)
		return
	}

	// Get user info
	user, err := h.pg.GetUserByID(ctx, userID)
	if err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	// Hash new password
	hashedPassword, err := hashPassword(req.NewPassword)
	if err != nil {
		log.Printf("[admin] Failed to hash password: %v", err)
		http.Error(w, `{"error":"Failed to process password"}`, http.StatusInternalServerError)
		return
	}

	// Update password
	query := "UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2"
	_, err = h.pg.DB().ExecContext(ctx, query, hashedPassword, userID)
	if err != nil {
		log.Printf("[admin] Failed to update password: %v", err)
		http.Error(w, `{"error":"Failed to reset password"}`, http.StatusInternalServerError)
		return
	}

	// Revoke all user sessions to force re-login
	h.redisCache.DeleteUserSessions(ctx, userID)

	// Log audit action
	h.auditLogger.LogAdminAction(ctx, adminID, "PASSWORD_RESET", "user", userID, map[string]interface{}{
		"username": user.Username,
	}, GetClientIP(r))

	log.Printf("[admin] Password reset for user %s by admin %s", user.Username, adminID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Password reset successfully. User sessions revoked.",
	})
}

// HandleForceLogoutUser revokes all sessions for a specific user
func (h *AdminHandler) HandleForceLogoutUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := chi.URLParam(r, "id")
	adminID := r.Context().Value("userID").(string)

	if userID == "" {
		http.Error(w, `{"error":"User ID required"}`, http.StatusBadRequest)
		return
	}

	// Get user info
	user, err := h.pg.GetUserByID(ctx, userID)
	if err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	// Delete all user sessions from Redis
	count, err := h.redisCache.DeleteUserSessions(ctx, userID)
	if err != nil {
		log.Printf("[admin] Failed to revoke user sessions: %v", err)
		http.Error(w, `{"error":"Failed to logout user"}`, http.StatusInternalServerError)
		return
	}

	// Log audit action
	h.auditLogger.LogAdminAction(ctx, adminID, "FORCE_LOGOUT", "user", userID, map[string]interface{}{
		"username":         user.Username,
		"sessions_revoked": count,
	}, GetClientIP(r))

	log.Printf("[admin] Force logged out user %s (%d sessions revoked) by %s", user.Username, count, adminID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":          "User logged out successfully",
		"sessions_revoked": count,
	})
}

// HandleGetAuditLogs returns paginated audit logs
func (h *AdminHandler) HandleGetAuditLogs(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Pagination parameters
	limit := 50
	offset := 0
	// TODO: Add query parameters for limit/offset/filtering

	query := `
		SELECT 
			al.id,
			al.actor_id,
			al.action,
			al.target_type,
			al.target_id,
			al.metadata,
			al.ip_address,
			al.created_at,
			u.username as actor_username
		FROM audit_logs al
		LEFT JOIN users u ON al.actor_id = u.id
		ORDER BY al.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := h.pg.DB().QueryContext(ctx, query, limit, offset)
	if err != nil {
		log.Printf("[admin] Failed to get audit logs: %v", err)
		http.Error(w, `{"error":"Failed to get audit logs"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type AuditLogEntry struct {
		ID            string         `json:"id"`
		ActorID       string         `json:"actor_id"`
		ActorUsername string         `json:"actor_username"`
		Action        string         `json:"action"`
		TargetType    sql.NullString `json:"target_type"`
		TargetID      sql.NullString `json:"target_id"`
		Metadata      []byte         `json:"metadata"`
		IPAddress     sql.NullString `json:"ip_address"`
		CreatedAt     string         `json:"created_at"`
	}

	var logs []AuditLogEntry
	for rows.Next() {
		var log AuditLogEntry
		var createdAt sql.NullTime

		err := rows.Scan(
			&log.ID,
			&log.ActorID,
			&log.Action,
			&log.TargetType,
			&log.TargetID,
			&log.Metadata,
			&log.IPAddress,
			&createdAt,
			&log.ActorUsername,
		)
		if err != nil {
			log := log
			_ = log
			continue
		}

		if createdAt.Valid {
			log.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		}

		logs = append(logs, log)
	}

	if logs == nil {
		logs = []AuditLogEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
		"limit": limit,
	})
}

// HandleGetAllFiles returns all files in the system (admin view)
func (h *AdminHandler) HandleGetAllFiles(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	query := `
		SELECT 
			f.id,
			f.user_id,
			f.file_name,
			f.size,
			f.mime_type,
			f.created_at,
			u.username
		FROM files f
		LEFT JOIN users u ON f.user_id = u.id
		ORDER BY f.created_at DESC
		LIMIT 100
	`

	rows, err := h.pg.DB().QueryContext(ctx, query)
	if err != nil {
		log.Printf("[admin] Failed to get all files: %v", err)
		http.Error(w, `{"error":"Failed to get files"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type FileEntry struct {
		ID          string         `json:"id"`
		UserID      string         `json:"user_id"`
		Username    sql.NullString `json:"username"`
		Filename    string         `json:"filename"`
		Size        int64          `json:"size"`
		ContentType string         `json:"content_type"`
		CreatedAt   string         `json:"created_at"`
	}

	var files []FileEntry
	for rows.Next() {
		var file FileEntry
		var createdAt sql.NullTime

		err := rows.Scan(
			&file.ID,
			&file.UserID,
			&file.Filename,
			&file.Size,
			&file.ContentType,
			&createdAt,
			&file.Username,
		)
		if err != nil {
			continue
		}

		if createdAt.Valid {
			file.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		}

		files = append(files, file)
	}

	if files == nil {
		files = []FileEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files": files,
		"count": len(files),
	})
}

// HandleDeleteAnyFile allows admin to delete any file (bypass owner check)
func (h *AdminHandler) HandleDeleteAnyFile(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	fileID := chi.URLParam(r, "id")
	adminID := r.Context().Value("userID").(string)

	if fileID == "" {
		http.Error(w, `{"error":"File ID required"}`, http.StatusBadRequest)
		return
	}

	// Get file info
	file, err := h.pg.GetFileMetadata(ctx, fileID)
	if err != nil {
		http.Error(w, `{"error":"File not found"}`, http.StatusNotFound)
		return
	}

	// Delete from MinIO
	err = h.minioStore.DeleteFile(ctx, file.MinIOPath)
	if err != nil {
		log.Printf("[admin] Failed to delete file from MinIO: %v", err)
		// Continue with DB deletion even if MinIO fails
	}

	// Delete from database
	query := "DELETE FROM files WHERE id = $1"
	_, err = h.pg.DB().ExecContext(ctx, query, fileID)
	if err != nil {
		log.Printf("[admin] Failed to delete file from database: %v", err)
		http.Error(w, `{"error":"Failed to delete file"}`, http.StatusInternalServerError)
		return
	}

	// Log audit action
	h.auditLogger.LogAdminAction(ctx, adminID, "FILE_DELETED", "file", fileID, map[string]interface{}{
		"filename": file.FileName,
		"owner_id": file.UserID,
	}, GetClientIP(r))

	log.Printf("[admin] Admin %s deleted file %s (owner: %s)", adminID, file.FileName, file.UserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "File deleted successfully",
	})
}

// ================================================================
// HELPER FUNCTIONS
// ================================================================

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
