package types

// JobType indicates the type of job
type JobType string

// Job Types Constants
const (
	JobPlanType  JobType = "plan"
	JobApplyType JobType = "apply"
)

// GetJobInput is the input to specify a single job to fetch.
type GetJobInput struct {
	ID string `json:"id"`
}

// GetJobLogsInput is the input to query a chunk of job logs
type GetJobLogsInput struct {
	ID          string `json:"id"`
	StartOffset int32  `json:"startOffset"`
	Limit       int32  `json:"limit"`
}

// JobLogSubscriptionInput is the input for subscribing to job log events.
type JobLogSubscriptionInput struct {
	RunID           string // Used to make sure run hasn't already finished.
	LastSeenLogSize *int32 `json:"lastSeenLogSize"`
	JobID           string `json:"jobId"`
}

// JobCancellationEventSubscriptionInput is the input for Job cancellation event subscription
type JobCancellationEventSubscriptionInput struct {
	JobID string `json:"jobId"`
}

// SaveJobLogsInput is the input for saving job logs.
type SaveJobLogsInput struct {
	Logs        string `json:"logs"`
	JobID       string `json:"jobId"`
	StartOffset int32  `json:"startOffset"`
}

// Job holds information about a Tharsis job.
// It is used as input to and output from some operations.
// ID resides in the metadata
type Job struct {
	Metadata        ResourceMetadata
	Status          string
	Type            JobType
	RunID           string
	WorkspacePath   string
	LogSize         int
	MaxJobDuration  int32
	CancelRequested bool
}

// CancellationEvent represents a job cancellation event
type CancellationEvent struct {
	Job Job
}

// GetJobLogsOutput is the output for retrieving job logs.
type GetJobLogsOutput struct {
	Logs    string
	LogSize int32
}

// JobLogEvent represents a job log event.
type JobLogEvent struct {
	Action string `json:"action"`
	Size   int32  `json:"size"`
}

// The End.
