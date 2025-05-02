package types

// Supporting structs for the FederatedRegistry:

// CreateFederatedRegistryTokensInput is the input for creating new federated registry tokens.
type CreateFederatedRegistryTokensInput struct {
	JobID string `json:"jobId"`
}

// FederatedRegistryToken is the output for each new federated registry token.
type FederatedRegistryToken struct {
	Hostname string `json:"hostname"`
	Token    string `json:"token"`
}
