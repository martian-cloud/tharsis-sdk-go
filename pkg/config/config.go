// Package config package
package config

import (
	"fmt"
	"log"
	"os"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/qiangxue/go-env"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/auth"
)

// Config holds a configuration.
type Config struct {
	Logger        *log.Logger
	TokenProvider auth.TokenProvider
	Endpoint      string
}

// Validate validates a config.
func (c Config) Validate() error {

	return validation.ValidateStruct(&c,
		validation.Field(&c.Endpoint, validation.Required, is.URL),
	)
}

// Load loads a configuration from a file and/or environment variables.
func Load(optFns ...func(*LoadOptions) error) (*Config, error) {

	// Run the options functions.
	loadOptions := LoadOptions{}
	for _, optFn := range optFns {
		err := optFn(&loadOptions)
		if err != nil {
			return nil, err
		}
	}

	// Copy from load options to config.
	c := Config(loadOptions)

	// If no logger, make a default one.
	if c.Logger == nil {
		c.Logger = log.Default()
	}

	// Environment variables override load options.
	if err := env.New("THARSIS_", log.Printf).Load(&c); err != nil {
		return nil, fmt.Errorf("failed to load env variables: %w", err)
	}

	// If no token provider already, try to make one.
	if c.TokenProvider == nil {

		// Get the environment variable values.
		serviceAccountName := os.Getenv("THARSIS_SERVICE_ACCOUNT_PATH")
		serviceAccountToken := os.Getenv("THARSIS_SERVICE_ACCOUNT_TOKEN")
		staticToken := os.Getenv("THARSIS_STATIC_TOKEN")

		// next preference: a service account provider from environment variables
		if (serviceAccountName != "") && (serviceAccountToken != "") {
			serviceAccountProvider, err := auth.NewServiceAccountTokenProvider(c.Endpoint,
				serviceAccountName, serviceAccountToken)
			if err != nil {
				return nil, fmt.Errorf("failed to obtain a token provider for service account %s: %v",
					serviceAccountName, err)
			}
			c.TokenProvider = serviceAccountProvider
		}

		// last option: a static token provider from an environment variable
		// Checking c.TokenProvider again, because it might have changed after the earlier check.
		if (c.TokenProvider == nil) && (staticToken != "") {
			staticProvider, err := auth.NewStaticTokenProvider(staticToken)
			if err != nil {
				return nil, fmt.Errorf("failed to obtain a static token provider: %v", err)
			}
			c.TokenProvider = staticProvider
		}
	}

	// If still no token provider, install a noop token provider that will return an error if GetToken is called.
	if c.TokenProvider == nil {
		c.TokenProvider = auth.NewNoopTokenProvider()
	}

	// Validate the config.
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &c, nil
}

// The End.
