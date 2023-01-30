//go:build integration
// +build integration

package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// This module tests getting a list of module versions with_OUT_ use of a paginator.
// Three module versions are created directly under the top-level group.
// Prefix 'gmv' stands for 'get module versions' to avoid collision with other tests.

type gmvModuleVersionInfo struct {
	modulePath string
	version    string
	shasum     string
}

const (
	gmaVersionCount = 3
	randomShasum    = "ae82ae9511c9a00f5c4acb444b79081fb18d57814f64cbdfbbe01c3bdcd303f"
)

func TestGetModuleVersions(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	require.Nil(t, err)
	require.NotNil(t, client)

	// Build the versions and module names, descriptions, and paths.
	versionInfo, moduleInfo := buildInfoForGetModuleVersions(gmaVersionCount)

	// Create the versions and module.
	versionIDs, moduleID, err := setupForGetModuleVersions(ctx, client, versionInfo, moduleInfo)
	require.Nil(t, err)
	assert.NotEmpty(t, moduleID)
	assert.Equal(t, gmaVersionCount, len(versionIDs))

	defer teardownFromGetModuleVersions(ctx, client, t, moduleID)

	// Get the versions.
	toSort := types.TerraformModuleVersionSortableFieldCreatedAtAsc
	pageLimit := pageLimit20
	foundVersionsOutput, err := client.TerraformModuleVersion.GetModuleVersions(ctx,
		&types.GetTerraformModuleVersionsInput{
			Sort: &toSort,
			PaginationOptions: &types.PaginationOptions{
				Limit: &pageLimit,
			},
			TerraformModuleID: moduleID,
		})
	require.Nil(t, err)
	assert.NotNil(t, foundVersionsOutput)
	foundVersions := foundVersionsOutput.ModuleVersions

	// Check the IDs.
	foundIDs := []string{}
	for _, foundVersion := range foundVersions {
		foundIDs = append(foundIDs, foundVersion.Metadata.ID)
	}

	assert.Equal(t, versionIDs, foundIDs)
}

func buildInfoForGetModuleVersions(count int) ([]gmvModuleVersionInfo, gmaModuleInfo) {
	result := []gmvModuleVersionInfo{}

	// Create info for the module first.
	moduleInfo := gmaModuleInfo{
		name:      "gmv-module-1",
		system:    "aws",
		groupPath: topGroupName,
	}

	for ix := 1; ix <= count; ix++ {
		info := gmvModuleVersionInfo{
			modulePath: fmt.Sprintf("%s/%s/%s", topGroupName, moduleInfo.name, moduleInfo.system),
			shasum:     randomShasum + fmt.Sprint(ix), // Make 'different' digests.
			version:    fmt.Sprintf("0.%d.0", ix),
		}
		result = append(result, info)
	}

	return result, moduleInfo
}

// setupForGetModuleVersions returns the IDs of all the module versions it creates.
func setupForGetModuleVersions(ctx context.Context, client *tharsis.Client,
	versionInfo []gmvModuleVersionInfo, moduleInfo gmaModuleInfo) ([]string, string, error) {
	result := []string{}

	// Create the module first.
	module, err := client.TerraformModule.CreateModule(ctx, &types.CreateTerraformModuleInput{
		Name:      moduleInfo.name,
		System:    moduleInfo.system,
		GroupPath: moduleInfo.groupPath,
	})
	if err != nil {
		return nil, "", err
	}

	for _, info := range versionInfo {
		version, err := gmaCreateOneModuleVersion(ctx, client, info)
		if err != nil {
			return nil, "", err
		}
		result = append(result, version.Metadata.ID)
	}

	return result, module.Metadata.ID, nil
}

func gmaCreateOneModuleVersion(ctx context.Context, client *tharsis.Client,
	info gmvModuleVersionInfo) (*types.TerraformModuleVersion, error) {
	return client.TerraformModuleVersion.CreateModuleVersion(ctx, &types.CreateTerraformModuleVersionInput{
		ModulePath: info.modulePath,
		SHASum:     info.shasum,
		Version:    info.version,
	})
}

func teardownFromGetModuleVersions(ctx context.Context, client *tharsis.Client, t *testing.T, moduleID string) {
	err := client.TerraformModule.DeleteModule(ctx, &types.DeleteTerraformModuleInput{
		ID: moduleID,
	})
	assert.Nil(t, err)
}
