//go:build integration
// +build integration

package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
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
		OIDCTrustPolicies: []types.OIDCTrustPolicy{
			{
				Issuer: trustPolicyIssuer,
				BoundClaims: map[string]string{
					boundClaimName: boundClaimValue,
				},
			},
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, createdServiceAccount)

	// Verify all the fields except metadata.
	expectTrustPolicies := []types.OIDCTrustPolicy{
		{
			Issuer:      trustPolicyIssuer,
			BoundClaims: map[string]string{boundClaimName: boundClaimValue},
		},
	}
	assert.Equal(t, topGroupName+"/"+serviceAccountName, createdServiceAccount.ResourcePath)
	assert.Equal(t, serviceAccountName, createdServiceAccount.Name)
	assert.Equal(t, serviceAccountDescription, createdServiceAccount.Description)
	assert.Equal(t, expectTrustPolicies, createdServiceAccount.OIDCTrustPolicies)
	assert.Equal(t, 1, len(createdServiceAccount.OIDCTrustPolicies))
	assert.Equal(t, 1, len(createdServiceAccount.OIDCTrustPolicies[0].BoundClaims))

	// Get/read and verify the service account.
	readServiceAccount, err := client.ServiceAccount.GetServiceAccount(ctx, &types.GetServiceAccountInput{
		ID: createdServiceAccount.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, createdServiceAccount, readServiceAccount)

	// Update the service account.
	updatedServiceAccount, err := client.ServiceAccount.UpdateServiceAccount(ctx,
		&types.UpdateServiceAccountInput{
			ID:          readServiceAccount.Metadata.ID,
			Description: updatedDescription,
			OIDCTrustPolicies: []types.OIDCTrustPolicy{
				{
					Issuer: updatedTrustPolicyIssuer,
					BoundClaims: map[string]string{
						updatedBoundClaimName: updatedBoundClaimValue,
					},
				},
			},
		})
	assert.Nil(t, err)

	// Verify the claimed update.
	assert.Equal(t, updatedDescription, updatedServiceAccount.Description)

	// Retrieve and verify the updated service account to make sure it persisted.
	read2ServiceAccount, err := client.ServiceAccount.GetServiceAccount(ctx, &types.GetServiceAccountInput{
		ID: createdServiceAccount.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, updatedServiceAccount, read2ServiceAccount)

	// Delete the service account.
	err = client.ServiceAccount.DeleteServiceAccount(ctx,
		&types.DeleteServiceAccountInput{
			ID: read2ServiceAccount.Metadata.ID,
		})
	assert.Nil(t, err)

	// Verify the service account is gone.
	_, err = client.ServiceAccount.GetServiceAccount(ctx, &types.GetServiceAccountInput{
		ID: read2ServiceAccount.Metadata.ID,
	})
	assert.True(t, tharsis.IsNotFoundError(err))
}

// The End.
