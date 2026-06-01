package errors

import (
	"fmt"
	"net/http"
)

// Code represents an error code
type Code string

// Error codes
const (
	CodeValidation       Code = "VALIDATION_ERROR"
	CodeInvalidToken     Code = "INVALID_TOKEN"
	CodeTokenExpired     Code = "TOKEN_EXPIRED"
	CodeInvalidCredentials Code = "INVALID_CREDENTIALS"
	CodeForbidden        Code = "FORBIDDEN"
	CodeNotFound         Code = "NOT_FOUND"
	CodeEmailExists      Code = "EMAIL_EXISTS"
	CodePhoneExists      Code = "PHONE_EXISTS"
	CodeDevicePaired     Code = "DEVICE_ALREADY_PAIRED"
	CodeRateLimited      Code = "RATE_LIMITED"
	CodeInternal         Code = "INTERNAL_ERROR"
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"
	CodeSafetyGate      Code = "SAFETY_GATE_FAILED"
	CodeCycleExpired     Code = "CYCLE_EXPIRED"
	CodeDoseTooHigh      Code = "DOSE_TOO_HIGH"
	CodeGlucoseTooLow    Code = "GLUCOSE_TOO_LOW"
)

// AppError represents an application error
type AppError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	HTTP    int     `json:"-"`
	Details []FieldError `json:"details,omitempty"`
}

// FieldError represents a validation field error
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// New creates a new error
func New(code Code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		HTTP:    httpStatus,
	}
}

// WithDetails adds field errors
func (e *AppError) WithDetails(details ...FieldError) *AppError {
	e.Details = details
	return e
}

// Common errors
var (
	ErrValidation       = New(CodeValidation, "Validation failed", http.StatusBadRequest)
	ErrInvalidToken     = New(CodeInvalidToken, "Invalid token", http.StatusUnauthorized)
	ErrTokenExpired     = New(CodeTokenExpired, "Token expired", http.StatusUnauthorized)
	ErrInvalidCredentials = New(CodeInvalidCredentials, "Invalid credentials", http.StatusUnauthorized)
	ErrForbidden        = New(CodeForbidden, "Access denied", http.StatusForbidden)
	ErrNotFound         = New(CodeNotFound, "Resource not found", http.StatusNotFound)
	ErrEmailExists      = New(CodeEmailExists, "Email already exists", http.StatusConflict)
	ErrPhoneExists      = New(CodePhoneExists, "Phone already exists", http.StatusConflict)
	ErrDevicePaired     = New(CodeDevicePaired, "Device already paired", http.StatusConflict)
	ErrRateLimited      = New(CodeRateLimited, "Too many requests", http.StatusTooManyRequests)
	ErrInternal         = New(CodeInternal, "Internal server error", http.StatusInternalServerError)
	ErrServiceUnavailable = New(CodeServiceUnavailable, "Service unavailable", http.StatusServiceUnavailable)
)

// Is checks if error matches
func Is(err, target error) bool {
	if e, ok := err.(*AppError); ok {
		if t, ok := target.(*AppError); ok {
			return e.Code == t.Code
		}
	}
	return err == target
}

// FromCode creates error from code
func FromCode(code Code) *AppError {
	switch code {
	case CodeValidation:
		return ErrValidation
	case CodeInvalidToken:
		return ErrInvalidToken
	case CodeTokenExpired:
		return ErrTokenExpired
	case CodeInvalidCredentials:
		return ErrInvalidCredentials
	case CodeForbidden:
		return ErrForbidden
	case CodeNotFound:
		return ErrNotFound
	case CodeEmailExists:
		return ErrEmailExists
	case CodePhoneExists:
		return ErrPhoneExists
	case CodeDevicePaired:
		return ErrDevicePaired
	case CodeRateLimited:
		return ErrRateLimited
	default:
		return ErrInternal
	}
}

// Wrap wraps an error with code
func Wrap(err error, code Code, message string) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:    code,
		Message: message,
		HTTP:    FromCode(code).HTTP,
	}
}
