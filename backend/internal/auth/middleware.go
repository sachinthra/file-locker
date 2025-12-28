package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sachinthra/file-locker/backend/internal/storage"
)

type AuthMiddleware struct {
	jwtService *JWTService
	redisCache *storage.RedisCache
}

// NewAuthMiddleware creates auth middleware
func NewAuthMiddleware(jwtService *JWTService, redisCache *storage.RedisCache) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		redisCache: redisCache,
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

		// 7. Set userID in context
		ctx = context.WithValue(r.Context(), "userID", claims.UserID)

		// 7. Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RateLimitMiddleware limits requests per user
func (a *AuthMiddleware) RateLimitMiddleware(requests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Get userID from context (set by RequireAuth)
			userID := r.Context().Value("userID")
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
				fmt.Fprintf(w, `{"error":"Rate limit exceeded","retry_after":%d}`, int(window.Seconds()))
				return
			}

			// 6. Otherwise allow request
			next.ServeHTTP(w, r)
		})
	}
}
