package config

import (
	"errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type TestConfigStruct struct {
	AppName string `mapstructure:"app_name"`
	Port    int    `mapstructure:"port"`
	SomeKey struct {
		ClientName string `mapstructure:"clientName"`
		ClientPort int    `mapstructure:"ClientPort"`
		ClientKey  string `mapstructure:"Client_Key"`
	} `mapstructure:"someKey"`
	DatabaseAPP struct {
		Host string `mapstructure:"host"`
	} `mapstructure:"databaseAPP"`

	RedisAPP struct {
		Host string `mapstructure:"Host"`
	} `mapstructure:"RedisAPP"`
}

type TestConfig struct {
	Name      string    `config:"name"`
	Age       int       `config:"age"`
	Rate32    float32   `config:"rate32"`
	Rate64    float64   `config:"rate64"`
	Active    bool      `config:"active"`
	Tags      []string  `config:"tags"`
	Numbers   []int     `config:"numbers"`
	Threshold []float64 `config:"threshold"`
}

func setupEnv(vars map[string]string) func() {
	for k, v := range vars {
		os.Setenv(k, v)
	}
	return func() {
		for k := range vars {
			os.Unsetenv(k)
		}
	}
}

func TestNewConfig_MissingFile(t *testing.T) {
	cfg := &TestConfigStruct{}
	err := NewConfig(&Config{
		Path:       "./not_exist",
		ConfigType: "yaml",
		Dest:       cfg,
	})
	assert.Error(t, err)
	assert.True(t, errors.As(err, &viper.ConfigFileNotFoundError{}))
}

func TestNewConfig_InvalidDest(t *testing.T) {
	err := NewConfig(&Config{
		Path:       "./testdata",
		ConfigType: "yaml",
		Dest:       TestConfigStruct{},
	})
	assert.Error(t, err)
	assert.Equal(t, "must be a pointer", err.Error())
}

func TestNewConfig_LoadYAML_Success(t *testing.T) {
	cfg := &TestConfigStruct{}
	err := NewConfig(&Config{
		Path:       "./testdata",
		ConfigType: "yaml",
		Dest:       cfg,
		Profile:    "test",
	})
	assert.NoError(t, err)
	assert.Equal(t, "demo-app", cfg.AppName)
	assert.Equal(t, 8080, cfg.Port)
}

func TestNewConfig_AutoEnvOverride(t *testing.T) {
	cleanup := setupEnv(map[string]string{
		"GO_PROFILE": "test_env",
		"APP_NAME":   "env-app",
		"PORT":       "9090",
	})
	defer cleanup()

	cfg := &TestConfigStruct{}
	err := NewConfig(&Config{
		Path:       "./testdata",
		ConfigType: "yaml",
		Dest:       cfg,
		Profile:    "test_env",
		AutoEnv:    true,
	})
	assert.NoError(t, err)
	assert.Equal(t, "env-app", cfg.AppName)
	assert.Equal(t, 9090, cfg.Port)
}

func TestNewConfig_AutoEnv_MixedKeys(t *testing.T) {
	cleanup := setupEnv(map[string]string{
		"SOMEKEY_CLIENTNAME": "envName",
		"SOMEKEY_CLIENTPORT": "8888",
		"SOMEKEY_CLIENT_KEY": "xyz123",
		"DATABASEAPP_HOST":   "dbHost",
		"REDISAPP_HOST":      "redisHost",
	})
	defer cleanup()

	cfg := &TestConfigStruct{}
	err := NewConfig(&Config{
		Path:       "./testdata",
		ConfigType: "yaml",
		Dest:       cfg,
		Profile:    "test_env",
		AutoEnv:    true,
	})
	assert.NoError(t, err)
	assert.Equal(t, "envName", cfg.SomeKey.ClientName)
	assert.Equal(t, 8888, cfg.SomeKey.ClientPort)
	assert.Equal(t, "xyz123", cfg.SomeKey.ClientKey)
	assert.Equal(t, "dbHost", cfg.DatabaseAPP.Host)
	assert.Equal(t, "redisHost", cfg.RedisAPP.Host)
}

