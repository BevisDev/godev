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

func (c *Config) GetProfile() string {
	if c.Profile != "" {
		return c.Profile
	}
	if p := os.Getenv("GO_PROFILE"); p != "" {
		return p
	}
	return "dev"
}

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
