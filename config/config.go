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
	DDragonURL           string `toml:"ddragon_url"`
	DDragonURLData       string `toml:"-"` // Derived after patch is set
	DDragonVersionURL    string `toml:"ddragon_version_url"`
	Debug                bool   `toml:"debug"`
	HTTPClientTimeout    int    `toml:"http_client_timeout"`
	CachePath            string `toml:"cache_path"`

	Logger     *log.Logger  `toml:"-"` // Exclude from TOML
	Cache      *cache.Cache `toml:"-"`
	HTTPClient *http.Client `toml:"-"`
}

// New returns an AppConfig with default values.
func New() *AppConfig {
   return &AppConfig{
       // Default to localhost
       ListenAddr:           "127.0.0.1",
       Port:                 1337,
		Debug:                true,
		LanguageCode:         "en_US",
		DDragonURL:           "https://ddragon.leagueoflegends.com/cdn/",
		DDragonVersionURL:    "https://ddragon.leagueoflegends.com/api/versions.json",
		LevenshteinThreshold: 3,
		CachePath:            "cache.gob",
		HTTPClientTimeout:    10,
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

// SetDDragonDataURL builds the full URL for champion data after patch is known.
func (cfg *AppConfig) SetDDragonDataURL() {
	cfg.DDragonURLData = fmt.Sprintf("%s%s/data/%s/champion/", cfg.DDragonURL, cfg.PatchNumber, cfg.LanguageCode)
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
