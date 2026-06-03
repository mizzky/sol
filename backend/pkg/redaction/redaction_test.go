package redaction

import (
	"log/slog"
	"testing"
)

func TestRedactAttr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
		in   string
		want string
	}{
		{
			name: "passwordはREDACTEDに置換される",
			key:  "password",
			in:   "super-secret",
			want: "[REDACTED]",
		},
		{
			name: "tokenはREDACTEDに置換される",
			key:  "token",
			in:   "jwt-token-value",
			want: "[REDACTED]",
		},
		{
			name: "access_tokenはREDACTEDに置換される",
			key:  "access_token",
			in:   "access-token-value",
			want: "[REDACTED]",
		},
		{
			name: "refresh_tokenはREDACTEDに置換される",
			key:  "refresh_token",
			in:   "refresh-token-value",
			want: "[REDACTED]",
		},
		{
			name: "authorizationはREDACTEDに置換される",
			key:  "authorization",
			in:   "Bearer jwt-token-value",
			want: "[REDACTED]",
		},
		{
			name: "emailは既存互換の形式でマスクされる",
			key:  "email",
			in:   "user@example.com",
			want: "u****@example.com",
		},
		{
			name: "対象外キーはそのまま",
			key:  "event",
			in:   "user_login_succeeded",
			want: "user_login_succeeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := RedactAttr(nil, slog.String(tt.key, tt.in))

			if got.Value.String() != tt.want {
				t.Fatalf("RedactAttr() = %q want %q", got.Value.String(), tt.want)
			}
		})
	}
}

func TestMaskEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "通常のメールアドレス",
			in:   "user@example.com",
			want: "u****@example.com",
		},
		{
			name: "ローカルパートが1文字",
			in:   "a@example.com",
			want: "a****@example.com",
		},
		{
			name: "メール形式でない文字列",
			in:   "invalid-email",
			want: "invalid-email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			t.Parallel()

			if got := MaskEmail(tt.in); got != tt,want {
				t.Fatalf("MaskEmail() = %q want %q", got, tt.want)
			}
		})
	}
}
