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
	ID            string `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	IsActive      bool   `json:"is_active"`
	AccountStatus string `json:"account_status"`
	FileCount     int    `json:"file_count"`
	TotalStorage  int64  `json:"total_storage"`
	CreatedAt     string `json:"created_at"`
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
			u.account_status,
			u.created_at,
			COALESCE(COUNT(f.id), 0) as file_count,
			COALESCE(SUM(f.size), 0) as total_storage
		FROM users u
		LEFT JOIN files f ON u.id = f.user_id
		GROUP BY u.id, u.username, u.email, u.role, u.is_active, u.account_status, u.created_at
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
			&user.AccountStatus,
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
// USER APPROVAL MANAGEMENT
// ================================================================

// HandleGetPendingUsers returns list of users awaiting approval
func (h *AdminHandler) HandleGetPendingUsers(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	query := `
		SELECT 
			u.id,
			u.username,
			u.email,
			u.role,
			u.account_status,
			u.created_at
		FROM users u
		WHERE u.account_status = 'pending'
		ORDER BY u.created_at ASC
	`

	rows, err := h.pg.DB().QueryContext(ctx, query)
	if err != nil {
		log.Printf("[admin] Failed to get pending users: %v", err)
		http.Error(w, `{"error":"Failed to get pending users"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type PendingUser struct {
		ID            string `json:"id"`
		Username      string `json:"username"`
		Email         string `json:"email"`
		Role          string `json:"role"`
		AccountStatus string `json:"account_status"`
		CreatedAt     string `json:"created_at"`
	}

	var users []PendingUser
	for rows.Next() {
		var user PendingUser
		var createdAt sql.NullTime
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.AccountStatus,
			&createdAt,
		)
		if err != nil {
			log.Printf("[admin] Failed to scan pending user: %v", err)
			continue
		}
		if createdAt.Valid {
			user.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		}
		users = append(users, user)
	}

	if users == nil {
		users = []PendingUser{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pending_users": users,
		"count":         len(users),
	})
}

// HandleApproveUser approves a pending user account
func (h *AdminHandler) HandleApproveUser(w http.ResponseWriter, r *http.Request) {
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

	// Check if user is in pending state
	if user.AccountStatus != "pending" {
		http.Error(w, `{"error":"User is not in pending state"}`, http.StatusBadRequest)
		return
	}

	// Approve user
	query := "UPDATE users SET account_status = 'active', is_active = true, updated_at = NOW() WHERE id = $1"
	_, err = h.pg.DB().ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("[admin] Failed to approve user: %v", err)
		http.Error(w, `{"error":"Failed to approve user"}`, http.StatusInternalServerError)
		return
	}

	// Log audit action
	h.auditLogger.LogAdminAction(ctx, adminID, "USER_APPROVED", "user", userID, map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
	}, GetClientIP(r))

	log.Printf("[admin] User %s approved by %s", user.Username, adminID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User approved successfully",
		"user_id": userID,
	})
}

// HandleRejectUser rejects a pending user account
func (h *AdminHandler) HandleRejectUser(w http.ResponseWriter, r *http.Request) {
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

	// Check if user is in pending state
	if user.AccountStatus != "pending" {
		http.Error(w, `{"error":"User is not in pending state"}`, http.StatusBadRequest)
		return
	}

	// Reject user (set to rejected status)
	query := "UPDATE users SET account_status = 'rejected', is_active = false, updated_at = NOW() WHERE id = $1"
	_, err = h.pg.DB().ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("[admin] Failed to reject user: %v", err)
		http.Error(w, `{"error":"Failed to reject user"}`, http.StatusInternalServerError)
		return
	}

	// Log audit action
	h.auditLogger.LogAdminAction(ctx, adminID, "USER_REJECTED", "user", userID, map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
	}, GetClientIP(r))

	log.Printf("[admin] User %s rejected by %s", user.Username, adminID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User rejected successfully",
		"user_id": userID,
	})
}

