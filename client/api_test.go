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

func TestFetchSummonerByName(t *testing.T) {
	const jsonBody = `{"id":"1","accountId":"2","puuid":"3","name":"TestSumm","profileIconId":10,"revisionDate":123,"summonerLevel":5}`

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
			summ, err := c.FetchSummonerByName(context.Background(), "x", "r", "k")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if summ.Name != "TestSumm" || summ.ID != "1" {
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
