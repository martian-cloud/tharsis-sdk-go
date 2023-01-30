// Package tharsis provides functions for interfacing with the Tharsis API.
package tharsis

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/hashicorp/go-retryablehttp"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/config"
)

const (
	graphQLSuffix = "graphql"
)

// Client provides access for the client/user to access the SDK functions.
// Note: When adding a new field here, make sure to assign the field near the end of the NewClient function.
type Client struct {
	cfg                        *config.Config // not currently essential but could become so
	logger                     *log.Logger
	httpClient                 *http.Client
	graphqlClient              graphqlClient
	graphqlSubscriptionClient  subscriptionClient
	ConfigurationVersion       ConfigurationVersion
	Group                      Group
	Job                        Job
	ManagedIdentity            ManagedIdentity
	Plan                       Plan
	Apply                      Apply
	Run                        Run
	ServiceAccount             ServiceAccount
	StateVersion               StateVersion
	Variable                   Variable
	Workspaces                 Workspaces
	TerraformProvider          TerraformProvider
	TerraformProviderVersion   TerraformProviderVersion
	TerraformModule            TerraformModule
	TerraformModuleVersion     TerraformModuleVersion
	TerraformModuleAttestation TerraformModuleAttestation
	TerraformProviderPlatform  TerraformProviderPlatform
	TerraformCLIVersions       TerraformCLIVersion
}

// NewClient returns a TharsisClient.
func NewClient(cfg *config.Config) (*Client, error) {

	graphQLEndpoint, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	graphQLEndpoint.Path = path.Join(graphQLEndpoint.Path, graphQLSuffix)

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.RequestLogHook = func(_ retryablehttp.Logger, r *http.Request, i int) {
		if i > 0 {
			cfg.Logger.Printf("%s %s failed. Retry attempt %d", r.Method, r.URL, i)
		}
	}
	retryClient.Logger = nil

	httpClient := retryClient.StandardClient()

	wrappedGraphqlClient := newGraphqlClientWrapper(graphQLEndpoint.String(), httpClient, cfg.TokenProvider, cfg.Logger)

	subscriptionClient, err := newLazySubscriptionClient(cfg, httpClient, graphQLEndpoint.String())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize graphql subscription client %w", err)
	}

	client := &Client{
		cfg:                       cfg,
		logger:                    cfg.Logger,
		httpClient:                httpClient,
		graphqlClient:             wrappedGraphqlClient,
		graphqlSubscriptionClient: subscriptionClient,
	}

	client.ConfigurationVersion = NewConfigurationVersion(client)
	client.Group = NewGroup(client)
	client.Job = NewJob(client)
	client.ManagedIdentity = NewManagedIdentity(client)
	client.Plan = NewPlan(client)
	client.Apply = NewApply(client)
	client.Run = NewRun(client)
	client.ServiceAccount = NewServiceAccount(client)
	client.StateVersion = NewStateVersion(client)
	client.Variable = NewVariable(client)
	client.Workspaces = NewWorkspaces(client)
	client.TerraformProvider = NewTerraformProvider(client)
	client.TerraformProviderVersion = NewTerraformProviderVersion(client)
	client.TerraformProviderPlatform = NewTerraformProviderPlatform(client)
	client.TerraformModule = NewTerraformModule(client)
	client.TerraformModuleVersion = NewTerraformModuleVersion(client)
	client.TerraformModuleAttestation = NewTerraformModuleAttestation(client)
	client.TerraformCLIVersions = NewTerraformCLIVersion(client)

	return client, nil
}

// Close client connections
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return c.graphqlSubscriptionClient.Close()
}

// The End.
