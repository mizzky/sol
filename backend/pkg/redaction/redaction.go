package redaction

import (
	"log/slog"
	"strings"
)

const Redacted = "[REDACTED]"

func RedactAttr(_ []string, attr slog.Attr) slog.Attr {
	if attr.Value.Kind() != slog.KindString {
		return attr
	}

	switch strings.ToLower(attr.Key) {
	case "password", "token", "access_token", "refresh_token", "authorization":
		return slog.String(attr.Key, Redacted)
	case "email":
		return slog.String(attr.Key, MaskEmail(attr.Value.String()))
	default:
		return attr
	}
}

func MaskEmail(s string) string {
	local, domain, ok := strings.Cut(s, "@")
	if !ok || local == "" || domain == "" {
		return s
	}

	return local[:1] + "****@" + domain
}
