package tharsis

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// Team implements functions related to Tharsis teams.
type Team interface {
	CreateTeam(ctx context.Context, input *types.CreateTeamInput) (*types.Team, error)
	DeleteTeam(ctx context.Context, input *types.DeleteTeamInput) error
	AddTeamMember(ctx context.Context, input *types.AddUserToTeamInput) (*types.TeamMember, error)
}

type team struct {
	client *Client
}

// NewTeam returns a Team.
func NewTeam(client *Client) Team {
	return &team{client: client}
}

// CreateTeam creates a new team and returns its content.
func (t *team) CreateTeam(ctx context.Context, input *types.CreateTeamInput) (*types.Team, error) {
	var wrappedCreate struct {
		CreateTeam struct {
			Team     graphQLTeam
			Problems []internal.GraphQLProblem
		} `graphql:"createTeam(input: $input)"`
	}

	// Creating a new object requires the wrapped object above
	// but with all the contents in a struct in the variables.
	variables := map[string]interface{}{
		"input": *input,
	}

	err := t.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedCreate.CreateTeam.Problems); err != nil {
		return nil, err
	}

	created := teamFromGraphQL(wrappedCreate.CreateTeam.Team)
	return &created, nil
}

func (t *team) DeleteTeam(ctx context.Context, input *types.DeleteTeamInput) error {
	var wrappedDelete struct {
		DeleteTeam struct {
			Team     graphQLTeam
			Problems []internal.GraphQLProblem
		} `graphql:"deleteTeam(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := t.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables)
	if err != nil {
		return err
	}

	return errors.ErrorFromGraphqlProblems(wrappedDelete.DeleteTeam.Problems)
}

func (t *team) AddTeamMember(ctx context.Context, input *types.AddUserToTeamInput) (*types.TeamMember, error) {
	var wrappedAddTeamMember struct {
		AddTeamMember struct {
			TeamMember graphQLTeamMember
			Problems   []internal.GraphQLProblem
		} `graphql:"addUserToTeam(input: $input)"`
	}

	// Creating a new object requires the wrapped object above
	// but with all the contents in a struct in the variables.
	variables := map[string]interface{}{
		"input": *input,
	}

	err := t.client.graphqlClient.Mutate(ctx, true, &wrappedAddTeamMember, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedAddTeamMember.AddTeamMember.Problems); err != nil {
		return nil, err
	}

	added := teamMemberFromGraphQL(wrappedAddTeamMember.AddTeamMember.TeamMember)
	return &added, nil
}

//////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

// graphQLTeam represents a Team with GraphQL types.
type graphQLTeam struct {
	ID             graphql.String
	Metadata       internal.GraphQLMetadata
	Name           graphql.String
	Description    graphql.String
	SCIMExternalID graphql.String
}

// graphQLTeamMember represents one team member with GraphQL types.
type graphQLTeamMember struct {
	ID           graphql.String
	Metadata     internal.GraphQLMetadata
	User         graphQLUser
	Team         graphQLTeam
	IsMaintainer graphql.Boolean
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

// teamMemberFromGraphQL converts a GraphQL team member to external team member.
func teamMemberFromGraphQL(g graphQLTeamMember) types.TeamMember {
	return types.TeamMember{
		Metadata:     internal.MetadataFromGraphQL(g.Metadata, g.ID),
		UserID:       string(g.User.ID),
		TeamID:       string(g.Team.ID),
		IsMaintainer: bool(g.IsMaintainer),
	}
}
