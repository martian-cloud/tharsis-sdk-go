package tharsis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal/errors"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

const (
	// defaultLogLimit is the log limit used when nothing is defined.
	defaultLogLimit int32 = 1024 * 1024
)

// JobLogStreamSubscriptionInput is the input for subscribing to job log events.
type JobLogStreamSubscriptionInput struct {
	LastSeenLogSize *int32 `json:"lastSeenLogSize"`
	JobID           string `json:"jobId"`
}

// jobLogStreamEvent represents a job log event.
type jobLogStreamEvent struct {
	Completed bool  `json:"completed"`
	Size      int32 `json:"size"`
}

// getJobLogsInput is the input to query a chunk of job logs
type getJobLogsInput struct {
	limit       *int32
	id          string
	startOffset int32
}

// getJobLogsOutput is the output for retrieving job logs.
type getJobLogsOutput struct {
	logs    string
	logSize int32
}

// Job implements functions related to Tharsis jobs.
type Job interface {
	GetJob(ctx context.Context, input *types.GetJobInput) (*types.Job, error)
	ClaimJob(ctx context.Context, input *types.ClaimJobInput) (*types.ClaimJobResponse, error)
	SubscribeToJobCancellationEvent(ctx context.Context, input *types.JobCancellationEventSubscriptionInput) (<-chan *types.CancellationEvent, error)
	SaveJobLogs(ctx context.Context, input *types.SaveJobLogsInput) error
	SubscribeToJobLogs(ctx context.Context, input *types.JobLogsSubscriptionInput) (<-chan *types.JobLogsEvent, error)
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
		return nil, errors.NewError(types.ErrNotFound, "job with id %s not found", input.ID)
	}

	result := jobFromGraphQL(*target.Job)
	return &result, nil
}

func (j *job) ClaimJob(ctx context.Context, input *types.ClaimJobInput) (*types.ClaimJobResponse, error) {
	var req struct {
		ClaimJob struct {
			JobID    *string
			Token    *string
			Problems []internal.GraphQLProblem
		} `graphql:"claimJob(input: $input)"`
	}

	// Creating a new object requires the wrapped object above
	// but with all the contents in a struct in the variables.
	variables := map[string]interface{}{
		"input": *input,
	}

	err := j.client.graphqlClient.Mutate(ctx, true, &req, variables)
	if err != nil {
		return nil, err
	}

	if err = errors.ErrorFromGraphqlProblems(req.ClaimJob.Problems); err != nil {
		return nil, err
	}

	response := types.ClaimJobResponse{
		Token: *req.ClaimJob.Token,
		JobID: *req.ClaimJob.JobID,
	}

	return &response, nil
}

// SubscribeToJobCancellationEvent queries for a job cancellation event returns its content.
func (j *job) SubscribeToJobCancellationEvent(_ context.Context, input *types.JobCancellationEventSubscriptionInput) (<-chan *types.CancellationEvent, error) {

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
			if err = json.Unmarshal(message, &event); err != nil {
				return err
			}

			eventChannel <- cancellationEventFromGraphQL(event.CancellationEvent)
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

	return errors.ErrorFromGraphqlProblems(wrappedSave.SaveLogs.Problems)
}

