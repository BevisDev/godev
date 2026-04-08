package config

import (
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"

	"github.com/BevisDev/godev/utils/str"
	"github.com/BevisDev/godev/utils/validate"
	"github.com/go-viper/mapstructure/v2"
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
	Data     T
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
	if validate.IsNilOrEmpty(settings) {
		return Response[T]{}, fmt.Errorf("[config] settings is empty")
	}

	var t T
	err := v.Unmarshal(&t)
	if err != nil {
		return Response[T]{}, fmt.Errorf("[config] failed to unmarshal: %v", err)
	}

	out.Data = t
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

func MapStruct[T any](m map[string]string) (T, error) {
	var dest T
	err := mapStruct(&dest, m)
	return dest, err
}

// MapStructAny maps values from map[string]interface{} into a struct based on `config` tags.
//
// Notes:
// - Similar behavior to MapStruct: it only sets fields that have tag `config:"key"`.
// - For slices, it supports:
//   - string values (comma-separated, e.g. "a,b,c")
//   - slice/array values (e.g. []interface{}{"a","b"}), converting elements to the target element kind.
func MapStructAny[T any](m map[string]any) (T, error) {
	var dest T
	err := mapStructAny(&dest, m)
	return dest, err
}

// MapStructViper maps map values into a struct using Viper's Unmarshal
// (which internally uses mapstructure).
//
// By default it uses struct tags `config:"..."` (not `mapstructure:"..."`).
// It also enables WeaklyTypedInput and provides a DecodeHook for:
// - string "a,b,c" -> []T (splits by ',' and trims spaces)
// - string "yes"/"y" -> true (when target kind is bool)
func MapStructViper[T any](m map[string]any) (T, error) {
	return MapStructViperTag[T](m, "config")
}

// MapStructViperTag is the same as MapStructViper but lets you choose the struct tag name.
func MapStructViperTag[T any](m map[string]any, tagName string) (T, error) {
	var dest T

	v := viper.New()
	hook := func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		// string -> bool (support common "yes"/"y" variants)
		if from.Kind() == reflect.String && to.Kind() == reflect.Bool {
			s := strings.ToLower(strings.TrimSpace(data.(string)))
			switch s {
			case "yes", "y":
				return "true", nil
			case "no", "n":
				return "false", nil
			default:
				return data, nil
			}
		}

		// string -> slice (CSV)
		if from.Kind() == reflect.String && to.Kind() == reflect.Slice {
			s := data.(string)
			parts := strings.Split(s, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			return parts, nil // let mapstructure/weak decode convert []string -> []int/[]float/... as needed
		}

		return data, nil
	}

	err := v.Unmarshal(
		m,
		func(dc *mapstructure.DecoderConfig) {
			dc.TagName = tagName
			dc.WeaklyTypedInput = true
			dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				hook,
			)
		},
	)

	return dest, err
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

func mapStructAny(target interface{}, cfMap map[string]interface{}) error {
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
			if err := mapStructAny(field.Addr().Interface(), cfMap); err != nil {
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
				if err := mapStructAny(field.Interface(), cfMap); err != nil {
					return err
				}
				continue
			}
		}

		key := t.Field(i).Tag.Get("config")
		if key == "" {
			continue
		}

		val, ok := cfMap[key]
		if !ok {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(coerceToString(val))

		case reflect.Int, reflect.Int32, reflect.Int64:
			field.SetInt(coerceToInt64(val))

		case reflect.Float32, reflect.Float64:
			field.SetFloat(coerceToFloat64(val))

		case reflect.Bool:
			field.SetBool(coerceToBool(val))

		case reflect.Slice:
			setSliceFromAny(field, val)
		default:
			// unsupported kind => ignore
		}
	}

	return nil
}

func coerceToString(val interface{}) string {
	if val == nil {
		return ""
	}
	return str.ToString(val)
}

func coerceToInt64(val interface{}) int64 {
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case float32:
		f := float64(v)
		if f != math.Trunc(f) {
			return 0
		}
		return int64(f)
	case float64:
		if v != math.Trunc(v) {
			return 0
		}
		return int64(v)
	case bool:
		if v {
			return 1
		}
		return 0
	case string:
		return str.ToInt[int64](v)
	default:
		return str.ToInt[int64](coerceToString(val))
	}
}

func coerceToFloat64(val interface{}) float64 {
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case float32:
		return float64(v)
	case float64:
		return v
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case bool:
		if v {
			return 1
		}
		return 0
	case string:
		return str.ToFloat[float64](v)
	default:
		return str.ToFloat[float64](coerceToString(val))
	}
}

func coerceToBool(val interface{}) bool {
	if val == nil {
		return false
	}

	switch v := val.(type) {
	case bool:
		return v
	case int, int8, int16, int32, int64:
		return coerceToInt64(v) != 0
	case uint, uint8, uint16, uint32, uint64:
		return coerceToInt64(v) != 0
	case float32, float64:
		return coerceToFloat64(v) != 0
	case string:
		lower := strings.ToLower(strings.TrimSpace(v))
		switch lower {
		case "true", "1", "yes", "y":
			return true
		default:
			return false
		}
	default:
		// fallback: try parsing stringified value
		return coerceToBool(coerceToString(val))
	}
}

func setSliceFromAny(field reflect.Value, val interface{}) {
	if val == nil {
		return
	}

	elemType := field.Type().Elem()
	elemKind := elemType.Kind()

	// Case 1: "a,b,c" as comma-separated string.
	if s, ok := val.(string); ok {
		parts := strings.Split(s, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}

		out := reflect.MakeSlice(field.Type(), 0, len(parts))
		for _, p := range parts {
			cv := coerceToElemValue(elemType, elemKind, p)
			out = reflect.Append(out, cv)
		}
		field.Set(out)
		return
	}

	// Case 2: slice/array.
	rv := reflect.ValueOf(val)
	if !rv.IsValid() || (rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array) {
		// unsupported => ignore
		return
	}

	out := reflect.MakeSlice(field.Type(), 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		cv := coerceToElemValue(elemType, elemKind, item)
		out = reflect.Append(out, cv)
	}
	field.Set(out)
}

func coerceToElemValue(elemType reflect.Type, elemKind reflect.Kind, val interface{}) reflect.Value {
	switch elemKind {
	case reflect.String:
		return reflect.ValueOf(coerceToString(val)).Convert(elemType)

	case reflect.Int:
		return reflect.ValueOf(int(coerceToInt64(val))).Convert(elemType)
	case reflect.Int32:
		return reflect.ValueOf(int32(coerceToInt64(val))).Convert(elemType)
	case reflect.Int64:
		return reflect.ValueOf(coerceToInt64(val)).Convert(elemType)

	case reflect.Float32:
		return reflect.ValueOf(float32(coerceToFloat64(val))).Convert(elemType)
	case reflect.Float64:
		return reflect.ValueOf(coerceToFloat64(val)).Convert(elemType)
	default:
		return reflect.Zero(elemType)
	}
}
