package tharsis

import (
	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// Related types and conversion functions:

// graphQLUser represents a Tharsis user with graphQL types.
type graphQLUser struct {
	ID             graphql.String
	Metadata       internal.GraphQLMetadata
	Username       graphql.String
	Email          graphql.String
	SCIMExternalID graphql.String
	Admin          graphql.Boolean
	Active         graphql.Boolean
}

// userFromGraphQL converts a graphQL user to an external Tharsis user.
func userFromGraphQL(g graphQLUser) types.User {
	return types.User{
		Metadata:       internal.MetadataFromGraphQL(g.Metadata, g.ID),
		Username:       string(g.Username),
		Email:          string(g.Email),
		SCIMExternalID: string(g.SCIMExternalID),
		Admin:          bool(g.Admin),
		Active:         bool(g.Active),
	}
}
