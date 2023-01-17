package auth

import "fmt"

// Provides no token but an error if GetToken is called.

// noopTokenProvider implements Provider.
type noopTokenProvider struct {
}

// NewNoopTokenProvider returns a new instance of this provider.
func NewNoopTokenProvider() TokenProvider {
	return &noopTokenProvider{}
}

func (p *noopTokenProvider) GetToken() (string, error) {
	return "", fmt.Errorf("unable to create a token provider: %s; %s",
		"to use a service account token, set environment variables THARSIS_SERVICE_ACCOUNT_PATH and THARSIS_SERVICE_ACCOUNT_TOKEN",
		"to use a static token, set environment variable THARSIS_STATIC_TOKEN")
}

// The End.
