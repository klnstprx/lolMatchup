package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/log"
)

type AppConfig struct {
	ListenAddr     string      `toml:"listen_addr"`
	Port           int         `toml:"port"`
	PatchNumber    string      `toml:"patch_number"`
	LanguageCode   string      `toml:"language_code"`
	DDragonURL     string      `toml:"ddragon_url"`
	DDragonURLData string      `toml:"-"`
	Debug          bool        `toml:"debug"`
	Logger         *log.Logger `toml:"-"` // Exclude Logger from TOML output
}

var App AppConfig

// Creates global AppConfig
func New() {
	App = AppConfig{}
}

// default config
func (cfg *AppConfig) Default() {
	cfg.ListenAddr = ""
	cfg.Port = 1337
	cfg.Debug = true
	cfg.LanguageCode = "en_US"
	cfg.DDragonURL = "https://ddragon.leagueoflegends.com/cdn/"
	cfg.PatchNumber = "14.21.1"
}

func (cfg *AppConfig) Load() {
	_, err := toml.DecodeFile("config.toml", cfg)
	if err != nil {
		cfg.Logger.Fatalf("Error decoding config file: %v", err)
	}
}

func (cfg *AppConfig) LogConfig() {
	// Marshal cfg into TOML bytes
	tomlBytes, err := toml.Marshal(cfg)
	if err != nil {
		cfg.Logger.Errorf("Failed to marshal config to TOML:", err)
		return
	}

	// Unmarshal TOML bytes into a map
	var data map[string]interface{}
	err = toml.Unmarshal(tomlBytes, &data)
	if err != nil {
		cfg.Logger.Errorf("Failed to unmarshal TOML:", err)
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

func (cfg *AppConfig) SetDDragonDataURL() {
	cfg.DDragonURLData = fmt.Sprintf(cfg.DDragonURL+"%s/data/%s/champion/", cfg.PatchNumber, cfg.LanguageCode)
}

func (cfg *AppConfig) SetLogger() {
	cfg.Logger = log.New(os.Stderr)
	if cfg.Debug {
		cfg.Logger.SetLevel(log.DebugLevel)
	}
	cfg.Logger.SetReportTimestamp(true)
}
