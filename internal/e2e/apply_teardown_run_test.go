//go:build integration
// +build integration

package main

import (
	"context"
	"testing"

	"github.com/likexian/gokit/assert"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// This module applies and tears down a run.

// TODO: Code this module.

const (

// Put constants here for the resources that will be created and deleted.

)

// TestApplyTeardownRun applies and tears down a run.
func TestApplyTeardownRun(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	// Workspace name and path.
	workspaceName := "apply-teardown-run-workspace"
	workspacePath := topGroupName + "/" + workspaceName

	// Create the resources.
	workspaceID, err := setupForApplyTeardownRun(ctx, client, t, workspaceName, workspacePath)
	assert.Nil(t, err)
	assert.NotEqual(t, workspaceID, "")

	// Tear down the resources when the test has finished.
	defer teardownFromApplyTeardownRun(ctx, client, t, workspacePath)

	// Do the main work of this test.

	// Check anything that has not been checked above.

}

// So far, this returns only the workspace ID.
func setupForApplyTeardownRun(ctx context.Context, client *tharsis.Client,
	t *testing.T, wsName, wsPath string) (string, error) {

	// Create the workspace that will be used for the run.
	createdWorkspace, err := client.Workspaces.CreateWorkspace(ctx, &types.CreateWorkspaceInput{
		Name:        wsName,
		GroupPath:   topGroupName,
		Description: "This is the workspace that will be used for applying and tearing down a run.",
	})
	assert.Nil(t, err)
	assert.NotNil(t, createdWorkspace)

	// Do more resource creation here.

	return createdWorkspace.Metadata.ID, nil
}

// So far, this function receives only the workspace ID.
func teardownFromApplyTeardownRun(ctx context.Context, client *tharsis.Client,
	t *testing.T, wsPath string) error {

	// Do more resource deletion here.

	// Delete the new workspace.
	err := client.Workspaces.DeleteWorkspace(ctx, &types.DeleteWorkspaceInput{
		WorkspacePath: wsPath,
	})
	assert.Nil(t, err)

	return nil
}

// The End.