func (j *job) SubscribeToJobLogs(ctx context.Context, input *types.JobLogsSubscriptionInput) (<-chan *types.JobLogsEvent, error) {
	logChan := make(chan *types.JobLogsEvent)

	// Subscribe to run events so we know when to stop outputting logs.
	runEvents, err := j.client.Run.SubscribeToWorkspaceRunEvents(ctx, &types.RunSubscriptionInput{
		RunID:         &input.RunID,
		WorkspacePath: input.WorkspacePath,
	})
	if err != nil {
		return nil, err
	}

	// Subscribe to the job log events so we can fetch logs only when new ones are available.
	logEvents, err := j.subscribeToJobLogStreamEvents(ctx, &JobLogStreamSubscriptionInput{
		JobID:           input.JobID,
		LastSeenLogSize: input.LastSeenLogSize,
	})
	if err != nil {
		return nil, err
	}

	logFetcher := func() {
		defer close(logChan)

		var (
			currentOffset int32
			runCompleted  bool
		)

		run, err := j.client.Run.GetRun(ctx, &types.GetRunInput{ID: input.RunID})
		if err != nil {
			logChan <- toJobLogsEvent("", err)
			return
		}

		// Make sure the run hasn't already finished.
		// In which case the for loop will fetch all the pending logs and
		// close the channel once finished.
		switch run.Status {
		case types.RunApplied,
			types.RunCanceled,
			types.RunPlanned,
			types.RunPlannedAndFinished,
			types.RunErrored:
			runCompleted = true
		}

		for {
			// Retrieve the plan logs.
			output, err := j.getJobLogs(ctx, &getJobLogsInput{
				id:          input.JobID,
				startOffset: currentOffset,
				limit:       input.Limit,
			})
			if err != nil {
				logChan <- toJobLogsEvent("", err)
				return
			}

			// Update the offset to the new log size.
			currentOffset += int32(len(output.logs))

			if len(output.logs) > 0 {
				// Send the logs on the channel.
				logChan <- toJobLogsEvent(output.logs, nil)
			}

			if runCompleted {
				if currentOffset < output.logSize {
					// Since all the logs haven't been sent, keep looping.
					continue
				}

				// Run has finished and all pending logs are sent.
				return
			}

			select {
			case <-ctx.Done():
				logChan <- toJobLogsEvent("", ctx.Err())
				return
			case <-logEvents:
			// This is a failsafe in case the subscription connection is closed due to a network issue
			case <-time.After(time.Second * 30):
			case eventRun := <-runEvents:
				switch eventRun.Status {
				case types.RunApplied,
					types.RunCanceled,
					types.RunPlanned,
					types.RunPlannedAndFinished,
					types.RunErrored:

					runCompleted = true
				}
			}
		}
	}

	// Fetch logs in a goroutine.
	go logFetcher()

	return logChan, nil
}

func (j *job) subscribeToJobLogStreamEvents(_ context.Context,
	input *JobLogStreamSubscriptionInput) (<-chan *jobLogStreamEvent, error) {
	eventChannel := make(chan *jobLogStreamEvent)

	var target struct {
		JobLogStreamEvent jobLogStreamEvent `graphql:"jobLogStreamEvents(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// The embedded job log event callback function.
	jobLogStreamEventCallback := func(message []byte, err error) error {
		// Detect any incoming error.
		if err != nil {
			// close channel
			close(eventChannel)
			return err
		}

		var event struct {
			JobLogStreamEvent jobLogStreamEvent `json:"jobLogStreamEvents"`
		}

		if message != nil {
			if err = json.Unmarshal(message, &event); err != nil {
				return err
			}

			eventChannel <- &event.JobLogStreamEvent
		}

		return nil
	}

	// Create the subscription.
	_, err := j.client.graphqlSubscriptionClient.Subscribe(&target, variables, jobLogStreamEventCallback)
	if err != nil {
		return nil, err
	}

	return eventChannel, nil
}

func (j *job) getJobLogs(ctx context.Context, input *getJobLogsInput) (*getJobLogsOutput, error) {
	// Use default value for log limit if nil.
	limit := defaultLogLimit
	if input.limit != nil {
		limit = *input.limit
	}

	variables := map[string]interface{}{
		"id":          graphql.String(input.id),
		"startOffset": graphql.Int(input.startOffset),
		"limit":       graphql.Int(limit),
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
		return nil, errors.NewError(types.ErrNotFound, "Job with id %s not found", input.id)
	}

	return &getJobLogsOutput{
		logs:    string(target.Job.Logs),
		logSize: int32(target.Job.LogSize),
	}, nil
}

// toJobLogsEvent returns a JobLogsEvent struct.
func toJobLogsEvent(logs string, err error) *types.JobLogsEvent {
	return &types.JobLogsEvent{
		Logs:  logs,
		Error: err,
	}
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
		ID       graphql.String
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
		WorkspaceID:     string(r.Workspace.ID),
		CancelRequested: bool(r.CancelRequested),
		LogSize:         int(r.LogSize),
		MaxJobDuration:  int32(r.MaxJobDuration),
	}
	return result
}

// cancellationEventFromGraphQL converts a GraphQL Cancellation Event
// to external cancellation event
func cancellationEventFromGraphQL(r graphQLCancellationEvent) *types.CancellationEvent {
	result := &types.CancellationEvent{
		Job: jobFromGraphQL(r.Job),
	}

	return result
}

// The End.
