package domain

import "mishatkin/shared/pkg/errors"

// Domain errors
var (
	ErrUserNotFound       = errors.New(errors.CodeNotFound, "User not found", 404)
	ErrEmailExists        = errors.New(errors.CodeEmailExists, "Email already exists", 409)
	ErrPhoneExists        = errors.New(errors.CodePhoneExists, "Phone already exists", 409)
	ErrInvalidCredentials = errors.New(errors.CodeInvalidCredentials, "Invalid credentials", 401)
	ErrWeakPassword      = errors.New(errors.CodeValidation, "Password too weak", 400)
	ErrTokenExpired       = errors.New(errors.CodeTokenExpired, "Token expired", 401)
	ErrTokenRevoked       = errors.New(errors.CodeTokenExpired, "Token revoked", 401)
	ErrTokenNotFound      = errors.New(errors.CodeInvalidToken, "Token not found", 401)
	ErrInvalidToken       = errors.New(errors.CodeInvalidToken, "Invalid token", 401)
)

