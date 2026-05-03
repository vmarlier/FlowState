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

type Config struct {
	Server Server  `json:"server"`
	Routes []Route `json:"routes"`
}

type Server struct {
	ListenAddr     string `json:"listen_addr"`
	ReadTimeoutMs  int    `json:"read_timeout_ms,omitempty"`  // default 5000
	WriteTimeoutMs int    `json:"write_timeout_ms,omitempty"` // defualt 10000
	IdleTimeoutMs  int    `json:"idle_timeout_ms,omitempty"`  // default 60000
}

type Route struct {
	Host        string    `json:"host"`
	PathPrefix  string    `json:"path_prefix,omitempty"` // default /
	Methods     []string  `json:"methods,omitempty"`     // default GET, POST
	Middlewares []string  `json:"middlewares,omitempty"` // default none
	Backends    []Backend `json:"backends"`
}

type Backend struct {
	Url    string `json:"url"`
	Weight int    `json:"weight,omitempty"`
}

const (
	serverReadTimeoutMs  = 5000
	serverWriteTimeoutMs = 10000
	serverIdleTimeoutMs  = 60000
	routePathPrefix      = "/"
	routeBackendWeight   = 1
)

var routeMethods = []string{http.MethodGet, http.MethodPost}

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
	config.normalize()
	err = config.validate()
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) defaultValues() {
	if c.Server.ReadTimeoutMs == 0 {
		c.Server.ReadTimeoutMs = serverReadTimeoutMs
	}

	if c.Server.WriteTimeoutMs == 0 {
		c.Server.WriteTimeoutMs = serverWriteTimeoutMs
	}

	if c.Server.IdleTimeoutMs == 0 {
		c.Server.IdleTimeoutMs = serverIdleTimeoutMs
	}

	for r, route := range c.Routes {
		if route.PathPrefix == "" {
			c.Routes[r].PathPrefix = routePathPrefix
		}

		if route.Methods == nil || len(route.Methods) == 0 {
			c.Routes[r].Methods = routeMethods
		}

		for b, backend := range route.Backends {
			if backend.Weight == 0 {
				c.Routes[r].Backends[b].Weight = routeBackendWeight
			}
		}
	}
}

func (c *Config) normalize() {
	for r, route := range c.Routes {
		curatedMethods := normalizeMethods(route.Methods)
		c.Routes[r].Methods = curatedMethods
	}
}

func (c *Config) validate() error {
	if c.Server.ListenAddr == "" {
		return errors.New("Config: server.listen_addr must be set")
	}

	if c.Server.ReadTimeoutMs < 0 || c.Server.WriteTimeoutMs < 0 || c.Server.IdleTimeoutMs < 0 {
		return errors.New("Config: server.*_timeouts_ms must be positive values")
	}

	for _, route := range c.Routes {
		if route.Host == "" {
			return errors.New("Config: routes.host must be set")
		}

		err := validatePathPrefix(route.PathPrefix)
		if err != nil {
			return err
		}

		err = validateMethods(route.Methods)
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

func normalizeMethods(methods []string) []string {
	seen := map[string]struct{}{}
	curatedMethods := []string{}

	for _, method := range methods {
		m := strings.ToUpper(method)

		if _, ok := seen[m]; ok {
			continue
		}

		seen[m] = struct{}{}
		curatedMethods = append(curatedMethods, m)
	}

	return curatedMethods
}

func validatePathPrefix(s string) error {
	if !strings.HasPrefix(s, "/") {
		return errors.New("Config: routes.path_prefix must start with /")
	}

	if strings.ContainsAny(s, "?# ") {
		return errors.New("Config: routes.path_prefix must not contain query/fragments nor whitespaces")
	}

	return nil
}

func validateUrl(s string) error {
	url, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("Config: routes.backends.url must be a valid url: %w", err)
	}

	if url.Scheme == "" || url.Host == "" {
		return errors.New("Config: routes.backends.url must be a valid URL")
	}

	if url.Scheme != "http" && url.Scheme != "https" {
		return errors.New("Config: routes.backends.url must be http or https")
	}

	return nil
}

func validateMethods(methods []string) error {
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

	for _, method := range methods {
		if _, ok := validMethods[method]; !ok {
			return errors.New("Config: routes.methods must be a valid HTTP method")
		}
	}

	return nil
}
