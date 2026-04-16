package config

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/cache"
)

// AppConfig holds configuration options loaded from TOML or defaulted at runtime.
type AppConfig struct {
	ListenAddr           string `toml:"listen_addr"`
	Port                 int    `toml:"port"`
	PatchNumber          string `toml:"-"` // Set dynamically upon initialization
	LanguageCode         string `toml:"language_code"`
	LevenshteinThreshold int    `toml:"levenshtein_threshold"`
	MerakiURL            string `toml:"meraki_url"`
	DDragonVersionURL    string `toml:"ddragon_version_url"`
	Debug                bool   `toml:"debug"`
	HTTPClientTimeout    int    `toml:"http_client_timeout"`
	CachePath            string `toml:"cache_path"`

	Logger     *log.Logger  `toml:"-"` // Exclude from TOML
	Cache      *cache.Cache `toml:"-"`
	HTTPClient *http.Client `toml:"-"`
	// Riot API configuration
	RiotAPIKey string `toml:"riot_api_key"`
	RiotRegion string `toml:"riot_region"`
}

// New returns an AppConfig with default values.
func New() *AppConfig {
	return &AppConfig{
		// Default to localhost
		ListenAddr:           "127.0.0.1",
		Port:                 1337,
		Debug:                true,
		LanguageCode:         "en_US",
		MerakiURL:            "https://cdn.merakianalytics.com/riot/lol/resources/latest/en-US/",
		DDragonVersionURL:    "https://ddragon.leagueoflegends.com/api/versions.json",
		LevenshteinThreshold: 3,
		CachePath:            "cache.json",
		HTTPClientTimeout:    10,
	}
}

// Validate checks configuration for common issues and logs warnings.
func (cfg *AppConfig) Validate(logger *log.Logger) {
	if cfg.RiotAPIKey == "" || cfg.RiotAPIKey == "YOUR_RIOT_API_KEY_HERE" {
		logger.Warn("Riot API key is missing or placeholder — player lookup and live game features will not work")
	}
	validRegions := map[string]bool{
		"na1": true, "br1": true, "la1": true, "la2": true,
		"euw1": true, "eun1": true, "ru": true, "tr1": true,
		"kr": true, "jp1": true,
		"oc1": true, "sg2": true, "tw2": true, "vn2": true,
	}
	if cfg.RiotRegion != "" && !validRegions[cfg.RiotRegion] {
		logger.Warnf("Riot region %q is not a recognized region", cfg.RiotRegion)
	}
}

// Initialize sets up logger, gin mode, cache, and HTTP client.
func (cfg *AppConfig) Initialize() error {
	cfg.setLogger()
	cfg.setGinMode()
	cfg.setCache()
	cfg.setHTTPClient()
	return nil
}

// setHTTPClient configures the HTTP client with a timeout from the config.
func (cfg *AppConfig) setHTTPClient() {
	timeout := time.Duration(cfg.HTTPClientTimeout) * time.Second
	cfg.HTTPClient = &http.Client{
		Timeout: timeout,
	}
}

// Load reads a TOML file into AppConfig.
func (cfg *AppConfig) Load(path string) error {
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return fmt.Errorf("error decoding config file %s: %w", path, err)
	}
	return nil
}

// LogConfig logs configuration keys and values as TOML after marshalling.
func (cfg *AppConfig) LogConfig() {
	tomlBytes, err := toml.Marshal(*cfg)
	if err != nil {
		cfg.Logger.Errorf("Failed to marshal config to TOML: %v", err)
		return
	}

	var data map[string]interface{}
	if err = toml.Unmarshal(tomlBytes, &data); err != nil {
		cfg.Logger.Errorf("Failed to unmarshal TOML: %v", err)
		return
	}

	cfg.Logger.Info("Configuration:")
	cfg.logMap("", data)
}

// logMap logs nested data structures from a map, used by LogConfig.
func (cfg *AppConfig) logMap(prefix string, data map[string]interface{}) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		switch v := value.(type) {
		case map[string]interface{}:
			cfg.logMap(fullKey, v)
		default:
			cfg.Logger.Infof("%s: %v", fullKey, v)
		}
	}
}

// setGinMode sets Gin to debug or release mode based on cfg.Debug.
func (cfg *AppConfig) setGinMode() {
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

// setLogger configures cfg.Logger based on the debug setting.
func (cfg *AppConfig) setLogger() {
	logger := log.New(os.Stderr)
	if cfg.Debug {
		logger.SetLevel(log.DebugLevel)
	}
	logger.SetReportTimestamp(true)
	cfg.Logger = logger
}

// setCache initializes the cache with config path and threshold.
func (cfg *AppConfig) setCache() {
	cfg.Cache = cache.New(cfg.CachePath, cfg.LevenshteinThreshold)
}
