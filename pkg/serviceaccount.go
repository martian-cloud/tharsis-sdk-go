package tharsis

import (
	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// Related types and conversion functions:

// graphQLServiceAccount represents a service account with GraphQL types.
type graphQLServiceAccount struct {
	ID           graphql.String
	Metadata     internal.GraphQLMetadata
	ResourcePath graphql.String
	Name         graphql.String
	Description  graphql.String
}

// serviceAccountFromGraphQL converts a GraphQL service account to external service account.
func serviceAccountFromGraphQL(g graphQLServiceAccount) types.ServiceAccount {
	return types.ServiceAccount{
		Metadata:     internal.MetadataFromGraphQL(g.Metadata, g.ID),
		ResourcePath: string(g.ResourcePath),
		Name:         string(g.Name),
		Description:  string(g.Description),
	}
}

// The End.
