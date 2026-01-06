package tharsis

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// Me implements functions for getting the current authenticated subject.
type Me interface {
	GetCallerInfo(ctx context.Context) (any, error)
}

type me struct {
	client *Client
}

// NewMe returns a Me.
func NewMe(client *Client) Me {
	return &me{client: client}
}

// GetCallerInfo returns the currently authenticated subject (User or ServiceAccount).
func (m *me) GetCallerInfo(ctx context.Context) (any, error) {
	var query struct {
		Me struct {
			Typename       graphql.String        `graphql:"__typename"`
			User           graphQLUser           `graphql:"... on User"`
			ServiceAccount graphQLServiceAccount `graphql:"... on ServiceAccount"`
		}
	}

	if err := m.client.graphqlClient.Query(ctx, true, &query, nil); err != nil {
		return nil, err
	}

	typename := string(query.Me.Typename)

	switch typename {
	case "User":
		user := userFromGraphQL(query.Me.User)
		return &user, nil
	case "ServiceAccount":
		sa := serviceAccountFromGraphQL(query.Me.ServiceAccount)
		return &sa, nil
	default:
		return nil, &types.Error{Code: types.ErrBadRequest, Msg: "unknown caller type: " + typename}
	}
}
