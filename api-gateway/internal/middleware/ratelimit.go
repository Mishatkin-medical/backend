package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimitMiddleware struct {
	client          *redis.Client
	limitPerUser    int
	limitPerIP      int
}

func NewRateLimitMiddleware(redisURL string, limitPerUser, limitPerIP int) *RateLimitMiddleware {
	client := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	return &RateLimitMiddleware{
		client:       client,
		limitPerUser: limitPerUser,
		limitPerIP:   limitPerIP,
	}
}

func (m *RateLimitMiddleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get identifier (user ID or IP)
			userID := getUserIDFromContext(r.Context())
			ip := getClientIP(r)

			identifier := ip
			if userID != "" {
				identifier = "user:" + userID
			}

			// Check rate limit
			key := "ratelimit:" + identifier
			
			ctx := context.Background()
			count, err := m.client.Incr(ctx, key).Result()
			if err != nil {
				// Redis error, allow request
				next.ServeHTTP(w, r)
				return
			}

			// Set expiry on first request
			if count == 1 {
				m.client.Expire(ctx, key, time.Minute)
			}

			// Check limit
			limit := m.limitPerUser
			if userID == "" {
				limit = m.limitPerIP
			}

			w.Header().Set("X-RateLimit-Limit", string(rune(limit)))
			w.Header().Set("X-RateLimit-Remaining", string(rune(limit-int(count))))

			if count > int64(limit) {
				m.writeError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *RateLimitMiddleware) writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}
	return r.RemoteAddr
}

func getUserIDFromContext(ctx interface{}) string {
	return ""
}
