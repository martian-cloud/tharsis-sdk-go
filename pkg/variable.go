package tharsis

import (
	"context"
	"fmt"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// Variable implements functions related to Tharsis Variable.
type Variable interface {
	CreateVariable(ctx context.Context,
		input *types.CreateNamespaceVariableInput) (*types.NamespaceVariable, error)
	GetVariable(ctx context.Context,
		input *types.GetNamespaceVariableInput) (*types.NamespaceVariable, error)
	UpdateVariable(ctx context.Context,
		input *types.UpdateNamespaceVariableInput) (*types.NamespaceVariable, error)
	DeleteVariable(ctx context.Context,
		input *types.DeleteNamespaceVariableInput) error
	SetVariables(ctx context.Context, input *types.SetNamespaceVariablesInput) error
}

type variable struct {
	client *Client
}

// NewVariable returns a new Variable.
func NewVariable(client *Client) Variable {
	return &variable{client: client}
}

//////////////////////////////////////////////////////////////////////////////

// The Variable paginator will go here.

//////////////////////////////////////////////////////////////////////////////

// CreateVariable creates a variable.
func (m *variable) CreateVariable(ctx context.Context,
	input *types.CreateNamespaceVariableInput) (*types.NamespaceVariable, error) {

	// The createNamespaceVariable mutation returns the whole namespace object.
	// After retrieving the namespace object, we will need to find the variable in question.
	var wrappedCreate struct {
		CreateNamespaceVariable struct {
			Namespace graphQLNamespace
			Problems  []internal.GraphQLProblem
		} `graphql:"createNamespaceVariable(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedCreate.CreateNamespaceVariable.Problems); err != nil {
		return nil, err
	}

	// Find the variable within the namespace object.
	variable := variableFromGraphQLNamespace(wrappedCreate.CreateNamespaceVariable.Namespace, input.Key)
	if variable == nil {
		return nil, fmt.Errorf("failed to find variable just created: %s.%s", input.NamespacePath, input.Key)
	}

	return variable, nil
}

// GetVariable reads a variable.
func (m *variable) GetVariable(ctx context.Context,
	input *types.GetNamespaceVariableInput) (*types.NamespaceVariable, error) {

	var target struct {
		Node *struct {
			NamespaceVariable graphQLNamespaceVariable `graphql:"...on NamespaceVariable"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.String(input.ID)}

	err := m.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}
	if target.Node == nil {
		return nil, errors.NewError(types.ErrNotFound, "variable with id %s not found", input.ID)
	}

	result := variableFromGraphQL(target.Node.NamespaceVariable)
	return &result, nil
}

// UpdateVariable updates a variable.
func (m *variable) UpdateVariable(ctx context.Context,
	input *types.UpdateNamespaceVariableInput) (*types.NamespaceVariable, error) {

	// The updateNamespaceVariable mutation returns the whole namespace object.
	// After retrieving the namespace object, we will need to find the variable in question.
	var wrappedUpdate struct {
		UpdateNamespaceVariable struct {
			Namespace graphQLNamespace
			Problems  []internal.GraphQLProblem
		} `graphql:"updateNamespaceVariable(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedUpdate, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedUpdate.UpdateNamespaceVariable.Problems); err != nil {
		return nil, err
	}

	// Find the variable within the namespace object.
	variable := variableFromGraphQLNamespace(wrappedUpdate.UpdateNamespaceVariable.Namespace, input.Key)
	if variable == nil {
		return nil, fmt.Errorf("failed to find variable just updated: %s", input.Key)
	}

	return variable, nil
}

// DeleteVariable deletes a variable.
func (m *variable) DeleteVariable(ctx context.Context,
	input *types.DeleteNamespaceVariableInput) error {

	var wrappedDelete struct {
		DeleteNamespaceVariable struct {
			Problems []internal.GraphQLProblem
		} `graphql:"deleteNamespaceVariable(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedDelete, variables)
	if err != nil {
		return err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedDelete.DeleteNamespaceVariable.Problems); err != nil {
		return err
	}

	return nil
}

func (m *variable) SetVariables(ctx context.Context, input *types.SetNamespaceVariablesInput) error {
	var wrappedSet struct {
		SetNamespaceVariables struct {
			Problems []internal.GraphQLProblem
		} `graphql:"setNamespaceVariables(input: $input)"`
	}

	// Creating a new object requires the wrapped object above
	// but with all the contents in a struct in the variables.
	variables := map[string]interface{}{
		"input": *input,
	}

	err := m.client.graphqlClient.Mutate(ctx, true, &wrappedSet, variables)
	if err != nil {
		return err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedSet.SetNamespaceVariables.Problems); err != nil {
		return err
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

// graphQLNamespace represents the namespace in which a newly-created variable resides
// For this module, we need only the Variables field.
type graphQLNamespace struct {
	Variables []graphQLNamespaceVariable
}

// graphQLNamespaceVariable represents a variable with GraphQL types.
type graphQLNamespaceVariable struct {
	ID            graphql.String
	Metadata      internal.GraphQLMetadata
	Value         *graphql.String
	NamespacePath graphql.String
	Key           graphql.String
	Category      graphql.String
	HCL           graphql.Boolean
}

// variableFromGraphQLNamespace finds the specified variable in the namespace object
// and returns a non-GraphQL version of it.
func variableFromGraphQLNamespace(v graphQLNamespace, key string) *types.NamespaceVariable {
	graphQLKey := graphql.String(key)

	for _, v := range v.Variables {
		if v.Key == graphQLKey {
			result := variableFromGraphQL(v)
			return &result
		}
	}

	// Variable not found.
	return nil
}

// variableFromGraphQL converts a GraphQL variable to a plain variable.
func variableFromGraphQL(v graphQLNamespaceVariable) types.NamespaceVariable {
	return types.NamespaceVariable{
		Metadata:      internal.MetadataFromGraphQL(v.Metadata, v.ID),
		Key:           string(v.Key),
		Value:         (*string)(v.Value),
		Category:      types.VariableCategory(v.Category),
		HCL:           bool(v.HCL),
		NamespacePath: string(v.NamespacePath),
	}
}

// The End.
