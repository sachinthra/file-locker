package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachinthra/file-locker/backend/internal/constants"
	"golang.org/x/crypto/bcrypt"

	"github.com/sachinthra/file-locker/backend/internal/storage"
)

type TokensHandler struct {
	DB *sql.DB
}

func NewTokensHandler(pg *storage.PostgresStore) *TokensHandler {
	return &TokensHandler{DB: pg.DB()}
}

type createTokenReq struct {
	Name          string `json:"name"`
	ExpiresInDays int    `json:"expires_in_days"`
}

// POST /api/auth/tokens
func (h *TokensHandler) HandleCreateToken(w http.ResponseWriter, r *http.Request) {
	uid, _ := r.Context().Value(constants.UserIDKey).(string)
	log.Printf("[tokens] %s %s CreateToken request by user=%s from=%s", r.Method, r.URL.Path, uid, r.RemoteAddr)
	var req createTokenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[tokens] CreateToken decode error: %v", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	// generate raw token: fl_ + 32 chars random
	rawUUID := strings.ReplaceAll(uuid.New().String(), "-", "")
	raw := "fl_" + rawUUID[:32]
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "failed generate token", http.StatusInternalServerError)
		return
	}

	var expiresAt *time.Time
	if req.ExpiresInDays > 0 {
		t := time.Now().Add(time.Duration(req.ExpiresInDays) * 24 * time.Hour).UTC()
		expiresAt = &t
	}

	id := uuid.New().String()
	createdAt := time.Now().UTC()
	_, err = h.DB.Exec(`INSERT INTO personal_access_tokens (id, user_id, name, token_hash, created_at, expires_at) VALUES ($1,$2,$3,$4,$5,$6)`, id, uid, req.Name, string(hashed), createdAt, expiresAt)
	if err != nil {
		log.Printf("[tokens] DB insert error for user=%s: %v", uid, err)
		http.Error(w, "failed save token", http.StatusInternalServerError)
		return
	}

	res := map[string]interface{}{
		"id":         id,
		"name":       req.Name,
		"created_at": createdAt,
		"expires_at": expiresAt,
		"token":      raw,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// GET /api/auth/tokens
func (h *TokensHandler) HandleListTokens(w http.ResponseWriter, r *http.Request) {
	uid, _ := r.Context().Value(constants.UserIDKey).(string)
	log.Printf("[tokens] %s %s ListTokens request by user=%s from=%s", r.Method, r.URL.Path, uid, r.RemoteAddr)
	rows, err := h.DB.Query(`SELECT id, name, created_at, last_used_at, expires_at FROM personal_access_tokens WHERE user_id = $1 ORDER BY created_at DESC`, uid)
	if err != nil {
		log.Printf("[tokens] DB list error for user=%s: %v", uid, err)
		http.Error(w, "failed list tokens", http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()
	out := []map[string]interface{}{}
	for rows.Next() {
		var id, name string
		var created time.Time
		var lastUsed sql.NullTime
		var expires sql.NullTime
		if err := rows.Scan(&id, &name, &created, &lastUsed, &expires); err != nil {
			continue
		}
		rec := map[string]interface{}{"id": id, "name": name, "created_at": created}
		if lastUsed.Valid {
			rec["last_used_at"] = lastUsed.Time
		} else {
			rec["last_used_at"] = nil
		}
		if expires.Valid {
			rec["expires_at"] = expires.Time
		} else {
			rec["expires_at"] = nil
		}
		out = append(out, rec)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"tokens": out})
}

// DELETE /api/auth/tokens/{id}
func (h *TokensHandler) HandleRevokeToken(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	uid, _ := r.Context().Value(constants.UserIDKey).(string)
	log.Printf("[tokens] %s %s RevokeToken request id=%s by user=%s from=%s", r.Method, r.URL.Path, id, uid, r.RemoteAddr)
	res, err := h.DB.Exec(`DELETE FROM personal_access_tokens WHERE id = $1 AND user_id = $2`, id, uid)
	if err != nil {
		log.Printf("[tokens] DB delete error for id=%s user=%s: %v", id, uid, err)
		http.Error(w, "failed revoke", http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
