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

func newTestPageHandler() *PageHandler {
	return &PageHandler{
		Logger: log.New(os.Stderr),
	}
}

func TestPageHandlers(t *testing.T) {
	h := newTestPageHandler()

	ph := newTestPlayerHandler(nil)

	cfg := config.New()
	cfg.Logger = log.New(os.Stderr)
	cfg.Cache = cache.New("", 3)
	apiClient := &client.Client{HTTPClient: &http.Client{}, Logger: cfg.Logger}
	ch := NewChampionHandler(cfg, apiClient)
	lgh := NewLiveGameHandler(cfg, apiClient)

	tests := []struct {
		name    string
		path    string
		handler gin.HandlerFunc
	}{
		{"HomePageGET", "/", h.HomePageGET},
		{"ChampionPageGET", "/champion", ch.ChampionPageGET},
		{"PlayerPageGET", "/player", ph.PlayerPageGET},
		{"LiveGamePageGET", "/livegame", lgh.LiveGamePageGET},
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
		name           string
		query          string
		htmx           bool
		wantStatus     int
		wantRedirect   string // Location header for non-HTMX, HX-Redirect for HTMX
	}{
		{
			name:         "empty query redirects to home",
			query:        "",
			wantStatus:   http.StatusFound,
			wantRedirect: "/",
		},
		{
			name:         "champion query redirects to champion search",
			query:        "Ahri",
			wantStatus:   http.StatusFound,
			wantRedirect: "/champion-search?champion=Ahri",
		},
		{
			name:         "player query redirects to player search",
			query:        "Faker#KR",
			wantStatus:   http.StatusFound,
			wantRedirect: "/player-search?riotID=Faker%23KR",
		},
		{
			name:         "HTMX champion query returns HX-Redirect",
			query:        "Ahri",
			htmx:         true,
			wantStatus:   http.StatusOK,
			wantRedirect: "/champion-search?champion=Ahri",
		},
		{
			name:         "HTMX player query returns HX-Redirect",
			query:        "Faker#KR",
			htmx:         true,
			wantStatus:   http.StatusOK,
			wantRedirect: "/player-search?riotID=Faker%23KR",
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

			if tt.htmx {
				got := w.Header().Get("HX-Redirect")
				if got != tt.wantRedirect {
					t.Errorf("expected HX-Redirect %q, got %q", tt.wantRedirect, got)
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
