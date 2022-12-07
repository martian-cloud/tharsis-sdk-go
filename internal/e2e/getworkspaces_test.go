//go:build integration
// +build integration

package main

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// This module tests getting a list of workspaces with_OUT_ use of a paginator.
// Three workspaces are created directly under the top-level group.
// Prefix 'gw' stands for 'get workspaces' to avoid collision with other tests.

type gwWorkspaceInfo struct {
	name        string
	description string
	path        string
}

const (
	gwWorkspaceCount = 3
)

// TestGetWorkspaces gets multiple workspaces.
func TestGetWorkspaces(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	// Build the workspace names, descriptions, and paths.
	workspacesInfo := buildInfoForGetWorkspaces(gwWorkspaceCount)

	// Create the workspaces.
	workspacePaths, err := setupForGetWorkspaces(ctx, client, workspacesInfo)
	assert.Nil(t, err)
	assert.Equal(t, gwWorkspaceCount, len(workspacePaths))

	// Tear down the workspaces when the test has finished.
	defer teardownFromGetWorkspaces(ctx, client, t, workspacePaths)

	// Get the workspaces.
	toSort := types.WorkspaceSortableFieldFullPathAsc
	pageLimit := pageLimit20
	wantGroup := topGroupName
	foundWorkspacesOutput, err := client.Workspaces.GetWorkspaces(ctx,
		&types.GetWorkspacesInput{
			Sort: &toSort,
			PaginationOptions: &types.PaginationOptions{
				Limit: &pageLimit,
			},
			Filter: &types.WorkspaceFilter{GroupPath: &wantGroup},
		})
	assert.Nil(t, err)
	assert.NotNil(t, foundWorkspacesOutput)
	foundWorkspaces := foundWorkspacesOutput.Workspaces

	// Check the paths.
	foundPaths := []string{}
	for _, foundWorkspace := range foundWorkspaces {
		foundPaths = append(foundPaths, foundWorkspace.FullPath)
	}

	// Expect gw workspace IDs.
	expectPaths := []string{}
	for _, info := range workspacesInfo {
		expectPaths = append(expectPaths, info.path)
	}

	assert.True(t, reflect.DeepEqual(expectPaths, foundPaths))
}

func buildInfoForGetWorkspaces(count int) []gwWorkspaceInfo {
	result := []gwWorkspaceInfo{}

	for ix := 1; ix <= count; ix++ {
		name := fmt.Sprintf("gw-workspace-%d", ix)
		info := gwWorkspaceInfo{
			name:        name,
			description: fmt.Sprintf("This is gw test workspace %d.", ix),
			path:        fmt.Sprintf("%s/%s", topGroupName, name),
		}
		result = append(result, info)
	}

	return result
}

// setupForGetWorkspaces returns the IDs of all the workspaces it creates.
func setupForGetWorkspaces(ctx context.Context, client *tharsis.Client,
	workspacesInfo []gwWorkspaceInfo) ([]string, error) {
	result := []string{}

	for _, info := range workspacesInfo {
		workspace, err := gwCreateOneWorkspace(ctx, client, topGroupName, info)
		if err != nil {
			return nil, err
		}
		result = append(result, workspace.FullPath)
	}

	return result, nil
}

func gwCreateOneWorkspace(ctx context.Context, client *tharsis.Client,
	groupPath string, info gwWorkspaceInfo) (*types.Workspace, error) {
	return client.Workspaces.CreateWorkspace(ctx, &types.CreateWorkspaceInput{
		Name:        info.name,
		Description: info.description,
		GroupPath:   groupPath,
	})
}

func teardownFromGetWorkspaces(ctx context.Context, client *tharsis.Client, t *testing.T, paths []string) {
	for _, path := range paths {
		err := client.Workspaces.DeleteWorkspace(ctx, &types.DeleteWorkspaceInput{
			WorkspacePath: &path,
		})
		assert.Nil(t, err)
	}
}

// The End.
