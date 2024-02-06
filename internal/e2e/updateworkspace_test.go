//go:build integration
// +build integration

package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// This module tests creating, getting, updating, and deleting a workspace.

func TestUpdateWorkspace(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	// Name and path for the workspace.
	updateWorkspaceName := "update-workspace-01"
	updateWorkspacePath := topGroupName + "/" + updateWorkspaceName

	// Create the new workspace.
	newDescription := "This is a test workspace not yet updated, " + updateWorkspaceName
	createdWorkspace, err := client.Workspaces.CreateWorkspace(ctx, &types.CreateWorkspaceInput{
		Name:        updateWorkspaceName,
		GroupPath:   topGroupName,
		Description: newDescription,
	})
	assert.Nil(t, err)
	assert.NotNil(t, createdWorkspace)
	assert.Equal(t, updateWorkspaceName, createdWorkspace.Name)
	assert.Equal(t, newDescription, createdWorkspace.Description)
	assert.Equal(t, updateWorkspacePath, createdWorkspace.FullPath)

	// Get the newly-created workspace.
	toUpdateWorkspace, err := client.Workspaces.GetWorkspace(ctx,
		&types.GetWorkspaceInput{Path: &updateWorkspacePath})
	assert.Nil(t, err)
	assert.NotNil(t, toUpdateWorkspace)

	// Update the workspace's description and PreventDestroyPlan.
	newDescription = "This is a test workspace updated at " + time.Now().String()
	newPreventDestroyPlan := false
	updatedWorkspace, err := client.Workspaces.UpdateWorkspace(ctx,
		&types.UpdateWorkspaceInput{
			WorkspacePath:      &toUpdateWorkspace.FullPath,
			Description:        newDescription,
			PreventDestroyPlan: &newPreventDestroyPlan,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, newDescription, updatedWorkspace.Description)
	assert.Equal(t, newPreventDestroyPlan, updatedWorkspace.PreventDestroyPlan)

	// Delete the new workspace.
	err = client.Workspaces.DeleteWorkspace(ctx, &types.DeleteWorkspaceInput{
		WorkspacePath: &updatedWorkspace.FullPath,
	})
	assert.Nil(t, err)

}
