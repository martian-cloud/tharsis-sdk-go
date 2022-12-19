package internal

// These types are used internally for creating and updating service accounts.

// CreateServiceAccountInput is similar to types.CreateServiceAccountInput
type CreateServiceAccountInput struct {
	Name              string                               `json:"name"`
	Description       string                               `json:"description"`
	GroupPath         string                               `json:"groupPath"`
	OIDCTrustPolicies []ServiceAccountOIDCTrustPolicyInput `json:"oidcTrustPolicies"`
}

// UpdateServiceAccountInput is the input for updating a service account.
type UpdateServiceAccountInput struct {
	ID                string                               `json:"id"`
	Description       string                               `json:"description"`
	OIDCTrustPolicies []ServiceAccountOIDCTrustPolicyInput `json:"oidcTrustPolicies"`
}

// ServiceAccountOIDCTrustPolicyInput is similar to types.ServiceAccountOIDCTrustPolicyInput
type ServiceAccountOIDCTrustPolicyInput struct {
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