func TestNewConfig_AssignProfile_WithEnvOverride(t *testing.T) {
	cleanup := setupEnv(map[string]string{
		"APP_NAME": "env-app",
		"PORT":     "9090",
	})
	defer cleanup()

	cfg := &TestConfigStruct{}
	err := NewConfig(&Config{
		Path:       "./testdata",
		ConfigType: "yaml",
		Dest:       cfg,
		AutoEnv:    true,
		Profile:    "test_env",
	})
	assert.NoError(t, err)
	assert.Equal(t, "env-app", cfg.AppName)
	assert.Equal(t, 9090, cfg.Port)
}

func TestNewConfig_ReplaceEnv(t *testing.T) {
	cleanup := setupEnv(map[string]string{
		"APP_NAME": "expanded-app",
		"PORT":     "9090", // optional: ensure it doesn't override if not expanded manually
	})
	defer cleanup()

	cfg := &TestConfigStruct{}
	err := NewConfig(&Config{
		Path:       "./testdata",
		ConfigType: "yaml",
		Dest:       cfg,
		Profile:    "test_replace_env",
		ReplaceEnv: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, "expanded-app", cfg.AppName)
	assert.Equal(t, 8080, cfg.Port)
}

func equalSlice[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestReadValue_AllTypes(t *testing.T) {
	cfg := &TestConfig{}

	cfMap := map[string]string{
		"name":      "Alice",
		"age":       "30",
		"rate32":    "1.23",
		"rate64":    "4.56",
		"active":    "true",
		"tags":      "red, green ,blue",
		"numbers":   "1,2,3,4",
		"threshold": "0.1,0.5,1.5",
	}

	err := ReadValue(cfg, cfMap)
	if err != nil {
		t.Fatalf("ReadValue failed: %v", err)
	}

	// Check từng field
	if cfg.Name != "Alice" {
		t.Errorf("expected Name=Alice, got %s", cfg.Name)
	}
	if cfg.Age != 30 {
		t.Errorf("expected Age=30, got %d", cfg.Age)
	}
	if cfg.Rate32 < 1.229 || cfg.Rate32 > 1.231 { // float32 tolerance
		t.Errorf("expected Rate32 ~1.23, got %f", cfg.Rate32)
	}
	if cfg.Rate64 != 4.56 {
		t.Errorf("expected Rate64=4.56, got %f", cfg.Rate64)
	}
	if !cfg.Active {
		t.Errorf("expected Active=true, got false")
	}
	if !equalSlice(cfg.Tags, []string{"red", "green", "blue"}) {
		t.Errorf("expected Tags [red green blue], got %#v", cfg.Tags)
	}
	if !equalSlice(cfg.Numbers, []int{1, 2, 3, 4}) {
		t.Errorf("expected Numbers [1 2 3 4], got %#v", cfg.Numbers)
	}
	if !equalSlice(cfg.Threshold, []float64{0.1, 0.5, 1.5}) {
		t.Errorf("expected Threshold [0.1 0.5 1.5], got %#v", cfg.Threshold)
	}
}

func TestReadValue_InvalidCases(t *testing.T) {
	// Case 1: target không phải pointer
	cfg := TestConfig{}
	cfMap := map[string]string{}
	if err := ReadValue(cfg, cfMap); err == nil {
		t.Errorf("expected error when target is not pointer")
	}

	// Case 2: target nil
	var cfgNil *TestConfig
	if err := ReadValue(cfgNil, cfMap); err == nil {
		t.Errorf("expected error when target is nil")
	}

	// Case 3: parse lỗi (age = abc)
	cfg2 := &TestConfig{}
	cfMap2 := map[string]string{"age": "abc"}
	if err := ReadValue(cfg2, cfMap2); err != nil {
		t.Errorf("unexpected error for invalid int: %v", err)
	}
	// Expect Age=0 vì parse lỗi
	if cfg2.Age != 0 {
		t.Errorf("expected Age=0 for invalid int, got %d", cfg2.Age)
	}
}
