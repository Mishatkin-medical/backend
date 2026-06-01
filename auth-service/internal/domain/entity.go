package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user entity
type User struct {
	ID           uuid.UUID
	Email        *string
	Phone        *string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}

// NewUser creates a new user
func NewUser(email, phone *string, passwordHash string) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New(),
		Email:        email,
		Phone:        phone,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// UserResponse represents user data for API response
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     *string   `json:"email,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt,
	}
}
