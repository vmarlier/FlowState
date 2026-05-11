package config

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("loads valid config", func(t *testing.T) {
		path := tempConfig(t, `{
		  "server": {
			"listen_addr": ":8080",
			"read_timeout_ms": 5000,
			"write_timeout_ms": 10000,
			"idle_timeout_ms": 60000
		  },
		  "routes": [
			{
			  "host": "api.local",
			  "path_prefix": "/users",
			  "methods": [
				"GET",
				"POST"
			  ],
			  "middlewares": [
				"request_id",
				"logging"
			  ],
			  "backends": [
				{
				  "url": "http://127.0.0.1:9002",
				  "weight": 1
				}
			  ]
			}
		  ]
		}`)
		defer os.Remove(path)

		_, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v, want nil", err)
		}
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		path := tempConfig(t, `{
		  "server": {
			"listen_addr" ":8080",
			"read_timeout_ms": 5000,
			"write_timeout_ms": 10000,
			"idle_timeout_ms": 60000
		  },
		  "routes": [
			}
		  ]
		}`)
		defer os.Remove(path)

		_, err := Load(path)
		if err == nil {
			t.Fatalf("Load() succeeded but wanted err")
		}
	})
}

func TestDefaultValue(t *testing.T) {
	t.Run("default for timeout ms", func(t *testing.T) {
		path := tempConfig(t, `{
		  "server": {
			"listen_addr": ":8080",
			"read_timeout_ms": 0
		  },
		  "routes": [
			{
			  "host": "api.local",
			  "methods": [],
			  "middlewares": [
				"request_id",
				"logging"
			  ],
			  "backends": [
				{
				  "url": "http://127.0.0.1:9001"
				}
			  ]
			}
		  ]
		}`)
		defer os.Remove(path)

		config, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v, want nil", err)
		}

		config.defaultValues()

		if config.Server.ReadTimeoutMs != serverReadTimeoutMs {
			t.Fatalf("defaultValues(), ReadTimeoutMs got: %d, wanted: %d", config.Server.ReadTimeoutMs, serverReadTimeoutMs)

		}

		if config.Server.WriteTimeoutMs != serverWriteTimeoutMs {
			t.Fatalf("defaultValues(), WriteTimeoutMs got: %d, wanted: %d", config.Server.WriteTimeoutMs, serverWriteTimeoutMs)

		}

		if config.Server.IdleTimeoutMs != serverIdleTimeoutMs {
			t.Fatalf("defaultValues(), IdleTimeoutMs got: %d, wanted: %d", config.Server.IdleTimeoutMs, serverIdleTimeoutMs)

		}

		if config.Routes[0].PathPrefix != routePathPrefix {
			t.Fatalf("defaultValues(), Routes.PathPrefix got: %s, wanted: %s", config.Routes[0].PathPrefix, routePathPrefix)
		}

		if !reflect.DeepEqual(config.Routes[0].Methods, routeMethods) {
			t.Fatalf("defaultValues(), Routes.Methods got: %v, wanted: %v", config.Routes[0].Methods, routeMethods)
		}

		if config.Routes[0].Backends[0].Weight != routeBackendWeight {
			t.Fatalf("defaultValues(), Routes.Backends.Weight got: %d, wanted: %d", config.Routes[0].Backends[0].Weight, routeBackendWeight)
		}
	})

}

func TestNormalize(t *testing.T) {
	path := tempConfig(t, `{
		  "server": {
			"listen_addr": ":8080",
			"read_timeout_ms": 5000,
			"write_timeout_ms": 10000,
			"idle_timeout_ms": 60000
		  },
		  "routes": [
			{
			  "host": "api.local",
			  "path_prefix": "/users",
			  "methods": [
				"get",
				"posT",
				"Post",
				"POST"
			  ],
			  "middlewares": [
				"request_id",
				"logging"
			  ],
			  "backends": [
				{
				  "url": "http://127.0.0.1:9002",
				  "weight": 1
				}
			  ]
			}
		  ]
		}`)
	defer os.Remove(path)

	config, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	config.normalize()
	if len(config.Routes[0].Methods) != 2 || config.Routes[0].Methods[0] != "GET" || config.Routes[0].Methods[1] != "POST" {
		t.Fatalf("normalize(), Routes.Methods got: %v, wanted: %v", config.Routes[0].Methods, []string{"GET", "POST"})
	}
}

