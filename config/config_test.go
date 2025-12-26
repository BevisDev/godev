package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestMustLoad_MissingFile(t *testing.T) {
	assert.Panics(t, func() {
		_ = MustLoad[*TestConfigStruct](&Config{
			Path:      "./testdata1",
			Extension: "yaml",
			Profile:   "test",
		})
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

	assert.Equal(t, "demo-app", resp.Data.AppName)
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
		"APP_OVERRIDE":       "app_override", // it overrides
	})
	defer cleanup()

	out := MustLoad[TestConfigStruct](&Config{
		Path:      "./testdata",
		Extension: "yaml",
		Profile:   "test_env",
		AutoEnv:   true,
	})

	cfg := out.Data
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
		"PORT":     "9090", // optional: ensure it doesn't override if not expanded manually
	})
	defer cleanup()

	out := MustLoad[TestConfigStruct](&Config{
		Path:       "./testdata",
		Extension:  "yaml",
		Profile:    "test_replace_env",
		ReplaceEnv: true,
	})

	cfg := out.Data
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

func TestMustMapStruct_Panic(t *testing.T) {
	cfMap := map[string]string{
		"name": "abc",
	}

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
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("expected not panic, but function panics")
		}
	}()

	cfg := MustMapStruct[TestConfig](cfMap)
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

	// nested struct
	// struct
	if cfg.Server.Name != "ServerApp" {
		t.Errorf("expected Name=ServerApp, got %s", cfg.Server.Name)
	}
	if cfg.Server.Profile != "test_profile" {
		t.Errorf("expected profile=test_profile, got %s", cfg.Server.Profile)
	}

	// *struct
	if cfg.Client.Name != "ClientApp" {
		t.Errorf("expected Name=ClientApp, got %s", cfg.Client.Name)
	}
}
