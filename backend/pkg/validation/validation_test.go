package validation

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	cases := []struct {
		name   string
		email  string
		wantOk bool
	}{
		{"valid simple", "user@example.com", true},
		{"valid trimmed upper", "USER@Example.Com", true},
		{"no at", "no-at.example.com", false},
		{"double at", "user@@example.com", false},
		{"bad domain", "user@.com", false},
		{"subdomain", "user@sub.example.com", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateEmail(strings.TrimSpace(c.email))
			if (err == nil) != c.wantOk {
				t.Fatalf("ValidateEmail(%q) wantOk=%v err=%v", c.email, c.wantOk, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	long65 := strings.Repeat("a", 65)
	long64 := strings.Repeat("a", 64)

	cases := []struct {
		name   string
		pw     string
		wantOk bool
	}{
		{"too short", "short7", false},
		{"min ok", "Passw0rd", true},
		{"max ok", long64, true},
		{"too long", long65, false},
		{"contains space", "pass word1", false},
		{"contains control", "\x01abcdefg", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidatePassword(c.pw)
			if (err == nil) != c.wantOk {
				t.Fatalf("ValidatePassword(%q) wantOk=%v err=%v", c.pw, c.wantOk, err)
			}
		})
	}

}

func TestValidateName(t *testing.T) {
	long256 := strings.Repeat("a", 256)
	cases := []struct {
		name   string
		n      string
		wantOk bool
	}{
		{"empty", "", false},
		{"trimmed ok", " Alice ", true},
		{"too long", long256, false},
		{"valid", "太郎", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateName(strings.TrimSpace(c.n))
			if (err == nil) != c.wantOk {
				t.Fatalf("ValidationName(%q) wantOk=%v err=%v", c.n, c.wantOk, err)
			}
		})
	}
}

func TestValidateRegisterRequest(t *testing.T) {
	long65 := strings.Repeat("a", 65)
	long64 := strings.Repeat("a", 64)
	long256 := strings.Repeat("a", 256)

	cases := []struct {
		name   string
		n      string
		email  string
		pw     string
		wantOk bool
	}{
		{"valid", "Alice", "alice@example.com", "Passw0rd", true},
		{"empty name", "", "alice@example.com", "Passw0rd", false},
		{"empty email", "Alice", "", "Passw0rd", false},
		{"empty password", "Alice", "alice@example.com", "", false},
		{"invalid email", "Alice", "alice-at-example.com", "Passw0rd", false},
		{"short password", "Alice", "alice@example.com", "short7", false},
		{"pw contains space", "Alice", "alice@example.com", "pass word1", false},
		{"pw too long", "Alice", "alice@example.com", long65, false},
		{"name too long", long256, "alice@example.com", "Passw0rd", false},
		{"email trimmed upper valid", " Bob ", " BOB@Example.COM ", "Passw0rd", true},
		{"max password ok", "Alice", "alice@example.com", long64, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateRegisterRequest(strings.TrimSpace(c.n), strings.TrimSpace(c.email), c.pw)
			if (err == nil) != c.wantOk {
				t.Fatalf("ValidateRegisterRequest(name=%q, email=%q, pw=%q) wantOk=%v err=%v", c.n, c.email, c.pw, c.wantOk, err)
			}
		})
	}
}

func TestValidateRole(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		wantErr error
	}{
		{
			name:    "正常系：admin",
			role:    "admin",
			wantErr: nil,
		},
		{
			name:    "正常系：member",
			role:    "member",
			wantErr: nil,
		},
		{
			name:    "異常系：user(不正な値)",
			role:    "user",
			wantErr: ErrInvalidRole,
		},
		{
			name:    "異常系：空文字",
			role:    "",
			wantErr: ErrInvalidRole,
		},
		{
			name:    "異常系：大文字ADMIN",
			role:    "ADMIN",
			wantErr: ErrInvalidRole,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRole(tt.role)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
