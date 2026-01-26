package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/BevisDev/godev/utils/str"
	"github.com/spf13/viper"
)

// Config defines the input configuration for loading application settings from file and/or environment.
//
// It is typically used with Viper or similar tools to load a config file into a target struct.
type Config struct {
	Path       string // Path is the directory path where the config file is located (e.g., "./configs").
	Ext        string // Ext is the type of the config file (e.g., "yaml", "json", "toml").
	AutoEnv    bool   // AutoEnv is used for env overrides (APP_PORT overrides app.port)
	ReplaceEnv bool   // ReplaceEnv is used for replacing placeholders like "$DB_DSN"
	Profile    string // Profile is config file name (without extension), e.g., "dev", "prod".
}

type Response[T any] struct {
	Settings map[string]any
	Data     *T
}

// Load loads configuration and panics on failure.
// It reads the config file, applies env overrides, expands $VARS,
// and unmarshal the result into the target struct.
func Load[T any](cf *Config) (Response[T], error) {
	if cf == nil {
		return Response[T]{}, fmt.Errorf("config is nil")
	}

	v := viper.New()
	v.AddConfigPath(cf.Path)
	v.SetConfigName(cf.Profile)
	v.SetConfigType(cf.Ext)

	// BINDING ENV
	if cf.AutoEnv {
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()
	}

	// READ CONFIG
	if err := v.ReadInConfig(); err != nil {
		return Response[T]{}, fmt.Errorf("[config] failed to read: %v", err)
	}

	// ALL SETTINGS
	settings := v.AllSettings()

	// REPLACE ENV
	if cf.ReplaceEnv {
		replaceSettings(settings)
		err := v.MergeConfigMap(settings)
		if err != nil {
			return Response[T]{}, fmt.Errorf("[config] failed to merge: %v", err)
		}
	}

	// RETURN
	var out Response[T]
	if settings == nil || len(settings) <= 0 {
		return Response[T]{}, fmt.Errorf("[config] settings is empty")
	}

	var t T
	err := v.Unmarshal(&t)
	if err != nil {
		return Response[T]{}, fmt.Errorf("[config] failed to unmarshal: %v", err)
	}

	out.Data = &t
	out.Settings = v.AllSettings()
	return out, nil
}

func replaceSettings(data map[string]interface{}) {
	for k, v := range data {
		data[k] = replace(v)
	}
}

func replace(value interface{}) interface{} {
	switch val := value.(type) {
	case string:
		return os.ExpandEnv(val)

	case map[string]interface{}:
		for k, v := range val {
			val[k] = replace(v)
		}
		return val

	case []interface{}:
		for i, v := range val {
			val[i] = replace(v)
		}
		return val

	default:
		return value
	}
}

func MapStruct[T any](m map[string]string) (*T, error) {
	var dest T
	err := mapStruct(&dest, m)
	if err != nil {
		return nil, err
	}
	return &dest, nil
}

func mapStruct(target interface{}, cfMap map[string]string) error {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("[config] target must be a non-nil pointer to a struct")
	}

	v := rv.Elem()
	if v.Kind() != reflect.Struct {
		return errors.New("[config] target must point to a struct")
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		// struct
		if field.Kind() == reflect.Struct {
			if err := mapStruct(field.Addr().Interface(), cfMap); err != nil {
				return err
			}
			continue
		}

		// *struct
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}

			if field.Elem().Kind() == reflect.Struct {
				if err := mapStruct(field.Interface(), cfMap); err != nil {
					return err
				}
				continue
			}
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

			case reflect.Float32, reflect.Float64:
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
