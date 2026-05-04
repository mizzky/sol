package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
)

var uuidFormatRe = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestRequestIDMiddleware_SingleRequest(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestIDMiddleware())

	r.GET("/test", func(c *gin.Context) {
		requestIDAny, exists := c.Get("request_id")
		if !exists {
			t.Fatal("request_id not found in context")
		}

		requestID, ok := requestIDAny.(string)
		if !ok {
			t.Fatalf("request_id is not string: %T", requestIDAny)
		}

		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status mismatch: got=%d want=%d", w.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response :%v", err)
	}

	id := body["request_id"]
	if id == "" {
		t.Fatal("request_id is empty")
	}
	if !uuidFormatRe.MatchString(id) {
		t.Fatalf("request_id is not UUID format: %s", id)
	}
}

func TestRequestIDMiddleware_DifferentIDPerRequest(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestIDMiddleware())
	r.GET("/test", func(c *gin.Context) {
		requestIDAny, exists := c.Get("request_id")
		if !exists {
			t.Fatal("request_id not found in context")
		}

		requestID, ok := requestIDAny.(string)
		if !ok {
			t.Fatalf("request_id is not string: %T", requestIDAny)
		}

		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	getID := func(t *testing.T) string {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status mismatch: got=%d want=%d", w.Code, http.StatusOK)
		}
		var body map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		id := body["request_id"]
		if id == "" {
			t.Fatal("request_id is empty")
		}
		if !uuidFormatRe.MatchString(id) {
			t.Fatalf("request_id is not UUID format: %s", id)
		}
		return id
	}

	id1 := getID(t)
	id2 := getID(t)

	if id1 == id2 {
		t.Fatalf("request_id must be dirrefent across requests: id1=%s id2=%s", id1, id2)
	}
}

func TestRequestIDMiddleware_IDIsStableInNextHandler(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestIDMiddleware())

	r.GET("/chain",
		func(c *gin.Context) {
			requestIDAny, exists := c.Get("request_id")
			if !exists {
				t.Fatal("request_id not found in context(first handler)")
			}

			requestID, ok := requestIDAny.(string)
			if !ok {
				t.Fatalf("request_id is not string in first handler: %T", requestIDAny)
			}
			c.Set("first_seen_request_id", requestID)
			c.Next()
		},
		func(c *gin.Context) {
			firstAny, exists := c.Get("first_seen_request_id")
			if !exists {
				t.Fatal("first_seen_request_id not found")
			}
			firstID, ok := firstAny.(string)
			if !ok {
				t.Fatalf("first_seen_request_id is not string:%T", firstAny)
			}

			currentAny, exists := c.Get("request_id")
			if !exists {
				t.Fatal("request_id not found in context(second handler)")
			}
			currentID, ok := currentAny.(string)
			if !ok {
				t.Fatalf("request_id is not string in second handler: %T", currentAny)
			}
			c.JSON(http.StatusOK, gin.H{"first_seen_request_id": firstID, "current_seen_request_id": currentID})
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/chain", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status mismatch: got=%d want=%d", w.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	firstID := body["first_seen_request_id"]
	currentID := body["current_seen_request_id"]

	if firstID == "" || currentID == "" {
		t.Fatalf("request_id must not be empty: first=%s current=%s", firstID, currentID)
	}
	if firstID != currentID {
		t.Fatalf("request_id changed in next handler: first=%s current=%s", firstID, currentID)
	}
}
