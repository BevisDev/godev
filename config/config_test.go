package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestConfigStruct struct {
	AppName     string `mapstructure:"app_name"`
	Port        int    `mapstructure:"port"`
	AppOverride string `mapstructure:"app_override"`
	SomeKey     struct {
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
	Server    Server
	Client    *Client
}

type Server struct {
	Name    string `config:"server_name"`
	Profile string `config:"server_profile"`
}

type Client struct {
	Name string `config:"client_name"`
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

// =============================================================================
// MustLoad
// =============================================================================

func TestMustLoad_MissingFile(t *testing.T) {
	assert.Panics(t, func() {
		_ = MustLoad[*TestConfigStruct](&Config{
			Path:      "./testdata1",
			Extension: "yaml",
			Profile:   "test",
		})
	})
}

func TestMustLoad_NilConfig(t *testing.T) {
	assert.Panics(t, func() {
		_ = MustLoad[TestConfigStruct](nil)
	})
}

func TestMustLoad_InvalidTarget(t *testing.T) {
	assert.Panics(t, func() {
		_ = MustLoad[string](&Config{
			Path:      "./testdata",
			Extension: "yaml",
			Profile:   "test",
		})
	})
}

func TestMustLoad_Success(t *testing.T) {
	resp := MustLoad[TestConfigStruct](&Config{
		Path:      "./testdata",
		Extension: "yaml",
		Profile:   "test",
	})

	require.NotNil(t, resp.Data)
	assert.Equal(t, "demo-app", resp.Data.AppName)
	assert.NotEmpty(t, resp.Settings)
}

func TestMustLoad_AutoEnv(t *testing.T) {
	cleanup := setupEnv(map[string]string{
		"GO_PROFILE":         "test_env",
		"APP_NAME":           "env-app",
		"PORT":               "9090",
		"SOMEKEY_CLIENTNAME": "envName",
		"SOMEKEY_CLIENTPORT": "8888",
		"SOMEKEY_CLIENT_KEY": "xyz123",
		"DATABASEAPP_HOST":   "dbHost",
		"REDISAPP_HOST":      "redisHost",
		"APP_OVERRIDE":       "app_override",
	})
	defer cleanup()

	out := MustLoad[TestConfigStruct](&Config{
		Path:      "./testdata",
		Extension: "yaml",
		Profile:   "test_env",
		AutoEnv:   true,
	})

	cfg := out.Data
	require.NotNil(t, cfg)
	assert.Equal(t, "env-app", cfg.AppName)
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "envName", cfg.SomeKey.ClientName)
	assert.Equal(t, 8888, cfg.SomeKey.ClientPort)
	assert.Equal(t, "xyz123", cfg.SomeKey.ClientKey)
	assert.Equal(t, "dbHost", cfg.DatabaseAPP.Host)
	assert.Equal(t, "redisHost", cfg.RedisAPP.Host)
	assert.Equal(t, "app_override", cfg.AppOverride)
}

func TestMustLoad_ReplaceEnv(t *testing.T) {
	cleanup := setupEnv(map[string]string{
		"APP_NAME": "expanded-app",
		"PORT":     "9090",
	})
	defer cleanup()

	out := MustLoad[TestConfigStruct](&Config{
		Path:       "./testdata",
		Extension:  "yaml",
		Profile:    "test_replace_env",
		ReplaceEnv: true,
	})

	cfg := out.Data
	require.NotNil(t, cfg)
	assert.Equal(t, "expanded-app", cfg.AppName)
	assert.Equal(t, 8080, cfg.Port)
}

// =============================================================================
// MustMapStruct / MapStruct
// =============================================================================

func TestMustMapStruct_Panic(t *testing.T) {
	cfMap := map[string]string{"name": "abc"}

	assert.Panics(t, func() {
		_ = MustMapStruct[*TestConfig](cfMap)
	})
}

func TestMapStruct_AllTypes(t *testing.T) {
	cfMap := map[string]string{
		"name":           "Alice",
		"age":            "30",
		"rate32":         "1.23",
		"rate64":         "4.56",
		"active":         "true",
		"tags":           "red, green ,blue",
		"numbers":        "1,2,3,4",
		"threshold":      "0.1,0.5,1.5",
		"server_name":    "ServerApp",
		"server_profile": "test_profile",
		"client_name":    "ClientApp",
	}

	cfg := MustMapStruct[TestConfig](cfMap)
	require.NotNil(t, cfg)

	assert.Equal(t, "Alice", cfg.Name)
	assert.Equal(t, 30, cfg.Age)
	assert.InDelta(t, 1.23, float64(cfg.Rate32), 0.001)
	assert.Equal(t, 4.56, cfg.Rate64)
	assert.True(t, cfg.Active)
	assert.True(t, equalSlice(cfg.Tags, []string{"red", "green", "blue"}))
	assert.True(t, equalSlice(cfg.Numbers, []int{1, 2, 3, 4}))
	assert.True(t, equalSlice(cfg.Threshold, []float64{0.1, 0.5, 1.5}))
	assert.Equal(t, "ServerApp", cfg.Server.Name)
	assert.Equal(t, "test_profile", cfg.Server.Profile)
	require.NotNil(t, cfg.Client)
	assert.Equal(t, "ClientApp", cfg.Client.Name)
}

func TestMapStruct_EmptyMap(t *testing.T) {
	cfg := MustMapStruct[TestConfig](map[string]string{})
	require.NotNil(t, cfg)
	assert.Empty(t, cfg.Name)
	assert.Zero(t, cfg.Age)
	assert.False(t, cfg.Active)
	assert.Nil(t, cfg.Tags)
	assert.Nil(t, cfg.Numbers)
}

func TestMapStruct_UnknownKeysIgnored(t *testing.T) {
	cfMap := map[string]string{
		"name":        "Alice",
		"unknown_key": "ignored",
	}

	cfg := MustMapStruct[TestConfig](cfMap)
	require.NotNil(t, cfg)
	assert.Equal(t, "Alice", cfg.Name)
}

func TestMapStruct_BoolVariants(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"true", "true", true},
		{"1", "1", true},
		{"yes", "yes", true},
		{"y", "y", true},
		{"false", "false", false},
		{"0", "0", false},
		{"other", "other", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type C struct {
				Active bool `config:"active"`
			}
			cfg := MustMapStruct[C](map[string]string{"active": tt.input})
			require.NotNil(t, cfg)
			assert.Equal(t, tt.expected, cfg.Active)
		})
	}
}
