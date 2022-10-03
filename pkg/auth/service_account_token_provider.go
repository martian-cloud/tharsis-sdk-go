package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Provides tokens suitable for use by a service account, including automatic renewal.

const (
	loginPath = "v1/serviceaccounts/login"
	loginType = "service-account-token"
	loginJSON = "application/json"

	expirationGuardband = 30 * time.Second // seconds of guardband for expiration time
)

// Nested structures to use to log in:
// More proper to make each layer its own named type--according to StackOverflow.
type loginBody struct {
	Data struct {
		Type       string `json:"type"`
		Attributes struct {
			ServiceAccountPath string `json:"service-account-path"`
			Token              string `json:"token"`
		} `json:"attributes"`
	} `json:"data"`
}

// serviceAccountTokenProvider implements Provider.
type serviceAccountTokenProvider struct {
	loginURL string
	// The permanent/static values from constructor arguments, environment variables, etc.
	options *options
	// The temporary/dynamic service account token, with expiration.
	// For thread safety, the token and its expiration with a mutex are protected by a mutex.
	token *tokenInfo
}

type options struct {
	serviceAccountPath string
	firstTokenValue    string
}

type tokenInfo struct {
	mutex   sync.RWMutex
	token   string
	expires *time.Time
}

// NewServiceAccountTokenProvider returns a new instance of this provider.
//
// Constructor arguments for options service account path and token values take
// priority over the environment variables.  For now,
//
func NewServiceAccountTokenProvider(endpointURL, accountPath, token string) (TokenProvider, error) {

	if accountPath == "" {
		return nil, fmt.Errorf("service account path was empty")
	}

	if token == "" {
		return nil, fmt.Errorf("service account first token was empty")
	}

	serviceAccountProvider := serviceAccountTokenProvider{
		// For the login URL, do not use path.Join to combine the URL and the path.  It corrupts "//" to "/".
		loginURL: endpointURL + "/" + loginPath,
		options: &options{
			serviceAccountPath: accountPath,
			firstTokenValue:    token,
		},
		token: &tokenInfo{},
	}

	return &serviceAccountProvider, nil
}

// GetToken is the one required method for the Provider interface.
func (p *serviceAccountTokenProvider) GetToken() (string, error) {

	if p.isTokenExpired() {
		err := p.renewToken()
		if err != nil {
			return "", err
		}
	}

	p.token.mutex.RLock()
	defer p.token.mutex.RUnlock()
	return p.token.token, nil
}

//////////////////////////////////////////////////////////////////////////////

// Internal stuff:

// For debug/testing, have a source shell script that sets environment variables
// THARSIS_SERVICE_ACCOUNT_PATH to the path (_WITH_ group prefix "cts/") of my test service account and
// THARSIS_SERVICE_ACCOUNT_TOKEN to the token I get back from "terraform login api..."
// Then, the login HTTP request will give a token for the GraphQL requests.

// isTokenExpired returns true if a token was set but has expired, true if no token was ever set,
// and false if a token has been set and has not yet expired.
func (p *serviceAccountTokenProvider) isTokenExpired() bool {
	p.token.mutex.RLock()
	defer p.token.mutex.RUnlock()

	return p.token.expires == nil || !time.Now().Add(expirationGuardband).Before(*p.token.expires)
}

func (p *serviceAccountTokenProvider) renewToken() error {

	reqBody, err := json.Marshal(
		loginBody{
			Data: struct {
				Type       string `json:"type"`
				Attributes struct {
					ServiceAccountPath string `json:"service-account-path"`
					Token              string `json:"token"`
				} `json:"attributes"`
			}{
				Type: loginType,
				Attributes: struct {
					ServiceAccountPath string `json:"service-account-path"`
					Token              string `json:"token"`
				}{
					ServiceAccountPath: p.options.serviceAccountPath,
					Token:              p.options.firstTokenValue,
				},
			},
		})
	if err != nil {
		return err
	}

	resp, err := http.Post(p.loginURL, loginJSON, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Must check the status code.
	if (resp.StatusCode != http.StatusCreated) && (resp.StatusCode != http.StatusOK) {
		return fmt.Errorf("service account token renewal failed: %s", respBody)
	}

	var gotRespBody loginBody
	err = json.Unmarshal(respBody, &gotRespBody)
	if err != nil {
		return err
	}

	// FIXME: For now, assume the service account token expires in 300 second.
	// The expiration time will be added to the response structure later on.
	bogusExpire := time.Now().Add(300 * time.Second)

	// Store the (temporary) token and expiration time.
	p.token.mutex.Lock()
	p.token.token = gotRespBody.Data.Attributes.Token
	// FIXME: Fix this:
	p.token.expires = &bogusExpire
	p.token.mutex.Unlock()

	return nil
}

// The End.
