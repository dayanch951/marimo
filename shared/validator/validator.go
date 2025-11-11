package validator

import (
	"errors"
	"html"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrInvalidEmail            = errors.New("invalid email format")
	ErrWeakPassword            = errors.New("password does not meet requirements")
	ErrPasswordTooShort        = errors.New("password must be at least 8 characters")
	ErrPasswordNoUppercase     = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLowercase     = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit         = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecial       = errors.New("password must contain at least one special character")
	ErrInvalidInput            = errors.New("input contains invalid characters")
	ErrInputTooLong            = errors.New("input exceeds maximum length")
	ErrEmptyInput              = errors.New("input cannot be empty")
)

// Email regex pattern (RFC 5322 simplified)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// SQL injection patterns
var sqlInjectionPatterns = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|script|javascript|<script)`)

// XSS patterns
var xssPatterns = regexp.MustCompile(`(?i)(<script|javascript:|onerror=|onload=|<iframe|eval\()`)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return ErrEmptyInput
	}

	if len(email) > 255 {
		return ErrInputTooLong
	}

	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

// PasswordRequirements defines password validation rules
type PasswordRequirements struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
}

// DefaultPasswordRequirements returns default password requirements
func DefaultPasswordRequirements() PasswordRequirements {
	return PasswordRequirements{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
	}
}

// ValidatePassword validates password against requirements
func ValidatePassword(password string, reqs PasswordRequirements) error {
	if len(password) < reqs.MinLength {
		return ErrPasswordTooShort
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if reqs.RequireUpper && !hasUpper {
		return ErrPasswordNoUppercase
	}
	if reqs.RequireLower && !hasLower {
		return ErrPasswordNoLowercase
	}
	if reqs.RequireDigit && !hasDigit {
		return ErrPasswordNoDigit
	}
	if reqs.RequireSpecial && !hasSpecial {
		return ErrPasswordNoSpecial
	}

	return nil
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// HTML escape
	input = html.EscapeString(input)

	return input
}

// ValidateName validates a person's name
func ValidateName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return ErrEmptyInput
	}

	if len(name) > 255 {
		return ErrInputTooLong
	}

	// Allow letters, spaces, hyphens, apostrophes
	nameRegex := regexp.MustCompile(`^[a-zA-ZàáâäãåąčćęèéêëėįìíîïłńòóôöõøùúûüųūÿýżźñçčšžÀÁÂÄÃÅĄĆČĖĘÈÉÊËÌÍÎÏĮŁŃÒÓÔÖÕØÙÚÛÜŲŪŸÝŻŹÑßÇŒÆČŠŽ∂ð' -]+$`)

	if !nameRegex.MatchString(name) {
		return ErrInvalidInput
	}

	return nil
}

// CheckSQLInjection checks for SQL injection patterns
func CheckSQLInjection(input string) error {
	if sqlInjectionPatterns.MatchString(input) {
		return ErrInvalidInput
	}
	return nil
}

// CheckXSS checks for XSS patterns
func CheckXSS(input string) error {
	if xssPatterns.MatchString(input) {
		return ErrInvalidInput
	}
	return nil
}

// ValidateAndSanitize validates and sanitizes a string input
func ValidateAndSanitize(input string, maxLength int) (string, error) {
	input = strings.TrimSpace(input)

	if input == "" {
		return "", ErrEmptyInput
	}

	if len(input) > maxLength {
		return "", ErrInputTooLong
	}

	// Check for SQL injection
	if err := CheckSQLInjection(input); err != nil {
		return "", err
	}

	// Check for XSS
	if err := CheckXSS(input); err != nil {
		return "", err
	}

	// Sanitize
	input = SanitizeString(input)

	return input, nil
}

// ValidateUUID validates a UUID string
func ValidateUUID(uuid string) error {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	if !uuidRegex.MatchString(strings.ToLower(uuid)) {
		return ErrInvalidInput
	}

	return nil
}

// ValidateJSONField validates a generic JSON field
func ValidateJSONField(fieldName, value string, maxLength int) error {
	if value == "" {
		return errors.New(fieldName + " is required")
	}

	if len(value) > maxLength {
		return errors.New(fieldName + " is too long")
	}

	// Check for dangerous patterns
	if err := CheckSQLInjection(value); err != nil {
		return errors.New(fieldName + " contains invalid characters")
	}

	if err := CheckXSS(value); err != nil {
		return errors.New(fieldName + " contains invalid characters")
	}

	return nil
}
