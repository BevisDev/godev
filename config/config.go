package config

import (
	"errors"
	"github.com/BevisDev/godev/utils/str"
	"github.com/BevisDev/godev/utils/validate"
	"github.com/spf13/viper"
	"os"
	"reflect"
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
// Returns an error if the config file is missing, malformed, or the `Dest` is not a pointer.
//
// Example:
//
//		var appConfig AppConfig
//
//	 // to get profile flexible using environment
//		profile := os.Getenv("GO_PROFILE")
//
//		err := NewConfig(&Config{
//		  Path:       "./configs",
//		  ConfigType: "yaml",
//		  Dest:       &appConfig,
//		  Profile:    profile,
//		})
//		if err != nil {
//		  log.Fatalf("failed to load config: %v", err)
//		}
func NewConfig(cf *Config) error {
	if cf == nil {
		return errors.New("config is nil")
	}
	if !validate.IsPtr(cf.Dest) {
		return errors.New("must be a pointer")
	}

	v := viper.New()
	v.AddConfigPath(cf.Path)
	v.SetConfigName(cf.Profile)
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

func ReadValue(target interface{}, cfMap map[string]string) error {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("target must be a non-nil pointer to a struct")
	}

	v := rv.Elem()
	if v.Kind() != reflect.Struct {
		return errors.New("target must point to a struct")
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		key := t.Field(i).Tag.Get("config")
		if key == "" {
			continue
		}

		if val, ok := cfMap[key]; ok {
			switch field.Kind() {
			case reflect.String:
				field.SetString(val)

			case reflect.Int, reflect.Int32, reflect.Int64:
				n := str.ToInt[int64](val)
				field.SetInt(n)

			case reflect.Float32:
				f := str.ToFloat[float32](val)
				field.SetFloat(float64(f))

			case reflect.Float64:
				f := str.ToFloat[float64](val)
				field.SetFloat(f)

			case reflect.Bool:
				lower := strings.ToLower(strings.TrimSpace(val))
				switch lower {
				case "true", "1", "yes", "y":
					field.SetBool(true)
				default:
					field.SetBool(false)
				}

			case reflect.Slice:
				parts := strings.Split(val, ",")
				for j := range parts {
					parts[j] = strings.TrimSpace(parts[j])
				}
				elemKind := field.Type().Elem().Kind()

				switch elemKind {
				case reflect.String:
					field.Set(reflect.ValueOf(parts))

				case reflect.Int, reflect.Int32, reflect.Int64:
					var nums []int
					for _, p := range parts {
						n := str.ToInt[int](p)
						nums = append(nums, n)
					}
					field.Set(reflect.ValueOf(nums))

				case reflect.Float32, reflect.Float64:
					var floats []float64
					for _, p := range parts {
						f := str.ToFloat[float64](p)
						floats = append(floats, f)
					}
					field.Set(reflect.ValueOf(floats))
				default:
					continue
				}
			default:
				continue
			}
		}
	}

	return nil
}
