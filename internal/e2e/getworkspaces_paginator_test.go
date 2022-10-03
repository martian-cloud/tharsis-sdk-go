//go:build integration
// +build integration

package main

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/likexian/gokit/assert"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// This module tests getting a list of workspaces via a paginator.
// Three workspaces are created directly under the top-level group.
// Prefix 'gwp' stands for 'get workspaces paginator' to avoid collision with other tests.

type gwpWorkspaceInfo struct {
	name        string
	description string
	path        string
}

const (
	gwpWorkspaceCount = 3
)

// TestGetWorkspacesPaginator gets a list of workspaces.
func TestGetWorkspacesPaginator(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	// Build the workspace names, descriptions, and paths.
	workspacesInfo := buildInfoForGetWorkspacesPaginator(gwpWorkspaceCount)

	// Create the workspaces.
	workspacePaths, err := setupForGetWorkspacesPaginator(ctx, client, workspacesInfo)
	assert.Nil(t, err)
	assert.Equal(t, len(workspacePaths), gwpWorkspaceCount)

	// Tear down the workspaces when the test has finished.
	defer teardownFromGetWorkspacesPaginator(ctx, client, t, workspacePaths)

	// Get the paginator.
	toSort := types.WorkspaceSortableFieldFullPathAsc
	pageLimit := pageLimit2
	wantGroup := topGroupName
	workspacesPaginator, err := client.Workspaces.GetWorkspacePaginator(ctx,
		&types.GetWorkspacesInput{
			Sort: &toSort,
			PaginationOptions: &types.PaginationOptions{
				Limit: &pageLimit,
			},
			Filter: &types.WorkspaceFilter{GroupPath: &wantGroup},
		})
	assert.Nil(t, err)
	assert.NotNil(t, workspacesPaginator)

	// Scan the pages.
	expectLengths := []int{2, 1, 99999} // should never see the 99999
	foundPaths := []string{}
	for workspacesPaginator.HasMore() {
		getWorkspacesOutput, err := workspacesPaginator.Next(ctx)
		assert.Nil(t, err)

		// Make sure we're getting the correct number of groups on each page.
		var expectLength int
		expectLength, expectLengths = expectLengths[0], expectLengths[1:]

		assert.Equal(t, len(getWorkspacesOutput.Workspaces), expectLength)

		// Prepare to make sure we eventually get all the groups.
		for _, workspace := range getWorkspacesOutput.Workspaces {
			assert.NotNil(t, workspace)
			foundPaths = append(foundPaths, workspace.FullPath)
		}
	}

	// Expect gwp workspace IDs.
	expectPaths := []string{}
	for _, info := range workspacesInfo {
		expectPaths = append(expectPaths, info.path)
	}

	assert.True(t, reflect.DeepEqual(expectPaths, foundPaths))
}

func buildInfoForGetWorkspacesPaginator(count int) []gwpWorkspaceInfo {
	result := []gwpWorkspaceInfo{}

	for ix := 1; ix <= count; ix++ {
		name := fmt.Sprintf("gwp-workspace-%d", ix)
		info := gwpWorkspaceInfo{
			name:        name,
			description: fmt.Sprintf("This is gwp test workspace %d.", ix),
			path:        fmt.Sprintf("%s/%s", topGroupName, name),
		}
		result = append(result, info)
	}

	return result
}

// setupForGetWorkspacesPaginator returns the IDs of all the workspaces it creates.
func setupForGetWorkspacesPaginator(ctx context.Context, client *tharsis.Client,
	workspacesInfo []gwpWorkspaceInfo) ([]string, error) {
	result := []string{}

	for _, info := range workspacesInfo {
		workspace, err := gwpCreateOneWorkspace(ctx, client, topGroupName, info)
		if err != nil {
			return nil, err
		}
		result = append(result, workspace.FullPath)
	}

	return result, nil
}

func gwpCreateOneWorkspace(ctx context.Context, client *tharsis.Client,
	groupPath string, info gwpWorkspaceInfo) (*types.Workspace, error) {
	return client.Workspaces.CreateWorkspace(ctx, &types.CreateWorkspaceInput{
		Name:        info.name,
		Description: info.description,
		GroupPath:   groupPath,
	})
}

func teardownFromGetWorkspacesPaginator(ctx context.Context, client *tharsis.Client, t *testing.T, paths []string) {
	for _, path := range paths {
		err := client.Workspaces.DeleteWorkspace(ctx, &types.DeleteWorkspaceInput{
			WorkspacePath: path,
		})
		assert.Nil(t, err)
	}
}

// The End.
