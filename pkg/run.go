package tharsis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/paginators"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// Run implements functions related to Tharsis runs.
type Run interface {
	GetRun(ctx context.Context, input *types.GetRunInput) (*types.Run, error)
	GetRuns(ctx context.Context, input *types.GetRunsInput) (*types.GetRunsOutput, error)
	GetRunVariables(ctx context.Context, input *types.GetRunVariablesInput) ([]types.RunVariable, error)
	SetVariablesIncludedInTFConfig(ctx context.Context, input *types.SetVariablesIncludedInTFConfigInput) error
	GetRunPaginator(ctx context.Context, input *types.GetRunsInput) (*RunPaginator, error)
	CreateRun(ctx context.Context, input *types.CreateRunInput) (*types.Run, error)
	ApplyRun(ctx context.Context, input *types.ApplyRunInput) (*types.Run, error)
	CancelRun(ctx context.Context, input *types.CancelRunInput) (*types.Run, error)
	SubscribeToWorkspaceRunEvents(ctx context.Context, input *types.RunSubscriptionInput) (<-chan *types.Run, error)
}

type run struct {
	client *Client
}

// NewRun returns a Run.
func NewRun(client *Client) Run {
	return &run{client: client}
}

// GetRun returns everything about the run.
func (r *run) GetRun(ctx context.Context, input *types.GetRunInput) (*types.Run, error) {
	// Validate and resolve ID or TRN
	resolvedID, err := types.ValidateIDOrTRN(input.ID, input.TRN, "run")
	if err != nil {
		return nil, errors.NewError(types.ErrBadRequest, err.Error())
	}

	var target struct {
		Run *graphQLRun `graphql:"run(id: $id)"`
	}
	variables := map[string]interface{}{
		"id": graphql.String(resolvedID),
	}

	err = r.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}

	if target.Run == nil {
		return nil, errors.NewError(types.ErrNotFound, "run with id %s not found", resolvedID)
	}

	result := runFromGraphQL(*target.Run)
	return &result, nil
}

// GetRuns returns a list of run objects.
//
// Based on the 'first' and 'after' fields of the PaginationOptions within the GetRunsInput,
// it returns the first 'first' items after the 'after' element.  That can be equivalent to the
// first page from a paged query.
func (r *run) GetRuns(ctx context.Context,
	input *types.GetRunsInput) (*types.GetRunsOutput, error) {

	// Pass nil for after so the user's cursor value will be used.
	queryStruct, err := getRuns(ctx, r.client.graphqlClient, input, nil)
	if err != nil {
		return nil, err
	}

	// Convert and repackage the type-specific results.
	runResults := make([]types.Run, len(queryStruct.Runs.Edges))
	for ix, runCustom := range queryStruct.Runs.Edges {
		runResults[ix] = runFromGraphQL(runCustom.Node)
	}

	return &types.GetRunsOutput{
		PageInfo: &types.PageInfo{
			TotalCount:  int(queryStruct.Runs.TotalCount),
			HasNextPage: bool(queryStruct.Runs.PageInfo.HasNextPage),
			Cursor:      string(queryStruct.Runs.PageInfo.EndCursor),
		},
		Runs: runResults,
	}, nil
}

// GetRunVariables returns a list of run variables.
func (r *run) GetRunVariables(ctx context.Context, input *types.GetRunVariablesInput) ([]types.RunVariable, error) {
	variables := map[string]interface{}{
		"id": graphql.String(input.RunID),
	}

	if input.IncludeSensitiveValues {
		var target struct {
			Run struct {
				Variables               []graphQLRunVariable
				SensitiveVariableValues []graphQLRunVariableSensitiveValue
			} `graphql:"run(id: $id)"`
		}

		// Query for variables.
		err := r.client.graphqlClient.Query(ctx, true, &target, variables)
		if err != nil {
			return nil, err
		}
		if len(target.Run.Variables) == 0 {
			return nil, nil
		}

		// Build map of sensitive variable values
		sensitiveValues := make(map[string]string)
		for _, v := range target.Run.SensitiveVariableValues {
			sensitiveValues[string(v.VersionID)] = string(v.Value)
		}

		// Convert and repackage the type-specific results.
		variablesResult := make([]types.RunVariable, len(target.Run.Variables))
		for ix, varCustom := range target.Run.Variables {
			runVariable := runVariableFromGraphQL(varCustom)
			if runVariable.Sensitive {
				if value, ok := sensitiveValues[*runVariable.VersionID]; ok {
					runVariable.Value = &value
				} else {
					return nil, errors.NewError(types.ErrNotFound, "sensitive value for variable %s not found", runVariable.Key)
				}
			}
			variablesResult[ix] = runVariable
		}

		return variablesResult, nil
	}

	var target struct {
		Run struct {
			Variables []graphQLRunVariable
		} `graphql:"run(id: $id)"`
	}

	// Query for variables.
	err := r.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}
	if len(target.Run.Variables) == 0 {
		return nil, nil
	}

	// Convert and repackage the type-specific results.
	variablesResult := make([]types.RunVariable, len(target.Run.Variables))
	for ix, varCustom := range target.Run.Variables {
		variablesResult[ix] = runVariableFromGraphQL(varCustom)
	}

	return variablesResult, nil
}

