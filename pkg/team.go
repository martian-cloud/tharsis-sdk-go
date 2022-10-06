package tharsis

import (
	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// Related types and conversion functions:

// graphQLTeam represents a Team with GraphQL types.
type graphQLTeam struct {
	ID             graphql.String
	Metadata       internal.GraphQLMetadata
	Name           graphql.String
	Description    graphql.String
	SCIMExternalID graphql.String
}

// teamFromGraphQL converts a GraphQL team to external team.
func teamFromGraphQL(g graphQLTeam) types.Team {
	return types.Team{
		Metadata:       internal.MetadataFromGraphQL(g.Metadata, g.ID),
		Name:           string(g.Name),
		Description:    string(g.Description),
		SCIMExternalID: string(g.SCIMExternalID),
	}
}
