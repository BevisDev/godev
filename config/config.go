package config

import (
	"errors"
	"github.com/BevisDev/godev/utils/validate"
	"github.com/spf13/viper"
	"os"
	"strings"
)

// Config defines the input configuration for loading application settings from file and/or environment.
//
// It is typically used with Viper or similar tools to load a config file into a target struct.
type Config struct {
	// Path is the directory path where the config file is located (e.g., "./configs").
	Path string

	// ConfigType specifies the type of the config file (e.g., "yaml", "json", "toml").
	ConfigType string

	// Dest is a pointer to a struct that will be populated with the configuration data.
	// This must be a pointer; otherwise, loading will fail.
	Dest interface{}

	// AutoEnv enables Viper's automatic environment variable binding.
	//
	// When enabled, Viper will automatically try to map environment variables to configuration keys.
	// Keys are matched by transforming them (e.g., dots to underscores, camelCase to UPPER_SNAKE_CASE),
	// and matched variables will override corresponding keys in the config.
	//
	// Example:
	//	Config key: "app.port"
	//	Environment: APP_PORT=9000
	//	Result: viper.Get("app.port") == 9000
	AutoEnv bool

	// ReplaceEnv determines whether environment variable placeholders in the config values
	// (e.g., "$APP_NAME") should be expanded using os.Getenv.
	//
	// When enabled, after reading the config file, all string values in the configuration
	// will be recursively scanned and any "$VAR" will be replaced with the value of the corresponding environment variable.
	//
	// This is useful when your config file contains placeholders like:
	//	app_name: $APP_NAME
	ReplaceEnv bool

	// Profile is the name of the config file to load (without extension), e.g., "dev", "prod".
	// It will be combined with Path and ConfigType to locate the file.
	Profile string
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
func NewConfig(cf *Config) error {
	if cf == nil {
		return errors.New("config is nil")
	}
	if !validate.IsPtr(cf.Dest) {
		return errors.New("must be a pointer")
	}

	profile := cf.GetProfile()
	v := viper.New()
	v.AddConfigPath(cf.Path)
	v.SetConfigName(profile)
	v.SetConfigType(cf.ConfigType)
	if cf.AutoEnv {
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()
	}

	// read config
	if err := v.ReadInConfig(); err != nil {
		return err
	}

	// read environment
	if cf.ReplaceEnv {
		settings := v.AllSettings()
		replaceEnvVars(settings)
		err := v.MergeConfigMap(settings)
		if err != nil {
			return err
		}
	}

	return v.Unmarshal(&cf.Dest)
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