func (r *run) GetRunPaginator(_ context.Context,
	input *types.GetRunsInput) (*RunPaginator, error) {

	paginator := newRunPaginator(*r.client, input)
	return &paginator, nil
}

// SetVariablesIncludedInTFConfig sets variables that are included in the Terraform config.
func (r *run) SetVariablesIncludedInTFConfig(ctx context.Context, input *types.SetVariablesIncludedInTFConfigInput) error {
	var wrappedUpdate struct {
		SetVariablesIncludedInTFConfig struct {
			Problems []internal.GraphQLProblem
		} `graphql:"setVariablesIncludedInTFConfig(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	if err := r.client.graphqlClient.Mutate(ctx, true, &wrappedUpdate, variables); err != nil {
		return err
	}

	return errors.ErrorFromGraphqlProblems(wrappedUpdate.SetVariablesIncludedInTFConfig.Problems)
}

// CreateRun creates a new run and returns its content.
func (r *run) CreateRun(ctx context.Context, input *types.CreateRunInput) (*types.Run, error) {

	var wrappedCreate struct {
		CreateRun struct {
			Problems []internal.GraphQLProblem
			Run      graphQLRun
		} `graphql:"createRun(input: $input)"`
	}

	// Creating a new object requires the wrapped object above
	// but with all the contents in a struct in the variables.
	variables := map[string]interface{}{
		"input": *input,
	}

	err := r.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedCreate.CreateRun.Problems); err != nil {
		return nil, err
	}

	created := runFromGraphQL(wrappedCreate.CreateRun.Run)
	return &created, nil
}

// ApplyRun applies a run and returns its context
func (r *run) ApplyRun(ctx context.Context, input *types.ApplyRunInput) (*types.Run, error) {

	var wrappedApply struct {
		ApplyRun struct {
			Problems []internal.GraphQLProblem
			Run      graphQLRun
		} `graphql:"applyRun(input: $input)"`
	}

	// Applying an object requires the wrapped object above
	// but with all the contents in a struct in the variables.
	variables := map[string]interface{}{
		"input": *input,
	}

	err := r.client.graphqlClient.Mutate(ctx, true, &wrappedApply, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedApply.ApplyRun.Problems); err != nil {
		return nil, err
	}

	applied := runFromGraphQL(wrappedApply.ApplyRun.Run)
	return &applied, nil
}

// CancelRun cancels a run and returns its context
func (r *run) CancelRun(ctx context.Context, input *types.CancelRunInput) (*types.Run, error) {

	var wrappedCancel struct {
		CancelRun struct {
			Problems []internal.GraphQLProblem
			Run      graphQLRun
		} `graphql:"cancelRun(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := r.client.graphqlClient.Mutate(ctx, true, &wrappedCancel, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(wrappedCancel.CancelRun.Problems); err != nil {
		return nil, err
	}

	canceled := runFromGraphQL(wrappedCancel.CancelRun.Run)
	return &canceled, nil
}

