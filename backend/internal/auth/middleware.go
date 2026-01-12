package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"log"

	"github.com/sachinthra/file-locker/backend/internal/constants"
	"github.com/sachinthra/file-locker/backend/internal/storage"
)

type AuthMiddleware struct {
	jwtService *JWTService
	redisCache *storage.RedisCache
	pg         *storage.PostgresStore
}

// NewAuthMiddleware creates auth middleware
func NewAuthMiddleware(jwtService *JWTService, redisCache *storage.RedisCache, pg *storage.PostgresStore) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		redisCache: redisCache,
		pg:         pg,
	}
}

// RequireAuth is standard Chi middleware
func (a *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		// 1. Try to get token from Authorization header first
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// Check format: "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// 2. If not in header, try query parameter (for download/stream links)
		if tokenString == "" {
			tokenString = r.URL.Query().Get("token")
		}

		// 3. If still no token, return error
		if tokenString == "" {
			http.Error(w, `{"error":"Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		// If token looks like PAT (starts with fl_), verify against DB
		if strings.HasPrefix(tokenString, "fl_") {
			// delegate verification to PostgresStore helper
			if a.pg == nil {
				log.Printf("[auth] PAT lookup requested but PostgresStore not available from %s", r.RemoteAddr)
				http.Error(w, `{"error":"token lookup not available"}`, http.StatusInternalServerError)
				return
			}
			tokenID, userID, err := a.pg.VerifyPersonalAccessToken(context.Background(), tokenString)
			if err != nil {
				if err == sql.ErrNoRows {
					log.Printf("[auth] PAT verify failed: not found from %s", r.RemoteAddr)
					http.Error(w, `{"error":"Invalid token"}`, http.StatusUnauthorized)
					return
				}
				log.Printf("[auth] PAT verify error from %s: %v", r.RemoteAddr, err)
				http.Error(w, `{"error":"token lookup failed"}`, http.StatusInternalServerError)
				return
			}
			// token verified; set userID in context
			log.Printf("[auth] PAT accepted id=%s user=%s from=%s", tokenID, userID, r.RemoteAddr)
			ctx := context.WithValue(r.Context(), constants.UserIDKey, userID)
			// optionally attach token ID
			ctx = context.WithValue(ctx, constants.PatIDKey, tokenID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// 4. Validate token with jwtService
		claims, err := a.jwtService.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, `{"error":"Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// 5. Check if session exists in Redis (using token as key)
		ctx := context.Background()
		sessionUserID, err := a.redisCache.GetSession(ctx, tokenString)
		if err != nil {
			http.Error(w, `{"error":"Session not found or expired"}`, http.StatusUnauthorized)
			return
		}

		// 6. Verify session userID matches token claims
		if sessionUserID != claims.UserID {
			http.Error(w, `{"error":"Session mismatch"}`, http.StatusUnauthorized)
			return
		}

		// 7. Check if user account is active
		user, err := a.pg.GetUserByID(ctx, claims.UserID)
		if err != nil {
			log.Printf("[auth] Failed to get user for account status check: %v", err)
			http.Error(w, `{"error":"User not found"}`, http.StatusUnauthorized)
			return
		}

		// Check if account is suspended
		if !user.IsActive {
			log.Printf("[auth] Blocked request from suspended user: %s (%s)", user.Username, user.ID)
			http.Error(w, `{"error":"Account suspended. Contact administrator."}`, http.StatusForbidden)
			return
		}

		// 8. Set userID in context
		ctx = context.WithValue(r.Context(), constants.UserIDKey, claims.UserID)

		// 9. Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin middleware ensures the user is authenticated AND has admin role
func (a *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Get userID from context (set by RequireAuth)
		userID := r.Context().Value(constants.UserIDKey)
		if userID == nil {
			http.Error(w, `{"error":"User not authenticated"}`, http.StatusUnauthorized)
			return
		}

		// 2. Fetch user from database to check role
		ctx := context.Background()
		user, err := a.pg.GetUserByID(ctx, userID.(string))
		if err != nil {
			log.Printf("[auth] Failed to get user %s for admin check: %v", userID, err)
			http.Error(w, `{"error":"User not found"}`, http.StatusUnauthorized)
			return
		}

		// 3. Check if user has admin role
		if user.Role != "admin" {
			log.Printf("[auth] Access denied: user %s (role=%s) attempted to access admin endpoint", user.Username, user.Role)
			http.Error(w, `{"error":"Admin access required"}`, http.StatusForbidden)
			return
		}

		// 4. User is admin, proceed
		log.Printf("[auth] Admin access granted to user %s", user.Username)
		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware limits requests per user
func (a *AuthMiddleware) RateLimitMiddleware(requests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Get userID from context (set by RequireAuth)
			userID := r.Context().Value(constants.UserIDKey)
			if userID == nil {
				http.Error(w, `{"error":"User not authenticated"}`, http.StatusUnauthorized)
				return
			}

			// 2. Key: "ratelimit:{userID}:{window}"
			currentWindow := time.Now().Unix() / int64(window.Seconds())

			ctx := context.Background()

			// 3. Increment counter with INCR
			count, err := a.redisCache.IncrRateLimit(ctx, userID.(string), currentWindow)
			if err != nil {
				http.Error(w, `{"error":"Rate limit check failed"}`, http.StatusInternalServerError)
				return
			}

			// 4. Set expiration on first request
			if count == 1 {
				err = a.redisCache.SetRateLimit(ctx, userID.(string), currentWindow, "1", window)
				if err != nil {
					// Log error but don't block request
					fmt.Printf("Failed to set expiration: %v\n", err)
				}
			}

			// 5. If count > limit, return 429 Too Many Requests
			if count > int64(requests) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = fmt.Fprintf(w, `{"error":"Rate limit exceeded","retry_after":%d}`, int(window.Seconds()))
				return
			}

			// 6. Otherwise allow request
			next.ServeHTTP(w, r)
		})
	}
}
