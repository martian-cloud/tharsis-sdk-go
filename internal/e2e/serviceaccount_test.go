//go:build integration
// +build integration

package main

import (
	"context"
	"testing"

	"github.com/likexian/gokit/assert"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TestCRUDServiceAccount tests service account create, get, update, and delete.
func TestCRUDServiceAccount(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	serviceAccountName := "service-account-01"
	serviceAccountDescription := "service account description 01"
	updatedDescription := "service account description 01, updated description"
	trustPolicyIssuer := "http://trust/policy/issuer"
	boundClaimName := "bound-claim-name"
	boundClaimValue := "bound-claim-value"
	updatedTrustPolicyIssuer := "http://trust/policy/issuer/updated"
	updatedBoundClaimName := "updated-bound-claim-name"
	updatedBoundClaimValue := "updated-bound-claim-value"

	// Create the service account.
	createdServiceAccount, err := client.ServiceAccount.CreateServiceAccount(ctx, &types.CreateServiceAccountInput{
		Name:        serviceAccountName,
		Description: serviceAccountDescription,
		GroupPath:   topGroupName,
		OIDCTrustPolicies: []types.OIDCTrustPolicyInput{
			{
				Issuer: trustPolicyIssuer,
				BoundClaims: []types.JWTClaimInput{
					{
						Name:  boundClaimName,
						Value: boundClaimValue,
					},
				},
			},
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, createdServiceAccount)

	// Get/read and verify the service account.
	readServiceAccount, err := client.ServiceAccount.GetServiceAccount(ctx, &types.GetServiceAccountInput{
		ID: createdServiceAccount.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, readServiceAccount, createdServiceAccount)

	// Update the service account.
	updatedServiceAccount, err := client.ServiceAccount.UpdateServiceAccount(ctx,
		&types.UpdateServiceAccountInput{
			ID:          readServiceAccount.Metadata.ID,
			Description: updatedDescription,
			OIDCTrustPolicies: []types.OIDCTrustPolicyInput{
				{
					Issuer: updatedTrustPolicyIssuer,
					BoundClaims: []types.JWTClaimInput{
						{
							Name:  updatedBoundClaimName,
							Value: updatedBoundClaimValue,
						},
					},
				},
			},
		})
	assert.Nil(t, err)

	// Verify the claimed update.
	assert.Equal(t, updatedServiceAccount.Description, updatedDescription)

	// Retrieve and verify the updated service account to make sure it persisted.
	read2ServiceAccount, err := client.ServiceAccount.GetServiceAccount(ctx, &types.GetServiceAccountInput{
		ID: createdServiceAccount.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, read2ServiceAccount, updatedServiceAccount)

	// Delete the service account.
	err = client.ServiceAccount.DeleteServiceAccount(ctx,
		&types.DeleteServiceAccountInput{
			ID: read2ServiceAccount.Metadata.ID,
		})
	assert.Nil(t, err)

	// Verify the service account is gone.
	read3ServiceAccount, err := client.ServiceAccount.GetServiceAccount(ctx, &types.GetServiceAccountInput{
		ID: read2ServiceAccount.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, read3ServiceAccount, (*types.ServiceAccount)(nil))
}

// The End.
