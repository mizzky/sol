package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sol_coffeesys/backend/middleware"
	"sol_coffeesys/backend/pkg/apperror"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestErrorHandler(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		handler    gin.HandlerFunc
		wantStatus int
		wantBody   map[string]any
	}{
		{
			name: "ValidationErrorは400を返す",
			handler: func(c *gin.Context) {
				_ = c.Error(apperror.NewValidationError("email", "bad-email", "format", ""))
			},
			wantStatus: http.StatusBadRequest,
			wantBody: map[string]any{
				"error": apperror.ValidationMessageEmail,
			},
		},
		{
			name: "UnauthorizedErrorは401を返す",
			handler: func(c *gin.Context) {
				_ = c.Error(apperror.NewUnauthorizedError("token_not_found", apperror.UnauthorizedMessageAuth))
			},
			wantStatus: http.StatusUnauthorized,
			wantBody: map[string]any{
				"error": apperror.UnauthorizedMessageAuth,
			},
		},
		{
			name: "エラーなしは下流レスポンス維持",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			},
			wantStatus: http.StatusOK,
			wantBody: map[string]any{
				"ok": true,
			},
		},
		{
			name: "既に書き込み済みなら上書きしない",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusTeapot, gin.H{"before": "written"})
				_ = c.Error(apperror.NewInternalError("X", nil, apperror.InternalServerMessageCommon))
			},
			wantStatus: http.StatusTeapot,
			wantBody: map[string]any{
				"before": "written",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := gin.New()
			r.Use(middleware.ErrorHandler(apperror.ToHTTP))
			r.GET("/test", tt.handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			var got map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("failed to unmarshal body: %v body=%s", err, w.Body.String())
			}

			if w.Code != tt.wantStatus {
				t.Fatalf("status mismatch: got=%d want=%d body=%s", w.Code, tt.wantStatus, w.Body.String())
			}

			if !reflect.DeepEqual(got, tt.wantBody) {
				t.Fatalf("body mismatch: got=%v want=%v", got, tt.wantBody)
			}
		})
	}
}

func TestNewJSONLogger_Redaction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		attrs     []any
		assertion func(t *testing.T, got map[string]any)
	}{
		{
			name:  "passwordが[REDACTED]に置換される",
			attrs: []any{"password", "super-secret"},
			assertion: func(t *testing.T, got map[string]any) {
				t.Helper()
				if got["password"] != "[REDACTED]" {
					t.Fatalf("password mismatch: got=%v want=%v", got["password"], "[REDACTED]")
				}
			},
		},
		{
			name:  "tokenが[REDACTED]に置換される",
			attrs: []any{"token", "jwt-token-value"},
			assertion: func(t *testing.T, got map[string]any) {
				t.Helper()
				if got["token"] != "[REDACTED]" {
					t.Fatalf("token mismatch: got=%v want=%v", got["token"], "[REDACTED]")
				}
			},
		},
		{
			name:  "マスク対象外の属性はそのまま出力される",
			attrs: []any{"event", "login_failed", "status", "400"},
			assertion: func(t *testing.T, got map[string]any) {
				t.Helper()
				if got["event"] != "login_failed" {
					t.Fatalf("event mismatch: got=%v want=%v", got["event"], "login_failed")
				}
				if got["status"] != "400" {
					t.Fatalf("status mismatch: got=%v want=%v", got["status"], "400")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := middleware.NewJSONLogger(&buf, slog.LevelInfo)

			logger.Info("test-log", tt.attrs...)
			line := strings.TrimSpace(buf.String())
			if line == "" {
				t.Fatal("expected log output, got empty string")
			}

			var got map[string]any
			if err := json.Unmarshal([]byte(line), &got); err != nil {
				t.Fatalf("failed to unmarshal json log: %v line=%s", err, line)
			}

			tt.assertion(t, got)
		})
	}
}
