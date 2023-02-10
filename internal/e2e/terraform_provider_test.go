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

func TestTerraformProviderCRUD(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	require.Nil(t, err)
	assert.NotNil(t, client)

	// Values for various fields.
	tfpName := "tf-provider-name-1"
	tfpGroupPath := topGroupName
	tfpResourcePath := tfpGroupPath + "/" + tfpName
	tfpRegistryNamespace := tfpGroupPath
	tfpRepositoryURL := "tf-provider-repository-url"
	tfpPrivate := true

	updatedName := "tf-provider-name-1"
	updatedRepositoryURL := "tf-provider-repository-url"
	updatedPrivate := true

	// Create the Terraform provider.
	toCreate := &types.CreateTerraformProviderInput{
		Name:          tfpName,
		GroupPath:     tfpGroupPath,
		RepositoryURL: tfpRepositoryURL,
		Private:       tfpPrivate,
	}
	createdTerraformProvider, err := client.TerraformProvider.CreateProvider(ctx, toCreate)

	require.Nil(t, err)
	assert.NotNil(t, createdTerraformProvider)
	assert.Equal(t, tfpName, createdTerraformProvider.Name)
	assert.Equal(t, tfpResourcePath, createdTerraformProvider.ResourcePath)
	assert.Equal(t, tfpRegistryNamespace, createdTerraformProvider.RegistryNamespace)
	assert.Equal(t, tfpRepositoryURL, createdTerraformProvider.RepositoryURL)
	assert.Equal(t, tfpPrivate, createdTerraformProvider.Private)

	// Get the Terraform provider to make sure it persisted.
	gotTerraformProvider, err := client.TerraformProvider.GetProvider(ctx, &types.GetTerraformProviderInput{
		ID: createdTerraformProvider.Metadata.ID,
	})
	require.Nil(t, err)
	assert.NotNil(t, gotTerraformProvider)

	// Verify the returned contents are what they should be.
	assert.Equal(t, createdTerraformProvider.Metadata, gotTerraformProvider.Metadata)
	assert.Equal(t, tfpName, gotTerraformProvider.Name)
	assert.Equal(t, tfpResourcePath, gotTerraformProvider.ResourcePath)
	assert.Equal(t, tfpRegistryNamespace, gotTerraformProvider.RegistryNamespace)
	assert.Equal(t, tfpRepositoryURL, gotTerraformProvider.RepositoryURL)
	assert.Equal(t, tfpPrivate, gotTerraformProvider.Private)

	// Update the Terraform provider.
	toUpdate := &types.UpdateTerraformProviderInput{
		ID:            createdTerraformProvider.Metadata.ID,
		Name:          updatedName,
		RepositoryURL: updatedRepositoryURL,
		Private:       updatedPrivate,
	}
	updatedTerraformProvider, err := client.TerraformProvider.UpdateProvider(ctx, toUpdate)
	require.Nil(t, err)
	assert.NotNil(t, updatedTerraformProvider)
	assert.Equal(t, tfpName, updatedTerraformProvider.Name)
	assert.Equal(t, tfpResourcePath, updatedTerraformProvider.ResourcePath)
	assert.Equal(t, tfpRegistryNamespace, updatedTerraformProvider.RegistryNamespace)
	assert.Equal(t, tfpRepositoryURL, updatedTerraformProvider.RepositoryURL)
	assert.Equal(t, tfpPrivate, updatedTerraformProvider.Private)

	// Retrieve and verify the update that should have persisted.
	gotUpdated, err := client.TerraformProvider.GetProvider(ctx, &types.GetTerraformProviderInput{
		ID: createdTerraformProvider.Metadata.ID,
	})
	require.Nil(t, err)
	assert.NotNil(t, gotUpdated)
	// The metadata wouldn't be equal after the update.
	assert.Equal(t, tfpName, gotUpdated.Name)
	assert.Equal(t, tfpResourcePath, gotUpdated.ResourcePath)
	assert.Equal(t, tfpRegistryNamespace, gotUpdated.RegistryNamespace)
	assert.Equal(t, tfpRepositoryURL, gotUpdated.RepositoryURL)
	assert.Equal(t, tfpPrivate, gotUpdated.Private)

	// Delete the Terraform provider.
	deletedTerraformProvider, err := client.TerraformProvider.DeleteProvider(ctx, &types.DeleteTerraformProviderInput{
		ID: gotTerraformProvider.Metadata.ID,
	})
	require.Nil(t, err)
	assert.NotNil(t, deletedTerraformProvider)
	assert.Equal(t, updatedTerraformProvider.Metadata, deletedTerraformProvider.Metadata)
	assert.Equal(t, tfpName, deletedTerraformProvider.Name)
	assert.Equal(t, tfpResourcePath, deletedTerraformProvider.ResourcePath)
	assert.Equal(t, tfpRegistryNamespace, deletedTerraformProvider.RegistryNamespace)
	assert.Equal(t, tfpRepositoryURL, deletedTerraformProvider.RepositoryURL)
	assert.Equal(t, tfpPrivate, deletedTerraformProvider.Private)
}

// The End.
