package types

// ApplyStatus represents the various states for a Apply resource
type ApplyStatus string

// Apply Status Types
const (
	ApplyCanceled ApplyStatus = "canceled"
	ApplyCreated  ApplyStatus = "created"
	ApplyErrored  ApplyStatus = "errored"
	ApplyFinished ApplyStatus = "finished"
	ApplyPending  ApplyStatus = "pending"
	ApplyQueued   ApplyStatus = "queued"
	ApplyRunning  ApplyStatus = "running"
)

// Apply holds information about a Tharsis run.
// ID resides in the metadata
type Apply struct {
	Metadata     ResourceMetadata
	CurrentJobID *string
	Status       ApplyStatus
}

// UpdateApplyInput is the input for updating an apply.
type UpdateApplyInput struct {
	ID     string      `json:"id"`
	Status ApplyStatus `json:"status"`
}
