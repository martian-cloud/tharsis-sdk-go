package tharsis

import (
	"context"
	"time"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// ServiceAccount implements functions related to Tharsis ServiceAccount.
type ServiceAccount interface {
	CreateServiceAccount(ctx context.Context,
		input *types.CreateServiceAccountInput) (*types.ServiceAccount, error)
	GetServiceAccount(ctx context.Context,
		input *types.GetServiceAccountInput) (*types.ServiceAccount, error)
	UpdateServiceAccount(ctx context.Context,
		input *types.UpdateServiceAccountInput) (*types.ServiceAccount, error)
	DeleteServiceAccount(ctx context.Context,
		input *types.DeleteServiceAccountInput) error
	Login(ctx context.Context,
		input *types.ServiceAccountLoginInput) (*types.ServiceAccountLoginResponse, error)
}

type serviceAccount struct {
	client *Client
}

// NewServiceAccount returns a new ServiceAccount.
func NewServiceAccount(client *Client) ServiceAccount {
	return &serviceAccount{client: client}
}

//////////////////////////////////////////////////////////////////////////////

// The ServiceAccount paginator will go here.

//////////////////////////////////////////////////////////////////////////////

// CreateServiceAccount creates a service account.
func (m *serviceAccount) CreateServiceAccount(ctx context.Context,
	input *types.CreateServiceAccountInput) (*types.ServiceAccount, error) {

	// Must change bound claims from map[string]string to []JWTClaimInput
	modifiedInput := internal.CreateServiceAccountInput{
		Name:              input.Name,
		Description:       input.Description,
		GroupPath:         input.GroupPath,
		OIDCTrustPolicies: modifyTrustPolicies(input.OIDCTrustPolicies),
	}

	var wrappedCreate struct {
		CreateServiceAccount struct {
			ServiceAccount graphQLServiceAccount
			Problems       []internal.GraphQLProblem
		} `graphql:"createServiceAccount(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": modifiedInput,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	if err = errorFromGraphqlProblems(wrappedCreate.CreateServiceAccount.Problems); err != nil {
		return nil, err
	}

	serviceAccount := serviceAccountFromGraphQL(wrappedCreate.CreateServiceAccount.ServiceAccount)
	return &serviceAccount, nil
}

// GetServiceAccount reads a service account.
func (m *serviceAccount) GetServiceAccount(ctx context.Context,
	input *types.GetServiceAccountInput) (*types.ServiceAccount, error) {

	var target struct {
		ServiceAccount *graphQLServiceAccount `graphql:"serviceAccount(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.String(input.ID),
	}

	err := m.client.graphqlClient.Query(ctx, &target, variables)
	if err != nil {
		return nil, err
	}

	if target.ServiceAccount == nil {
		return nil, newError(ErrNotFound, "service account with id %s not found", input.ID)
	}

	serviceAccount := serviceAccountFromGraphQL(*target.ServiceAccount)
	return &serviceAccount, nil
}

// UpdateServiceAccount updates a service account.
func (m *serviceAccount) UpdateServiceAccount(ctx context.Context,
	input *types.UpdateServiceAccountInput) (*types.ServiceAccount, error) {

	// Must change bound claims from map[string]string to []JWTClaimInput
	// ID is used to find the service account.
	// Description and trust policies are modified.
	modifiedInput := internal.UpdateServiceAccountInput{
		ID:                input.ID,
		Description:       input.Description,
		OIDCTrustPolicies: modifyTrustPolicies(input.OIDCTrustPolicies),
	}

	var wrappedUpdate struct {
		UpdateServiceAccount struct {
			ServiceAccount graphQLServiceAccount
			Problems       []internal.GraphQLProblem
		} `graphql:"updateServiceAccount(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": modifiedInput,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, &wrappedUpdate, variables)
	if err != nil {
		return nil, err
	}

	if err = errorFromGraphqlProblems(wrappedUpdate.UpdateServiceAccount.Problems); err != nil {
		return nil, err
	}

	serviceAccount := serviceAccountFromGraphQL(wrappedUpdate.UpdateServiceAccount.ServiceAccount)
	return &serviceAccount, nil
}

// DeleteServiceAccount deletes a service account.
func (m *serviceAccount) DeleteServiceAccount(ctx context.Context,
	input *types.DeleteServiceAccountInput) error {

	var wrappedDelete struct {
		DeleteServiceAccount struct {
			Problems []internal.GraphQLProblem
		} `graphql:"deleteServiceAccount(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, &wrappedDelete, variables)
	if err != nil {
		return err
	}

	if err = errorFromGraphqlProblems(wrappedDelete.DeleteServiceAccount.Problems); err != nil {
		return err
	}

	return nil
}

// Login logs in to a service account.
func (m *serviceAccount) Login(ctx context.Context,
	input *types.ServiceAccountLoginInput) (*types.ServiceAccountLoginResponse, error) {

	var wrappedLogin struct {
		ServiceAccountLogin struct {
			Token     graphql.String
			ExpiresIn graphql.Int
			Problems  []internal.GraphQLProblem
		} `graphql:"serviceAccountLogin(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute the mutation request.
	err := m.client.graphqlClient.Mutate(ctx, &wrappedLogin, variables)
	if err != nil {
		return nil, err
	}

	if err = errorFromGraphqlProblems(wrappedLogin.ServiceAccountLogin.Problems); err != nil {
		return nil, err
	}

	// The API returns the duration to expiration in seconds.  This method returns a time.Duration.
	// The conversion of the int to time.Duration is required by the compiler.
	return &types.ServiceAccountLoginResponse{
		Token:     string(wrappedLogin.ServiceAccountLogin.Token),
		ExpiresIn: time.Duration(int(wrappedLogin.ServiceAccountLogin.ExpiresIn)) * time.Second,
	}, nil
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

// graphQLBoundClaim represents a bound claim with GraphQL types.
// If GraphQL supported maps, it would be simpler to use a map for bound claims.
type graphQLBoundClaim struct {
	Name  graphql.String
	Value graphql.String
}

// graphQLTrustPolicy represents a trust policy with GraphQL types.
type graphQLTrustPolicy struct {
	Issuer      graphql.String
	BoundClaims []graphQLBoundClaim
}

// graphQLServiceAccount represents a service account with GraphQL types.
type graphQLServiceAccount struct {
	ID                graphql.String
	Metadata          internal.GraphQLMetadata
	ResourcePath      graphql.String
	Name              graphql.String
	Description       graphql.String
	OIDCTrustPolicies []graphQLTrustPolicy
}

// serviceAccountFromGraphQL converts a GraphQL service account to external service account.
func serviceAccountFromGraphQL(g graphQLServiceAccount) types.ServiceAccount {
	trustPolicies := []types.OIDCTrustPolicy{}
	for _, trustPolicy := range g.OIDCTrustPolicies {
		trustPolicies = append(trustPolicies, trustPolicyFromGraphQL(trustPolicy))
	}
	return types.ServiceAccount{
		Metadata:          internal.MetadataFromGraphQL(g.Metadata, g.ID),
		ResourcePath:      string(g.ResourcePath),
		Name:              string(g.Name),
		Description:       string(g.Description),
		OIDCTrustPolicies: trustPolicies,
	}
}

// trustPolicyFromGraphQL converts a GraphQL trust policy to an external trust policy.
func trustPolicyFromGraphQL(tp graphQLTrustPolicy) types.OIDCTrustPolicy {
	boundClaims := make(map[string]string)
	for _, boundClaim := range tp.BoundClaims {
		boundClaims[string(boundClaim.Name)] = string(boundClaim.Value)
	}
	return types.OIDCTrustPolicy{
		Issuer:      string(tp.Issuer),
		BoundClaims: boundClaims,
	}
}

// modifyTrustPolicies converts a slice of external trust policies (with map[string]string for BoundClaims)
// to a slice of internal trust policies (with []JWTClaimInput for BoundClaims)
func modifyTrustPolicies(input []types.OIDCTrustPolicy) []internal.ServiceAccountOIDCTrustPolicyInput {
	result := []internal.ServiceAccountOIDCTrustPolicyInput{}

	for _, inputPolicy := range input {
		modifiedPolicy := internal.ServiceAccountOIDCTrustPolicyInput{
			Issuer: inputPolicy.Issuer,
		}

		for name, value := range inputPolicy.BoundClaims {
			modifiedPolicy.BoundClaims = append(modifiedPolicy.BoundClaims, internal.JWTClaimInput{
				Name:  name,
				Value: value,
			})
		}

		result = append(result, modifiedPolicy)
	}

	return result
}

// The End.