func TestValidate(t *testing.T) {
	t.Run("rejects route without backends", func(t *testing.T) {
		config := &Config{
			Server: Server{ListenAddr: ":8080"},
			Routes: []Route{
				{
					Host:       "api.local",
					PathPrefix: "/",
					Methods:    []string{"GET"},
					Backends:   []Backend{}, // no backends
				},
			},
		}
		err := config.validate()
		if err == nil {
			t.Fatal("validate() err = nil, want error for missing backends")
		}
		var agg *AggregatedConfigErrors
		if !errors.As(err, &agg) {
			t.Fatalf("error type = %T, want *AggregatedConfigErrors", err)
		}
		found := false
		for _, e := range agg.ConfigErrors {
			if e.FieldPath == "routes[0].backends" && e.Kind == ConfigErrorKindMissingRequired {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected missing required error for routes[0].backends, got errors: %#v", agg.ConfigErrors)
		}
	})

	t.Run("rejects negative timeout values", func(t *testing.T) {
		config := &Config{
			Server: Server{
				ListenAddr:     ":8080",
				ReadTimeoutMs:  -1,
				WriteTimeoutMs: 10000,
				IdleTimeoutMs:  60000,
			},
			Routes: []Route{
				{
					Host:       "api.local",
					PathPrefix: "/",
					Methods:    []string{"GET"},
					Backends:   []Backend{{Url: "http://localhost:9000"}},
				},
			},
		}
		err := config.validate()
		if err == nil {
			t.Fatal("validate() err = nil, want error for negative timeout")
		}
		var agg *AggregatedConfigErrors
		if !errors.As(err, &agg) {
			t.Fatalf("error type = %T, want *AggregatedConfigErrors", err)
		}
		found := false
		for _, e := range agg.ConfigErrors {
			if e.FieldPath == "server.*_timeout_ms" && e.Kind == ConfigErrorKindOutOfRange {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected out_of_range error for server.*_timeout_ms, got errors: %#v", agg.ConfigErrors)
		}
	})

	t.Run("rejects backend with empty url", func(t *testing.T) {
		config := &Config{
			Server: Server{ListenAddr: ":8080"},
			Routes: []Route{
				{
					Host:       "api.local",
					PathPrefix: "/",
					Methods:    []string{"GET"},
					Backends:   []Backend{{Url: ""}},
				},
			},
		}
		err := config.validate()
		if err == nil {
			t.Fatal("validate() err = nil, want error for empty backend url")
		}
		var agg *AggregatedConfigErrors
		if !errors.As(err, &agg) {
			t.Fatalf("error type = %T, want *AggregatedConfigErrors", err)
		}
		found := false
		for _, e := range agg.ConfigErrors {
			if e.FieldPath == "routes[0].backends[0].url" && e.Kind == ConfigErrorKindMissingRequired {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected missing required error for routes[0].backends[0].url, got errors: %#v", agg.ConfigErrors)
		}
	})

	t.Run("accepts valid config", func(t *testing.T) {
		config := &Config{
			Server: Server{ListenAddr: ":8080", ReadTimeoutMs: 1, WriteTimeoutMs: 2, IdleTimeoutMs: 3},
			Routes: []Route{
				{
					Host:       "api.local",
					PathPrefix: "/",
					Methods:    []string{"GET"},
					Backends:   []Backend{{Url: "http://localhost:9000"}},
				},
			},
		}
		err := config.validate()
		if err != nil {
			t.Fatalf("validate() error = %v, want nil", err)
		}
	})
}

func tempConfig(t *testing.T, content string) string {
	f, err := os.CreateTemp("", "config-test.json")
	if err != nil {
		t.Fatalf("CreateTemp() error: %v", err)
	}

	if _, err := f.WriteString(content); err != nil {
		f.Close()
		t.Fatalf("WriteString() error = %v", err)
	}

	if err := f.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	return f.Name()
}
