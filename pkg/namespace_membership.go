package tharsis

import (
	"context"

	"github.com/aws/smithy-go/ptr"
	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// NamespaceMembership implements functions related to Tharsis groups.
type NamespaceMembership interface {
	AddMembership(ctx context.Context, input *types.CreateNamespaceMembershipInput) (*types.NamespaceMembership, error)
	UpdateMembership(ctx context.Context, input *types.UpdateNamespaceMembershipInput) (*types.NamespaceMembership, error)
	DeleteMembership(ctx context.Context, input *types.DeleteNamespaceMembershipInput) (*types.NamespaceMembership, error)
}

type namespaceMembership struct {
	client *Client
}

// NewNamespaceMembership returns a NamespaceMembership.
func NewNamespaceMembership(client *Client) NamespaceMembership {
	return &namespaceMembership{client: client}
}

func (m *namespaceMembership) AddMembership(ctx context.Context, input *types.CreateNamespaceMembershipInput) (*types.NamespaceMembership, error) {
	var wrappedAddMembership struct {
		CreateNamespaceMembership struct {
			Membership graphQLNamespaceMembership
			Problems   []internal.GraphQLProblem
		} `graphql:"createNamespaceMembership(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedAddMembership, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedAddMembership.CreateNamespaceMembership.Problems); err != nil {
		return nil, err
	}

	added := namespaceMembershipFromGraphQL(wrappedAddMembership.CreateNamespaceMembership.Membership)
	return &added, nil
}

func (m *namespaceMembership) UpdateMembership(ctx context.Context, input *types.UpdateNamespaceMembershipInput) (*types.NamespaceMembership, error) {
	var wrappedUpdateMembership struct {
		UpdateNamespaceMembership struct {
			Membership graphQLNamespaceMembership
			Problems   []internal.GraphQLProblem
		} `graphql:"updateNamespaceMembership(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedUpdateMembership, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedUpdateMembership.UpdateNamespaceMembership.Problems); err != nil {
		return nil, err
	}

	updated := namespaceMembershipFromGraphQL(wrappedUpdateMembership.UpdateNamespaceMembership.Membership)
	return &updated, nil
}

func (m *namespaceMembership) DeleteMembership(ctx context.Context, input *types.DeleteNamespaceMembershipInput) (*types.NamespaceMembership, error) {
	var wrappedDeleteMembership struct {
		DeleteNamespaceMembership struct {
			Membership graphQLNamespaceMembership
			Problems   []internal.GraphQLProblem
		} `graphql:"deleteNamespaceMembership(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedDeleteMembership, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedDeleteMembership.DeleteNamespaceMembership.Problems); err != nil {
		return nil, err
	}

	deleted := namespaceMembershipFromGraphQL(wrappedDeleteMembership.DeleteNamespaceMembership.Membership)
	return &deleted, nil
}

//////////////////////////////////////////////////////////////////////////////

type graphQLNamespaceMembership struct {
	ID           graphql.String
	Metadata     internal.GraphQLMetadata
	ResourcePath graphql.String
	Member       graphQLMember
	Role         graphQLRole
}

type graphQLMember struct {
	Team           graphQLTeam           `graphql:"...on Team"`
	Typename       graphql.String        `graphql:"__typename"`
	ServiceAccount graphQLServiceAccount `graphql:"...on ServiceAccount"`
	User           graphQLUser           `graphql:"...on User"`
}

type graphQLRole struct {
	ID          graphql.String
	Name        graphql.String
	Permissions []graphql.String
}

// namespaceMembershipFromGraphQL converts a GraphQL NamespaceMembership to an external NamespaceMembership.
func namespaceMembershipFromGraphQL(g graphQLNamespaceMembership) types.NamespaceMembership {
	var userID, serviceAccountID, teamID *string

	switch g.Member.Typename {
	case "User":
		userID = ptr.String(string(g.Member.User.ID))
	case "ServiceAccount":
		serviceAccountID = ptr.String(string(g.Member.ServiceAccount.ID))
	case "Team":
		teamID = ptr.String(string(g.Member.Team.ID))
	}

	result := types.NamespaceMembership{
		Metadata:         internal.MetadataFromGraphQL(g.Metadata, g.ID),
		UserID:           userID,
		ServiceAccountID: serviceAccountID,
		TeamID:           teamID,
		Role:             string(g.Role.Name),
	}
	return result
}
