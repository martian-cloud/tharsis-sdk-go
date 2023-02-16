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

// JobLogsSubscriptionInput is the input for subscribing to job logs.
type JobLogsSubscriptionInput struct {
	Limit         *int32
	RunID         string
	WorkspacePath string
	JobID         string `json:"jobId"`
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

// JobLogsEvent is the output for subscribing to job logs.
type JobLogsEvent struct {
	Error error
	Logs  string
}

// The End.
