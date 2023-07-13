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

type gppVersionMirrorInfo struct {
	registryHostname  string
	registryNamespace string
	providerType      string
	version           string
}

const (
	gppVersionMirrorsCount = 3
)

func TestGetProviderVersionMirrors(t *testing.T) {
	ctx := context.Background()
	client, err := createClient()
	require.Nil(t, err)
	require.NotNil(t, client)

	mirrorsInfo := buildInfoForGetProviderVersionMirrorsPaginator(gppVersionMirrorsCount)

	// Create the version mirrors.
	mirrorIDs, err := setupForGetProviderVersionMirrorsPaginator(ctx, client, mirrorsInfo)
	require.Nil(t, err)
	assert.Equal(t, gppVersionMirrorsCount, len(mirrorIDs))

	defer teardownFromGetProviderVersionMirrorsPaginator(ctx, client, t, mirrorIDs)

	// Get the paginator.
	toSort := types.TerraformProviderVersionMirrorSortableFieldCreatedAtAsc
	pageLimit := pageLimit2
	mirrorsPaginator, err := client.TerraformProviderVersionMirror.GetProviderVersionMirrorPaginator(ctx,
		&types.GetTerraformProviderVersionMirrorsInput{
			Sort: &toSort,
			PaginationOptions: &types.PaginationOptions{
				Limit: &pageLimit,
			},
			GroupPath: topGroupName,
		})
	require.Nil(t, err)
	require.NotNil(t, mirrorsPaginator)

	// Scan the pages.
	expectLengths := []int{2, 1, 99999} // should never see the 99999
	foundIDs := []string{}
	for mirrorsPaginator.HasMore() {
		getMirrorsOutput, err := mirrorsPaginator.Next(ctx)
		require.Nil(t, err)

		var expectLength int
		expectLength, expectLengths = expectLengths[0], expectLengths[1:]

		assert.Equal(t, expectLength, len(getMirrorsOutput.VersionMirrors))

		// Prepare to make sure we eventually get all the groups.
		for _, mirror := range getMirrorsOutput.VersionMirrors {
			assert.NotNil(t, mirror)
			foundIDs = append(foundIDs, mirror.Metadata.ID)
		}
	}

	assert.Equal(t, mirrorIDs, foundIDs)
}

func buildInfoForGetProviderVersionMirrorsPaginator(count int) []gppVersionMirrorInfo {
	result := []gppVersionMirrorInfo{}

	for ix := 1; ix <= count; ix++ {
		result = append(result, gppVersionMirrorInfo{
			registryHostname:  "registry.terraform.io",
			registryNamespace: "hashicorp",
			providerType:      "aws",
			version:           fmt.Sprintf("5.%d.0", ix), // Should allow up to 7ish valid versions.
		})
	}

	return result
}

// setupForGetProviderVersionMirrorsPaginator returns the IDs of all the version mirrors it creates.
func setupForGetProviderVersionMirrorsPaginator(
	ctx context.Context,
	client *tharsis.Client,
	mirrorsInfo []gppVersionMirrorInfo,
) ([]string, error) {
	result := []string{}

	for _, info := range mirrorsInfo {
		versionMirror, err := gppCreateOneProviderVersionMirror(ctx, client, topGroupName, info)
		if err != nil {
			return nil, err
		}
		result = append(result, versionMirror.Metadata.ID)
	}

	return result, nil
}

func gppCreateOneProviderVersionMirror(
	ctx context.Context,
	client *tharsis.Client,
	groupPath string,
	info gppVersionMirrorInfo,
) (*types.TerraformProviderVersionMirror, error) {
	return client.TerraformProviderVersionMirror.CreateProviderVersionMirror(ctx, &types.CreateTerraformProviderVersionMirrorInput{
		RegistryHostname:  info.registryHostname,
		RegistryNamespace: info.registryNamespace,
		Type:              info.providerType,
		SemanticVersion:   info.version,
		GroupPath:         groupPath,
	})
}

func teardownFromGetProviderVersionMirrorsPaginator(
	ctx context.Context,
	client *tharsis.Client,
	t *testing.T,
	ids []string,
) {
	for _, id := range ids {
		err := client.TerraformProviderVersionMirror.DeleteProviderVersionMirror(ctx, &types.DeleteTerraformProviderVersionMirrorInput{
			ID: id,
		})
		assert.Nil(t, err)
	}
}
