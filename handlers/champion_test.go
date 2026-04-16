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

// fakeTransport implements http.RoundTripper for controlling HTTP responses in tests.
type fakeTransport struct {
	resp *http.Response
	err  error
}

func (f fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f.resp, f.err
}

func newTestChampionHandler(transport http.RoundTripper) *ChampionHandler {
	c := cache.New("", 3)
	c.SetChampionMap(map[string]string{
		"Aatrox": "Aatrox",
		"Ahri":   "Ahri",
	})
	cfg := config.New()
	cfg.Logger = log.New(os.Stderr)
	cfg.Cache = c

	httpClient := &http.Client{}
	if transport != nil {
		httpClient.Transport = transport
	}

	return &ChampionHandler{
		Logger: cfg.Logger,
		Cache:  c,
		Config: cfg,
		Client: &client.Client{
			HTTPClient:      httpClient,
			Logger:          cfg.Logger,
			ChampionDataURL: "http://fake.test/",
		},
	}
}

func makeChampionJSON(t *testing.T) string {
	t.Helper()
	champ := models.Champion{
		ID:    266,
		Key:   "Aatrox",
		Name:  "Aatrox",
		Title: "the Darkin Blade",
		Icon:  "https://example.com/icon.png",
	}
	data, err := json.Marshal(champ)
	if err != nil {
		t.Fatalf("failed to marshal champion JSON: %v", err)
	}
	return string(data)
}

func TestChampionGET_MissingParam(t *testing.T) {
	h := newTestChampionHandler(nil)

	tests := []struct {
		name  string
		query string
	}{
		{"missing param", "/champion"},
		{"empty param", "/champion?champion="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/champion", h.ChampionGET)

			req := httptest.NewRequest(http.MethodGet, tt.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", w.Code)
			}
			if !strings.Contains(w.Body.String(), "Champion name is required") {
				t.Errorf("expected body to mention required champion name, got: %s", w.Body.String())
			}
		})
	}
}

func TestChampionGET_CacheHit(t *testing.T) {
	h := newTestChampionHandler(nil)
	h.Cache.SetChampion(models.Champion{
		ID:    266,
		Key:   "Aatrox",
		Name:  "Aatrox",
		Title: "the Darkin Blade",
	})

	r := gin.New()
	r.GET("/champion", h.ChampionGET)

	req := httptest.NewRequest(http.MethodGet, "/champion?champion=Aatrox", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d; body: %s", w.Code, w.Body.String())
	}
	if w.Body.Len() == 0 {
		t.Error("expected non-empty response body")
	}
}

func TestChampionGET_FetchFromAPI(t *testing.T) {
	champJSON := makeChampionJSON(t)
	transport := fakeTransport{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(champJSON)),
		},
	}
	h := newTestChampionHandler(transport)

	r := gin.New()
	r.GET("/champion", h.ChampionGET)

	req := httptest.NewRequest(http.MethodGet, "/champion?champion=Aatrox", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d; body: %s", w.Code, w.Body.String())
	}
	if w.Body.Len() == 0 {
		t.Error("expected non-empty response body")
	}
}

func TestChampionGET_NotFound(t *testing.T) {
	transport := fakeTransport{
		resp: &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("")),
		},
	}
	h := newTestChampionHandler(transport)

	r := gin.New()
	r.GET("/champion", h.ChampionGET)

	req := httptest.NewRequest(http.MethodGet, "/champion?champion=NonExistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d; body: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "not found") {
		t.Errorf("expected body to mention 'not found', got: %s", w.Body.String())
	}
}
