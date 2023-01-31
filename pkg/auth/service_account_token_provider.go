package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
)

// Provides tokens suitable for use by a service account, including automatic renewal.

const (
	graphQLSuffix = "graphql"

	expirationGuardband = 30 * time.Second // seconds of guardband for expiration time
)

// createTokenBody is returned by the raw GraphQL mutation.
// For some API server error conditions, Errors is populated and Data is null.
type createTokenBody struct {
	Data struct {
		ServiceAccountCreateToken struct {
			Token     *string `json:"token"`
			ExpiresIn *int    `json:"expiresIn"`
			Problems  []struct {
				Message string   `json:"message"`
				Type    string   `json:"type"`
				Field   []string `json:"field"`
			} `json:"problems"`
		} `json:"serviceAccountCreateToken"`
	} `json:"data"`
	Errors []struct {
		Message    string `json:"message"`
		Extensions struct {
			Code string `json:"code"`
		} `json:"extensions"`
		Path []string `json:"path"`
	} `json:"errors"`
}

// serviceAccountTokenProvider implements Provider.
type serviceAccountTokenProvider struct {
	options *options
	// The temporary/dynamic service account token, with expiration.
	// For thread safety, the token and its expiration with a mutex are protected by a mutex.
	token       *tokenInfo
	endpointURL string
	// The permanent/static values from constructor arguments, environment variables, etc.
}

type options struct {
	serviceAccountPath string
	firstTokenValue    string
}

type tokenInfo struct {
	mutex   sync.RWMutex
	expires *time.Time
	token   string
}

// NewServiceAccountTokenProvider returns a new instance of this provider.
//
// Constructor arguments for options service account path and token values take
// priority over the environment variables.  For now,
func NewServiceAccountTokenProvider(endpointURL, accountPath, token string) (TokenProvider, error) {

	if accountPath == "" {
		return nil, fmt.Errorf("service account path was empty")
	}

	if token == "" {
		return nil, fmt.Errorf("service account first token was empty")
	}

	serviceAccountProvider := serviceAccountTokenProvider{
		// For the create token URL, do not use path.Join to combine the URL and the path.  It corrupts "//" to "/".
		endpointURL: endpointURL,
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

// isTokenExpired returns true if a token was set but has expired, true if no token was ever set,
// and false if a token has been set and has not yet expired.
func (p *serviceAccountTokenProvider) isTokenExpired() bool {
	p.token.mutex.RLock()
	defer p.token.mutex.RUnlock()

	return p.token.expires == nil || !time.Now().Add(expirationGuardband).Before(*p.token.expires)
}

func (p *serviceAccountTokenProvider) renewToken() error {

	// The request here is sent via an ordinary HTTP client rather than the GraphQL client used elsewhere,
	// because this module is created by config.Load, which does not have the GraphQL client available.
	// The normal GraphQL client is created in tharsis.NewClient.

	graphQLEndpoint, err := url.Parse(p.endpointURL)
	if err != nil {
		return err
	}

	graphQLEndpoint.Path = path.Join(graphQLEndpoint.Path, graphQLSuffix)

	mutationCore := fmt.Sprintf(
		`mutation {
			serviceAccountCreateToken(
				input:{
					serviceAccountPath: "%s"
					token:              "%s"
				}
			) {
				token
				expiresIn
				problems{
					message
					type
				}
			}
		}`,
		p.options.serviceAccountPath, p.options.firstTokenValue)

	type reqType struct {
		Query string `json:"query"`
	}

	reqBody, err := json.Marshal(reqType{Query: mutationCore})
	if err != nil {
		return err
	}

	resp, err := http.Post(graphQLEndpoint.String(), "", bytes.NewReader(reqBody)) // nosemgrep: gosec.G107-1
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

	var gotRespBody createTokenBody
	err = json.Unmarshal(respBody, &gotRespBody)
	if err != nil {
		return err
	}

	// Check for GraphQL errors in the response (even if the status code is 'ok').
	if len(gotRespBody.Errors) > 0 {
		return fmt.Errorf("service account token renewal failed: errors in response body: %#v", gotRespBody.Errors)
	}

	// Must check for GraphQL problems in the response.
	// All cases of user input should have been mapped by the API into GraphQL Problems.
	if len(gotRespBody.Data.ServiceAccountCreateToken.Problems) > 0 {

		// For now, parse the incoming problem into an error without the aid of the
		// errorFromGraphqlProblems from the errors module, while trying to mimic its function.
		// See below for a map, a type, and its methods copied from the errors module.

		var result error
		for _, problem := range gotRespBody.Data.ServiceAccountCreateToken.Problems {

			code, ok := graphqlErrorCodeToSDKErrorCode[problem.Type]
			if !ok {
				code = "internal error"
			}

			result = multierror.Append(result, &localError{
				Code: code,
				Msg:  problem.Message,
			})
		}

		return result
	}

	// If the API server is not working properly (no KMS access, for example), the pointer
	//  fields in the response structure can be nil.  Check for that to avoid a panic.
	if gotRespBody.Data.ServiceAccountCreateToken.Token == nil {
		return fmt.Errorf("service account token renewal failed: nil token field in response")
	}
	if gotRespBody.Data.ServiceAccountCreateToken.ExpiresIn == nil {
		return fmt.Errorf("service account token renewal failed: nil expiration field in response")
	}

	// Store the (temporary) token and expiration time.
	p.token.mutex.Lock()
	p.token.token = *gotRespBody.Data.ServiceAccountCreateToken.Token
	expiresWhen := time.Now().Add(time.Duration(*gotRespBody.Data.ServiceAccountCreateToken.ExpiresIn) * time.Second)
	p.token.expires = &expiresWhen
	p.token.mutex.Unlock()

	return nil
}

/////////////////////////////////////////////////////////////////////////////

// This map, type, and its methods are (for now) copied from the errors module.

var graphqlErrorCodeToSDKErrorCode = map[string]string{
	"INTERNAL_SERVER_ERROR": "internal error",
	"BAD_REQUEST":           "bad request",
	"NOT_IMPLEMENTED":       "not implemented",
	"CONFLICT":              "conflict",
	"OPTIMISTIC_LOCK":       "optimistic lock",
	"NOT_FOUND":             "not found",
	"FORBIDDEN":             "forbidden",
	"RATE_LIMIT_EXCEEDED":   "too many requests",
	"UNAUTHENTICATED":       "unauthorized",
	"UNAUTHORIZED":          "unauthorized",
}

// localError represents an error returned by the Tharsis API
type localError struct {
	Err  error
	Code string
	Msg  string
}

func (e *localError) Error() string {
	if e.Msg != "" && e.Err != nil {
		var b strings.Builder
		b.WriteString(e.Msg)
		b.WriteString(": ")
		b.WriteString(e.Err.Error())
		return b.String()
	} else if e.Msg != "" {
		return e.Msg
	} else if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("<%s>", e.Code)
}
func (e *localError) Unwrap() error {
	return e.Err
}

// The End.
