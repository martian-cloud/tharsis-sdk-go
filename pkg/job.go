package tharsis

import (
	"context"
	"encoding/json"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// Job implements functions related to Tharsis jobs.
type Job interface {
	GetJob(ctx context.Context, input *types.GetJobInput) (*types.Job, error)
	SubscribeToJobCancellationEvent(ctx context.Context, input *types.JobCancellationEventSubscriptionInput) (<-chan *types.CancellationEvent, error)
	SaveJobLogs(ctx context.Context, input *types.SaveJobLogsInput) error
	SubscribeToJobLogEvents(ctx context.Context, input *types.JobLogSubscriptionInput) (<-chan *types.JobLogEvent, error)
	GetJobLogs(ctx context.Context, input *types.GetJobLogsInput) (*types.GetJobLogsOutput, error)
}

type job struct {
	client *Client
}

// NewJob returns a Job.
func NewJob(client *Client) Job {
	return &job{client: client}
}

// GetJob returns everything about the job.
func (j *job) GetJob(ctx context.Context, input *types.GetJobInput) (*types.Job, error) {
	var target struct {
		Job *graphQLJob `graphql:"job(id: $id)"`
	}
	variables := map[string]interface{}{
		"id": graphql.String(input.ID),
	}

	err := j.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}
	if target.Job == nil {
		return nil, newError(ErrNotFound, "job with id %s not found", input.ID)
	}

	result := jobFromGraphQL(*target.Job)
	return &result, nil
}

// SubscribeToJobCancellationEvent queries for a job cancellation event returns its content.
func (j *job) SubscribeToJobCancellationEvent(ctx context.Context, input *types.JobCancellationEventSubscriptionInput) (<-chan *types.CancellationEvent, error) {

	eventChannel := make(chan *types.CancellationEvent)

	var target struct {
		CancellationEvent graphQLCancellationEvent `graphql:"jobCancellationEvent(input: $input)"`
	}
	variables := map[string]interface{}{
		"input": *input,
	}

	// The embedded cancellation event callback function.
	cancellationEventCallback := func(message []byte, err error) error {
		// Detect any incoming error.
		if err != nil {
			// close channel
			close(eventChannel)
			return err
		}

		var event struct {
			CancellationEvent graphQLCancellationEvent `json:"jobCancellationEvent"`
		}

		if message != nil {
			err = json.Unmarshal(message, &event)
			if err != nil {
				return err
			}

			ce := cancellationEventFromGraphQL(event.CancellationEvent)
			eventChannel <- &ce
		}

		return nil
	}

	// Create the subscription.
	_, err := j.client.graphqlSubscriptionClient.Subscribe(&target, variables, cancellationEventCallback)
	if err != nil {
		return nil, err
	}

	return eventChannel, nil
}

// SaveJobLogs saves the logs for a job.
func (j *job) SaveJobLogs(ctx context.Context, input *types.SaveJobLogsInput) error {
	var wrappedSave struct {
		SaveLogs struct {
			Problems []internal.GraphQLProblem
		} `graphql:"saveJobLogs(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	err := j.client.graphqlClient.Mutate(ctx, true, &wrappedSave, variables)
	if err != nil {
		return err
	}

	if err = errorFromGraphqlProblems(wrappedSave.SaveLogs.Problems); err != nil {
		return err
	}

	return nil
}

func (j *job) SubscribeToJobLogEvents(ctx context.Context, input *types.JobLogSubscriptionInput) (<-chan *types.JobLogEvent, error) {
	// Make sure the run hasn't already finished.
	run, err := j.client.Run.GetRun(ctx, &types.GetRunInput{ID: input.RunID})
	if err != nil {
		return nil, err
	}

	switch run.Status {
	case types.RunApplied,
		types.RunCanceled,
		types.RunPlanned,
		types.RunPlannedAndFinished,
		types.RunErrored:
		// Run has finished meaning no more job logs.
		return nil, newError(ErrBadRequest, "Run has already %s", run.Status)
	}

	eventChannel := make(chan *types.JobLogEvent)

	var target struct {
		JobLogEvent types.JobLogEvent `graphql:"jobLogEvents(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// The embedded job log event callback function.
	jobLogEventCallback := func(message []byte, err error) error {
		// Detect any incoming error.
		if err != nil {
			// close channel
			close(eventChannel)
			return err
		}

		var event struct {
			JobLogEvent types.JobLogEvent `json:"jobLogEvents"`
		}

		if message != nil {
			if err = json.Unmarshal(message, &event); err != nil {
				return err
			}

			eventChannel <- &event.JobLogEvent
		}

		return nil
	}

	// Create the subscription.
	_, err = j.client.graphqlSubscriptionClient.Subscribe(&target, variables, jobLogEventCallback)
	if err != nil {
		return nil, err
	}

	return eventChannel, nil
}

// GetJobLogs returns job logs.
func (j *job) GetJobLogs(ctx context.Context, input *types.GetJobLogsInput) (*types.GetJobLogsOutput, error) {
	variables := map[string]interface{}{
		"id":          graphql.String(input.ID),
		"startOffset": graphql.Int(input.StartOffset),
		"limit":       graphql.Int(input.Limit),
	}

	var target struct {
		Job *struct {
			Logs    graphql.String `graphql:"logs(startOffset: $startOffset, limit: $limit)"`
			LogSize graphql.Int
		} `graphql:"job(id: $id)"`
	}

	err := j.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}

	if target.Job == nil {
		return nil, newError(ErrNotFound, "Job with id %s not found", input.ID)
	}

	return &types.GetJobLogsOutput{
		Logs:    string(target.Job.Logs),
		LogSize: int32(target.Job.LogSize),
	}, nil
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

// graphQLJob represents the insides of the query structure,
// everything in the job object,
// and with graphql types.
type graphQLJob struct {
	ID       graphql.String
	Metadata internal.GraphQLMetadata
	Status   graphql.String
	Type     graphql.String
	Run      struct {
		ID graphql.String
	}
	Workspace struct {
		FullPath graphql.String
	}
	CancelRequested graphql.Boolean
	LogSize         graphql.Int
	MaxJobDuration  graphql.Int
}

type graphQLCancellationEvent struct {
	Job graphQLJob `json:"job"`
}

// jobFromGraphQL converts a GraphQL Job to an external Job.
func jobFromGraphQL(r graphQLJob) types.Job {
	result := types.Job{
		Metadata:        internal.MetadataFromGraphQL(r.Metadata, r.ID),
		Status:          string(r.Status),
		Type:            types.JobType(r.Type),
		RunID:           string(r.Run.ID),
		WorkspacePath:   string(r.Workspace.FullPath),
		CancelRequested: bool(r.CancelRequested),
		LogSize:         int(r.LogSize),
		MaxJobDuration:  int32(r.MaxJobDuration),
	}
	return result
}

// cancellationEventFromGraphQL converts a GraphQL Cancellation Event
// to external cancellation event
func cancellationEventFromGraphQL(r graphQLCancellationEvent) types.CancellationEvent {
	result := types.CancellationEvent{
		Job: jobFromGraphQL(r.Job),
	}

	return result
}

// The End.
