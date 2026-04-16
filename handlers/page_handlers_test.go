package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestPageHandler() *PageHandler {
	return &PageHandler{
		Logger: log.New(os.Stderr),
	}
}

func TestPageHandlers(t *testing.T) {
	h := newTestPageHandler()

	tests := []struct {
		name    string
		path    string
		handler gin.HandlerFunc
	}{
		{"HomePageGET", "/", h.HomePageGET},
		{"ChampionPageGET", "/champion", h.ChampionPageGET},
		{"PlayerPageGET", "/player", h.PlayerPageGET},
		{"LiveGamePageGET", "/livegame", h.LiveGamePageGET},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET(tt.path, tt.handler)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}
			if w.Body.Len() == 0 {
				t.Error("expected non-empty response body")
			}
			ct := w.Header().Get("Content-Type")
			if ct != "text/html; charset=utf-8" {
				t.Errorf("expected Content-Type text/html; charset=utf-8, got %q", ct)
			}
		})
	}
}
