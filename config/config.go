package config

import (
	"github.com/spf13/viper"
	"os"
)

func GetConfig(dest interface{}, path string) error {
	var (
		err     error
		profile = os.Getenv("GO_PROFILE")
	)
	if profile == "" {
		profile = "dev" // set default
	}

	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName(profile)
	v.SetConfigType("yml")
	v.AutomaticEnv()

	// read config
	if err = v.ReadInConfig(); err != nil {
		return err
	}

	// read environment
	if profile != "dev" {
		settings := v.AllSettings()
		replaceEnvVars(settings)
		err = v.MergeConfigMap(settings)
		if err != nil {
			return err
		}
	}

	err = v.Unmarshal(&dest)
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
