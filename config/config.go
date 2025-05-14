package config

import (
	"errors"
	"github.com/BevisDev/godev/utils/validate"
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	Path       string
	ConfigType string
	Dest       interface{}
	BindEnv    bool
	Profile    string
}

// NewConfig loads configuration from a file and optionally merges environment variables.
//
// It uses the given `Config` struct to determine the file path, config name (profile),
// config type (e.g., json, yaml), and whether to bind environment variables.
//
// The `Dest` field in `Config` must be a pointer to a struct, which will be filled
// with the parsed config data using `viper.Unmarshal`.
//
// If `BindEnv` is true and the profile is not "dev", environment variables are merged
// with file-based configuration (after being sanitized).
//
// Returns an error if the config file is missing, malformed, or the `Dest` is not a pointer.
//
// Example:
//
//	var appConfig AppConfig
//	err := NewConfig(&Config{
//	    Path:       "./configs",
//	    ConfigType: "yaml",
//	    BindEnv:    true,
//	    Dest:       &appConfig,
//	})
//
//	if err != nil {
//	    log.Fatalf("failed to load config: %v", err)
//	}
func NewConfig(config *Config) error {
	if config == nil {
		return errors.New("config is nil")
	}
	if !validate.IsPtr(config.Dest) {
		return errors.New("must be a pointer")
	}

	profile := config.GetProfile()
	v := viper.New()
	v.AddConfigPath(config.Path)
	v.SetConfigName(profile)
	v.SetConfigType(config.ConfigType)

	// read config
	if err := v.ReadInConfig(); err != nil {
		return err
	}

	// read environment
	if config.BindEnv && profile != "dev" {
		settings := v.AllSettings()

		replaceEnvVars(settings)

		err := v.MergeConfigMap(settings)
		if err != nil {
			return err
		}
	}

	return v.Unmarshal(&config.Dest)
}

func (c *Config) GetProfile() string {
	if c.Profile != "" {
		return c.Profile
	}
	if p := os.Getenv("GO_PROFILE"); p != "" {
		return p
	}
	return "dev"
}

func replaceEnvVars(data interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			v[key] = processValue(value)
		}
	case []interface{}:
		for i, value := range v {
			v[i] = processValue(value)
		}
	}
}

func processValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return os.ExpandEnv(v)
	case map[string]interface{}:
		replaceEnvVars(v)
	case []interface{}:
		replaceEnvVars(v)
	}
	return value
}
