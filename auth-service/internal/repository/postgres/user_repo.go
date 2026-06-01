package postgres

import (
	"context"
	"errors"
	"fmt"

	"mishatkin/auth-service/internal/domain"
	"mishatkin/shared/pkg/postgres"

	"github.com/google/uuid"
)

// UserRepository implements domain.UserRepository
type UserRepository struct {
	pool *postgres.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(pool *postgres.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create inserts a new user
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, phone, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

// GetByID retrieves user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, phone, password_hash, created_at, updated_at
		FROM users WHERE id = $1
	`
	var user domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, domain.ErrUserNotFound
		}
		return nil, domain.ErrUserNotFound
	}
	return &user, nil
}

// GetByEmail retrieves user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, phone, password_hash, created_at, updated_at
		FROM users WHERE email = $1
	`
	var user domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return &user, nil
}

// GetByPhone retrieves user by phone
func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
		SELECT id, email, phone, password_hash, created_at, updated_at
		FROM users WHERE phone = $1
	`
	var user domain.User
	err := r.pool.QueryRow(ctx, query, phone).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return &user, nil
}

// EmailExists checks if email exists
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, email).Scan(&exists)
	return exists, err
}

// PhoneExists checks if phone exists
func (r *UserRepository) PhoneExists(ctx context.Context, phone string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, phone).Scan(&exists)
	return exists, err
}

// TokenRepository implements domain.TokenRepository
type TokenRepository struct {
	pool *postgres.Pool
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(pool *postgres.Pool) *TokenRepository {
	return &TokenRepository{pool: pool}
}

// Create inserts a new token
func (r *TokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.Revoked,
		token.CreatedAt,
	)
	return err
}

// GetByHash retrieves token by hash
func (r *TokenRepository) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked, created_at
		FROM refresh_tokens WHERE token_hash = $1
	`
	var token domain.RefreshToken
	err := r.pool.QueryRow(ctx, query, hash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.Revoked,
		&token.CreatedAt,
	)
	if err != nil {
		return nil, domain.ErrTokenNotFound
	}
	return &token, nil
}

// Revoke marks token as revoked
func (r *TokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// RevokeAllUserTokens revokes all tokens for a user
func (r *TokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

