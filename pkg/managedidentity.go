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
	CreateManagedIdentityAlias(ctx context.Context,
		input *types.CreateManagedIdentityAliasInput) (*types.ManagedIdentity, error)
	DeleteManagedIdentityAlias(ctx context.Context,
		input *types.DeleteManagedIdentityAliasInput) error
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
	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
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

	err := m.client.graphqlClient.Query(ctx, true, &target, variables)
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
	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedUpdate, variables)
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
	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables)
	if err != nil {
		return err
	}

	return errorFromGraphqlProblems(wrappedDelete.DeleteManagedIdentity.Problems)
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
	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
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
	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedAssign, variables)
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
	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedUnassign, variables)
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

	err := m.client.graphqlClient.Query(ctx, true, &target, variables)
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
	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
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

	err := m.client.graphqlClient.Query(ctx, true, &target, variables)
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
	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedUpdate, variables)
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
	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables)
	if err != nil {
		return err
	}

	return errorFromGraphqlProblems(wrappedDelete.DeleteManagedIdentityAccessRule.Problems)
}

func (m *managedIdentity) CreateManagedIdentityAlias(ctx context.Context,
	input *types.CreateManagedIdentityAliasInput) (*types.ManagedIdentity, error) {

	var wrappedCreate struct {
		CreateManagedIdentityAlias struct {
			ManagedIdentity GraphQLManagedIdentity
			Problems        []internal.GraphQLProblem
		} `graphql:"createManagedIdentityAlias(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	if err = errorFromGraphqlProblems(wrappedCreate.CreateManagedIdentityAlias.Problems); err != nil {
		return nil, err
	}

	identity := identityFromGraphQL(wrappedCreate.CreateManagedIdentityAlias.ManagedIdentity)
	return &identity, nil
}

func (m *managedIdentity) DeleteManagedIdentityAlias(ctx context.Context,
	input *types.DeleteManagedIdentityAliasInput) error {

	var wrappedDelete struct {
		DeleteManagedIdentityAlias struct {
			Problems []internal.GraphQLProblem
		} `graphql:"deleteManagedIdentityAlias(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables)
	if err != nil {
		return err
	}

	return errorFromGraphqlProblems(wrappedDelete.DeleteManagedIdentityAlias.Problems)
}

// Related types and conversion functions:

// graphQLAccessRuleModuleAttestationPolicy represents a
// ManagedIdentityAccessRuleModuleAttestationPolicy with graphql types.
type graphQLAccessRuleModuleAttestationPolicy struct {
	PredicateType *graphql.String
	PublicKey     graphql.String
}

// graphQLManagedIdentityAccessRule represents a managed identity
// access rule with graphQL types.
type graphQLManagedIdentityAccessRule struct {
	Metadata                  internal.GraphQLMetadata
	ID                        graphql.String
	RunStage                  graphql.String
	Type                      graphql.String
	ModuleAttestationPolicies []graphQLAccessRuleModuleAttestationPolicy
	ManagedIdentity           GraphQLManagedIdentity
	AllowedUsers              []graphQLUser
	AllowedServiceAccounts    []graphQLServiceAccount
	AllowedTeams              []graphQLTeam
}

// GraphQLManagedIdentity represents the insides of the query structure,
// everything in the managed identity object, and with graphql types.
type GraphQLManagedIdentity struct {
	AliasSourceID *graphql.String
	Metadata      internal.GraphQLMetadata
	ID            graphql.String
	Type          graphql.String
	ResourcePath  graphql.String
	Name          graphql.String
	Description   graphql.String
	Data          graphql.String
	CreatedBy     graphql.String
	IsAlias       graphql.Boolean
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
	result := types.ManagedIdentity{
		Metadata:     internal.MetadataFromGraphQL(g.Metadata, g.ID),
		Type:         types.ManagedIdentityType(g.Type),
		ResourcePath: string(g.ResourcePath),
		Name:         string(g.Name),
		Description:  string(g.Description),
		Data:         string(g.Data),
		CreatedBy:    string(g.CreatedBy),
		IsAlias:      bool(g.IsAlias),
	}

	if g.AliasSourceID != nil {
		result.AliasSourceID = (*string)(g.AliasSourceID)
	}

	return result
}

// moduleAttestationPolicyFromGraphQL converts a GraphQL ManagedIdentityAccessRuleModuleAttestationPolicy
// to an external ManagedIdentityAccessRuleModuleAttestationPolicy.
func moduleAttestationPolicyFromGraphQL(g graphQLAccessRuleModuleAttestationPolicy) types.ManagedIdentityAccessRuleModuleAttestationPolicy {
	result := types.ManagedIdentityAccessRuleModuleAttestationPolicy{
		PublicKey: string(g.PublicKey),
	}

	if g.PredicateType != nil {
		result.PredicateType = (*string)(g.PredicateType)
	}

	return result
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

	attestationPolicies := []types.ManagedIdentityAccessRuleModuleAttestationPolicy{}
	for _, policy := range g.ModuleAttestationPolicies {
		attestationPolicies = append(attestationPolicies, moduleAttestationPolicyFromGraphQL(policy))
	}

	return types.ManagedIdentityAccessRule{
		Metadata:                  internal.MetadataFromGraphQL(g.Metadata, g.ID),
		RunStage:                  types.JobType(g.RunStage),
		AllowedUsers:              users,
		AllowedServiceAccounts:    serviceAccounts,
		AllowedTeams:              teams,
		ManagedIdentityID:         string(g.ManagedIdentity.ID),
		Type:                      types.ManagedIdentityAccessRuleType(g.Type),
		ModuleAttestationPolicies: attestationPolicies,
	}
}

// The End.
