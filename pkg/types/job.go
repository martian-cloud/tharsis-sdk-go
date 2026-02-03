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
	Limit           *int32
	LastSeenLogSize *int32
	RunID           string
	WorkspacePath   string
	JobID           string `json:"jobId"`
}

// GetJobLogsInput is the input to query a chunk of job logs
type GetJobLogsInput struct {
	Limit *int32
	JobID string
	Start int32
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
	WorkspaceID     string
	Tags            []string
	Properties      map[string]string
	LogSize         int
	MaxJobDuration  int32
	CancelRequested bool
}

// CancellationEvent represents a job cancellation event
type CancellationEvent struct {
	Job Job
}

// ClaimJobInput is the input for claiming a job
type ClaimJobInput struct {
	RunnerPath string `json:"runnerPath"`
}

// ClaimJobResponse is the response when claiming a job
type ClaimJobResponse struct {
	Token string
	JobID string
}

// JobLogsEvent is the output for subscribing to job logs.
type JobLogsEvent struct {
	Error error
	Logs  string
}

// JobLogs is the output for getting job logs after a job has finished.
type JobLogs struct {
	Logs string
	Size int32
}
