package logging

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func BuildAttrs(c *gin.Context, in EventInput) []slog.Attr {
	route := c.FullPath()
	if route == "" && c.Request != nil && c.Request.URL != nil {
		route = c.Request.URL.Path
	}

	method := ""
	if c.Request != nil {
		method = c.Request.Method
	}

	attrs := []slog.Attr{
		slog.String("request_id", c.GetString(CtxKeyRequestID)),
		slog.String("method", method),
		slog.String("route", route),
		slog.Int("status", in.Status),
		slog.String("event", in.Event),
	}

	if startedAtRaw, ok := c.Get(CtxKeyRequestStartedAt); ok {
		if startedAt, ok := startedAtRaw.(time.Time); ok {
			elapsed := time.Since(startedAt)
			attrs = append(attrs, slog.Float64("duration_ms", float64(elapsed.Microseconds())/1000))
		}
	}

	if rawUserID, ok := c.Get(CtxKeyUserID); ok {
		if userID, ok := rawUserID.(int64); ok {
			attrs = append(attrs, slog.Int64("user_id", userID))
		}
	}

	if in.Message != "" {
		attrs = append(attrs, slog.String("message", in.Message))
	}

	if len(in.Extra) > 0 {
		attrs = append(attrs, in.Extra...)
	}

	return attrs
}
