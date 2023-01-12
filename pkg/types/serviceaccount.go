package types

import "time"

// OIDCTrustPolicy models one trust policy under a service account.
type OIDCTrustPolicy struct {
	BoundClaims map[string]string `json:"boundClaims"`
	Issuer      string            `json:"issuer"`
}

// ServiceAccount provides M2M authentication
type ServiceAccount struct {
	Metadata          ResourceMetadata
	ResourcePath      string
	Name              string
	Description       string
	OIDCTrustPolicies []OIDCTrustPolicy
}

// CreateServiceAccountInput is the input for creating a service account.
type CreateServiceAccountInput struct {
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	GroupPath         string            `json:"groupPath"`
	OIDCTrustPolicies []OIDCTrustPolicy `json:"oidcTrustPolicies"`
}

// GetServiceAccountInput is the input for retrieving
// a service account.
type GetServiceAccountInput struct {
	ID string `json:"id"`
}

// UpdateServiceAccountInput is the input for updating a service account.
type UpdateServiceAccountInput struct {
	ID                string            `json:"id"`
	Description       string            `json:"description"`
	OIDCTrustPolicies []OIDCTrustPolicy `json:"oidcTrustPolicies"`
}

// DeleteServiceAccountInput is the input for deleting a service account.
type DeleteServiceAccountInput struct {
	ID string `json:"id"`
}

// ServiceAccountLoginInput is the input for logging in to a service account.
type ServiceAccountLoginInput struct {
	ServiceAccountPath string `json:"serviceAccountPath"`
	Token              string `json:"token"`
}

// ServiceAccountLoginResponse is the output from logging in to a service account.
type ServiceAccountLoginResponse struct {
	Token     string        `json:"token"`
	ExpiresIn time.Duration `json:"expiresIn"`
}

// The End.
