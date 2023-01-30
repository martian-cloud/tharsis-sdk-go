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

// This module tests getting a list of module attestations with_OUT_ use of a paginator.
// Three module attestations are created directly under the top-level group.
// Prefix 'gma' stands for 'get module attestations' to avoid collision with other tests.

type gmaModuleAttestationInfo struct {
	modulePath      string
	description     string
	attestationData string
}

type gmaModuleInfo struct {
	name      string
	system    string
	groupPath string
}

const (
	gmaAttestationCount = 1
	attestationData     = "eyJwYXlsb2FkVHlwZSI6ImFwcGxpY2F0aW9uL3ZuZC5pbi10b3RvK2pzb24iLCJwYXlsb2FkIjoiZXlKZmRIbHdaU0k2SW1oMGRIQnpPaTh2YVc0dGRHOTBieTVwYnk5VGRHRjBaVzFsYm5RdmRqQXVNU0lzSW5CeVpXUnBZMkYwWlZSNWNHVWlPaUpqYjNOcFoyNHVjMmxuYzNSdmNtVXVaR1YyTDJGMGRHVnpkR0YwYVc5dUwzWXhJaXdpYzNWaWFtVmpkQ0k2VzNzaWJtRnRaU0k2SW1Kc2IySWlMQ0prYVdkbGMzUWlPbnNpYzJoaE1qVTJJam9pTjJGbE5EY3haV1F4T0RNNU5UTXpPVFUzTW1ZMU1qWTFZamd6TlRnMk1HVXlPR0V5WmpnMU1ERTJORFUxTWpFMFkySXlNVFJpWVdabE5EUXlNbU0zWkNKOWZWMHNJbkJ5WldScFkyRjBaU0k2ZXlKRVlYUmhJam9pZTF3aWRtVnlhV1pwWldSY0lqcDBjblZsZlZ4dUlpd2lWR2x0WlhOMFlXMXdJam9pTWpBeU1pMHhNaTB4TWxReE5EbzFOam8wTVZvaWZYMD0iLCJzaWduYXR1cmVzIjpbeyJrZXlpZCI6IiIsInNpZyI6Ik1FVUNJUURIZGk2UkI2YktESVlPZ3duZkwvaVU5UlQ2a2xyaGRUaEt1NHkzK29JZGNBSWdaVmRQeUczaGhsQTJNZnJxYTkvVUsrOFF4c2d4T2pYcGxGd2JxWW1nQnkwPSJ9XX0="
	moduleDigest        = "7ae471ed18395339572f5265b835860e28a2f85016455214cb214bafe4422c7d"
)

func TestGetModuleAttestations(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	require.Nil(t, err)
	require.NotNil(t, client)

	// Build the attestations and module names, descriptions, and paths.
	attestationInfo, moduleInfo := buildInfoForGetModuleAttestations(gmaAttestationCount)

	// Create the attestations and module.
	attestationIDs, moduleID, err := setupForGetModuleAttestations(ctx, client, attestationInfo, moduleInfo)
	require.Nil(t, err)
	assert.NotEmpty(t, moduleID)
	assert.Equal(t, gmaAttestationCount, len(attestationIDs))

	defer teardownFromGetModuleAttestations(ctx, client, t, moduleID)

	// Get the attestations.
	toSort := types.TerraformModuleAttestationSortableFieldCreatedAtAsc
	pageLimit := pageLimit20
	wantDigest := moduleDigest
	foundAttestationsOutput, err := client.TerraformModuleAttestation.GetModuleAttestations(ctx,
		&types.GetTerraformModuleAttestationsInput{
			Sort: &toSort,
			PaginationOptions: &types.PaginationOptions{
				Limit: &pageLimit,
			},
			Filter: &types.TerraformModuleAttestationFilter{
				Digest:            &wantDigest,
				TerraformModuleID: &moduleID,
			},
		})
	require.Nil(t, err)
	require.NotNil(t, foundAttestationsOutput)
	foundAttestations := foundAttestationsOutput.ModuleAttestations

	// Check the IDs.
	foundIDs := []string{}
	for _, foundAttestation := range foundAttestations {
		foundIDs = append(foundIDs, foundAttestation.Metadata.ID)
	}

	assert.Equal(t, attestationIDs, foundIDs)
}

func buildInfoForGetModuleAttestations(count int) ([]gmaModuleAttestationInfo, gmaModuleInfo) {
	result := []gmaModuleAttestationInfo{}

	// Create info for the module first.
	moduleInfo := gmaModuleInfo{
		name:      "gma-module-1",
		system:    "aws",
		groupPath: topGroupName,
	}

	for ix := 1; ix <= count; ix++ {
		info := gmaModuleAttestationInfo{
			modulePath:      fmt.Sprintf("%s/%s/%s", topGroupName, moduleInfo.name, moduleInfo.system),
			description:     fmt.Sprintf("This is gma test module attestation %d.", ix),
			attestationData: attestationData,
		}
		result = append(result, info)
	}

	return result, moduleInfo
}

// setupForGetModuleAttestations returns the IDs of all the module attestations it creates.
func setupForGetModuleAttestations(ctx context.Context, client *tharsis.Client,
	attestationInfo []gmaModuleAttestationInfo, moduleInfo gmaModuleInfo) ([]string, string, error) {
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

	for _, info := range attestationInfo {
		attestation, err := gmaCreateOneAttestation(ctx, client, info)
		if err != nil {
			return nil, "", err
		}
		result = append(result, attestation.Metadata.ID)
	}

	return result, module.Metadata.ID, nil
}

func gmaCreateOneAttestation(ctx context.Context, client *tharsis.Client,
	info gmaModuleAttestationInfo) (*types.TerraformModuleAttestation, error) {
	return client.TerraformModuleAttestation.CreateModuleAttestation(ctx, &types.CreateTerraformModuleAttestationInput{
		ModulePath:      info.modulePath,
		Description:     info.description,
		AttestationData: info.attestationData,
	})
}

func teardownFromGetModuleAttestations(ctx context.Context, client *tharsis.Client, t *testing.T, moduleID string) {
	err := client.TerraformModule.DeleteModule(ctx, &types.DeleteTerraformModuleInput{
		ID: moduleID,
	})
	assert.Nil(t, err)
}
