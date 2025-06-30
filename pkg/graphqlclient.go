package tharsis

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/auth"
)

type graphqlClient interface {
	Query(ctx context.Context, withAuth bool, q interface{}, variables map[string]interface{},
		options ...graphql.Option) error
	Mutate(ctx context.Context, withAuth bool, m interface{}, variables map[string]interface{},
		options ...graphql.Option) error
}

// The two GraphQL clients are initialized in a lazy manner.  The mutex protects the two GraphQL clients.
type graphqlClientWrapper struct {
	tokenProvider  auth.TokenProvider
	httpClient     *http.Client
	logger         *log.Logger
	noAuthClient   *graphql.Client
	withAuthClient *graphql.Client
	endpoint       string
	mutex          sync.Mutex
}

// Ensure graphqlClientWrapper implements the graphqlClient interface.
var (
	_ graphqlClient = &graphqlClientWrapper{}
)

// newGraphqlClientWrapper creates and returns a new graphqlClientWrapper.
// Because the individual *graphql.Client fields are initialized in lazy fashion,
// does not do much.
func newGraphqlClientWrapper(endpoint string, httpClient *http.Client,
	tokenProvider auth.TokenProvider, logger *log.Logger) graphqlClient {
	return &graphqlClientWrapper{
		endpoint:      endpoint,
		httpClient:    httpClient,
		tokenProvider: tokenProvider,
		logger:        logger,
	}
}

func (g *graphqlClientWrapper) Query(ctx context.Context, withAuth bool,
	q interface{}, variables map[string]interface{}, options ...graphql.Option) error {
	gClient, err := g.getClient(withAuth)
	if err != nil {
		return err
	}

	err = gClient.Query(ctx, q, variables, options...)
	if err != nil {
		return errors.ErrorFromGraphqlError(err)
	}

	return nil
}

func (g *graphqlClientWrapper) Mutate(ctx context.Context, withAuth bool,
	m interface{}, variables map[string]interface{}, options ...graphql.Option) error {
	gClient, err := g.getClient(withAuth)
	if err != nil {
		return err
	}

	err = gClient.Mutate(ctx, m, variables, options...)
	if err != nil {
		return errors.ErrorFromGraphqlError(err)
	}

	return nil
}

// getClient does the atomic lazy initialization of whichever GraphQL client is requested.
func (g *graphqlClientWrapper) getClient(withAuth bool) (*graphql.Client, error) {

	// Synchronize access to the GraphQL clients.
	g.mutex.Lock()
	defer g.mutex.Unlock()

	switch {
	case !withAuth && (g.noAuthClient != nil):
		// If no auth is required and client already exists,
		return g.noAuthClient, nil
	case withAuth && (g.withAuthClient != nil):
		// If client already exists, just return it.
		return g.withAuthClient, nil
	case withAuth && (g.tokenProvider == nil):
		// If auth is required but no token provider is available, error out.
		return nil, fmt.Errorf("unable to create a token provider: %s; %s",
			"to use a service account token, set environment variables THARSIS_SERVICE_ACCOUNT_PATH and THARSIS_SERVICE_ACCOUNT_TOKEN",
			"to use a static token, set environment variable THARSIS_STATIC_TOKEN")
	}

	// Create a new (plain so far) GraphQL client.
	result := graphql.NewClient(g.endpoint, g.httpClient)

	// If required, add auth to the GraphQL client.
	if withAuth {
		result = result.WithRequestModifier(graphql.RequestModifier(
			func(req *http.Request) {
				authToken, gtErr := g.tokenProvider.GetToken()
				if gtErr != nil {
					g.logger.Printf("failed to get authentication token: %s", gtErr.Error())
					return
				}
				req.Header.Set("Authorization", "Bearer "+authToken)
			}))
		g.withAuthClient = result
	} else {
		g.noAuthClient = result
	}

	return result, nil
}
