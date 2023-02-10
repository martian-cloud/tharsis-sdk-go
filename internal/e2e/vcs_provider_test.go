//go:build integration
// +build integration

package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestVCSProviderCRUD(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	require.Nil(t, err)
	assert.NotNil(t, client)

	// Values for various fields.  Cannot verify CreatedBy.
	name := "vcs-provider-name-1"
	description := "vcs-provider-description-1"
	hostname := "vcs-provider-hostname-1"
	clientID := "vcs-provider-client-id-1"
	clientSecret := "vcs-provider-client-secret-1"
	resourcePath := topGroupName + "/" + name
	vcsProviderType := types.VCSProviderTypeGitlab
	autoCreateWebhooks := false

	updatedDescription := "vcs-provider-description-1--updated"
	updatedClientID := "vcs-provider-client-id-1--updated"
	updatedClientSecret := "vcs-provider-client-secret-1--updated"

	// Create the VCS provider.
	toCreate := &types.CreateVCSProviderInput{
		Name:               name,
		Description:        description,
		GroupPath:          topGroupName,
		Hostname:           &hostname,
		OAuthClientID:      clientID,
		OAuthClientSecret:  clientSecret,
		Type:               vcsProviderType,
		AutoCreateWebhooks: autoCreateWebhooks,
	}
	createdVCSProvider, err := client.VCSProvider.CreateProvider(ctx, toCreate)

	require.Nil(t, err)
	assert.NotNil(t, createdVCSProvider)
	assert.Equal(t, name, createdVCSProvider.Name)
	assert.Equal(t, description, createdVCSProvider.Description)
	assert.Equal(t, hostname, createdVCSProvider.Hostname)
	assert.Equal(t, resourcePath, createdVCSProvider.ResourcePath)
	assert.Equal(t, vcsProviderType, createdVCSProvider.Type)
	assert.Equal(t, autoCreateWebhooks, createdVCSProvider.AutoCreateWebhooks)

	// Get the VCS provider to make sure it persisted.
	gotVCSProvider, err := client.VCSProvider.GetProvider(ctx, &types.GetVCSProviderInput{
		ID: createdVCSProvider.Metadata.ID,
	})
	require.Nil(t, err)
	assert.NotNil(t, gotVCSProvider)

	// Verify the returned contents are what they should be.
	assert.Equal(t, createdVCSProvider.Metadata, gotVCSProvider.Metadata)
	assert.Equal(t, createdVCSProvider.CreatedBy, gotVCSProvider.CreatedBy)
	assert.Equal(t, name, gotVCSProvider.Name)
	assert.Equal(t, description, gotVCSProvider.Description)
	assert.Equal(t, hostname, gotVCSProvider.Hostname)
	assert.Equal(t, resourcePath, gotVCSProvider.ResourcePath)
	assert.Equal(t, vcsProviderType, gotVCSProvider.Type)
	assert.Equal(t, autoCreateWebhooks, gotVCSProvider.AutoCreateWebhooks)

	// Update the VCS provider.
	toUpdate := &types.UpdateVCSProviderInput{
		ID:                createdVCSProvider.Metadata.ID,
		Description:       &updatedDescription,
		OAuthClientID:     &updatedClientID,
		OAuthClientSecret: &updatedClientSecret,
	}
	updatedVCSProvider, err := client.VCSProvider.UpdateProvider(ctx, toUpdate)
	require.Nil(t, err)
	assert.NotNil(t, updatedVCSProvider)
	assert.Equal(t, name, updatedVCSProvider.Name)
	assert.Equal(t, updatedDescription, updatedVCSProvider.Description)
	assert.Equal(t, hostname, updatedVCSProvider.Hostname)
	// We don't get back the client ID and client secret, so we cannot check them.
	assert.Equal(t, resourcePath, updatedVCSProvider.ResourcePath)
	assert.Equal(t, vcsProviderType, updatedVCSProvider.Type)
	assert.Equal(t, autoCreateWebhooks, updatedVCSProvider.AutoCreateWebhooks)

	// Retrieve and verify the update that should have persisted.
	gotUpdated, err := client.VCSProvider.GetProvider(ctx, &types.GetVCSProviderInput{
		ID: createdVCSProvider.Metadata.ID,
	})
	require.Nil(t, err)
	assert.NotNil(t, gotUpdated)
	// The metadata wouldn't be equal after the update.
	assert.Equal(t, createdVCSProvider.CreatedBy, gotUpdated.CreatedBy)
	assert.Equal(t, name, gotUpdated.Name)
	assert.Equal(t, updatedDescription, gotUpdated.Description)
	assert.Equal(t, hostname, gotUpdated.Hostname)
	// We don't get back the client ID and client secret, so we cannot check them.
	assert.Equal(t, resourcePath, gotUpdated.ResourcePath)
	assert.Equal(t, vcsProviderType, gotUpdated.Type)
	assert.Equal(t, autoCreateWebhooks, gotUpdated.AutoCreateWebhooks)

	// Delete the VCS provider.
	deletedVCSProvider, err := client.VCSProvider.DeleteProvider(ctx, &types.DeleteVCSProviderInput{
		ID: gotVCSProvider.Metadata.ID,
	})
	require.Nil(t, err)
	assert.NotNil(t, deletedVCSProvider)
	assert.Equal(t, updatedVCSProvider.Metadata, deletedVCSProvider.Metadata)
	assert.Equal(t, createdVCSProvider.CreatedBy, deletedVCSProvider.CreatedBy)
	assert.Equal(t, name, deletedVCSProvider.Name)
	assert.Equal(t, updatedDescription, deletedVCSProvider.Description)
	assert.Equal(t, hostname, deletedVCSProvider.Hostname)
	assert.Equal(t, resourcePath, deletedVCSProvider.ResourcePath)
	assert.Equal(t, vcsProviderType, deletedVCSProvider.Type)
	assert.Equal(t, autoCreateWebhooks, deletedVCSProvider.AutoCreateWebhooks)
}

// The End.
