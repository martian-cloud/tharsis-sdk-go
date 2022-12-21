package tharsis

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// ManagedIdentity implements functions related to Tharsis ManagedIdentity.
type ManagedIdentity interface {
	CreateManagedIdentity(ctx context.Context,
		input *types.CreateManagedIdentityInput) (*types.ManagedIdentity, error)
	GetManagedIdentity(ctx context.Context,
		input *types.GetManagedIdentityInput) (*types.ManagedIdentity, error)
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
	GetManagedIdentityAccessRules(ctx context.Context,
		input *types.GetManagedIdentityInput) ([]types.ManagedIdentityAccessRule, error)
	CreateManagedIdentityAccessRule(ctx context.Context,
		input *types.CreateManagedIdentityAccessRuleInput) (*types.ManagedIdentityAccessRule, error)
	GetManagedIdentityAccessRule(ctx context.Context,
		input *types.GetManagedIdentityAccessRuleInput) (*types.ManagedIdentityAccessRule, error)
	UpdateManagedIdentityAccessRule(ctx context.Context,
		input *types.UpdateManagedIdentityAccessRuleInput) (*types.ManagedIdentityAccessRule, error)
	DeleteManagedIdentityAccessRule(ctx context.Context,
		input *types.DeleteManagedIdentityAccessRuleInput) error
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

	if err = errorFromGraphqlProblems(wrappedCreate.CreateManagedIdentity.Problems); err != nil {
		return nil, err
	}

	identity := identityFromGraphQL(wrappedCreate.CreateManagedIdentity.ManagedIdentity)
	return &identity, nil
}

// GetManagedIdentity reads a managed identity.
func (m *managedIdentity) GetManagedIdentity(ctx context.Context,
	input *types.GetManagedIdentityInput) (*types.ManagedIdentity, error) {

	var target struct {
		ManagedIdentity *GraphQLManagedIdentity `graphql:"managedIdentity(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.String(input.ID),
	}

	err := m.client.graphqlClient.Query(ctx, &target, variables)
	if err != nil {
		return nil, err
	}

	if target.ManagedIdentity == nil {
		return nil, newError(ErrNotFound, "managed identity with id %s not found", input.ID)
	}

	identity := identityFromGraphQL(*target.ManagedIdentity)
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

	if err = errorFromGraphqlProblems(wrappedUpdate.UpdateManagedIdentity.Problems); err != nil {
		return nil, err
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

	if err = errorFromGraphqlProblems(wrappedDelete.DeleteManagedIdentity.Problems); err != nil {
		return err
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

	if err = errorFromGraphqlProblems(wrappedCreate.CreateManagedIdentityCredentials.Problems); err != nil {
		return nil, err
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

	if err = errorFromGraphqlProblems(wrappedAssign.AssignManagedIdentity.Problems); err != nil {
		return nil, err
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

	if err = errorFromGraphqlProblems(wrappedUnassign.UnAssignManagedIdentity.Problems); err != nil {
		return nil, err
	}

	created, err := workspaceFromGraphQL(wrappedUnassign.UnAssignManagedIdentity.Workspace)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// GetManagedIdentityAccessRules returns the access rules that are tied to the specified managed identity.
func (m *managedIdentity) GetManagedIdentityAccessRules(ctx context.Context,
	input *types.GetManagedIdentityInput) ([]types.ManagedIdentityAccessRule, error) {
	var target struct {
		ManagedIdentity *struct {
			AccessRules []graphQLManagedIdentityAccessRule `graphql:"accessRules"`
		} `graphql:"managedIdentity(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.String(input.ID),
	}

	err := m.client.graphqlClient.Query(ctx, &target, variables)
	if err != nil {
		return nil, err
	}

	if target.ManagedIdentity == nil {
		return nil, nil
	}

	return accessRulesFromGraphQL(target.ManagedIdentity.AccessRules), nil
}

