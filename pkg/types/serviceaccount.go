package types

// ServiceAccount provides M2M authentication
type ServiceAccount struct {
	Metadata     ResourceMetadata
	ResourcePath string
	Name         string
	Description  string
}

// CreateServiceAccountInput is the input for creating a service account.
type CreateServiceAccountInput struct {
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	GroupPath         string                 `json:"groupPath"`
	OIDCTrustPolicies []OIDCTrustPolicyInput `json:"oidcTrustPolicies"`
}

// GetServiceAccountInput is the input for retrieving
// a service account.
type GetServiceAccountInput struct {
	ID string `json:"id"`
}

// UpdateServiceAccountInput is the input for updating a service account.
type UpdateServiceAccountInput struct {
	ID                string                 `json:"id"`
	Description       string                 `json:"description"`
	OIDCTrustPolicies []OIDCTrustPolicyInput `json:"oidcTrustPolicies"`
}

// DeleteServiceAccountInput is the input for deleting a service account.
type DeleteServiceAccountInput struct {
	ID string `json:"id"`
}

// OIDCTrustPolicyInput is the input for OIDC trust policies
// when created at the same time as the service account.
type OIDCTrustPolicyInput struct {
	Issuer      string          `json:"issuer"`
	BoundClaims []JWTClaimInput `json:"boundClaims"`
}

// JWTClaimInput is the input for JWT claims
// when created at the same time as the OIDC trust policy
type JWTClaimInput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// The End.
