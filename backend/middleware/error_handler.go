package middleware

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sol_coffeesys/backend/pkg/apperror"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewJSONLogger(w io.Writer, level slog.Leveler) *slog.Logger {
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: redactSensitiveAttr,
	})
	return slog.New(handler)
}

func redactSensitiveAttr(_ []string, attr slog.Attr) slog.Attr {
	switch strings.ToLower(attr.Key) {
	case "password", "token":
		return slog.String(attr.Key, "[REDACTED]")
	default:
		return attr
	}
}

func ErrorHandler(toHTTP func(error) (int, string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		err := c.Errors.Last().Err
		if err == nil {
			return
		}

		status, msg := toHTTP(err)
		logError(c, err, status, msg)

		if c.Writer.Written() {
			return
		}
		c.JSON(status, gin.H{"error": msg})

	}
}

func logError(c *gin.Context, err error, status int, msg string) {
	route := c.FullPath()
	if route == "" {
		route = c.Request.URL.Path
	}

	slog.LogAttrs(
		c.Request.Context(),
		logLevelForError(err),
		"http_error",
		slog.String("event", "http_error"),
		slog.String("message", msg),
		slog.String("error_type", errorTypeName(err)),
		slog.Int("status", status),
		slog.String("method", c.Request.Method),
		slog.String("route", route),
	)
}

func logLevelForError(err error) slog.Level {
	switch {
	case isValidationError(err), isNotFoundError(err), isConflictError(err), isBusinessLogicError(err):
		return slog.LevelInfo
	case isUnauthorizedError(err), isForbiddenError(err):
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}

func errorTypeName(err error) string {
	switch {
	case isValidationError(err):
		return "ValidationError"
	case isNotFoundError(err):
		return "NotFoundError"
	case isConflictError(err):
		return "ConflictError"
	case isUnauthorizedError(err):
		return "UnauthorizedError"
	case isForbiddenError(err):
		return "ForbiddenError"
	case isBusinessLogicError(err):
		return "BusinessLogicError"
	case isInternalError(err):
		return "InternalError"
	default:
		return strings.TrimPrefix(fmt.Sprintf("%T", err), "*")
	}
}

func isValidationError(err error) bool {
	var target *apperror.ValidationError
	return errors.As(err, &target)
}

func isNotFoundError(err error) bool {
	var target *apperror.NotFoundError
	return errors.As(err, &target)
}

func isConflictError(err error) bool {
	var target *apperror.ConflictError
	return errors.As(err, &target)
}

func isUnauthorizedError(err error) bool {
	var target *apperror.UnauthorizedError
	return errors.As(err, &target)
}

func isForbiddenError(err error) bool {
	var target *apperror.ForbiddenError
	return errors.As(err, &target)
}

func isBusinessLogicError(err error) bool {
	var target *apperror.BusinessLogicError
	return errors.As(err, &target)
}

func isInternalError(err error) bool {
	var target *apperror.InternalError
	return errors.As(err, &target)
}
