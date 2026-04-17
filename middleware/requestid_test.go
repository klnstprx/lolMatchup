package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("generates ID when none provided", func(t *testing.T) {
		r := gin.New()
		r.Use(RequestIDMiddleware())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)

		id := w.Header().Get("X-Request-ID")
		if id == "" {
			t.Error("expected X-Request-ID header, got empty")
		}
		if len(id) != 16 {
			t.Errorf("expected 16-char hex ID, got %q (len %d)", id, len(id))
		}
	})

	t.Run("preserves client-provided ID", func(t *testing.T) {
		r := gin.New()
		r.Use(RequestIDMiddleware())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", "client-id-123")
		r.ServeHTTP(w, req)

		id := w.Header().Get("X-Request-ID")
		if id != "client-id-123" {
			t.Errorf("expected preserved ID %q, got %q", "client-id-123", id)
		}
	})

	t.Run("unique IDs per request", func(t *testing.T) {
		r := gin.New()
		r.Use(RequestIDMiddleware())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		ids := make(map[string]bool)
		for range 10 {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			r.ServeHTTP(w, req)
			id := w.Header().Get("X-Request-ID")
			if ids[id] {
				t.Errorf("duplicate request ID: %s", id)
			}
			ids[id] = true
		}
	})
}
