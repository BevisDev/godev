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
}

func TestNewConfig_InvalidDest(t *testing.T) {
	err := NewConfig(&Config{
		Path:       "./testdata",
		ConfigType: "yaml",
		Dest:       TestConfigStruct{}, // not pointer
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
		AutoEnv:    false,
	})
	assert.NoError(t, err)
	assert.Equal(t, "demo-app", cfg.AppName)
	assert.Equal(t, 8080, cfg.Port)
}

func TestNewConfig_AutoEnvOverride(t *testing.T) {
	os.Setenv("GO_PROFILE", "test")
	os.Setenv("APP_NAME", "env-app")
	os.Setenv("PORT", "9090")

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

func TestNewConfig_AssignProfile_WithEnvOverride(t *testing.T) {
	os.Setenv("APP_NAME", "env-app")
	os.Setenv("PORT", "9090")

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