// ================================================================
// SETTINGS MANAGEMENT
// ================================================================

// HandleGetSettings returns system settings
func (h *AdminHandler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	query := `SELECT key, value, description FROM settings ORDER BY key`
	rows, err := h.pg.DB().QueryContext(ctx, query)
	if err != nil {
		log.Printf("[admin] Failed to get settings: %v", err)
		http.Error(w, `{"error":"Failed to get settings"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Setting struct {
		Key         string `json:"key"`
		Value       string `json:"value"`
		Description string `json:"description"`
	}

	settings := make(map[string]Setting)
	for rows.Next() {
		var s Setting
		if err := rows.Scan(&s.Key, &s.Value, &s.Description); err != nil {
			continue
		}
		settings[s.Key] = s
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"settings": settings,
	})
}

// HandleUpdateSetting updates a system setting
func (h *AdminHandler) HandleUpdateSetting(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	adminID := r.Context().Value("userID").(string)

	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, `{"error":"Setting key required"}`, http.StatusBadRequest)
		return
	}

	// Update setting
	query := "UPDATE settings SET value = $1, updated_at = NOW(), updated_by = $2 WHERE key = $3"
	result, err := h.pg.DB().ExecContext(ctx, query, req.Value, adminID, req.Key)
	if err != nil {
		log.Printf("[admin] Failed to update setting: %v", err)
		http.Error(w, `{"error":"Failed to update setting"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error":"Setting not found"}`, http.StatusNotFound)
		return
	}

	// Log audit action
	h.auditLogger.LogAdminAction(ctx, adminID, "SETTING_UPDATED", "system", "", map[string]interface{}{
		"key":   req.Key,
		"value": req.Value,
	}, GetClientIP(r))

	log.Printf("[admin] Setting %s updated to %s by %s", req.Key, req.Value, adminID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Setting updated successfully",
		"key":     req.Key,
		"value":   req.Value,
	})
}

// ================================================================
// ANNOUNCEMENTS MANAGEMENT
// ================================================================

// HandleGetAnnouncements returns active announcements (optionally filtered by un-dismissed for user)
func (h *AdminHandler) HandleGetAnnouncements(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value("userID")

	// Check if we should filter by un-dismissed for this user
	filterUndismissed := r.URL.Query().Get("undismissed") == "true"

	var query string
	var args []interface{}

	if filterUndismissed && userID != nil {
		// Get announcements not dismissed by this user
		query = `
			SELECT 
				a.id,
				a.title,
				a.message,
				a.type,
				a.target_type,
				a.target_user_ids,
				a.is_active,
				a.expires_at,
				a.created_by,
				a.created_at,
				u.username as creator_username
			FROM announcements a
			LEFT JOIN users u ON a.created_by = u.id
			WHERE a.is_active = true
			  AND (a.expires_at IS NULL OR a.expires_at > NOW())
			  AND a.id NOT IN (
				  SELECT announcement_id 
				  FROM user_announcement_dismissals 
				  WHERE user_id = $1
			  )
			  AND (
				  a.target_type = 'all'
				  OR (a.target_type = 'specific_users' AND $1 = ANY(a.target_user_ids))
			  )
			ORDER BY a.created_at DESC
		`
		args = []interface{}{userID.(string)}
	} else {
		// Get all active announcements (admin view)
		query = `
			SELECT 
				a.id,
				a.title,
				a.message,
				a.type,
				a.target_type,
				a.target_user_ids,
				a.is_active,
				a.expires_at,
				a.created_by,
				a.created_at,
				u.username as creator_username
			FROM announcements a
			LEFT JOIN users u ON a.created_by = u.id
			WHERE a.is_active = true
			  AND (a.expires_at IS NULL OR a.expires_at > NOW())
			ORDER BY a.created_at DESC
		`
	}

	rows, err := h.pg.DB().QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("[admin] Failed to get announcements: %v", err)
		http.Error(w, `{"error":"Failed to get announcements"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Announcement struct {
		ID              string       `json:"id"`
		Title           string       `json:"title"`
		Message         string       `json:"message"`
		Type            string       `json:"type"`
		TargetType      string       `json:"target_type"`
		TargetUserIDs   []string     `json:"target_user_ids,omitempty"`
		IsActive        bool         `json:"is_active"`
		ExpiresAt       sql.NullTime `json:"expires_at,omitempty"`
		CreatedBy       string       `json:"created_by"`
		CreatorUsername string       `json:"creator_username"`
		CreatedAt       string       `json:"created_at"`
	}

	var announcements []Announcement
	for rows.Next() {
		var ann Announcement
		var createdAt sql.NullTime
		var targetUserIDs sql.NullString

		err := rows.Scan(
			&ann.ID,
			&ann.Title,
			&ann.Message,
			&ann.Type,
			&ann.TargetType,
			&targetUserIDs,
			&ann.IsActive,
			&ann.ExpiresAt,
			&ann.CreatedBy,
			&createdAt,
			&ann.CreatorUsername,
		)
		if err != nil {
			log.Printf("[admin] Failed to scan announcement: %v", err)
			continue
		}

		if createdAt.Valid {
			ann.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		}

		// Parse target_user_ids if present (PostgreSQL array as string)
		// This is a simplified parser; production code might need more robust handling
		if targetUserIDs.Valid && targetUserIDs.String != "" {
			// Remove {}, split by comma
			// Note: This is basic; for production, consider using a proper array scanner
		}

		announcements = append(announcements, ann)
	}

	if announcements == nil {
		announcements = []Announcement{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"announcements": announcements,
		"count":         len(announcements),
	})
}

// HandleCreateAnnouncement creates a new announcement
func (h *AdminHandler) HandleCreateAnnouncement(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	adminID := r.Context().Value("userID").(string)

	var req struct {
		Title         string   `json:"title"`
		Message       string   `json:"message"`
		Type          string   `json:"type"`        // 'info', 'warning', 'critical'
		TargetType    string   `json:"target_type"` // 'all', 'specific_users'
		TargetUserIDs []string `json:"target_user_ids,omitempty"`
		ExpiresAt     *string  `json:"expires_at,omitempty"` // ISO 8601 format
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Title == "" || req.Message == "" {
		http.Error(w, `{"error":"Title and message are required"}`, http.StatusBadRequest)
		return
	}

	// Validate type
	if req.Type != "info" && req.Type != "warning" && req.Type != "critical" {
		req.Type = "info" // Default
	}

	// Validate target_type
	if req.TargetType != "all" && req.TargetType != "specific_users" {
		req.TargetType = "all" // Default
	}

	// Build query
	var query string
	var args []interface{}

	if req.ExpiresAt != nil {
		query = `
			INSERT INTO announcements (title, message, type, target_type, target_user_ids, expires_at, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at
		`
		args = []interface{}{req.Title, req.Message, req.Type, req.TargetType, req.TargetUserIDs, *req.ExpiresAt, adminID}
	} else {
		query = `
			INSERT INTO announcements (title, message, type, target_type, target_user_ids, created_by)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, created_at
		`
		args = []interface{}{req.Title, req.Message, req.Type, req.TargetType, req.TargetUserIDs, adminID}
	}

	var announcementID string
	var createdAt sql.NullTime
	err := h.pg.DB().QueryRowContext(ctx, query, args...).Scan(&announcementID, &createdAt)
	if err != nil {
		log.Printf("[admin] Failed to create announcement: %v", err)
		http.Error(w, `{"error":"Failed to create announcement"}`, http.StatusInternalServerError)
		return
	}

	// Log audit action
	h.auditLogger.LogAdminAction(ctx, adminID, "ANNOUNCEMENT_CREATED", "announcement", announcementID, map[string]interface{}{
		"title": req.Title,
		"type":  req.Type,
	}, GetClientIP(r))

	log.Printf("[admin] Announcement created by %s: %s", adminID, req.Title)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Announcement created successfully",
		"id":      announcementID,
	})
}

// HandleDeleteAnnouncement deletes (deactivates) an announcement
func (h *AdminHandler) HandleDeleteAnnouncement(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	announcementID := chi.URLParam(r, "id")
	adminID := r.Context().Value("userID").(string)

	if announcementID == "" {
		http.Error(w, `{"error":"Announcement ID required"}`, http.StatusBadRequest)
		return
	}

	// Deactivate announcement
	query := "UPDATE announcements SET is_active = false, updated_at = NOW() WHERE id = $1"
	result, err := h.pg.DB().ExecContext(ctx, query, announcementID)
	if err != nil {
		log.Printf("[admin] Failed to delete announcement: %v", err)
		http.Error(w, `{"error":"Failed to delete announcement"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error":"Announcement not found"}`, http.StatusNotFound)
		return
	}

	// Log audit action
	h.auditLogger.LogAdminAction(ctx, adminID, "ANNOUNCEMENT_DELETED", "announcement", announcementID, nil, GetClientIP(r))

	log.Printf("[admin] Announcement %s deleted by %s", announcementID, adminID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Announcement deleted successfully",
	})
}

// HandleDismissAnnouncement allows a user to dismiss an announcement
func (h *AdminHandler) HandleDismissAnnouncement(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	announcementID := chi.URLParam(r, "id")
	userID := r.Context().Value("userID").(string)

	if announcementID == "" {
		http.Error(w, `{"error":"Announcement ID required"}`, http.StatusBadRequest)
		return
	}

	// Insert dismissal record
	query := `
		INSERT INTO user_announcement_dismissals (user_id, announcement_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, announcement_id) DO NOTHING
	`
	_, err := h.pg.DB().ExecContext(ctx, query, userID, announcementID)
	if err != nil {
		log.Printf("[user] Failed to dismiss announcement: %v", err)
		http.Error(w, `{"error":"Failed to dismiss announcement"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("[user] User %s dismissed announcement %s", userID, announcementID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Announcement dismissed successfully",
	})
}

// ================================================================
// STORAGE CLEANUP
// ================================================================

// HandleAnalyzeStorage analyzes storage for orphaned files
func (h *AdminHandler) HandleAnalyzeStorage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	adminID := r.Context().Value("userID").(string)

	log.Printf("[admin] Starting storage analysis by admin %s", adminID)

	// Get all files from database
	dbQuery := `SELECT id, minio_path FROM files`
	dbRows, err := h.pg.DB().QueryContext(ctx, dbQuery)
	if err != nil {
		log.Printf("[admin] Failed to query database files: %v", err)
		http.Error(w, `{"error":"Failed to query database"}`, http.StatusInternalServerError)
		return
	}
	defer dbRows.Close()

	dbFiles := make(map[string]string) // minio_path -> file_id
	for dbRows.Next() {
		var fileID, minioPath string
		if err := dbRows.Scan(&fileID, &minioPath); err != nil {
			continue
		}
		dbFiles[minioPath] = fileID
	}

	// Get all objects from MinIO bucket
	minioObjects, err := h.minioStore.ListAllObjects(ctx)
	if err != nil {
		log.Printf("[admin] Failed to list MinIO objects: %v", err)
		http.Error(w, `{"error":"Failed to list MinIO objects"}`, http.StatusInternalServerError)
		return
	}

	// Find orphaned files (in MinIO but not in DB)
	type OrphanedFile struct {
		Path string `json:"path"`
		Size int64  `json:"size"`
	}

	var orphanedFiles []OrphanedFile
	var orphanedTotalSize int64

	for _, obj := range minioObjects {
		if _, exists := dbFiles[obj.Key]; !exists {
			orphanedFiles = append(orphanedFiles, OrphanedFile{
				Path: obj.Key,
				Size: obj.Size,
			})
			orphanedTotalSize += obj.Size
		}
	}

	// Find ghost records (in DB but not in MinIO)
	type GhostRecord struct {
		FileID string `json:"file_id"`
		Path   string `json:"path"`
	}

	minioSet := make(map[string]bool)
	for _, obj := range minioObjects {
		minioSet[obj.Key] = true
	}

	var ghostRecords []GhostRecord
	for path, fileID := range dbFiles {
		if !minioSet[path] {
			ghostRecords = append(ghostRecords, GhostRecord{
				FileID: fileID,
				Path:   path,
			})
		}
	}

	// Log audit action
	h.auditLogger.LogAdminAction(ctx, adminID, "STORAGE_ANALYZED", "system", "", map[string]interface{}{
		"total_db_files":      len(dbFiles),
		"total_minio_objects": len(minioObjects),
		"orphaned_files":      len(orphanedFiles),
		"ghost_records":       len(ghostRecords),
	}, GetClientIP(r))

	log.Printf("[admin] Storage analysis complete: %d orphaned files, %d ghost records",
		len(orphanedFiles), len(ghostRecords))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_db_files":      len(dbFiles),
		"total_minio_objects": len(minioObjects),
		"orphaned_files":      orphanedFiles,
		"orphaned_count":      len(orphanedFiles),
		"orphaned_total_size": orphanedTotalSize,
		"ghost_records":       ghostRecords,
		"ghost_count":         len(ghostRecords),
	})
}

// HandleCleanupStorage cleans up orphaned files from MinIO
func (h *AdminHandler) HandleCleanupStorage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	adminID := r.Context().Value("userID").(string)

	var req struct {
		OrphanedPaths []string `json:"orphaned_paths"`
		GhostFileIDs  []string `json:"ghost_file_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	log.Printf("[admin] Starting storage cleanup by admin %s: %d orphaned, %d ghosts",
		adminID, len(req.OrphanedPaths), len(req.GhostFileIDs))

	// Delete orphaned files from MinIO
	deletedOrphaned := 0
	var orphanedErrors []string
	for _, path := range req.OrphanedPaths {
		if err := h.minioStore.DeleteFile(ctx, path); err != nil {
			log.Printf("[admin] Failed to delete orphaned file %s: %v", path, err)
			orphanedErrors = append(orphanedErrors, path)
		} else {
			deletedOrphaned++
		}
	}

	// Delete ghost records from database
	deletedGhosts := 0
	var ghostErrors []string
	for _, fileID := range req.GhostFileIDs {
		query := "DELETE FROM files WHERE id = $1"
		if _, err := h.pg.DB().ExecContext(ctx, query, fileID); err != nil {
			log.Printf("[admin] Failed to delete ghost record %s: %v", fileID, err)
			ghostErrors = append(ghostErrors, fileID)
		} else {
			deletedGhosts++
		}
	}

	// Log audit action
	h.auditLogger.LogAdminAction(ctx, adminID, "STORAGE_CLEANED", "system", "", map[string]interface{}{
		"deleted_orphaned": deletedOrphaned,
		"deleted_ghosts":   deletedGhosts,
		"orphaned_errors":  len(orphanedErrors),
		"ghost_errors":     len(ghostErrors),
	}, GetClientIP(r))

	log.Printf("[admin] Storage cleanup complete: %d orphaned deleted, %d ghosts deleted",
		deletedOrphaned, deletedGhosts)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":          "Storage cleanup completed",
		"deleted_orphaned": deletedOrphaned,
		"deleted_ghosts":   deletedGhosts,
		"orphaned_errors":  orphanedErrors,
		"ghost_errors":     ghostErrors,
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
