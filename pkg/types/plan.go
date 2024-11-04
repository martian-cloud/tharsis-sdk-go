package types

// PlanStatus represents the various states for a Plan resource
type PlanStatus string

// Run Status Types
const (
	PlanCanceled PlanStatus = "canceled"
	PlanQueued   PlanStatus = "queued"
	PlanErrored  PlanStatus = "errored"
	PlanFinished PlanStatus = "finished"
	PlanPending  PlanStatus = "pending"
	PlanRunning  PlanStatus = "running"
)

// Plan holds information about a Tharsis plan.
// ID resides in the metadata
type Plan struct {
	CurrentJobID         *string
	Metadata             ResourceMetadata
	Status               PlanStatus
	ResourceAdditions    int
	ResourceChanges      int
	ResourceDestructions int
	HasChanges           bool
	ErrorMessage         *string
}

// UpdatePlanInput is the input for updating a plan
type UpdatePlanInput struct {
	ID           string     `json:"id"`
	Status       PlanStatus `json:"status"`
	HasChanges   bool       `json:"hasChanges"`
	ErrorMessage *string    `json:"errorMessage"`
}
