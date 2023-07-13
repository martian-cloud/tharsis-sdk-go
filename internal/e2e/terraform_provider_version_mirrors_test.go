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

func TestProviderVersionMirrorCRUD(t *testing.T) {
	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	toCreate := &types.CreateTerraformProviderVersionMirrorInput{
		RegistryHostname:  "registry.terraform.io",
		RegistryNamespace: "hashicorp",
		Type:              "aws",
		SemanticVersion:   "5.7.0",
		GroupPath:         topGroupName,
	}

	createdMirror, err := client.TerraformProviderVersionMirror.CreateProviderVersionMirror(ctx, toCreate)
	require.Nil(t, err)
	require.NotNil(t, createdMirror)
	assert.Equal(t, toCreate.RegistryHostname, createdMirror.RegistryHostname)
	assert.Equal(t, toCreate.RegistryNamespace, createdMirror.RegistryNamespace)
	assert.Equal(t, toCreate.SemanticVersion, createdMirror.SemanticVersion)
	assert.Equal(t, toCreate.Type, createdMirror.Type)

	toGet := &types.GetTerraformProviderVersionMirrorInput{
		ID: createdMirror.Metadata.ID,
	}

	// First, see if id retrieval works.
	gotMirror, err := client.TerraformProviderVersionMirror.GetProviderVersionMirror(ctx, toGet)
	require.Nil(t, err)
	require.NotNil(t, gotMirror)
	assert.Equal(t, toCreate.RegistryHostname, gotMirror.RegistryHostname)
	assert.Equal(t, toCreate.RegistryNamespace, gotMirror.RegistryNamespace)
	assert.Equal(t, toCreate.SemanticVersion, gotMirror.SemanticVersion)
	assert.Equal(t, toCreate.Type, gotMirror.Type)

	toGetByAddress := &types.GetTerraformProviderVersionMirrorByAddressInput{
		RegistryHostname:  toCreate.RegistryHostname,
		RegistryNamespace: toCreate.RegistryNamespace,
		Type:              toCreate.Type,
		Version:           toCreate.SemanticVersion,
		GroupPath:         toCreate.GroupPath,
	}

	// Now check if we get the same result using the address.
	gotMirror, err = client.TerraformProviderVersionMirror.GetProviderVersionMirrorByAddress(ctx, toGetByAddress)
	require.Nil(t, err)
	require.NotNil(t, gotMirror)
	assert.Equal(t, toCreate.RegistryHostname, gotMirror.RegistryHostname)
	assert.Equal(t, toCreate.RegistryNamespace, gotMirror.RegistryNamespace)
	assert.Equal(t, toCreate.SemanticVersion, gotMirror.SemanticVersion)
	assert.Equal(t, toCreate.Type, gotMirror.Type)

	// Teardown.
	err = client.TerraformProviderVersionMirror.DeleteProviderVersionMirror(ctx, &types.DeleteTerraformProviderVersionMirrorInput{
		ID: gotMirror.Metadata.ID,
	})
	require.Nil(t, err)
}
