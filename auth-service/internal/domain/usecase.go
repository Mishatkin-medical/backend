package usecase

import (
	"context"
	"errors"
	"time"

	"mishatkin/auth-service/internal/domain"
	"mishatkin/shared/pkg/errors"
	"mishatkin/shared/pkg/jwt"
	"mishatkin/shared/pkg/validator"

	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase handles authentication operations
type AuthUseCase struct {
	userRepo   domain.UserRepository
	tokenRepo domain.TokenRepository
	tokenCache domain.TokenCache
	jwtService *jwt.JWT
	accessTTL  time.Duration
	refreshTTL time.Duration
	bcryptCost int
}

// NewAuthUseCase creates a new auth use case
func NewAuthUseCase(
	userRepo domain.UserRepository,
	tokenRepo domain.TokenRepository,
	tokenCache domain.TokenCache,
	jwtSecret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	bcryptCost int,
) *AuthUseCase {
	jwtSvc := jwt.NewWithSecret(jwtSecret)
	return &AuthUseCase{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		tokenCache: tokenCache,
		jwtService: jwtSvc,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		bcryptCost: bcryptCost,
	}
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Email    string `validate:"omitempty,email"`
	Phone    string `validate:"omitempty,e164"`
	Password string `validate:"required,password"`
}

// RegisterResponse represents registration response
type RegisterResponse struct {
	User        *domain.UserResponse
	AccessToken string
	ExpiresIn   int64
}

// Register creates a new user account
func (uc *AuthUseCase) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// Validate request
	if err := validator.Validate(req); err != nil {
		return nil, errors.Wrap(err, errors.CodeValidation, "Validation failed")
	}

	// Check email
	if req.Email != "" {
		exists, err := uc.userRepo.EmailExists(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, domain.ErrEmailExists
		}
	}

	// Check phone
	if req.Phone != "" {
		exists, err := uc.userRepo.PhoneExists(ctx, req.Phone)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, domain.ErrPhoneExists
		}
	}

	// Check at least one identifier
	if req.Email == "" && req.Phone == "" {
		return nil, errors.New(errors.CodeValidation, "Email or phone required", 400)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), uc.bcryptCost)
	if err != nil {
		return nil, err
	}

	// Create user
	var email, phone *string
	if req.Email != "" {
		email = &req.Email
	}
	if req.Phone != "" {
		phone = &req.Phone
	}

	user := domain.NewUser(email, phone, string(hashedPassword))
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate tokens
	return uc.generateTokens(ctx, user)
}

// LoginRequest represents login request
type LoginRequest struct {
	Identifier string `validate:"required"`
	Password  string `validate:"required"`
}

// Login authenticates a user
func (uc *AuthUseCase) Login(ctx context.Context, req *LoginRequest) (*RegisterResponse, error) {
	// Validate request
	if err := validator.Validate(req); err != nil {
		return nil, errors.Wrap(err, errors.CodeValidation, "Validation failed")
	}

	// Find user by email or phone
	var user *domain.User
	var err error

	if isEmail(req.Identifier) {
		user, err = uc.userRepo.GetByEmail(ctx, req.Identifier)
	} else {
		user, err = uc.userRepo.GetByPhone(ctx, req.Identifier)
	}

	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Generate tokens
	return uc.generateTokens(ctx, user)
}

// RefreshRequest represents token refresh request
type RefreshRequest struct {
	RefreshToken string
}

// RefreshTokens refreshes access token
func (uc *AuthUseCase) RefreshTokens(ctx context.Context, req *RefreshRequest) (*RegisterResponse, error) {
	if req.RefreshToken == "" {
		return nil, domain.ErrTokenNotFound
	}

	// Hash token for lookup
	tokenHash := jwt.HashRefresh(req.RefreshToken)

	// Get token from DB
	token, err := uc.tokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrTokenNotFound) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, err
	}

	// Check if revoked
	if token.Revoked {
		return nil, domain.ErrTokenRevoked
	}

	// Check if expired
	if time.Now().After(token.ExpiresAt) {
		return nil, domain.ErrTokenExpired
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return nil, err
	}

	// Revoke old token
	if err := uc.tokenRepo.Revoke(ctx, token.ID); err != nil {
		return nil, err
	}

	// Generate new tokens
	return uc.generateTokens(ctx, user)
}

// Logout invalidates refresh token
func (uc *AuthUseCase) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return domain.ErrTokenNotFound
	}

	tokenHash := jwt.HashRefresh(refreshToken)
	token, err := uc.tokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrTokenNotFound) {
			return nil // Already logged out
		}
		return err
	}

	return uc.tokenRepo.Revoke(ctx, token.ID)
}

// DeleteAccount deletes user account
func (uc *AuthUseCase) DeleteAccount(ctx context.Context, userID string) error {
	// Convert string to UUID
	id, err := parseUUID(userID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	// Revoke all tokens
	if err := uc.tokenRepo.RevokeAllUserTokens(ctx, id); err != nil {
		return err
	}

	// TODO: Delete user from userRepo
	// For now, just revoke tokens

	return nil
}

// ValidateAccessToken validates access token and returns claims
func (uc *AuthUseCase) ValidateAccessToken(ctx context.Context, token string) (*jwt.Claims, error) {
	claims, err := uc.jwtService.ValidateAccess(token)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	return claims, nil
}

// Helper methods

func (uc *AuthUseCase) generateTokens(ctx context.Context, user *domain.User) (*RegisterResponse, error) {
	// Generate access token
	accessToken, err := uc.jwtService.IssueAccess(user.ID, getString(user.Email), uc.accessTTL)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := uc.jwtService.IssueRefresh()
	if err != nil {
		return nil, err
	}

	// Store refresh token hash in DB
	tokenHash := jwt.HashRefresh(refreshToken)
	token := &domain.RefreshToken{
		ID:        newUUID(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(uc.refreshTTL),
		CreatedAt: time.Now(),
	}
	if err := uc.tokenRepo.Create(ctx, token); err != nil {
		return nil, err
	}

	return &RegisterResponse{
		User:        user.ToResponse(),
		AccessToken: accessToken,
		ExpiresIn:   int64(uc.accessTTL.Seconds()),
	}, nil
}

func isEmail(s string) bool {
	return len(s) > 0 && (s[0] >= 'a' && s[0] <= 'z' || s[0] >= 'A' && s[0] <= 'Z')
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func newUUID() interface{ String() string } {
	return &uuidWrapper{}
}

type uuidWrapper struct{}
func (u *uuidWrapper) String() string { return "generated-uuid" }

func parseUUID(s string) (interface{ String() string }, error) {
	return &uuidWrapper{s}, nil
}

// Helper functions
func isEmailIdentifier(s string) bool {
	atIndex := -1
	for i, c := range s {
		if c == '@' {
			atIndex = i
			break
		}
	}
	return atIndex > 0 && atIndex < len(s)-1
}

func getUserEmail(user *domain.User) string {
	if user.Email == nil {
		return ""
	}
	return *user.Email
}

