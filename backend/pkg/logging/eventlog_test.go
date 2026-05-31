package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestBuildAttrs_ContainRequireFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		setupContext      func(*gin.Context)
		input             EventInput
		wantRequestID     string
		wantMethod        string
		wantRoute         string
		wantStatus        int
		wantEvent         string
		wantUserID        int64
		wantUserIDPresent bool
		checkDuration     bool
	}{
		{
			name: "必須フィールドが出る",
			setupContext: func(c *gin.Context) {
				c.Request = httptest.NewRequest(http.MethodGet, "/api/login", nil)
				c.Set(CtxKeyRequestID, "req-1")
				c.Set(CtxKeyRequestStartedAt, time.Now().Add(-10*time.Millisecond))
			},
			input: EventInput{
				Event:  "auth_login_succeeded",
				Status: http.StatusOK,
				Level:  slog.LevelInfo,
			},
			wantRequestID:     "req-1",
			wantMethod:        http.MethodGet,
			wantRoute:         "/api/login",
			wantStatus:        http.StatusOK,
			wantEvent:         "auth_login_succeeded",
			wantUserIDPresent: false,
			checkDuration:     true,
		},
		{
			name: "user_idはある時だけ出る",
			setupContext: func(c *gin.Context) {
				c.Request = httptest.NewRequest(http.MethodGet, "/api/me", nil)
				c.Set(CtxKeyRequestID, "req-2")
				c.Set(CtxKeyRequestStartedAt, time.Now().Add(-10*time.Millisecond))
				c.Set(CtxKeyUserID, int64(42))
			},
			input: EventInput{
				Event:  "user_profile_fetched",
				Status: http.StatusOK,
				Level:  slog.LevelInfo,
			},
			wantRequestID:     "req-2",
			wantMethod:        http.MethodGet,
			wantRoute:         "/api/me",
			wantStatus:        http.StatusOK,
			wantEvent:         "user_profile_fetched",
			wantUserID:        int64(42),
			wantUserIDPresent: true,
			checkDuration:     true,
		},
		{
			name: "started_atが無い時もpanicしない",
			setupContext: func(c *gin.Context) {
				c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)
				c.Set(CtxKeyRequestID, "req-3")
			},
			input: EventInput{
				Event:  "health_checked",
				Status: http.StatusOK,
				Level:  slog.LevelInfo,
			},
			wantRequestID:     "req-3",
			wantMethod:        http.MethodGet,
			wantRoute:         "/health",
			wantStatus:        http.StatusOK,
			wantEvent:         "health_checked",
			wantUserIDPresent: false,
			checkDuration:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setupContext(c)

			attrs := BuildAttrs(c, tt.input)
			got := attrsToMap(attrs)

			if got["request_id"] != tt.wantRequestID {
				t.Fatalf("request_id mismatch: got=%v want=%v", got["request_id"], tt.wantRequestID)
			}
			if got["method"] != tt.wantMethod {
				t.Fatalf("method mismatch: got=%v want=%v", got["method"], tt.wantMethod)
			}
			if got["route"] != tt.wantRoute {
				t.Fatalf("route mismatch: got=%v want=%v", got["route"], tt.wantRoute)
			}
			gotStatus, ok := got["status"].(int64)
			if !ok {
				t.Fatalf("status type mismatch: got=%T", got["status"])
			}
			if int(gotStatus) != tt.wantStatus {
				t.Fatalf("status mismatch: got=%v want=%v", got["status"], tt.wantStatus)
			}
			if got["event"] != tt.wantEvent {
				t.Fatalf("event mismatch: got=%v want=%v", got["event"], tt.wantEvent)
			}

			rawUserID, exists := got["user_id"]
			if exists != tt.wantUserIDPresent {
				t.Fatalf("user_id presence mismatch: got=%v want=%v", exists, tt.wantUserIDPresent)
			}
			if tt.wantUserIDPresent && rawUserID != tt.wantUserID {
				t.Fatalf("user_id mismatch: got=%v want=%v", rawUserID, tt.wantUserID)
			}

			if tt.checkDuration {
				durationMS, ok := got["duration_ms"].(float64)
				if !ok {
					t.Fatalf("duration_ms type mismatch: got=%T", got["duration_ms"])
				}
				if durationMS < 0 {
					t.Fatalf("duration_ms should not be negative: got=%v", durationMS)
				}
			}

		})
	}
}

func attrsToMap(attrs []slog.Attr) map[string]any {
	result := make(map[string]any, len(attrs))
	for _, attr := range attrs {
		result[attr.Key] = attr.Value.Any()
	}
	return result
}

func TestLogEvent_UsesEventAsMessageWhenMessageEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orig := slog.Default()
	t.Cleanup(func() { slog.SetDefault(orig) })

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/login", nil)
	c.Set(CtxKeyRequestID, "req-1")
	c.Set(CtxKeyRequestStartedAt, time.Now().Add(-10*time.Millisecond))
	c.Set(CtxKeyUserID, int64(42))

	LogEvent(c, EventInput{
		Event:  "auth_login_succeeded",
		Status: http.StatusOK,
		Level:  slog.LevelInfo,
		// Messageは空
	})

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode log json: %v raw=%s", err, buf.String())
	}

	// slogのmessageキーは "msg"
	if got["msg"] != "auth_login_succeeded" {
		t.Fatalf("msg mismatch: got=%v want=%v", got["msg"], "auth_login_succeeded")
	}
	if got["event"] != "auth_login_succeeded" {
		t.Fatalf("event mismatch: got=%v want=%v", got["event"], "auth_login_succeeded")
	}
	if got["request_id"] != "req-1" {
		t.Fatalf("request_id mismatch: got=%v want=%v", got["request_id"], "req-1")
	}
}

func TestLogEvent_DefaultLevelInfoWhenLevelNotSpecified(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orig := slog.Default()
	t.Cleanup(func() { slog.SetDefault(orig) })

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)
	c.Set(CtxKeyRequestID, "req-2")
	c.Set(CtxKeyRequestStartedAt, time.Now().Add(-5*time.Millisecond))

	// Level未指定（ゼロ値）
	LogEvent(c, EventInput{
		Event:  "health_checked",
		Status: http.StatusOK,
	})

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode log json: %v raw=%s", err, buf.String())
	}

	if got["level"] != "INFO" {
		t.Fatalf("level mismatch: got=%v want=INFO", got["level"])
	}
	if got["msg"] != "health_checked" {
		t.Fatalf("msg mismatch: got=%v want=health_checked", got["msg"])
	}
}
