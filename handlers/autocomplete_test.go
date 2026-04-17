package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/cache"
	"github.com/klnstprx/lolMatchup/config"
)

func newTestAutocompleteHandler() *AutocompleteHandler {
	c := cache.New("", 3)
	c.SetChampionMap(map[string]string{
		"Aatrox":     "Aatrox",
		"Ahri":       "Ahri",
		"Ashe":       "Ashe",
		"Akali":      "Akali",
		"Blitzcrank": "Blitzcrank",
		"Brand":      "Brand",
	})
	return &AutocompleteHandler{
		Logger: log.New(os.Stderr),
		Cache:  c,
		Config: &config.AppConfig{PatchNumber: "15.9.1"},
	}
}

func TestAutocompleteGET(t *testing.T) {
	h := newTestAutocompleteHandler()

	tests := []struct {
		name         string
		query        string
		wantStatus   int
		wantEmpty    bool   // expect empty/minimal body (no suggestions)
		wantContains string // substring expected in body if non-empty
	}{
		{
			name:       "empty query returns empty suggestions",
			query:      "",
			wantStatus: http.StatusOK,
			wantEmpty:  true,
		},
		{
			name:       "whitespace-only query returns empty suggestions",
			query:      "   ",
			wantStatus: http.StatusOK,
			wantEmpty:  true,
		},
		{
			name:         "matching query returns suggestions",
			query:        "Aa",
			wantStatus:   http.StatusOK,
			wantContains: "Aatrox",
		},
		{
			name:         "case insensitive match",
			query:        "ash",
			wantStatus:   http.StatusOK,
			wantContains: "Ashe",
		},
		{
			name:       "no match returns empty suggestions",
			query:      "Xyzzyplugh",
			wantStatus: http.StatusOK,
			wantEmpty:  true,
		},
		{
			name:         "prefix match on B",
			query:        "Bl",
			wantStatus:   http.StatusOK,
			wantContains: "Blitzcrank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/autocomplete", h.AutocompleteGET)

			target := "/autocomplete?champion=" + url.QueryEscape(tt.query)
			req := httptest.NewRequest(http.MethodGet, target, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}

			body := w.Body.String()
			if tt.wantContains != "" && !strings.Contains(body, tt.wantContains) {
				t.Errorf("expected body to contain %q, got:\n%s", tt.wantContains, body)
			}
			// For "empty" cases the body is just the empty component output;
			// we check it does NOT contain any champion name from our map.
			if tt.wantEmpty {
				for _, name := range []string{"Aatrox", "Ahri", "Ashe", "Akali", "Blitzcrank", "Brand"} {
					if strings.Contains(body, name) {
						t.Errorf("expected no suggestions, but body contains %q", name)
					}
				}
			}
		})
	}
}
