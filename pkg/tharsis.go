// Package tharsis provides functions for interfacing with the Tharsis API.
package tharsis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/config"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	graphQLSuffix         = "graphql"
	websocketWriteTimeout = 30 * time.Minute
)

// Client provides access for the client/user to access the SDK functions.
// Note: When adding a new field here, make sure to assign the field near the end of the NewClient function.
type Client struct {
	cfg           *config.Config // not currently essential but could become so
	logger        *log.Logger
	httpClient    *http.Client
	graphqlClient graphql.Client

	// TODO: Update subscription client to be a lazy connection which only
	// starts up when first subscription is created.
	graphqlSubscriptionClient *graphql.SubscriptionClient
	ConfigurationVersion      ConfigurationVersion
	Group                     Group
	Job                       Job
	ManagedIdentity           ManagedIdentity
	Plan                      Plan
	Apply                     Apply
	Run                       Run
	ServiceAccount            ServiceAccount
	StateVersion              StateVersion
	Variable                  Variable
	Workspaces                Workspaces
	TerraformProvider         TerraformProvider
	TerraformProviderVersion  TerraformProviderVersion
	TerraformModule           TerraformModule
	TerraformModuleVersion    TerraformModuleVersion
	TerraformProviderPlatform TerraformProviderPlatform
	TerraformCLIVersions      TerraformCLIVersion
}

// NewClient returns a TharsisClient.
func NewClient(cfg *config.Config) (*Client, error) {

	graphQLEndpoint, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	graphQLEndpoint.Path = path.Join(graphQLEndpoint.Path, graphQLSuffix)

	authToken, err := cfg.TokenProvider.GetToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %s", err)
	}

	connectParams := map[string]interface{}{
		"Authorization": "Bearer " + authToken,
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.RequestLogHook = func(_ retryablehttp.Logger, r *http.Request, i int) {
		if i > 0 {
			cfg.Logger.Printf("%s %s failed. Retry attempt %d", r.Method, r.URL, i)
		}
	}
	retryClient.Logger = nil
	client := &Client{
		cfg:        cfg,
		logger:     cfg.Logger,
		httpClient: retryClient.StandardClient(),
		graphqlClient: *graphql.NewClient(graphQLEndpoint.String(), retryClient.StandardClient()).
			WithRequestModifier(graphql.RequestModifier(
				func(req *http.Request) {
					req.Header.Set("Authorization", "Bearer "+authToken)
				})),
		graphqlSubscriptionClient: graphql.NewSubscriptionClient(graphQLEndpoint.String()).
			WithConnectionParams(connectParams).
			WithTimeout(websocketWriteTimeout).
			WithWebSocket(buildWebsocketConn(retryClient.StandardClient(), cfg.Logger)),
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
	client.TerraformCLIVersions = NewTerraformCLIVersion(client)

	client.graphqlSubscriptionClient.OnError(func(_ *graphql.SubscriptionClient, err error) error {
		msg := fmt.Sprintf("subscription client reported an error: %v", err)
		if err != nil {
			client.logger.Print(msg)
		}
		client.logger.Print("Terminating websocket connection")

		// Always return a new error here to terminate the web socket connection
		return errors.New(msg)
	})

	// Start the GraphQL subscription client
	go func() {
		err = client.graphqlSubscriptionClient.Run()
		if err != nil {
			client.cfg.Logger.Printf("error from attempt to run the subscription client: %s", err)
		}
	}()

	return client, nil
}

// Close client connections
func (c *Client) Close() error {
	// Set onError function to nil before closing to prevent false positive error logs
	c.graphqlSubscriptionClient.OnError(nil)
	return c.graphqlSubscriptionClient.Close()
}

type websocketHandler struct {
	*websocket.Conn
	ctx                 context.Context
	cancelKeepAliveFunc func()
	logger              *log.Logger
	timeout             time.Duration
}

func (wh *websocketHandler) WriteJSON(v interface{}) error {
	ctx, cancel := context.WithTimeout(wh.ctx, wh.timeout)
	defer cancel()

	return wsjson.Write(ctx, wh.Conn, v)
}

func (wh *websocketHandler) ReadJSON(v interface{}) error {
	ctx, cancel := context.WithTimeout(wh.ctx, wh.timeout)
	defer cancel()
	return wsjson.Read(ctx, wh.Conn, v)
}

func (wh *websocketHandler) Close() error {
	if wh.cancelKeepAliveFunc != nil {
		wh.cancelKeepAliveFunc()
		wh.cancelKeepAliveFunc = nil
	}

	return wh.Conn.Close(websocket.StatusNormalClosure, "close websocket")
}

func (wh *websocketHandler) startKeepalive() func() {
	stop := make(chan bool)

	go func() {
		for {
			msg := graphql.OperationMessage{
				Type: graphql.GQL_CONNECTION_KEEP_ALIVE,
			}

			if err := wh.WriteJSON(msg); err != nil {
				wh.logger.Printf("Failed to send keep alive %v", err)
			}

			select {
			case <-time.After(time.Minute):
			case <-wh.ctx.Done():
				return
			case <-stop:
				return
			}
		}
	}()

	return func() {
		stop <- true
	}
}

func buildWebsocketConn(httpClient *http.Client, logger *log.Logger) func(sc *graphql.SubscriptionClient) (graphql.WebsocketConn, error) {
	return func(sc *graphql.SubscriptionClient) (graphql.WebsocketConn, error) {
		options := &websocket.DialOptions{
			Subprotocols: []string{"graphql-ws"},
			HTTPClient:   httpClient,
		}

		c, _, err := websocket.Dial(sc.GetContext(), sc.GetURL(), options)
		if err != nil {
			return nil, err
		}

		handler := &websocketHandler{
			ctx:     sc.GetContext(),
			Conn:    c,
			timeout: sc.GetTimeout(),
			logger:  logger,
		}

		sc.OnConnected(func() {
			if handler.cancelKeepAliveFunc == nil {
				// Start websocket keep alive messages
				handler.cancelKeepAliveFunc = handler.startKeepalive()
			}
		})

		return handler, nil
	}
}

// The End.
