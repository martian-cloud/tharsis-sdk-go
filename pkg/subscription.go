package tharsis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/auth"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/config"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	websocketWriteTimeout = 30 * time.Minute
)

type subscriptionClient interface {
	Subscribe(v interface{}, variables map[string]interface{}, handler func(message []byte, err error) error, options ...graphql.Option) (string, error)
	Close() (err error)
}

type lazySubscriptionClient struct {
	mutex         sync.Mutex
	tokenProvider auth.TokenProvider
	client        *graphql.SubscriptionClient
	logger        *log.Logger
	isRunning     int64
}

func newLazySubscriptionClient(cfg *config.Config, httpClient *http.Client, graphQLEndpoint string) (*lazySubscriptionClient, error) {
	client := graphql.NewSubscriptionClient(graphQLEndpoint).
		WithTimeout(websocketWriteTimeout).
		WithWebSocket(buildWebsocketConn(httpClient, cfg.Logger))

	client.OnError(func(_ *graphql.SubscriptionClient, err error) error {
		msg := fmt.Sprintf("subscription client reported an error: %s", err)
		if err != nil {
			cfg.Logger.Print(msg)
		}

		// Always return a new error here to terminate the web socket connection
		return errors.New(msg)
	})

	return &lazySubscriptionClient{
		logger:        cfg.Logger,
		tokenProvider: cfg.TokenProvider,
		client:        client,
	}, nil
}

func (s *lazySubscriptionClient) Subscribe(v interface{}, variables map[string]interface{}, handler func(message []byte, err error) error, options ...graphql.Option) (string, error) {
	subscriptionID, err := s.client.Subscribe(v, variables, handler, options...)
	if err != nil {
		return "", err
	}

	s.lazyRun()

	return subscriptionID, nil
}

func (s *lazySubscriptionClient) Close() error {
	// Set onError function to nil before closing to prevent false positive error logs
	s.client.OnError(nil)
	return s.client.Close()
}

func (s *lazySubscriptionClient) lazyRun() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Start connection if it's not already running
	if atomic.LoadInt64(&s.isRunning) > 0 {
		return
	}

	s.setIsRunning(true)

	s.client.OnDisconnected(func() {
		s.setIsRunning(false)
	})

	// Start the GraphQL subscription client
	go func() {
		authToken, err := s.tokenProvider.GetToken()
		if err != nil {
			s.logger.Printf("failed to get auth token: %v", err)
			s.setIsRunning(false)
			return
		}

		connectParams := map[string]interface{}{
			"Authorization": "Bearer " + authToken,
		}

		// Reset connection params before reconnecting
		s.client.WithConnectionParams(connectParams)

		if err := s.client.Run(); err != nil {
			s.logger.Printf("error from attempt to run the subscription client: %s", err)
			s.setIsRunning(false)
		}
	}()
}

func (s *lazySubscriptionClient) setIsRunning(value bool) {
	if value {
		atomic.StoreInt64(&s.isRunning, 1)
	} else {
		atomic.StoreInt64(&s.isRunning, 0)
	}
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
