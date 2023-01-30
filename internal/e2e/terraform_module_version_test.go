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

func TestTerraformModuleVersionVersions(t *testing.T) {
	ctx := context.Background()
	client, err := createClient()
	require.Nil(t, err)

	// Create module
	module, err := client.TerraformModule.CreateModule(ctx, &types.CreateTerraformModuleInput{
		GroupPath: topGroupName,
		Name:      "tharsis-sdk-e2e-test",
		System:    "tharsis",
		Private:   true,
	})
	require.Nil(t, err)

	moduleVersion, err := client.TerraformModuleVersion.CreateModuleVersion(ctx, &types.CreateTerraformModuleVersionInput{
		ModulePath: module.ResourcePath,
		Version:    "1.0.0",
		SHASum:     "7ae471ed18395339572f5265b835860e28a2f85016455214cb214bafe4422c7d",
	})
	require.Nil(t, err)

	assert.Equal(t, "1.0.0", moduleVersion.Version)
	assert.Equal(t, "pending", moduleVersion.Status)

	// Get the moduleVersion.
	moduleVersion2, err := client.TerraformModuleVersion.GetModuleVersion(ctx, &types.GetTerraformModuleVersionInput{
		ID: &moduleVersion.Metadata.ID,
	})
	require.Nil(t, err)

	assert.Equal(t, "1.0.0", moduleVersion2.Version)
	assert.Equal(t, "pending", moduleVersion2.Status)

	// Delete the moduleVersion
	err = client.TerraformModuleVersion.DeleteModuleVersion(ctx, &types.DeleteTerraformModuleVersionInput{
		ID: moduleVersion.Metadata.ID,
	})
	require.Nil(t, err)

	// Delete the module.
	err = client.TerraformModule.DeleteModule(ctx, &types.DeleteTerraformModuleInput{
		ID: module.Metadata.ID,
	})
	require.Nil(t, err)
}
