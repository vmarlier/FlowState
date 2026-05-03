package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// TODO:
// validate backend URL host
// simplify/fix path prefix check implementation

type Config struct {
	Server Server  `json:"server"`
	Routes []Route `json:"routes"`
}

type Server struct {
	Listen_addr      string `json:"listen_addr"`
	Read_timeout_ms  int    `json:"read_timeout_ms,omitempty"`  // default 5000
	Write_timeout_ms int    `json:"write_timeout_ms,omitempty"` // defualt 10000
	Idle_timeout_ms  int    `json:"idle_timeout_ms,omitempty"`  // default 60000
}

type Route struct {
	Host        string    `json:"host"`
	Path_prefix string    `json:"path_prefix,omitempty"` // default /
	Methods     []string  `json:"methods,omitempty"`     // default GET, POST
	Middlewares []string  `json:"middlewares,omitempty"` // default none
	Backends    []Backend `json:"backends"`
}

type Backend struct {
	Url    string `json:"url"`
	Weight int    `json:"weight,omitempty"`
}

const (
	server_read_timeout_ms  = 5000
	server_write_timeout_ms = 10000
	server_idle_timeout_ms  = 60000
	route_path_prefix       = "/"
	route_backend_weight    = 1
)

var route_methods = []string{http.MethodGet, http.MethodPost}

func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, err
	}

	config.defaultValues()
	err = config.validate()
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) defaultValues() {
	if c.Server.Read_timeout_ms == 0 {
		c.Server.Read_timeout_ms = server_read_timeout_ms
	}

	if c.Server.Write_timeout_ms == 0 {
		c.Server.Write_timeout_ms = server_write_timeout_ms
	}

	if c.Server.Idle_timeout_ms == 0 {
		c.Server.Idle_timeout_ms = server_idle_timeout_ms
	}

	for r, route := range c.Routes {
		if route.Path_prefix == "" {
			c.Routes[r].Path_prefix = route_path_prefix
		}

		if route.Methods == nil || len(route.Methods) == 0 {
			c.Routes[r].Methods = route_methods
		}

		for b, backend := range route.Backends {
			if backend.Weight == 0 {
				c.Routes[r].Backends[b].Weight = route_backend_weight
			}
		}
	}
}

func (c *Config) validate() error {
	if c.Server.Listen_addr == "" {
		return errors.New("Config: server.listen_addr must be set")
	}

	if c.Server.Read_timeout_ms < 0 || c.Server.Write_timeout_ms < 0 || c.Server.Idle_timeout_ms < 0 {
		return errors.New("Config: server.*_timeouts_ms must be positive values")
	}

	for _, route := range c.Routes {
		if route.Host == "" {
			return errors.New("Config: routes.host must be set")
		}

		err := validatePathPrefix(route.Path_prefix)
		if err != nil {
			return err
		}

		if len(route.Backends) == 0 {
			return errors.New("Config: each route must have at least one backend configured")
		}

		for _, backend := range route.Backends {
			if backend.Url == "" {
				return errors.New("Config: routes.backends.url must be set")
			}

			err := validateUrl(backend.Url)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func validatePathPrefix(s string) error {
	charPrefix := strings.Split(s, "")

	if charPrefix[0] != "/" {
		return errors.New("Config: routes.path_prefix must start with /")
	}

	if strings.ContainsAny(s, "?:# ") {
		return errors.New("Config: routes.path_prefix must not contain query/fragments nor whitespaces and must not be a url")
	}

	return nil
}

func validateUrl(s string) error {
	url, err := url.Parse(s)
	if err != nil {
		return errors.New("Config: routes.backends.url must be a valid url: " + fmt.Sprintf("%v", err.Error))
	}

	if url.Scheme != "http" && url.Scheme != "https" {
		return errors.New("Config: routes.backends.url must be http or https")
	}

	return nil
}

func validateMethods(methods []string) ([]string, error) {
	validMethods := map[string]struct{}{
		http.MethodGet:     struct{}{},
		http.MethodHead:    struct{}{},
		http.MethodPost:    struct{}{},
		http.MethodPut:     struct{}{},
		http.MethodPatch:   struct{}{},
		http.MethodDelete:  struct{}{},
		http.MethodConnect: struct{}{},
		http.MethodOptions: struct{}{},
		http.MethodTrace:   struct{}{},
	}

	seen := map[string]struct{}{}
	curatedMethods := []string{}

	for _, method := range methods {
		m := strings.ToUpper(method)

		if _, ok := validMethods[m]; !ok {
			return nil, errors.New("Config: routes.methods must be a valid HTTP method")
		}

		if _, ok := seen[m]; ok {
			continue
		}

		seen[m] = struct{}{}
		curatedMethods = append(curatedMethods, m)
	}

	return curatedMethods, nil
}
