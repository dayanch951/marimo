package validator

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"Valid email", "user@example.com", false},
		{"Valid email with subdomain", "user@mail.example.com", false},
		{"Valid email with plus", "user+tag@example.com", false},
		{"Valid email with dash", "user-name@example.com", false},
		{"Empty email", "", true},
		{"Missing @", "userexample.com", true},
		{"Missing domain", "user@", true},
		{"Missing local part", "@example.com", true},
		{"Invalid characters", "user name@example.com", true},
		{"Too long", "a" + string(make([]byte, 300)) + "@example.com", true},
		{"No TLD", "user@example", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	reqs := DefaultPasswordRequirements()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Valid password", "SecurePass123!", false},
		{"Valid with special chars", "P@ssw0rd!", false},
		{"Too short", "Pass1!", true},
		{"No uppercase", "password123!", true},
		{"No lowercase", "PASSWORD123!", true},
		{"No digit", "Password!", true},
		{"No special char", "Password123", true},
		{"Empty", "", true},
		{"Only letters", "PasswordOnly", true},
		{"Valid complex", "MyC0mpl3x!P@ss", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password, reqs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid simple name", "John Doe", false},
		{"Valid with hyphen", "Mary-Jane", false},
		{"Valid with apostrophe", "O'Brien", false},
		{"Valid Unicode", "José García", false},
		{"Empty name", "", true},
		{"Too long", string(make([]byte, 300)), true},
		{"Numbers", "John123", true},
		{"Special chars", "John@Doe", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckSQLInjection(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Clean input", "normal text", false},
		{"SQL SELECT", "SELECT * FROM users", true},
		{"SQL UNION", "1 UNION SELECT password", true},
		{"SQL DROP", "DROP TABLE users", true},
		{"SQL INSERT", "INSERT INTO users", true},
		{"Case insensitive", "SeLeCt * FrOm", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckSQLInjection(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckSQLInjection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckXSS(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Clean input", "normal text", false},
		{"Script tag", "<script>alert('xss')</script>", true},
		{"JavaScript protocol", "javascript:alert('xss')", true},
		{"Onerror attribute", "<img onerror='alert(1)'>", true},
		{"Iframe tag", "<iframe src='evil.com'>", true},
		{"Eval function", "eval('malicious')", true},
		{"Normal HTML", "<p>This is a paragraph</p>", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckXSS(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckXSS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Clean input", "hello", "hello"},
		{"With spaces", "  hello  ", "hello"},
		{"HTML tags", "<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"Special chars", "test & \"quote\"", "test &amp; &#34;quote&#34;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name    string
		uuid    string
		wantErr bool
	}{
		{"Valid UUID", "550e8400-e29b-41d4-a716-446655440000", false},
		{"Valid lowercase", "550e8400-e29b-41d4-a716-446655440000", false},
		{"Valid uppercase", "550E8400-E29B-41D4-A716-446655440000", false},
		{"Invalid format", "not-a-uuid", true},
		{"Missing hyphens", "550e8400e29b41d4a716446655440000", true},
		{"Too short", "550e8400-e29b-41d4", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUUID(tt.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUUID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkValidateEmail(b *testing.B) {
	email := "user@example.com"
	for i := 0; i < b.N; i++ {
		ValidateEmail(email)
	}
}

func BenchmarkValidatePassword(b *testing.B) {
	password := "SecurePass123!"
	reqs := DefaultPasswordRequirements()
	for i := 0; i < b.N; i++ {
		ValidatePassword(password, reqs)
	}
}
