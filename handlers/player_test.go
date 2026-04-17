package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/models"
)

func newTestPlayerHandler(transport http.RoundTripper) *PlayerHandler {
	cfg := config.New()
	cfg.Logger = log.New(os.Stderr)
	cfg.RiotRegion = "na1"
	cfg.RiotAPIKey = "test-api-key"

	httpClient := &http.Client{}
	if transport != nil {
		httpClient.Transport = transport
	}

	return &PlayerHandler{
		Logger: cfg.Logger,
		Config: cfg,
		Client: &client.Client{
			HTTPClient: httpClient,
			Logger:     cfg.Logger,
		},
	}
}

func TestPlayerGET(t *testing.T) {
	const acctJSON = `{"puuid":"test-puuid","gameName":"TestPlayer","tagLine":"NA1"}`
	const summonerJSON = `{"puuid":"test-puuid","profileIconId":1,"revisionDate":1700000000000,"summonerLevel":30}`

	tests := []struct {
		name       string
		query      string
		htmx       bool
		transport  http.RoundTripper
		wantStatus int
		wantBody   string
	}{
		{
			name:       "missing riotID HTMX returns 400",
			query:      "/player",
			htmx:       true,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Summoner identifier is required",
		},
		{
			name:       "missing riotID full page returns 200 with search form",
			query:      "/player",
			htmx:       false,
			wantStatus: http.StatusOK,
			wantBody:   "Player Lookup",
		},
		{
			name:       "invalid format HTMX returns error",
			query:      "/player?riotID=SomePlayer",
			htmx:       true,
			wantStatus: http.StatusOK,
			wantBody:   "Invalid format",
		},
		{
			name:       "invalid format full page returns page with error",
			query:      "/player?riotID=SomePlayer",
			htmx:       false,
			wantStatus: http.StatusOK,
			wantBody:   "Invalid format",
		},
		{
			name:  "player found HTMX returns 200 fragment",
			query: "/player?riotID=TestPlayer%23NA1",
			htmx:  true,
			transport: multiTransport{routes: map[string]*http.Response{
				"by-riot-id": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(acctJSON)),
				},
				"by-puuid": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(summonerJSON)),
				},
				"league": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("[]")),
				},
			}},
			wantStatus: http.StatusOK,
		},
		{
			name:  "player found full page returns 200 with layout",
			query: "/player?riotID=TestPlayer%23NA1",
			htmx:  false,
			transport: multiTransport{routes: map[string]*http.Response{
				"by-riot-id": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(acctJSON)),
				},
				"by-puuid": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(summonerJSON)),
				},
				"league": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("[]")),
				},
			}},
			wantStatus: http.StatusOK,
			wantBody:   "Player Lookup",
		},
		{
			name:  "account not found returns page with error",
			query: "/player?riotID=Unknown%23NA1",
			htmx:  false,
			transport: multiTransport{routes: map[string]*http.Response{
				"by-riot-id": {
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("")),
				},
			}},
			wantStatus: http.StatusOK,
			wantBody:   "not found",
		},
		{
			name:  "account permission denied returns page with error",
			query: "/player?riotID=TestPlayer%23NA1",
			htmx:  false,
			transport: multiTransport{routes: map[string]*http.Response{
				"by-riot-id": {
					StatusCode: http.StatusForbidden,
					Body:       io.NopCloser(strings.NewReader("Forbidden")),
				},
			}},
			wantStatus: http.StatusOK,
			wantBody:   "Permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestPlayerHandler(tt.transport)

			r := gin.New()
			r.GET("/player", h.PlayerGET)

			req := httptest.NewRequest(http.MethodGet, tt.query, nil)
			if tt.htmx {
				req.Header.Set("HX-Request", "true")
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d; body: %s", tt.wantStatus, w.Code, w.Body.String())
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got:\n%s", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestComputeChampionPool(t *testing.T) {
	matches := []models.MatchSummary{
		{ChampionName: "Ahri", ChampionID: 103, Win: true, Kills: 8, Deaths: 2, Assists: 10},
		{ChampionName: "Ahri", ChampionID: 103, Win: true, Kills: 6, Deaths: 3, Assists: 8},
		{ChampionName: "Ahri", ChampionID: 103, Win: false, Kills: 2, Deaths: 5, Assists: 4},
		{ChampionName: "Zed", ChampionID: 238, Win: true, Kills: 12, Deaths: 4, Assists: 3},
		{ChampionName: "Lux", ChampionID: 99, Win: false, Kills: 1, Deaths: 6, Assists: 12},
	}

	pool := computeChampionPool(matches, 5)

	if len(pool) != 3 {
		t.Fatalf("expected 3 champions, got %d", len(pool))
	}
	// First entry should be Ahri (3 games)
	if pool[0].ChampionName != "Ahri" {
		t.Errorf("expected first champion to be Ahri, got %s", pool[0].ChampionName)
	}
	if pool[0].Games != 3 {
		t.Errorf("expected 3 games for Ahri, got %d", pool[0].Games)
	}
	if pool[0].Wins != 2 {
		t.Errorf("expected 2 wins for Ahri, got %d", pool[0].Wins)
	}

	// Test topN limiting
	pool2 := computeChampionPool(matches, 1)
	if len(pool2) != 1 {
		t.Fatalf("expected 1 champion with topN=1, got %d", len(pool2))
	}

	// Test empty input
	pool3 := computeChampionPool(nil, 5)
	if len(pool3) != 0 {
		t.Errorf("expected 0 champions for nil input, got %d", len(pool3))
	}
}
