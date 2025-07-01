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

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
)

// Provides tokens suitable for use by a service account, including automatic renewal.

const (
	graphQLSuffix = "graphql"

	expirationGuardband = 30 * time.Second // seconds of guardband for expiration time
)

// OIDCTokenGetter is a callback for returning the OIDC token that is used to login to the service account
type OIDCTokenGetter func() (string, error)

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
	serviceAccountPath string
	oidcTokenGetter    OIDCTokenGetter
	// The temporary/dynamic service account token, with expiration.
	// For thread safety, the token and its expiration with a mutex are protected by a mutex.
	token       *tokenInfo
	endpointURL string
	// The permanent/static values from constructor arguments, environment variables, etc.
}

type tokenInfo struct {
	expires *time.Time
	token   string
	mutex   sync.RWMutex
}

// NewServiceAccountTokenProvider returns a new instance of this provider.
//
// Constructor arguments for options service account path and token values take
// priority over the environment variables.  For now,
func NewServiceAccountTokenProvider(endpointURL string, accountPath string, oidcTokenGetter OIDCTokenGetter) (TokenProvider, error) {

	if accountPath == "" {
		return nil, fmt.Errorf("service account path was empty")
	}

	serviceAccountProvider := serviceAccountTokenProvider{
		// For the create token URL, do not use path.Join to combine the URL and the path.  It corrupts "//" to "/".
		endpointURL:        endpointURL,
		serviceAccountPath: accountPath,
		oidcTokenGetter:    oidcTokenGetter,
		token:              &tokenInfo{},
	}

	return &serviceAccountProvider, nil
}

// GetToken is the one required method for the Provider interface.
func (p *serviceAccountTokenProvider) GetToken() (string, error) {

	if p.isTokenExpired() {
		err := p.renewToken()
		if err != nil {
			return "", fmt.Errorf("service account token renewal failed: %w", err)
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

	oidcToken, err := p.oidcTokenGetter()
	if err != nil {
		return fmt.Errorf("failed to get oidc token: %v", err)
	}

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
		p.serviceAccountPath, oidcToken)

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
		return errors.ErrorFromHTTPResponse(resp)
	}

	var gotRespBody createTokenBody
	err = json.Unmarshal(respBody, &gotRespBody)
	if err != nil {
		return err
	}

	// Check for GraphQL errors in the response (even if the status code is 'ok').
	if len(gotRespBody.Errors) > 0 {
		var errMsgs []string
		for _, err := range gotRespBody.Errors {
			errMsgs = append(errMsgs, err.Message)
		}
		return fmt.Errorf("errors in response body: %s", strings.Join(errMsgs, "; "))
	}

	// Must check for GraphQL problems in the response.
	// All cases of user input should have been mapped by the API into GraphQL Problems.
	gotProblems := gotRespBody.Data.ServiceAccountCreateToken.Problems
	if len(gotProblems) > 0 {
		// Convert from JSON struct to GraphQLProblem type.
		problemsSlice := []internal.GraphQLProblem{}
		for _, problem := range gotProblems {
			graphQLProblem := internal.GraphQLProblem{
				Message: graphql.String(problem.Message),
				Type:    internal.GraphQLProblemType(problem.Type),
			}

			// Convert the field.
			for _, field := range problem.Field {
				graphQLProblem.Field = append(graphQLProblem.Field, graphql.String(field))
			}

			problemsSlice = append(problemsSlice, graphQLProblem)
		}

		return errors.ErrorFromGraphqlProblems(problemsSlice)
	}

	// If the API server is not working properly (no KMS access, for example), the pointer
	//  fields in the response structure can be nil.  Check for that to avoid a panic.
	if gotRespBody.Data.ServiceAccountCreateToken.Token == nil {
		return fmt.Errorf("nil token field in response")
	}
	if gotRespBody.Data.ServiceAccountCreateToken.ExpiresIn == nil {
		return fmt.Errorf("nil expiration field in response")
	}

	// Store the (temporary) token and expiration time.
	p.token.mutex.Lock()
	p.token.token = *gotRespBody.Data.ServiceAccountCreateToken.Token
	expiresWhen := time.Now().Add(time.Duration(*gotRespBody.Data.ServiceAccountCreateToken.ExpiresIn) * time.Second)
	p.token.expires = &expiresWhen
	p.token.mutex.Unlock()

	return nil
}
