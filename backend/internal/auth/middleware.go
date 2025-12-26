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
		// 1. Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		// 2. Check format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"Invalid authorization format"}`, http.StatusUnauthorized)
			return
		}

		// 3. Extract token
		tokenString := parts[1]

		// 4. Validate token with jwtService
		claims, err := a.jwtService.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, `{"error":"Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// 5. Check if session exists in Redis
		sessionKey := fmt.Sprintf("session:%s", claims.UserID)
		ctx := context.Background()
		exists, err := a.redisCache.Exists(ctx, sessionKey)
		if err != nil || !exists {
			http.Error(w, `{"error":"Session not found or expired"}`, http.StatusUnauthorized)
			return
		}

		// 6. Set userID in context
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
			rateLimitKey := fmt.Sprintf("ratelimit:%s:%d", userID, currentWindow)

			ctx := context.Background()

			// 3. Increment counter with INCR
			count, err := a.redisCache.Incr(ctx, rateLimitKey)
			if err != nil {
				http.Error(w, `{"error":"Rate limit check failed"}`, http.StatusInternalServerError)
				return
			}

			// 4. Set expiration on first request
			if count == 1 {
				err = a.redisCache.Set(ctx, rateLimitKey, "1", window)
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
