package config

import (
	"log"
	"net/http"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/auth"
)

// LoadOptions holds the options for loading a configuration.
type LoadOptions struct {
	Logger        *log.Logger
	TokenProvider auth.TokenProvider
	Endpoint      string
	HTTPClient    *http.Client
}

// LoadOptionsFunc is a type alias for the type of function that adds a load option.
type LoadOptionsFunc func(*LoadOptions) error

// WithLogger adds the specified logger.
func WithLogger(v *log.Logger) LoadOptionsFunc {
	return func(o *LoadOptions) error {
		o.Logger = v
		return nil
	}
}

// WithTokenProvider adds the specified authentication provider.
func WithTokenProvider(v auth.TokenProvider) LoadOptionsFunc {
	return func(o *LoadOptions) error {
		o.TokenProvider = v
		return nil
	}
}

// WithEndpoint add/sets the specified endpoint, replacing THARSIS_ENDPOINT
func WithEndpoint(endpoint string) LoadOptionsFunc {
	return func(o *LoadOptions) error {
		o.Endpoint = endpoint
		return nil
	}
}

// WithHTTPClient overrides the default HTTP client used by the SDK
func WithHTTPClient(client *http.Client) LoadOptionsFunc {
	return func(o *LoadOptions) error {
		o.HTTPClient = client
		return nil
	}
}
