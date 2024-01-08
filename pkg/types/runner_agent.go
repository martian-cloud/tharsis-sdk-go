package types

// RunnerAgent represents a Tharsis Runner
type RunnerAgent struct {
	Metadata     ResourceMetadata
	Name         string
	Description  string
	GroupPath    string
	ResourcePath string
	CreatedBy    string
	Type         string
}

// GetRunnerInput is the input for retrieving a runner agent
type GetRunnerInput struct {
	ID string `json:"id"`
}

// CreateRunnerInput is the input for creating a runner agent
type CreateRunnerInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	GroupPath   string `json:"groupPath"`
}

// UpdateRunnerInput is the input for updating a runner agent
type UpdateRunnerInput struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// DeleteRunnerInput is the input for deleting a runner agent
type DeleteRunnerInput struct {
	ID string `json:"id"`
}

// AssignServiceAccountToRunnerInput is the input for assigning / un-assigning a service account to / from a runner agent
type AssignServiceAccountToRunnerInput struct {
	RunnerPath         string `json:"runnerPath"`
	ServiceAccountPath string `json:"serviceAccountPath"`
}
