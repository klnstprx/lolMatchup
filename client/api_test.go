package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// fakeTransport implements http.RoundTripper for testing.
type fakeTransport struct {
	resp *http.Response
	err  error
}

func (f fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f.resp, f.err
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(strings.NewReader(tt.body)),
			}
			c := &Client{HTTPClient: &http.Client{Transport: fakeTransport{resp: resp, err: nil}}}
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

func TestFetchActiveGame(t *testing.T) {
	const jsonBody = `{"gameId":42,"participants":[{"summonerName":"A"}]}`

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    error
	}{
		{"success", http.StatusOK, jsonBody, nil},
		{"not found", http.StatusNotFound, "", ErrGameNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(strings.NewReader(tt.body)),
			}
			c := &Client{HTTPClient: &http.Client{Transport: fakeTransport{resp: resp, err: nil}}}
			gm, err := c.FetchActiveGame(context.Background(), "id", "r", "k")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gm["gameId"] != float64(42) {
				t.Errorf("unexpected game data: %+v", gm)
			}
		})
	}
}

// TestDummy ensures the test suite runs at least one test.
func TestDummy(t *testing.T) {}