func (m *managedIdentity) CreateManagedIdentityAccessRule(ctx context.Context,
	input *types.CreateManagedIdentityAccessRuleInput) (*types.ManagedIdentityAccessRule, error) {

	var wrappedCreate struct {
		CreateManagedIdentityAccessRule struct {
			AccessRule graphQLManagedIdentityAccessRule
			Problems   []internal.GraphQLProblem
		} `graphql:"createManagedIdentityAccessRule(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	if err = errorFromGraphqlProblems(wrappedCreate.CreateManagedIdentityAccessRule.Problems); err != nil {
		return nil, err
	}

	accessRule := accessRuleFromGraphQL(wrappedCreate.CreateManagedIdentityAccessRule.AccessRule)
	return &accessRule, nil
}

// GetManagedIdentityAccessRule returns the managed identity access rule with the specified ID.
func (m *managedIdentity) GetManagedIdentityAccessRule(ctx context.Context,
	input *types.GetManagedIdentityAccessRuleInput) (*types.ManagedIdentityAccessRule, error) {

	var target struct {
		Node *struct {
			ManagedIdentityAccessRule graphQLManagedIdentityAccessRule `graphql:"...on ManagedIdentityAccessRule"`
		} `graphql:"node(id: $id)"`
	}

	variables := map[string]interface{}{"id": graphql.String(input.ID)}

	err := m.client.graphqlClient.Query(ctx, &target, variables)
	if err != nil {
		return nil, err
	}

	if target.Node == nil {
		return nil, newError(ErrNotFound, "managed identity access rule with id %s not found", input.ID)
	}

	accessRule := accessRuleFromGraphQL(target.Node.ManagedIdentityAccessRule)
	return &accessRule, nil
}

func (m *managedIdentity) UpdateManagedIdentityAccessRule(ctx context.Context,
	input *types.UpdateManagedIdentityAccessRuleInput) (*types.ManagedIdentityAccessRule, error) {

	var wrappedUpdate struct {
		UpdateManagedIdentityAccessRule struct {
			AccessRule graphQLManagedIdentityAccessRule
			Problems   []internal.GraphQLProblem
		} `graphql:"updateManagedIdentityAccessRule(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, &wrappedUpdate, variables)
	if err != nil {
		return nil, err
	}

	if err = errorFromGraphqlProblems(wrappedUpdate.UpdateManagedIdentityAccessRule.Problems); err != nil {
		return nil, err
	}

	accessRule := accessRuleFromGraphQL(wrappedUpdate.UpdateManagedIdentityAccessRule.AccessRule)
	return &accessRule, nil
}

func (m *managedIdentity) DeleteManagedIdentityAccessRule(ctx context.Context,
	input *types.DeleteManagedIdentityAccessRuleInput) error {

	var wrappedDelete struct {
		DeleteManagedIdentityAccessRule struct {
			Problems []internal.GraphQLProblem
		} `graphql:"deleteManagedIdentityAccessRule(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, &wrappedDelete, variables)
	if err != nil {
		return err
	}

	if err = errorFromGraphqlProblems(wrappedDelete.DeleteManagedIdentityAccessRule.Problems); err != nil {
		return err
	}

	return nil
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
	ManagedIdentity        GraphQLManagedIdentity
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
	CreatedBy    graphql.String
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
	for _, accessRule := range g {
		accessRules = append(accessRules, accessRuleFromGraphQL(accessRule))
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
		CreatedBy:    string(g.CreatedBy),
	}
}

// accessRuleFromGraphQL converts a GraphQL Managed Identity Access Rule
// to an external managed identity access rule.
func accessRuleFromGraphQL(g graphQLManagedIdentityAccessRule) types.ManagedIdentityAccessRule {

	users := []types.User{}
	serviceAccounts := []types.ServiceAccount{}
	teams := []types.Team{}

	// Convert users.
	for _, user := range g.AllowedUsers {
		users = append(users, userFromGraphQL(user))
	}

	// Convert service accounts.
	for _, sa := range g.AllowedServiceAccounts {
		serviceAccounts = append(serviceAccounts, serviceAccountFromGraphQL(sa))
	}

	// Convert teams.
	for _, team := range g.AllowedTeams {
		teams = append(teams, teamFromGraphQL(team))
	}

	return types.ManagedIdentityAccessRule{
		Metadata:               internal.MetadataFromGraphQL(g.Metadata, g.ID),
		RunStage:               types.JobType(g.RunStage),
		AllowedUsers:           users,
		AllowedServiceAccounts: serviceAccounts,
		AllowedTeams:           teams,
		ManagedIdentityID:      string(g.ManagedIdentity.ID),
	}
}

// The End.
