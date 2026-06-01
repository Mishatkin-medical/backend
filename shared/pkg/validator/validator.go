package validator

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	phoneRegex *regexp.Regexp
)

func init() {
	validate = validator.New()
	
	// Register phone number validation
	validate.RegisterValidation("e164", validateE164Phone)
	
	// Register password validation
	validate.RegisterValidation("password", validatePassword)
}

// Validate validates a struct
func Validate(s interface{}) error {
	return validate.Struct(s)
}

// ValidateField validates a single field
func ValidateField(s interface{}, field string) error {
	return validate.VarCtx(nil, field, "")
}

// RegisterCustomValidators registers custom validators
func RegisterCustomValidators(v *validator.Validate) {
	v.RegisterValidation("e164", validateE164Phone)
	v.RegisterValidation("password", validatePassword)
}

func validateE164Phone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true // allow empty (optional)
	}
	
	// E.164 format: +[country code][number]
	matched, _ := regexp.MatchString(`^\+[1-9]\d{6,14}$`, phone)
	return matched
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
	if len(password) < 8 {
		return false
	}
	
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	
	for _, ch := range password {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		case strings.ContainsAny(string(ch), "!@#$%^&*(),.?\":{}|<>"):
			hasSpecial = true
		}
	}
	
	return hasUpper && hasLower && hasDigit && hasSpecial
}

// FormatErrors converts validation errors to field errors
func FormatErrors(err error) []FieldError {
	if err == nil {
		return nil
	}

	var errors []FieldError
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, FieldError{
				Field:   toSnakeCase(e.Field()),
				Message: getErrorMessage(e),
			})
		}
	}

	return errors
}

func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "e164":
		return "Phone must be in E.164 format (e.g., +79001234567)"
	case "password":
		return "Password must contain uppercase, lowercase, digit and special character"
	case "uuid":
		return "Invalid UUID format"
	default:
		return e.Tag()
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// FieldError represents a field validation error
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
