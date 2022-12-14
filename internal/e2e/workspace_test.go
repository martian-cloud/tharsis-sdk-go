//go:build integration
// +build integration

package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TODO: This module has e2e tests only for newer method(s) added in December, 2022.
// The other methods should also have e2e tests added, including a TestGetWorkspaceByPath.

func TestGetWorkspaceByID(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	getWorkspaceName := "get-workspace-01"
	getWorkspaceFullPath := topGroupName + "/" + getWorkspaceName

	// Create the workspace.
	newDescription := "This is a test workspace, " + getWorkspaceName
	toCreate := &types.CreateWorkspaceInput{
		Name:        getWorkspaceName,
		GroupPath:   topGroupName,
		Description: newDescription,
	}
	createdWorkspace, err := client.Workspaces.CreateWorkspace(ctx, toCreate)
	assert.Nil(t, err)
	assert.NotNil(t, createdWorkspace)
	assert.Equal(t, getWorkspaceName, createdWorkspace.Name)
	assert.Equal(t, getWorkspaceFullPath, createdWorkspace.FullPath)
	assert.Equal(t, newDescription, createdWorkspace.Description)

	// Get the workspace.
	gotWorkspace, err := client.Workspaces.GetWorkspace(ctx, &types.GetWorkspaceInput{
		ID: &createdWorkspace.Metadata.ID,
	})
	assert.Nil(t, err)

	// Verify the returned contents are what they should be.
	assert.Equal(t, toCreate.Name, gotWorkspace.Name)
	assert.Equal(t, getWorkspaceFullPath, gotWorkspace.FullPath)
	assert.Equal(t, toCreate.Description, gotWorkspace.Description)

	// Delete the workspace.
	err = client.Workspaces.DeleteWorkspace(ctx, &types.DeleteWorkspaceInput{
		WorkspacePath: &gotWorkspace.FullPath,
	})
	assert.Nil(t, err)
}

// The End.
