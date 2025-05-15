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
	os.Setenv("GO_PROFILE", "test")
	cfg := &TestConfigStruct{}
	err := NewConfig(&Config{
		Path:       "./testdata",
		ConfigType: "yaml",
		Dest:       cfg,
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
