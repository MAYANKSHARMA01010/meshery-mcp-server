// Package config loads and validates meshery-mcp-server configuration.
//
// Precedence, lowest to highest: defaults < config file < environment variables.
package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

const (
	keyMesheryURL   = "meshery.url"
	keyMesheryToken = "meshery.token"
	keyLogLevel     = "log.level"

	envMesheryURL   = "MESHERY_SERVER_URL"
	envMesheryToken = "MESHERY_API_TOKEN" //nolint:gosec // G101: env var name, not a credential
	envLogLevel     = "LOG_LEVEL"

	defaultMesheryURL = "http://localhost:9081"
	defaultLogLevel   = LogLevelInfo
)

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// Config is the fully resolved, validated configuration.
type Config struct {
	Meshery MesheryConfig `mapstructure:"meshery"`
	Log     LogConfig     `mapstructure:"log"`
}

// MesheryConfig describes how to reach the Meshery server.
type MesheryConfig struct {
	URL   string `mapstructure:"url"`
	Token Secret `mapstructure:"token"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level string `mapstructure:"level"` // one of the LogLevel* constants
}

// Load builds the configuration. configPath is optional; an empty string
// means no file is read.
func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetDefault(keyMesheryURL, defaultMesheryURL)
	v.SetDefault(keyMesheryToken, "")
	v.SetDefault(keyLogLevel, defaultLogLevel)

	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, ErrReadConfigFile(err, configPath)
		}
	}

	// MustBindEnv only fails when called without a key, which would be a bug
	// here, not something a user can fix; panicking is appropriate.
	v.MustBindEnv(keyMesheryURL, envMesheryURL)
	v.MustBindEnv(keyMesheryToken, envMesheryToken)
	v.MustBindEnv(keyLogLevel, envLogLevel)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, ErrDecodeConfig(err)
	}
	cfg.Log.Level = strings.ToLower(strings.TrimSpace(cfg.Log.Level))

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	u, err := url.Parse(c.Meshery.URL)
	if err != nil {
		return ErrInvalidMesheryURL("The URL could not be parsed")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrInvalidMesheryURL(fmt.Sprintf("Scheme of %v must be http or https", u.Redacted()))
	}
	if u.Host == "" {
		return ErrInvalidMesheryURL(fmt.Sprintf("URL %v has no host", u.Redacted()))
	}

	switch c.Log.Level {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
	default:
		return ErrInvalidLogLevel(c.Log.Level)
	}
	return nil
}