func (r *run) SubscribeToWorkspaceRunEvents(_ context.Context, input *types.RunSubscriptionInput) (<-chan *types.Run, error) {
	eventChannel := make(chan *types.Run)

	var target struct {
		RunEvent struct {
			Run graphQLRun
		} `graphql:"workspaceRunEvents(input: $input)"`
	}
	variables := map[string]interface{}{
		"input": *input,
	}

	// The embedded run event callback function.
	runEventCallback := func(message []byte, err error) error {
		// Detect any incoming error.
		if err != nil {
			// Close channel
			close(eventChannel)
			return err
		}

		var event struct {
			RunEvent struct {
				Run graphQLRun `json:"run"`
			} `json:"workspaceRunEvents"`
		}

		if message != nil {
			err = json.Unmarshal(message, &event)
			if err != nil {
				return err
			}

			run := runFromGraphQL(event.RunEvent.Run)
			eventChannel <- &run
		}

		return nil
	}

	// Create the subscription.
	_, err := r.client.graphqlSubscriptionClient.Subscribe(&target, variables, runEventCallback)
	if err != nil {
		return nil, err
	}

	return eventChannel, nil
}

//////////////////////////////////////////////////////////////////////////////

// The GetRuns paginator:

// RunPaginator is a type-specific paginator.
type RunPaginator struct {
	generic paginators.Paginator
}

// newRunPaginator returns a new run paginator.
func newRunPaginator(client Client, input *types.GetRunsInput) RunPaginator {
	inputCopy := &types.GetRunsInput{
		Sort:              input.Sort,
		PaginationOptions: input.PaginationOptions,
		Filter:            input.Filter,
	}

	// First return value is a GetRunsOutput, which implements PaginatedResponse.
	queryCallback := func(ctx context.Context, after *string) (interface{}, error) {
		inputCopy.PaginationOptions.Cursor = after
		return client.Run.GetRuns(ctx, inputCopy)
	}

	genericPaginator := paginators.NewPaginator(queryCallback)

	return RunPaginator{
		generic: genericPaginator,
	}
}

// HasMore returns a boolean, whether there is another page (or more):
func (rp *RunPaginator) HasMore() bool {
	return rp.generic.HasMore()
}

// Next returns the next page of results:
func (rp *RunPaginator) Next(ctx context.Context) (*types.GetRunsOutput, error) {

	// The generic paginator runs the query.
	untyped, err := rp.generic.Next(ctx)
	if err != nil {
		return nil, err
	}

	// We know the returned data is a *GetRunsOutput:
	return untyped.(*types.GetRunsOutput), nil
}

//////////////////////////////////////////////////////////////////////////////

// getRuns executes the query and returns the results.
func getRuns(ctx context.Context, client graphqlClient,
	input *types.GetRunsInput, after *string) (*getRunsQuery, error) {

	// Must generate a new query structure for each page to
	// avoid the reflect slice index out of range panic.
	queryStructP := &getRunsQuery{}

	// Build the variables for filtering, sorting, and pagination.
	variables := map[string]interface{}{}

	// Shared input variables--possible candidates to factor out:
	if input.PaginationOptions.Limit != nil {
		variables["first"] = graphql.Int(*input.PaginationOptions.Limit)
	}
	if input.PaginationOptions.Cursor == nil {
		variables["after"] = (*graphql.String)(nil)
	} else {
		variables["after"] = graphql.String(*input.PaginationOptions.Cursor)
	}

	// after overrides input
	if after != nil {
		variables["after"] = graphql.String(*after)
	}

	// Resource type specific settings:
	// Make sure to pass the expected types for these variables.
	var workspacePath, workspaceID *graphql.String
	if input.Filter != nil {
		if input.Filter.WorkspacePath != nil {
			workspacePathString := graphql.String(*input.Filter.WorkspacePath)
			workspacePath = &workspacePathString
		}
		if input.Filter.WorkspaceID != nil {
			workspaceIDString := graphql.String(*input.Filter.WorkspaceID)
			workspaceID = &workspaceIDString
		}
	} else {
		workspacePath = nil
		workspaceID = nil
	}
	if workspacePath != nil {
		variables["workspacePath"] = workspacePath
	}
	if workspaceID != nil {
		variables["workspaceID"] = workspaceID
	}

	type RunSort string
	variables["sort"] = RunSort(*input.Sort)

	// Now, do the query.
	err := client.Query(ctx, true, queryStructP, variables)
	if err != nil {
		return nil, err
	}
	return queryStructP, nil
}

//////////////////////////////////////////////////////////////////////////////

// The query structure:

