package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew_Defaults(t *testing.T) {
	cfg := New()

	checks := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"ListenAddr", cfg.ListenAddr, "127.0.0.1"},
		{"Port", cfg.Port, 1337},
		{"Debug", cfg.Debug, true},
		{"LanguageCode", cfg.LanguageCode, "en_US"},
		{"LevenshteinThreshold", cfg.LevenshteinThreshold, 3},
		{"CachePath", cfg.CachePath, "cache.json"},
		{"HTTPClientTimeout", cfg.HTTPClientTimeout, 10},
		{"MerakiURL", cfg.MerakiURL, "https://cdn.merakianalytics.com/riot/lol/resources/latest/en-US/"},
		{"DDragonVersionURL", cfg.DDragonVersionURL, "https://ddragon.leagueoflegends.com/api/versions.json"},
	}

	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestLoad_ValidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `listen_addr = "0.0.0.0"
port = 8080
debug = false
riot_api_key = "RGAPI-test"
riot_region = "euw1"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	cfg := New()
	if err := cfg.Load(path); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.ListenAddr != "0.0.0.0" {
		t.Errorf("ListenAddr: got %q, want %q", cfg.ListenAddr, "0.0.0.0")
	}
	if cfg.Port != 8080 {
		t.Errorf("Port: got %d, want %d", cfg.Port, 8080)
	}
	if cfg.Debug {
		t.Error("Debug: expected false after load")
	}
	if cfg.RiotAPIKey != "RGAPI-test" {
		t.Errorf("RiotAPIKey: got %q, want %q", cfg.RiotAPIKey, "RGAPI-test")
	}
	if cfg.RiotRegion != "euw1" {
		t.Errorf("RiotRegion: got %q, want %q", cfg.RiotRegion, "euw1")
	}
	// Unset fields should retain defaults
	if cfg.CachePath != "cache.json" {
		t.Errorf("CachePath should retain default, got %q", cfg.CachePath)
	}
}

func TestLoad_InvalidFile(t *testing.T) {
	cfg := New()
	err := cfg.Load("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestLoad_MalformedTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(path, []byte("{{{{not toml"), 0644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	cfg := New()
	err := cfg.Load(path)
	if err == nil {
		t.Error("expected error for malformed TOML, got nil")
	}
}

func TestInitialize(t *testing.T) {
	cfg := New()
	cfg.CachePath = filepath.Join(t.TempDir(), "cache.json")

	if err := cfg.Initialize(); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	if cfg.Logger == nil {
		t.Error("Logger is nil after Initialize")
	}
	if cfg.Cache == nil {
		t.Error("Cache is nil after Initialize")
	}
	if cfg.HTTPClient == nil {
		t.Error("HTTPClient is nil after Initialize")
	}
	if cfg.HTTPClient.Timeout.Seconds() != float64(cfg.HTTPClientTimeout) {
		t.Errorf("HTTPClient timeout: got %v, want %ds", cfg.HTTPClient.Timeout, cfg.HTTPClientTimeout)
	}
}

func TestValidate_MissingKey(t *testing.T) {
	cfg := New()
	cfg.Initialize()
	// Should not panic with empty key
	cfg.Validate(cfg.Logger)
}

func TestValidate_PlaceholderKey(t *testing.T) {
	cfg := New()
	cfg.Initialize()
	cfg.RiotAPIKey = "YOUR_RIOT_API_KEY_HERE"
	// Should not panic
	cfg.Validate(cfg.Logger)
}

func TestValidate_InvalidRegion(t *testing.T) {
	cfg := New()
	cfg.Initialize()
	cfg.RiotRegion = "invalid_region"
	// Should not panic
	cfg.Validate(cfg.Logger)
}
