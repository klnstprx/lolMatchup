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
	const summonerJSON = `{"id":"abc","accountId":"def","puuid":"ghi","name":"TestPlayer","profileIconId":1,"revisionDate":1700000000000,"summonerLevel":30}`

	tests := []struct {
		name       string
		query      string
		transport  *fakeTransport
		wantStatus int
		wantBody   string
	}{
		{
			name:       "missing riotID returns 400",
			query:      "/player",
			wantStatus: http.StatusBadRequest,
			wantBody:   "Summoner identifier is required",
		},
		{
			name:       "empty riotID returns 400",
			query:      "/player?riotID=",
			wantStatus: http.StatusBadRequest,
			wantBody:   "Summoner identifier is required",
		},
		{
			name:       "invalid format without hash returns 400",
			query:      "/player?riotID=SomePlayer",
			wantStatus: http.StatusBadRequest,
			wantBody:   "Invalid format",
		},
		{
			name:  "summoner found returns 200",
			query: "/player?riotID=TestPlayer%23NA1",
			transport: &fakeTransport{
				resp: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(summonerJSON)),
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "summoner not found returns 404",
			query: "/player?riotID=Unknown%23NA1",
			transport: &fakeTransport{
				resp: &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("")),
				},
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "not found",
		},
		{
			name:  "permission denied returns 403",
			query: "/player?riotID=TestPlayer%23NA1",
			transport: &fakeTransport{
				resp: &http.Response{
					StatusCode: http.StatusForbidden,
					Body:       io.NopCloser(strings.NewReader("Forbidden")),
				},
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "Permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var transport http.RoundTripper
			if tt.transport != nil {
				transport = tt.transport
			}
			h := newTestPlayerHandler(transport)

			r := gin.New()
			r.GET("/player", h.PlayerGET)

			req := httptest.NewRequest(http.MethodGet, tt.query, nil)
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
