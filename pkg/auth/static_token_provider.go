package auth

import (
	"fmt"
)

// Provides a static token, where the user supplies the static token at initialization time.

// staticTokenProvider implements Provider.
type staticTokenProvider struct {
	token string
}

// NewStaticTokenProvider returns a new instance of this provider.
func NewStaticTokenProvider(token string) (TokenProvider, error) {
	if token == "" {
		return nil, fmt.Errorf("static token was empty")
	}

	staticProvider := staticTokenProvider{
		token: token,
	}
	var provider TokenProvider = &staticProvider
	return provider, nil
}

func (p *staticTokenProvider) GetToken() (string, error) {
	return p.token, nil
}
