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

type AppConfig struct {
	ListenAddr           string       `toml:"listen_addr"`
	Port                 int          `toml:"port"`
	PatchNumber          string       `toml:"-"`
	LanguageCode         string       `toml:"language_code"`
	LevenshteinThreshold int          `toml:"levenshtein_threshold"`
	DDragonURL           string       `toml:"ddragon_url"`
	DDragonURLData       string       `toml:"-"`
	DDragonVersionURL    string       `toml:"ddragon_version_url"`
	Debug                bool         `toml:"debug"`
	HTTPClientTimeout    int          `toml:"http_client_timeout"`
	CachePath            string       `toml:"cache_path"`
	Logger               *log.Logger  `toml:"-"` // Exclude from TOML
	Cache                *cache.Cache `toml:"-"` // Exclude from TOML
	HTTPClient           *http.Client `toml:"-"` // Exclude from TOML
}

// initialize a new AppConfig struct with default values
func New() *AppConfig {
	return &AppConfig{
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

func (cfg *AppConfig) Initialize() error {
	cfg.SetLogger()
	cfg.SetGinMode()
	cfg.SetCache()
	cfg.SetHTTPClient()
	return nil
}

func (cfg *AppConfig) SetHTTPClient() {
	timeout := time.Duration(cfg.HTTPClientTimeout) * time.Second
	cfg.HTTPClient = &http.Client{
		Timeout: timeout,
	}
}

func (cfg *AppConfig) Load(path string) error {
	_, err := toml.DecodeFile(path, cfg)
	if err != nil {
		return fmt.Errorf("error decoding config file: %s", err)
	}
	return nil
}

func (cfg *AppConfig) LogConfig() {
	// Marshal cfg into TOML bytes
	tomlBytes, err := toml.Marshal(cfg)
	if err != nil {
		cfg.Logger.Errorf("Failed to marshal config to TOML: %s", err)
		return
	}

	// Unmarshal TOML bytes into a map
	var data map[string]interface{}
	err = toml.Unmarshal(tomlBytes, &data)
	if err != nil {
		cfg.Logger.Errorf("Failed to unmarshal TOML: %s", err)
		return
	}

	cfg.Logger.Info("Printing configuration...")
	// Iterate over the map and log each field individually
	cfg.logMap("", data)
}

// Recursive function to handle nested maps
func (cfg *AppConfig) logMap(prefix string, data map[string]interface{}) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		switch v := value.(type) {
		case map[string]interface{}:
			// Recursively log nested maps
			cfg.logMap(fullKey, v)
		default:
			// Log the field name and value
			cfg.Logger.Infof("%s: %v", fullKey, v)
		}
	}
}

func (cfg *AppConfig) SetPatchNumber(patch string) {
	cfg.PatchNumber = patch
}

func (cfg *AppConfig) SetDDragonDataURL() {
	cfg.DDragonURLData = fmt.Sprintf(cfg.DDragonURL+"%s/data/%s/champion/", cfg.PatchNumber, cfg.LanguageCode)
}

func (cfg *AppConfig) SetGinMode() {
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

func (cfg *AppConfig) SetLogger() {
	cfg.Logger = log.New(os.Stderr)
	if cfg.Debug {
		cfg.Logger.SetLevel(log.DebugLevel)
	}
	cfg.Logger.SetReportTimestamp(true)
}

func (cfg *AppConfig) SetCache() {
	cfg.Cache = cache.New(cfg.CachePath, cfg.LevenshteinThreshold)
}
