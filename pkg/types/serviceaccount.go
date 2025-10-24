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
	GroupPath         string
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
	ID  string  `json:"id"`
	TRN *string `json:"trn"`
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

// ServiceAccountCreateTokenInput is the input for logging in to a service account.
type ServiceAccountCreateTokenInput struct {
	ServiceAccountPath string `json:"serviceAccountPath"`
	Token              string `json:"token"`
}

// ServiceAccountCreateTokenResponse is the output from logging in to a service account.
type ServiceAccountCreateTokenResponse struct {
	Token     string        `json:"token"`
	ExpiresIn time.Duration `json:"expiresIn"`
}
