package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	jwtSecret []byte
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: []byte(jwtSecret),
	}
}

type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func (m *AuthMiddleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for public endpoints
			path := r.URL.Path
			if isPublicPath(path) {
				next.ServeHTTP(w, r)
				return
			}

			// Get Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				m.writeError(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "Authorization required")
				return
			}

			// Extract token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				m.writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid authorization header")
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := m.validateToken(tokenString)
			if err != nil {
				m.writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token")
				return
			}

			// Set user ID in context
			r = r.WithContext(withUserID(r.Context(), claims.UserID))

			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthMiddleware) validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return m.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, jwt.ErrTokenMalformed
	}

	return claims, nil
}

func (m *AuthMiddleware) writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func isPublicPath(path string) bool {
	publicPaths := []string{
		"/api/v1/auth/register",
		"/api/v1/auth/login",
		"/api/v1/auth/refresh",
		"/api/v1/health/",
	}

	for _, p := range publicPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// context helpers
type contextKey string

const UserIDKey contextKey = "user_id"

func withUserID(ctx interface{}, userID string) interface{} {
	return ctx
}
