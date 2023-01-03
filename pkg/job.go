package tharsis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// Job implements functions related to Tharsis jobs.
type Job interface {
	GetJob(ctx context.Context, input *types.GetJobInput) (*types.Job, error)
	SubscribeToJobCancellationEvent(ctx context.Context, input *types.JobCancellationEventSubscriptionInput) (<-chan *types.CancellationEvent, error)
	SaveJobLogs(ctx context.Context, input *types.SaveJobLogsInput) error
	GetJobLogs(ctx context.Context, input *types.GetJobLogsInput) (chan string, error)
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

	err := j.client.graphqlClient.Query(ctx, &target, variables)
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

	err := j.client.graphqlClient.Mutate(ctx, &wrappedSave, variables)
	if err != nil {
		return err
	}

	if err = errorFromGraphqlProblems(wrappedSave.SaveLogs.Problems); err != nil {
		return err
	}

	return nil
}

// GetJobLogs launches a goroutine that periodically fetches job logs and sends
// them to the channel.  It closes the channel after the job has finished.
// It passes the log strings through with _NO_ attempt to glue together split
// lines or any other fancy processing.
func (j *job) GetJobLogs(ctx context.Context, input *types.GetJobLogsInput) (chan string, error) {
	startOffset := input.StartOffset
	logChannel := make(chan string)
	pollForLogs := func() {
		defer close(logChannel)
		for {

			variables := map[string]interface{}{
				"id":          graphql.String(input.ID),
				"startOffset": graphql.Int(startOffset),
				"limit":       graphql.Int(input.Limit),
			}

			var target struct {
				Job *struct {
					ID     graphql.String
					Status graphql.String
					Logs   graphql.String `graphql:"logs(startOffset: $startOffset, limit: $limit)"`
				} `graphql:"job(id: $id)"`
			}

			err := j.client.graphqlClient.Query(ctx, &target, variables)
			if err != nil {
				j.client.cfg.Logger.Printf("error: failed to query job: %s", err)
				return
			}
			if target.Job == nil {
				j.client.cfg.Logger.Printf("error: job not found: %s", input.ID)
				return
			}

			logs := string(target.Job.Logs)
			status := string(target.Job.Status)

			logChannel <- logs
			startOffset += len(logs)
			if status == jobFinished {
				return
			}

			time.Sleep(jobLogQuerySleep)
		}
	}

	go func() {
		pollForLogs()
	}()

	return logChannel, nil
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
