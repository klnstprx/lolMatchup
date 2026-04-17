package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
)

// fakeTransport implements http.RoundTripper for testing.
type fakeTransport struct {
	resp *http.Response
	err  error
}

func (f fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f.resp, f.err
}

func newTestClient(transport fakeTransport) *Client {
	return &Client{
		HTTPClient:        &http.Client{Transport: transport},
		Logger:            log.New(os.Stderr),
		ChampionDataURL:   "http://fake.test/",
		DDragonVersionURL: "http://fake.test/versions.json",
	}
}

func TestFetchSummonerByPUUID(t *testing.T) {
	const jsonBody = `{"puuid":"test-puuid","profileIconId":10,"revisionDate":123,"summonerLevel":5}`

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    error
	}{
		{"success", http.StatusOK, jsonBody, nil},
		{"not found", http.StatusNotFound, "", ErrSummonerNotFound},
		{"forbidden", http.StatusForbidden, "Forbidden", ErrPermissionDenied},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(fakeTransport{
				resp: &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(strings.NewReader(tt.body)),
				},
			})
			summ, err := c.FetchSummonerByPUUID(context.Background(), "test-puuid", "euw1", "k")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if summ.PUUID != "test-puuid" || summ.SummonerLevel != 5 {
				t.Errorf("bad summoner parsed: %+v", summ)
			}
		})
	}
}

func TestFetchAccountByRiotID(t *testing.T) {
	const acctJSON = `{"puuid":"abc","gameName":"Player","tagLine":"NA1"}`

	tests := []struct {
		name       string
		statusCode int
		body       string
		region     string
		wantErr    error
	}{
		{"success", http.StatusOK, acctJSON, "na1", nil},
		{"not found", http.StatusNotFound, "", "euw1", ErrAccountNotFound},
		{"forbidden", http.StatusForbidden, "Forbidden", "kr", ErrPermissionDenied},
		{"unknown region fallback", http.StatusOK, acctJSON, "custom", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(fakeTransport{
				resp: &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(strings.NewReader(tt.body)),
				},
			})
			acct, err := c.FetchAccountByRiotID(context.Background(), "p", "t", tt.region, "k")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if acct.PUUID != "abc" {
				t.Errorf("expected PUUID 'abc', got %q", acct.PUUID)
			}
		})
	}
}

func TestFetchCurrentGameByPUUID(t *testing.T) {
	const gameJSON = `{"gameId":123,"participants":[{"championId":266,"teamId":100,"riotId":"P#1"}]}`

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    error
	}{
		{"success", http.StatusOK, gameJSON, nil},
		{"not found", http.StatusNotFound, "", ErrGameNotFound},
		{"forbidden", http.StatusForbidden, "Forbidden", ErrPermissionDenied},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(fakeTransport{
				resp: &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(strings.NewReader(tt.body)),
				},
			})
			game, err := c.FetchCurrentGameByPUUID(context.Background(), "puuid", "na1", "k")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if game.GameID != 123 {
				t.Errorf("expected GameID 123, got %d", game.GameID)
			}
			if len(game.Participants) != 1 {
				t.Errorf("expected 1 participant, got %d", len(game.Participants))
			}
		})
	}
}

func TestFetchChampionData(t *testing.T) {
	const champJSON = `{"id":266,"key":"Aatrox","name":"Aatrox","title":"the Darkin Blade"}`

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    error
	}{
		{"success", http.StatusOK, champJSON, nil},
		{"not found", http.StatusNotFound, "", ErrChampionNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(fakeTransport{
				resp: &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(strings.NewReader(tt.body)),
				},
			})
			champ, err := c.FetchChampionData(context.Background(), "Aatrox")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if champ.Name != "Aatrox" {
				t.Errorf("expected name 'Aatrox', got %q", champ.Name)
			}
		})
	}
}

func TestFetchChampionList(t *testing.T) {
	const listJSON = `{"Aatrox":{"id":266,"key":"Aatrox","name":"Aatrox"}}`

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    bool
	}{
		{"success", http.StatusOK, listJSON, false},
		{"server error", http.StatusInternalServerError, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(fakeTransport{
				resp: &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(strings.NewReader(tt.body)),
				},
			})
			champs, err := c.FetchChampionList(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if _, ok := champs["Aatrox"]; !ok {
				t.Error("expected Aatrox in champion list")
			}
		})
	}
}

func TestFetchLatestPatch(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantPatch  string
		wantErr    bool
	}{
		{"success", http.StatusOK, `["14.10.1","14.9.1"]`, "14.10.1", false},
		{"empty versions", http.StatusOK, `[]`, "", true},
		{"server error", http.StatusInternalServerError, "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(fakeTransport{
				resp: &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(strings.NewReader(tt.body)),
				},
			})
			patch, err := c.FetchLatestPatch(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if patch != tt.wantPatch {
				t.Errorf("expected patch %q, got %q", tt.wantPatch, patch)
			}
		})
	}
}

