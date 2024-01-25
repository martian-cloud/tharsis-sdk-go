package types

import "time"

// CreateRunnerSessionInput is the input for creating a runner session.
type CreateRunnerSessionInput struct {
	RunnerPath string `json:"runnerPath"`
}

// RunnerSessionHeartbeatInput is the input for sending a runner session heartbeat.
type RunnerSessionHeartbeatInput struct {
	RunnerSessionID string `json:"runnerSessionId"`
}

// CreateRunnerSessionErrorInput is the input for sending a runner session error.
type CreateRunnerSessionErrorInput struct {
	RunnerSessionID string `json:"runnerSessionId"`
	ErrorMessage    string `json:"errorMessage"`
}

// RunnerSession represents a Tharsis Runner Session.
type RunnerSession struct {
	Runner        *RunnerAgent
	LastContacted *time.Time
	Metadata      ResourceMetadata
	ErrorCount    int
	Internal      bool
}
