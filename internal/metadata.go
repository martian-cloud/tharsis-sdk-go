package internal

import (
	"time"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// GraphQLMetadata represents the insides of the query structure,
// everything in the metadata object, and with graphql types.
//
// In the GraphQL structs, the ID field is in the parent rather than in the metadata.
type GraphQLMetadata struct {
	CreatedAt *time.Time     `json:"createdAt"`
	UpdatedAt *time.Time     `json:"updatedAt"`
	Version   graphql.String `json:"version"`
	TRN       graphql.String `json:"trn"`
}

// MetadataFromGraphQL converts GraphQL Metadata to an external metadata.
func MetadataFromGraphQL(g GraphQLMetadata, ID graphql.String) types.ResourceMetadata {
	return types.ResourceMetadata{
		ID:                   string(ID),
		Version:              string(g.Version),
		CreationTimestamp:    g.CreatedAt,
		LastUpdatedTimestamp: g.UpdatedAt,
		TRN:                  string(g.TRN),
	}
}

// MetadataToGraphQL converts an external metadata to a GraphQL Metadata.
func MetadataToGraphQL(m types.ResourceMetadata) (GraphQLMetadata, graphql.String) {
	return GraphQLMetadata{
		Version:   graphql.String(m.Version),
		CreatedAt: m.CreationTimestamp,
		UpdatedAt: m.LastUpdatedTimestamp,
		TRN:       graphql.String(m.TRN),
	}, graphql.String(m.ID)
}