// capturingTransport records the request URL for inspection.
type capturingTransport struct {
	lastURL string
	resp    *http.Response
}

func (ct *capturingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ct.lastURL = req.URL.String()
	return ct.resp, nil
}

func TestURLEncoding(t *testing.T) {
	const okJSON = `{"id":"1","accountId":"2","puuid":"3","name":"Test","profileIconId":1,"revisionDate":0,"summonerLevel":1}`
	const acctJSON = `{"puuid":"abc","gameName":"Player","tagLine":"NA1"}`
	const gameJSON = `{"gameId":1,"participants":[]}`

	tests := []struct {
		name      string
		call      func(c *Client)
		wantInURL string
	}{
		{
			name: "summoner puuid with slash",
			call: func(c *Client) {
				c.FetchSummonerByPUUID(context.Background(), "abc/def", "na1", "k")
			},
			wantInURL: "abc%2Fdef",
		},
		{
			name: "riot id gameName with space",
			call: func(c *Client) {
				c.FetchAccountByRiotID(context.Background(), "Some Player", "NA1", "na1", "k")
			},
			wantInURL: "Some%20Player",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := &capturingTransport{
				resp: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(okJSON)),
				},
			}
			c := &Client{
				HTTPClient:        &http.Client{Transport: ct},
				Logger:            log.New(os.Stderr),
				ChampionDataURL:   "http://fake.test/",
				DDragonVersionURL: "http://fake.test/versions.json",
			}
			// Use acctJSON for account tests
			if strings.Contains(tt.name, "riot id") {
				ct.resp = &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(acctJSON)),
				}
			}
			if strings.Contains(tt.name, "game") {
				ct.resp = &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(gameJSON)),
				}
			}
			tt.call(c)
			if !strings.Contains(ct.lastURL, tt.wantInURL) {
				t.Errorf("expected URL to contain %q, got %q", tt.wantInURL, ct.lastURL)
			}
		})
	}
}

func TestFetchMatchIDs(t *testing.T) {
	const idsJSON = `["EUW1_123","EUW1_456","EUW1_789"]`

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantCount  int
		wantErr    bool
	}{
		{"success", http.StatusOK, idsJSON, 3, false},
		{"server error", http.StatusInternalServerError, "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(fakeTransport{
				resp: &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(strings.NewReader(tt.body)),
				},
			})
			ids, err := c.FetchMatchIDs(context.Background(), "puuid", "euw1", "k", 3, 0)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(ids) != tt.wantCount {
				t.Errorf("expected %d IDs, got %d", tt.wantCount, len(ids))
			}
		})
	}
}

func TestFetchMatch(t *testing.T) {
	const matchJSON = `{"metadata":{"matchId":"EUW1_123"},"info":{"gameDuration":1684,"gameMode":"CLASSIC","queueId":420,"gameStartTimestamp":1776410475512,"participants":[{"puuid":"test","championName":"Skarner","win":true,"kills":8,"deaths":4,"assists":20}]}}`

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    error
	}{
		{"success", http.StatusOK, matchJSON, nil},
		{"not found", http.StatusNotFound, "", ErrMatchNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(fakeTransport{
				resp: &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(strings.NewReader(tt.body)),
				},
			})
			match, err := c.FetchMatch(context.Background(), "EUW1_123", "euw1", "k")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if match.Metadata.MatchID != "EUW1_123" {
				t.Errorf("expected matchId EUW1_123, got %q", match.Metadata.MatchID)
			}
			if match.Info.GameDuration != 1684 {
				t.Errorf("expected duration 1684, got %d", match.Info.GameDuration)
			}
		})
	}
}

func TestRiotURL(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		hostPrefix string
		want       string
	}{
		{"default uses real API", "", "euw1", "https://euw1.api.riotgames.com"},
		{"custom base URL", "http://localhost:9090", "euw1", "http://localhost:9090"},
		{"trailing slash stripped", "http://localhost:9090/", "euw1", "http://localhost:9090"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{RiotAPIBaseURL: tt.baseURL}
			got := c.riotURL(tt.hostPrefix)
			if got != tt.want {
				t.Errorf("riotURL(%q) = %q, want %q", tt.hostPrefix, got, tt.want)
			}
		})
	}
}

func TestAPIError(t *testing.T) {
	err := &APIError{StatusCode: 500, Body: "internal"}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error string to contain status code, got: %s", err.Error())
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("expected errors.As to match *APIError")
	}
}
