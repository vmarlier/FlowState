package config

import (
	"encoding/json"
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
		return nil, NewSystemError("failed to open config file", err)
	}

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, NewSystemError("failed to decode config file", err)
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
	errs := NewAggregatedConfigErrors()

	if c.Server.ListenAddr == "" {
		errs.AddConfigError(*NewConfigError("server.listen_addr", "listen_addr must be set", ConfigErrorKindMissingRequired, nil))
	}

	if c.Server.ReadTimeoutMs < 0 || c.Server.WriteTimeoutMs < 0 || c.Server.IdleTimeoutMs < 0 {
		errs.AddConfigError(*NewConfigError("server.*_timeout_ms", "timeouts must be positive values", ConfigErrorKindOutOfRange, nil))
	}

	for r, route := range c.Routes {
		if route.Host == "" {
			errs.AddConfigError(*NewConfigError(fmt.Sprintf("routes[%d].host", r), "host must be set", ConfigErrorKindMissingRequired, nil))
		}

		err := validatePathPrefix(route.PathPrefix)
		if err != nil {
			errs.AddConfigError(*NewConfigError(fmt.Sprintf("routes[%d].path_prefix", r), err.Error(), ConfigErrorKindInvalidFormat, nil))
		}

		err = validateMethods(route.Methods)
		if err != nil {
			errs.AddConfigError(*NewConfigError(fmt.Sprintf("routes[%d].methods", r), err.Error(), ConfigErrorKindInvalidValue, nil))
		}

		if len(route.Backends) == 0 {
			errs.AddConfigError(*NewConfigError(fmt.Sprintf("routes[%d].backends", r), "each route must have at least one backend configured", ConfigErrorKindMissingRequired, nil))
		}

		for b, backend := range route.Backends {
			if backend.Url == "" {
				errs.AddConfigError(*NewConfigError(fmt.Sprintf("routes[%d].backends[%d].url", r, b), "url must be set", ConfigErrorKindMissingRequired, nil))
				continue
			}

			err := validateUrl(backend.Url)
			if err != nil {
				errs.AddConfigError(*NewConfigError(fmt.Sprintf("routes[%d].backends[%d].url", r, b), err.Error(), ConfigErrorKindInvalidFormat, nil))
			}
		}
	}

	if len(errs.ConfigErrors) > 0 {
		return errs
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
		return fmt.Errorf("\"%s\" path_prefix must start with /", s)
	}

	if strings.ContainsAny(s, "?# ") {
		return fmt.Errorf("\"%s\" path_prefix must not contain query, fragments nor whitespaces", s)
	}

	return nil
}

func validateUrl(s string) error {
	url, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("\"%s\" can't be parsed as a URL: %w", s, err)
	}

	if url.Scheme == "" || url.Host == "" {
		return fmt.Errorf("\"%s\" must contain scheme and host to be a valid URL", url)
	}

	if url.Scheme != "http" && url.Scheme != "https" {
		return fmt.Errorf("\"%s\" must be http or https", url)
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
			return fmt.Errorf("%s is not a valid HTTP method", method)
		}
	}

	return nil
}
