package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/cache"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/models"
)

// multiTransport dispatches responses based on request URL substring.
type multiTransport struct {
	routes map[string]*http.Response
}

func (m multiTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	for pattern, resp := range m.routes {
		if strings.Contains(url, pattern) {
			return resp, nil
		}
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

func newTestLiveGameHandler(transport http.RoundTripper) *LiveGameHandler {
	c := cache.New("", 3)
	c.SetChampionMap(map[string]string{
		"Aatrox": "Aatrox",
		"Ahri":   "Ahri",
	})
	c.SetChampionKeyMap(map[string]string{
		"266": "Aatrox",
		"103": "Ahri",
	})
	c.SetSummonerSpells(map[string]models.SummonerSpell{
		"4":  {Name: "Flash", Key: "4", ImageFull: "SummonerFlash.png", Cooldown: 300},
		"14": {Name: "Ignite", Key: "14", ImageFull: "SummonerDot.png", Cooldown: 180},
		"11": {Name: "Smite", Key: "11", ImageFull: "SummonerSmite.png", Cooldown: 15},
		"12": {Name: "Teleport", Key: "12", ImageFull: "SummonerTeleport.png", Cooldown: 360},
	})

	cfg := config.New()
	cfg.Logger = log.New(os.Stderr)
	cfg.Cache = c
	cfg.RiotRegion = "na1"
	cfg.RiotAPIKey = "test-key"

	httpClient := &http.Client{}
	if transport != nil {
		httpClient.Transport = transport
	}

	return NewLiveGameHandler(cfg, &client.Client{
		HTTPClient: httpClient,
		Logger:     cfg.Logger,
	})
}

func TestLiveGameGET_Validation(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantBody   string
	}{
		{"missing riotID", "/livegame", http.StatusBadRequest, "Summoner identifier is required"},
		{"empty riotID", "/livegame?riotID=", http.StatusBadRequest, "Summoner identifier is required"},
		{"no hash separator", "/livegame?riotID=PlayerNA1", http.StatusBadRequest, "Invalid format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestLiveGameHandler(nil)
			r := gin.New()
			r.GET("/livegame", h.LiveGameGET)

			req := httptest.NewRequest(http.MethodGet, tt.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got: %s", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestLiveGameGET_AccountNotFound(t *testing.T) {
	transport := fakeTransport{
		resp: &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("")),
		},
	}
	h := newTestLiveGameHandler(transport)
	r := gin.New()
	r.GET("/livegame", h.LiveGameGET)

	req := httptest.NewRequest(http.MethodGet, "/livegame?riotID=Unknown%23NA1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "not found") {
		t.Errorf("expected 'not found' in body, got: %s", w.Body.String())
	}
}

func TestLiveGameGET_AccountForbidden(t *testing.T) {
	transport := fakeTransport{
		resp: &http.Response{
			StatusCode: http.StatusForbidden,
			Body:       io.NopCloser(strings.NewReader("Forbidden")),
		},
	}
	h := newTestLiveGameHandler(transport)
	r := gin.New()
	r.GET("/livegame", h.LiveGameGET)

	req := httptest.NewRequest(http.MethodGet, "/livegame?riotID=Player%23NA1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestLiveGameGET_GameNotFound(t *testing.T) {
	acctJSON := `{"puuid":"abc-123","gameName":"Player","tagLine":"NA1"}`
	transport := multiTransport{
		routes: map[string]*http.Response{
			"account": {
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(acctJSON)),
			},
			"spectator": {
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("")),
			},
		},
	}
	h := newTestLiveGameHandler(transport)
	r := gin.New()
	r.GET("/livegame", h.LiveGameGET)

	req := httptest.NewRequest(http.MethodGet, "/livegame?riotID=Player%23NA1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d; body: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "not currently in a game") {
		t.Errorf("expected 'not currently in a game' in body, got: %s", w.Body.String())
	}
}

func TestLiveGameGET_Success(t *testing.T) {
	acctJSON := `{"puuid":"abc-123","gameName":"Player","tagLine":"NA1"}`

	game := models.CurrentGameInfo{
		GameStartTime: 1713300000000,
		Participants: []models.CurrentGameParticipant{
			{ChampionID: 266, TeamID: 100, RiotID: "Player#NA1", Spell1ID: 4, Spell2ID: 12},
			{ChampionID: 103, TeamID: 200, RiotID: "Enemy#NA1", Spell1ID: 4, Spell2ID: 14},
		},
		BannedChampions: []models.BannedChampion{
			{PickTurn: 1, ChampionID: 103, TeamID: 100},
			{PickTurn: 2, ChampionID: 266, TeamID: 200},
		},
	}
	gameJSON, err := json.Marshal(game)
	if err != nil {
		t.Fatalf("failed to marshal game: %v", err)
	}

	transport := multiTransport{
		routes: map[string]*http.Response{
			"account": {
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(acctJSON)),
			},
			"spectator": {
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(string(gameJSON))),
			},
		},
	}
	h := newTestLiveGameHandler(transport)
	r := gin.New()
	r.GET("/livegame", h.LiveGameGET)

	req := httptest.NewRequest(http.MethodGet, "/livegame?riotID=Player%23NA1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	if w.Body.Len() == 0 {
		t.Error("expected non-empty body")
	}
}
