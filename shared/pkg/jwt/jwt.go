package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents JWT claims
type Claims struct {
	UserID uuid.UUID `json:"sub"`
	Email  string    `json:"email,omitempty"`
	jwt.RegisteredClaims
}

// Tokens represents access + refresh tokens
type Tokens struct {
	AccessToken  string
	RefreshToken string
	AccessTTL    time.Duration
}

// IssueInput contains data for token generation
type IssueInput struct {
	UserID    uuid.UUID
	Email     string
	AccessTTL time.Duration
}

// New creates a new logger instance
func New(serviceName string) *JWT {
	return &JWT{
		secret: []byte(os.Getenv("JWT_SECRET")),
	}
}

// JWT handles token operations
type JWT struct {
	secret []byte
}

// NewWithSecret creates JWT with specific secret
func NewWithSecret(secret string) *JWT {
	return &JWT{
		secret: []byte(secret),
	}
}

// SetSecret sets the JWT secret
func (j *JWT) SetSecret(secret string) {
	j.secret = []byte(secret)
}

// IssueAccess creates an access token
func (j *JWT) IssueAccess(userID uuid.UUID, email string, ttl time.Duration) (string, error) {
	if ttl == 0 {
		ttl = 15 * time.Minute
	}

	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mishatkin",
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// IssueRefresh creates a refresh token (64 bytes random)
func (j *JWT) IssueRefresh() (string, error) {
	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashRefresh hashes a refresh token for storage
func HashRefresh(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}

// ValidateAccess validates an access token
func (j *JWT) ValidateAccess(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// ValidateRefresh validates a refresh token hash
func (j *JWT) ValidateRefresh(hashedToken, providedToken string) bool {
	return HashRefresh(providedToken) == hashedToken
}

// GetUserID extracts user ID from claims
func (c *Claims) GetUserID() uuid.UUID {
	return c.UserID
}

