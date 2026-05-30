package logging

import (
	"log/slog"
)

const (
	CtxKeyRequestStartedAt = "request_started_at"
	CtxKeyRequestID        = "request_id"
	CtxKeyUserID           = "userID"
)

type EventInput struct {
	Event   string
	Status  int
	Message string // optional
	Level   slog.Level
	Extra   []slog.Attr
}
