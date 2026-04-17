package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/cache"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/config"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestConfig() *config.AppConfig {
	cfg := config.New()
	cfg.Logger = log.New(os.Stderr)
	cfg.Cache = cache.New("", 3)
	return cfg
}

func newTestPageHandler() *PageHandler {
	cfg := newTestConfig()
	apiClient := &client.Client{HTTPClient: &http.Client{}, Logger: cfg.Logger}
	ch := NewChampionHandler(cfg, apiClient)
	ph := NewPlayerHandler(cfg, apiClient)
	return &PageHandler{
		Logger:          cfg.Logger,
		ChampionHandler: ch,
		PlayerHandler:   ph,
	}
}

func TestPageHandlers(t *testing.T) {
	h := newTestPageHandler()

	cfg := newTestConfig()
	apiClient := &client.Client{HTTPClient: &http.Client{}, Logger: cfg.Logger}
	ch := NewChampionHandler(cfg, apiClient)
	ph := newTestPlayerHandler(nil)

	tests := []struct {
		name    string
		path    string
		handler gin.HandlerFunc
	}{
		{"HomePageGET", "/", h.HomePageGET},
		{"ChampionGET empty", "/champion", ch.ChampionGET},
		{"PlayerGET empty", "/player", ph.PlayerGET},
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

func TestSearchGET(t *testing.T) {
	h := newTestPageHandler()

	tests := []struct {
		name         string
		query        string
		htmx         bool
		wantStatus   int
		wantRedirect string // Location header for non-HTMX redirects
		wantBody     bool   // true if we expect an HTML body (HTMX proxy case)
	}{
		{
			name:         "empty query redirects to home",
			query:        "",
			wantStatus:   http.StatusFound,
			wantRedirect: "/",
		},
		{
			name:         "champion query redirects to champion route",
			query:        "Ahri",
			wantStatus:   http.StatusFound,
			wantRedirect: "/champion?champion=Ahri",
		},
		{
			name:         "player query redirects to player route",
			query:        "Faker#KR",
			wantStatus:   http.StatusFound,
			wantRedirect: "/player?riotID=Faker%23KR",
		},
		{
			name:       "HTMX champion query proxies to champion handler",
			query:      "Ahri",
			htmx:       true,
			wantStatus: http.StatusOK, // error rendered inline in response body
			wantBody:   true,
		},
		{
			name:       "HTMX player query proxies to player handler",
			query:      "Faker#KR",
			htmx:       true,
			wantStatus: http.StatusOK, // error rendered inline in response body
			wantBody:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/search", h.SearchGET)

			target := "/search?q=" + tt.query
			req := httptest.NewRequest(http.MethodGet, target, nil)
			if tt.htmx {
				req.Header.Set("HX-Request", "true")
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.wantBody {
				if w.Body.Len() == 0 {
					t.Error("expected non-empty response body for HTMX proxy")
				}
			} else if tt.wantStatus == http.StatusFound {
				got := w.Header().Get("Location")
				if got != tt.wantRedirect {
					t.Errorf("expected Location %q, got %q", tt.wantRedirect, got)
				}
			}
		})
	}
}
