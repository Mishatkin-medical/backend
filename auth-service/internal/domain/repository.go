package domain

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository defines user data operations
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	PhoneExists(ctx context.Context, phone string) (bool, error)
}

// TokenRepository defines token operations
type TokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
}

// TokenCache defines token caching operations
type TokenCache interface {
	Set(ctx context.Context, key string, value interface{}, expirySeconds int) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, seconds int) error
}
