package data

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/klnstprx/lolMatchup/cache"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/models"
)

// routingTransport dispatches responses based on URL substring matching.
type routingTransport struct {
	routes map[string]*http.Response
}

func (rt *routingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for pattern, resp := range rt.routes {
		if strings.Contains(req.URL.String(), pattern) {
			return resp, nil
		}
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

func makeResp(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

const versionsJSON = `["15.1.1","14.9.1"]`
const champListJSON = `{"Aatrox":{"id":266,"key":"Aatrox","name":"Aatrox"},"Ahri":{"id":103,"key":"Ahri","name":"Ahri"}}`

func newTestLoader(t *testing.T, transport http.RoundTripper, cachedPatch string) *DataLoader {
	t.Helper()
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "cache.json")

	cfg := config.New()
	cfg.Logger = log.New(io.Discard)
	cfg.Cache = cache.New(cachePath, 3)
	cfg.HTTPClient = &http.Client{Transport: transport}

	if cachedPatch != "" {
		cfg.Cache.SetPatch(cachedPatch)
	}

	apiClient := &client.Client{
		HTTPClient:        cfg.HTTPClient,
		Logger:            cfg.Logger,
		ChampionDataURL:   "http://fake.test/",
		DDragonVersionURL: "http://fake.test/versions.json",
	}

	return NewDataLoader(cfg, apiClient, cfg.Cache)
}

func TestInitialize_PatchChange(t *testing.T) {
	transport := &routingTransport{routes: map[string]*http.Response{
		"versions.json":  makeResp(200, versionsJSON),
		"champions.json": makeResp(200, champListJSON),
	}}

	dl := newTestLoader(t, transport, "14.9.1")

	if err := dl.Initialize(context.Background()); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	if got := dl.Cache.GetPatch(); got != "15.1.1" {
		t.Errorf("Patch: got %q, want %q", got, "15.1.1")
	}
	if dl.Config.PatchNumber != "15.1.1" {
		t.Errorf("PatchNumber: got %q, want %q", dl.Config.PatchNumber, "15.1.1")
	}
	if dl.Cache.GetChampionMapLen() != 2 {
		t.Errorf("ChampionMapLen: got %d, want 2", dl.Cache.GetChampionMapLen())
	}
}

func TestInitialize_SamePatch(t *testing.T) {
	transport := &routingTransport{routes: map[string]*http.Response{
		"versions.json": makeResp(200, `["15.1.1"]`),
	}}

	dl := newTestLoader(t, transport, "15.1.1")
	// Pre-populate champion map
	dl.Cache.SetChampionMap(map[string]string{"Aatrox": "Aatrox"})

	if err := dl.Initialize(context.Background()); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	// Should not have invalidated — champion map should still have Aatrox
	if dl.Cache.GetChampionMapLen() != 1 {
		t.Errorf("ChampionMapLen: got %d, want 1 (unchanged)", dl.Cache.GetChampionMapLen())
	}
}

func TestInitialize_SamePatchEmptyMap(t *testing.T) {
	transport := &routingTransport{routes: map[string]*http.Response{
		"versions.json":  makeResp(200, `["15.1.1"]`),
		"champions.json": makeResp(200, champListJSON),
	}}

	dl := newTestLoader(t, transport, "15.1.1")
	// ChampionMap is empty — should trigger repopulation

	if err := dl.Initialize(context.Background()); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	if dl.Cache.GetChampionMapLen() != 2 {
		t.Errorf("ChampionMapLen: got %d, want 2 (repopulated)", dl.Cache.GetChampionMapLen())
	}
}

func TestInitialize_OfflineFallback(t *testing.T) {
	transport := &routingTransport{routes: map[string]*http.Response{
		"versions.json": makeResp(500, "server error"),
	}}

	dl := newTestLoader(t, transport, "14.9.1")

	err := dl.Initialize(context.Background())
	if err != nil {
		t.Fatalf("expected no error with offline fallback, got: %v", err)
	}
	if dl.Config.PatchNumber != "14.9.1" {
		t.Errorf("PatchNumber: got %q, want cached %q", dl.Config.PatchNumber, "14.9.1")
	}
}

func TestInitialize_OfflineNoCachedPatch(t *testing.T) {
	transport := &routingTransport{routes: map[string]*http.Response{
		"versions.json": makeResp(500, "server error"),
	}}

	dl := newTestLoader(t, transport, "")

	err := dl.Initialize(context.Background())
	if err == nil {
		t.Fatal("expected error with no cached patch and offline, got nil")
	}
}

func TestBuildChampionMaps(t *testing.T) {
	champions := map[string]models.Champion{
		"Aatrox": {ID: 266, Key: "Aatrox", Name: "Aatrox"},
		"Ahri":   {ID: 103, Key: "Ahri", Name: "Ahri"},
	}

	nameMap, keyMap := buildChampionMaps(champions)

	if len(nameMap) != 2 {
		t.Errorf("nameMap len: got %d, want 2", len(nameMap))
	}
	if nameMap["Aatrox"] != "Aatrox" {
		t.Errorf("nameMap[Aatrox]: got %q, want %q", nameMap["Aatrox"], "Aatrox")
	}
	if nameMap["Ahri"] != "Ahri" {
		t.Errorf("nameMap[Ahri]: got %q, want %q", nameMap["Ahri"], "Ahri")
	}

	if len(keyMap) != 2 {
		t.Errorf("keyMap len: got %d, want 2", len(keyMap))
	}
	if keyMap["266"] != "Aatrox" {
		t.Errorf("keyMap[266]: got %q, want %q", keyMap["266"], "Aatrox")
	}
	if keyMap["103"] != "Ahri" {
		t.Errorf("keyMap[103]: got %q, want %q", keyMap["103"], "Ahri")
	}
}