// getRunsQuery is the query structure for GetRuns.
// It contains the tag with the include-everything argument list.
type getRunsQuery struct {
	Runs struct {
		PageInfo struct {
			EndCursor   graphql.String
			HasNextPage graphql.Boolean
		}
		Edges      []struct{ Node graphQLRun }
		TotalCount graphql.Int
	} `graphql:"runs(first: $first, after: $after, workspacePath: $workspacePath, sort: $sort)"`
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

// graphQLRun represents the insides of the query structure,
// everything in the run object,
// and with graphql types.
type graphQLRun struct {
	Metadata               internal.GraphQLMetadata
	ModuleSource           *graphql.String
	Apply                  *graphQLApply
	ConfigurationVersion   *struct{ ID graphql.String }
	ModuleVersion          *graphql.String
	ModuleDigest           *graphql.String
	ForceCanceledBy        *graphql.String
	ForceCancelAvailableAt *graphql.String
	StateVersion           *struct{ ID graphql.String }
	Workspace              struct {
		ID       graphql.String
		FullPath graphql.String
	}
	TargetAddresses  []graphql.String
	CreatedBy        graphql.String
	Status           graphql.String
	ID               graphql.String
	TerraformVersion graphql.String
	Plan             graphQLPlan
	IsDestroy        graphql.Boolean
	ForceCanceled    graphql.Boolean
	Refresh          graphql.Boolean
	RefreshOnly      graphql.Boolean
	Speculative      graphql.Boolean
}

type graphQLRunVariable struct {
	Value              *graphql.String
	NamespacePath      *graphql.String
	Key                graphql.String
	Category           graphql.String
	Hcl                graphql.Boolean
	Sensitive          graphql.Boolean
	VersionID          *graphql.String
	IncludedInTFConfig *bool
}

type graphQLRunVariableSensitiveValue struct {
	VersionID graphql.String
	Value     graphql.String
}

// runFromGraphQL converts a GraphQL Run to an external Run.
func runFromGraphQL(g graphQLRun) types.Run {
	result := types.Run{
		Metadata:               internal.MetadataFromGraphQL(g.Metadata, g.ID),
		CreatedBy:              string(g.CreatedBy),
		Status:                 types.RunStatus(g.Status),
		IsDestroy:              bool(g.IsDestroy),
		WorkspaceID:            string(g.Workspace.ID),
		WorkspacePath:          string(g.Workspace.FullPath),
		ModuleSource:           (*string)(g.ModuleSource),
		ModuleVersion:          (*string)(g.ModuleVersion),
		ModuleDigest:           (*string)(g.ModuleDigest),
		ForceCanceledBy:        (*string)(g.ForceCanceledBy),
		ForceCancelAvailableAt: timeFromGraphQL(g.ForceCancelAvailableAt),
		ForceCanceled:          bool(g.ForceCanceled),
		TerraformVersion:       string(g.TerraformVersion),
		Refresh:                bool(g.Refresh),
		RefreshOnly:            bool(g.RefreshOnly),
		Speculative:            bool(g.Speculative),
	}
	result.Plan = planFromGraphQL(g.Plan)
	if a := g.Apply; a != nil {
		result.Apply = applyFromGraphQL(a)
	}

	if g.ConfigurationVersion != nil {
		cvID := string(g.ConfigurationVersion.ID)
		result.ConfigurationVersionID = &cvID
	}

	if g.TargetAddresses != nil {
		result.TargetAddresses = make([]string, len(g.TargetAddresses))
		for i, v := range g.TargetAddresses {
			result.TargetAddresses[i] = string(v)
		}
	}

	if g.StateVersion != nil {
		svID := string(g.StateVersion.ID)
		result.StateVersionID = &svID
	}

	return result
}

// runVariableFromGraphQL
func runVariableFromGraphQL(v graphQLRunVariable) types.RunVariable {
	result := types.RunVariable{
		Key:                string(v.Key),
		Value:              (*string)(v.Value),
		Category:           types.VariableCategory(v.Category),
		NamespacePath:      (*string)(v.NamespacePath),
		Sensitive:          bool(v.Sensitive),
		VersionID:          (*string)(v.VersionID),
		IncludedInTFConfig: v.IncludedInTFConfig,
	}
	return result
}

// timeFromGraphQL converts a GraphQL String to a time.Time
// in case of a parsing/conversion error, return nil (shouldn't ever happen)
func timeFromGraphQL(s *graphql.String) *time.Time {
	if s == nil {
		return nil
	}

	result, err := time.Parse(time.RFC3339, string(*s))
	if err != nil {
		return nil
	}
	return &result
}
