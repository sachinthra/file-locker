package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/sachinthra/file-locker/backend/internal/storage"
)

// AuditLogger handles logging of admin actions
type AuditLogger struct {
	pg *storage.PostgresStore
}

// NewAuditLogger creates a new audit logger instance
func NewAuditLogger(pg *storage.PostgresStore) *AuditLogger {
	return &AuditLogger{pg: pg}
}

// LogAdminAction logs an admin action to the audit_logs table
func (a *AuditLogger) LogAdminAction(ctx context.Context, actorID, action, targetType, targetID string, metadata map[string]interface{}, ipAddress string) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		log.Printf("[audit] Failed to marshal metadata: %v", err)
		metadataJSON = []byte("{}")
	}

	query := `
		INSERT INTO audit_logs (actor_id, action, target_type, target_id, metadata, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = a.pg.DB().ExecContext(ctx, query, actorID, action, targetType, targetID, metadataJSON, ipAddress)
	if err != nil {
		log.Printf("[audit] Failed to log action: %v", err)
		return err
	}

	log.Printf("[audit] %s by %s on %s/%s", action, actorID, targetType, targetID)
	return nil
}

// GetClientIP extracts IP address from request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (proxy/nginx)
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

// AuditLog represents a single audit log entry
type AuditLog struct {
	ID         string                 `json:"id"`
	ActorID    string                 `json:"actor_id"`
	Action     string                 `json:"action"`
	TargetType string                 `json:"target_type"`
	TargetID   string                 `json:"target_id"`
	Metadata   map[string]interface{} `json:"metadata"`
	IPAddress  string                 `json:"ip_address"`
	CreatedAt  string                 `json:"created_at"`
}
