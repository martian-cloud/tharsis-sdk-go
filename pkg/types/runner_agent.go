package types

// RunnerType indicates the type of runner
type RunnerType string

// Runner Types Constants (currently lowercase to match the API and the API's model)
const (
	RunnerTypeShared RunnerType = "shared"
	RunnerTypeGroup  RunnerType = "group"
)

// RunnerAgent represents a Tharsis Runner
type RunnerAgent struct {
	Metadata        ResourceMetadata
	Name            string
	Description     string
	GroupPath       string
	ResourcePath    string
	CreatedBy       string
	Type            RunnerType
	Tags            []string
	RunUntaggedJobs bool
}

// GetRunnerInput is the input for retrieving a runner agent
type GetRunnerInput struct {
	ID string `json:"id"`
}

// CreateRunnerInput is the input for creating a runner agent
type CreateRunnerInput struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	GroupPath       string   `json:"groupPath"`
	Tags            []string `json:"tags"`
	RunUntaggedJobs bool     `json:"runUntaggedJobs"`
}

// UpdateRunnerInput is the input for updating a runner agent
type UpdateRunnerInput struct {
	Tags            *[]string `json:"tags"`
	RunUntaggedJobs *bool     `json:"runUntaggedJobs"`
	ID              string    `json:"id"`
	Description     string    `json:"description"`
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
