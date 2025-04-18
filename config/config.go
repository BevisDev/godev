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
	AutoEnv    bool
	Profile    string
}

func NewConfig(config *Config) error {
	var (
		err     error
		profile = config.Profile
	)
	if !validate.IsPtr(config.Dest) {
		return errors.New("must be a pointer")
	}
	if profile == "" {
		p := os.Getenv("GO_PROFILE")
		if p == "" {
			profile = "dev"
		} else {
			profile = p
		}
	}

	v := viper.New()
	v.AddConfigPath(config.Path)
	v.SetConfigName(profile)
	v.SetConfigType(config.ConfigType)
	if config.AutoEnv {
		v.AutomaticEnv()
	}

	// read config
	if err = v.ReadInConfig(); err != nil {
		return err
	}

	// read environment
	if config.AutoEnv && profile != "dev" {
		settings := v.AllSettings()
		replaceEnvVars(settings)
		err = v.MergeConfigMap(settings)
		if err != nil {
			return err
		}
	}

	err = v.Unmarshal(&config.Dest)
	return err
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
