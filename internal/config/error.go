package config

import "fmt"

type AggregatedConfigErrors struct {
	ConfigErrors []ConfigError
}

type ConfigError struct {
	Component string
	FieldPath string
	Kind      Kind
	Message   string
	Cause     error
}

type Kind string

const (
	ConfigErrorKindMissingRequired  Kind = "missing_required"
	ConfigErrorKindInvalidValue     Kind = "invalid_value"
	ConfigErrorKindInvalidFormat    Kind = "invalid_format"
	ConfigErrorKindUnsupportedValue Kind = "unsupported_value"
	ConfigErrorKindOutOfRange       Kind = "out_of_range"
	ConfigErrorKindDuplicateValue   Kind = "duplicate_value"
)

type SystemError struct {
	Component string
	Message   string
	Cause     error
}

const component string = "config"

func NewAggregatedConfigErrors() *AggregatedConfigErrors {
	return &AggregatedConfigErrors{}
}

func NewConfigError(fieldPath, message string, kind Kind, err error) *ConfigError {
	return &ConfigError{
		Component: component,
		FieldPath: fieldPath,
		Kind:      kind,
		Message:   message,
		Cause:     err,
	}
}

func NewSystemError(message string, err error) *SystemError {
	return &SystemError{
		Component: component,
		Message:   message,
		Cause:     err,
	}
}

func (a *AggregatedConfigErrors) AddConfigError(c ConfigError) {
	a.ConfigErrors = append(a.ConfigErrors, c)
}

func (a *AggregatedConfigErrors) Error() string {
	return fmt.Sprintf("configuration validation failed with %d errors", len(a.ConfigErrors))
}

func (s *SystemError) Error() string {
	return fmt.Sprintf("component=%s message=%s cause=%w", s.Component, s.Message, s.Cause)
}
