//go:build integration
// +build integration

package main

import (
	"context"
	"testing"

	"github.com/likexian/gokit/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestTerraformModules(t *testing.T) {
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

	assert.Equal(t, "tharsis-sdk-e2e-test", module.Name)
	assert.Equal(t, "tharsis", module.System)
	assert.Equal(t, true, module.Private)

	// Get the module.
	module2, err := client.TerraformModule.GetModule(ctx, &types.GetTerraformModuleInput{
		ID: module.Metadata.ID,
	})
	require.Nil(t, err)

	assert.Equal(t, "tharsis-sdk-e2e-test", module2.Name)
	assert.Equal(t, "tharsis", module2.System)
	assert.Equal(t, true, module2.Private)

	// Update the module
	module3, err := client.TerraformModule.UpdateModule(ctx, &types.UpdateTerraformModuleInput{
		ID:   module.Metadata.ID,
		Name: "tharsis-sdk-e2e-test-updated",
	})
	require.Nil(t, err)

	assert.Equal(t, "tharsis-sdk-e2e-test-updated", module3.Name)

	// Delete the module
	err = client.TerraformModule.DeleteModule(ctx, &types.DeleteTerraformModuleInput{
		ID: module.Metadata.ID,
	})
	require.Nil(t, err)
}
