package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
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

func TestErrorHandler_LogOutput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                     string
		err                      error
		route                    string
		requestPath              string
		omitRequestIDMiddleware  bool
		wantRequestIDEmpty       bool
		checkRequestIDWithHeader bool
		wantLevel                string
		wantStatus               int
		wantErrorType            string
		wantMessage              string
		wantRoute                string
	}{
		{
			name:          "ValidationErrorはINFOで出力",
			err:           apperror.NewValidationError("email", "bad-email", "format", ""),
			wantLevel:     "INFO",
			wantStatus:    http.StatusBadRequest,
			wantErrorType: "ValidationError",
			wantMessage:   apperror.ValidationMessageEmail,
		},
		{
			name:          "UnauthorizedErrorはWARNで出力",
			err:           apperror.NewUnauthorizedError("token_not_found", apperror.UnauthorizedMessageAuth),
			wantLevel:     "WARN",
			wantStatus:    http.StatusUnauthorized,
			wantErrorType: "UnauthorizedError",
			wantMessage:   apperror.UnauthorizedMessageAuth,
		},
		{
			name:          "InternalErrorはERRORで出力",
			err:           apperror.NewInternalError("CreateUser", errors.New("db"), apperror.InternalServerMessageCommon),
			wantLevel:     "ERROR",
			wantStatus:    http.StatusInternalServerError,
			wantErrorType: "InternalError",
			wantMessage:   apperror.InternalServerMessageCommon,
		},
		{
			name:          "NotFoundErrorはINFOで出力",
			err:           apperror.NewNotFoundError("user", 0, apperror.NotFoundMessageUser),
			wantLevel:     "INFO",
			wantStatus:    http.StatusNotFound,
			wantErrorType: "NotFoundError",
			wantMessage:   apperror.NotFoundMessageUser,
		},
		{
			name:          "ConflictErrorはINFOで出力",
			err:           apperror.NewConflictError("sku", "ABC", apperror.ConflictMessageSku),
			wantLevel:     "INFO",
			wantStatus:    http.StatusConflict,
			wantErrorType: "ConflictError",
			wantMessage:   apperror.ConflictMessageSku,
		},
		{
			name:          "BusinessLogicErrorはINFOで出力",
			err:           apperror.NewBusinessLogicError(apperror.BusinessLogicMessageGeneric),
			wantLevel:     "INFO",
			wantStatus:    http.StatusBadRequest,
			wantErrorType: "BusinessLogicError",
			wantMessage:   apperror.BusinessLogicMessageGeneric,
		},
		{
			name:          "ForbiddenはWARNで出力",
			err:           apperror.NewForbiddenError("admin", "user", apperror.ForbiddenMessageAdmin),
			wantLevel:     "WARN",
			wantStatus:    http.StatusForbidden,
			wantErrorType: "ForbiddenError",
			wantMessage:   apperror.ForbiddenMessageAdmin,
		},

		// /user/42ではなく/user/:idのように識別子ではなくプレースホルダ付きで返すこと
		{
			name:          "pathではなくrouteを出力する",
			route:         "/users/:id",
			requestPath:   "/users/42",
			err:           apperror.NewNotFoundError("user", 42, apperror.NotFoundMessageUser),
			wantLevel:     "INFO",
			wantStatus:    http.StatusNotFound,
			wantErrorType: "NotFoundError",
			wantMessage:   apperror.NotFoundMessageUser,
			wantRoute:     "/users/:id",
		},
		{
			name:                    "request_idなしで空文字が出る",
			err:                     apperror.NewUnauthorizedError("token_not_found", apperror.UnauthorizedMessageAuth),
			omitRequestIDMiddleware: true,
			wantRequestIDEmpty:      true,
			wantLevel:               "WARN",
			wantStatus:              http.StatusUnauthorized,
			wantErrorType:           "UnauthorizedError",
			wantMessage:             apperror.UnauthorizedMessageAuth,
		},
		{
			name:                     "request_idがヘッダと一致する",
			err:                      apperror.NewUnauthorizedError("token_not_found", apperror.UnauthorizedMessageAuth),
			omitRequestIDMiddleware:  false,
			checkRequestIDWithHeader: true,
			wantLevel:                "WARN",
			wantStatus:               http.StatusUnauthorized,
			wantErrorType:            "UnauthorizedError",
			wantMessage:              apperror.UnauthorizedMessageAuth,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalLogger := slog.Default()
			t.Cleanup(func() { slog.SetDefault(originalLogger) })

			var buf bytes.Buffer
			logger := middleware.NewJSONLogger(&buf, slog.LevelInfo)
			slog.SetDefault(logger)

			route := tt.route
			if route == "" {
				route = "/test"
			}
			requestPath := tt.requestPath
			if requestPath == "" {
				requestPath = route
			}

			r := gin.New()
			if !tt.omitRequestIDMiddleware {
				r.Use(middleware.RequestIDMiddleware())
			}
			r.Use(middleware.ErrorHandler(apperror.ToHTTP))
			r.GET(route, func(c *gin.Context) {
				_ = c.Error(tt.err)
			})

			req := httptest.NewRequest(http.MethodGet, requestPath, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			var got map[string]any
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("failed to decode log: %v raw =%v", err, buf.String())
			}

			if got["level"] != tt.wantLevel {
				t.Fatalf("level mismatch: got=%v want=%v", got["level"], tt.wantLevel)
			}
			if got["message"] != tt.wantMessage {
				t.Fatalf("message mismatch: got=%v want=%v", got["message"], tt.wantMessage)
			}
			if got["error_type"] != tt.wantErrorType {
				t.Fatalf("error_type mismatch: got=%v want=%v", got["error_type"], tt.wantErrorType)
			}

			if int(got["status"].(float64)) != tt.wantStatus {
				t.Fatalf("status mismatch: got=%v want=%v", got["status"], tt.wantStatus)
			}
			if got["method"] != http.MethodGet {
				t.Fatalf("method mismatch: got=%v want=%v", got["method"], http.MethodGet)
			}

			wantRoute := tt.wantRoute
			if wantRoute == "" {
				wantRoute = route
			}
			if got["route"] != wantRoute {
				t.Fatalf("route mismatch: got=%v want=%v", got["route"], wantRoute)
			}
			requestID, ok := got["request_id"].(string)
			if !ok {
				t.Fatal("request_id is missing or not string")
			}
			if tt.wantRequestIDEmpty {
				if requestID != "" {
					t.Fatalf("request_id mismatch: got=%q want empty", requestID)
				}
			} else {
				if requestID == "" {
					t.Fatal("request_id should not be empty")
				}
			}

			if tt.checkRequestIDWithHeader {
				headerID := w.Header().Get("X-Request-ID")
				if headerID == "" {
					t.Fatal("X-Request-ID header is empty")
				}
				if requestID != headerID {
					t.Fatalf("request_id mismatch: log=%q header=%q", requestID, headerID)
				}
			}

			durationMS, ok := got["duration_ms"].(float64)
			if !ok {
				t.Fatal("duration_ms is missing or not numeric")
			}
			if durationMS < 0 {
				t.Fatalf("duration_ms is negative: %v", durationMS)
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
