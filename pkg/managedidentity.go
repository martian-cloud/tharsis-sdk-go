package tharsis

import (
	"context"
	"fmt"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// ManagedIdentity implements functions related to Tharsis ManagedIdentity.
type ManagedIdentity interface {
	CreateManagedIdentity(ctx context.Context,
		input *types.CreateManagedIdentityInput) (*types.ManagedIdentity, error)
	UpdateManagedIdentity(ctx context.Context,
		input *types.UpdateManagedIdentityInput) (*types.ManagedIdentity, error)
	DeleteManagedIdentity(ctx context.Context,
		input *types.DeleteManagedIdentityInput) error
	CreateManagedIdentityCredentials(ctx context.Context,
		input *types.CreateManagedIdentityCredentialsInput) ([]byte, error)
	AssignManagedIdentityToWorkspace(ctx context.Context,
		input *types.AssignManagedIdentityInput) (*types.Workspace, error)
	UnassignManagedIdentityFromWorkspace(ctx context.Context,
		input *types.AssignManagedIdentityInput) (*types.Workspace, error)
}

type managedIdentity struct {
	client *Client
}

// NewManagedIdentity returns a new ManagedIdentity.
func NewManagedIdentity(client *Client) ManagedIdentity {
	return &managedIdentity{client: client}
}

//////////////////////////////////////////////////////////////////////////////

// The ManagedIdentity paginator will go here.

//////////////////////////////////////////////////////////////////////////////

// CreateManagedIdentity creates a managed identity.
func (m *managedIdentity) CreateManagedIdentity(ctx context.Context,
	input *types.CreateManagedIdentityInput) (*types.ManagedIdentity, error) {

	var wrappedCreate struct {
		CreateManagedIdentity struct {
			ManagedIdentity GraphQLManagedIdentity
			Problems        []internal.GraphQLProblem
		} `graphql:"createManagedIdentity(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	err = internal.ProblemsToError(wrappedCreate.CreateManagedIdentity.Problems)
	if err != nil {
		return nil, fmt.Errorf("problems creating managed identity: %v", err)
	}

	identity := identityFromGraphQL(wrappedCreate.CreateManagedIdentity.ManagedIdentity)
	return &identity, nil
}

// UpdateManagedIdentity updates a managed identity.
func (m *managedIdentity) UpdateManagedIdentity(ctx context.Context,
	input *types.UpdateManagedIdentityInput) (*types.ManagedIdentity, error) {

	var wrappedUpdate struct {
		UpdateManagedIdentity struct {
			ManagedIdentity GraphQLManagedIdentity
			Problems        []internal.GraphQLProblem
		} `graphql:"updateManagedIdentity(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, &wrappedUpdate, variables)
	if err != nil {
		return nil, err
	}

	err = internal.ProblemsToError(wrappedUpdate.UpdateManagedIdentity.Problems)
	if err != nil {
		return nil, fmt.Errorf("problems updating managed identity: %v", err)
	}

	identity := identityFromGraphQL(wrappedUpdate.UpdateManagedIdentity.ManagedIdentity)
	return &identity, nil
}

// DeleteManagedIdentity deletes a managed identity.
func (m *managedIdentity) DeleteManagedIdentity(ctx context.Context,
	input *types.DeleteManagedIdentityInput) error {

	var wrappedDelete struct {
		DeleteManagedIdentity struct {
			Problems []internal.GraphQLProblem
		} `graphql:"deleteManagedIdentity(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, &wrappedDelete, variables)
	if err != nil {
		return err
	}

	err = internal.ProblemsToError(wrappedDelete.DeleteManagedIdentity.Problems)
	if err != nil {
		return fmt.Errorf("problems deleting managed identity: %v", err)
	}

	return nil
}

// CreateManagedIdentityCredentials returns new managed identity credentials.
func (m *managedIdentity) CreateManagedIdentityCredentials(ctx context.Context,
	input *types.CreateManagedIdentityCredentialsInput) ([]byte, error) {

	var wrappedCreate struct {
		CreateManagedIdentityCredentials struct {
			ManagedIdentityCredentials struct {
				Data graphql.String
			}
			Problems []internal.GraphQLProblem
		} `graphql:"createManagedIdentityCredentials(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	err = internal.ProblemsToError(wrappedCreate.CreateManagedIdentityCredentials.Problems)
	if err != nil {
		return nil, fmt.Errorf("problems creating managed identity credentials: %v", err)
	}

	return []byte(wrappedCreate.CreateManagedIdentityCredentials.ManagedIdentityCredentials.Data), nil
}

// AssignManagedIdentityToWorkspace assigns the given identity to a workspace.
func (m *managedIdentity) AssignManagedIdentityToWorkspace(ctx context.Context,
	input *types.AssignManagedIdentityInput) (*types.Workspace, error) {
	var wrappedAssign struct {
		AssignManagedIdentity struct {
			Problems  []internal.GraphQLProblem
			Workspace graphQLWorkspace
		} `graphql:"assignManagedIdentity(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, &wrappedAssign, variables)
	if err != nil {
		return nil, err
	}

	err = internal.ProblemsToError(wrappedAssign.AssignManagedIdentity.Problems)
	if err != nil {
		return nil, fmt.Errorf("problems assigning managed identity to workspace: %v", err)
	}

	created, err := workspaceFromGraphQL(wrappedAssign.AssignManagedIdentity.Workspace)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// UnassignManagedIdentityFromWorkspace un-assigns the given identity from a workspace.
func (m *managedIdentity) UnassignManagedIdentityFromWorkspace(ctx context.Context,
	input *types.AssignManagedIdentityInput) (*types.Workspace, error) {
	var wrappedUnassign struct {
		UnAssignManagedIdentity struct {
			Problems  []internal.GraphQLProblem
			Workspace graphQLWorkspace
		} `graphql:"unassignManagedIdentity(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, &wrappedUnassign, variables)
	if err != nil {
		return nil, err
	}

	err = internal.ProblemsToError(wrappedUnassign.UnAssignManagedIdentity.Problems)
	if err != nil {
		return nil, fmt.Errorf("problems assigning managed identity to workspace: %v", err)
	}

	created, err := workspaceFromGraphQL(wrappedUnassign.UnAssignManagedIdentity.Workspace)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// Related types and conversion functions:

// graphQLManagedIdentityAccessRule represents a managed identity
// access rule with graphQL types.
type graphQLManagedIdentityAccessRule struct {
	ID                     graphql.String
	Metadata               internal.GraphQLMetadata
	RunStage               graphql.String
	AllowedUsers           []graphQLUser
	AllowedServiceAccounts []graphQLServiceAccount
	AllowedTeams           []graphQLTeam
}

// GraphQLManagedIdentity represents the insides of the query structure,
// everything in the managed identity object, and with graphql types.
type GraphQLManagedIdentity struct {
	ID           graphql.String
	Metadata     internal.GraphQLMetadata
	Type         graphql.String
	ResourcePath graphql.String
	Name         graphql.String
	Description  graphql.String
	Data         graphql.String
	AccessRules  []graphQLManagedIdentityAccessRule
}

// TODO: Some of these functions may not be needed.

// sliceManagedIdentitiesFromGraphQL converts a slice of GraphQL Managed Identities
// to a slice of external managed identities.
func sliceManagedIdentitiesFromGraphQL(inputs []GraphQLManagedIdentity) []types.ManagedIdentity {
	result := make([]types.ManagedIdentity, len(inputs))
	for ix, input := range inputs {
		result[ix] = identityFromGraphQL(input)
	}
	return result
}

// accessRulesFromGraphQL converts a managed identity access rule to external access rule.
func accessRulesFromGraphQL(g []graphQLManagedIdentityAccessRule) []types.ManagedIdentityAccessRule {
	accessRules := []types.ManagedIdentityAccessRule{}

	// Convert the fields.
	for _, accessRule := range g {
		users := []types.User{}
		serviceAccounts := []types.ServiceAccount{}
		teams := []types.Team{}

		// Convert users.
		for _, user := range accessRule.AllowedUsers {
			users = append(users, userFromGraphQL(user))
		}

		// Convert service accounts.
		for _, sa := range accessRule.AllowedServiceAccounts {
			serviceAccounts = append(serviceAccounts, serviceAccountFromGraphQL(sa))
		}

		// Convert teams.
		for _, team := range accessRule.AllowedTeams {
			teams = append(teams, teamFromGraphQL(team))
		}

		accessRules = append(accessRules, types.ManagedIdentityAccessRule{
			Metadata:               internal.MetadataFromGraphQL(accessRule.Metadata, accessRule.ID),
			RunStage:               types.JobType(accessRule.RunStage),
			AllowedUsers:           users,
			AllowedServiceAccounts: serviceAccounts,
			AllowedTeams:           teams,
		})
	}

	return accessRules
}

// identityFromGraphQL converts a GraphQL Managed Identity to an external managed identity.
func identityFromGraphQL(g GraphQLManagedIdentity) types.ManagedIdentity {
	return types.ManagedIdentity{
		Metadata:     internal.MetadataFromGraphQL(g.Metadata, g.ID),
		Type:         types.ManagedIdentityType(g.Type),
		ResourcePath: string(g.ResourcePath),
		Name:         string(g.Name),
		Description:  string(g.Description),
		Data:         string(g.Data),
		AccessRules:  accessRulesFromGraphQL(g.AccessRules),
	}
}

// The End.
