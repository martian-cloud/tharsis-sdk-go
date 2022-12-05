// example/workspace/main.go contains examples for using workspace-related functions.

// Package main contains workspace examples
package main

import (
	"context"

	"github.com/aws/smithy-go/ptr"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/auth"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/config"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// ExampleGetWorkspace demonstrates GetWorkspace:
func ExampleGetWorkspace() error {

	// For authentication, this example uses an explicitly specified static token provider.
	staticToken := "==insert-static-token-here=="
	staticTokenProvider, err := auth.NewStaticTokenProvider(staticToken)
	if err != nil {
		return err
	}

	cfg, err := config.Load(config.WithTokenProvider(staticTokenProvider))
	if err != nil {
		return err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	workspace, err := client.Workspaces.GetWorkspace(ctx,
		&types.GetWorkspaceInput{Path: ptr.String("wild-space/rmr-dummy-2")})
	if err != nil {
		return err
	}
	_ = workspace

	return nil
}

// ExampleGetWorkspaces demonstrates GetWorkspaces:
func ExampleGetWorkspaces() error {

	// For authentication, this example uses an explicitly specified service account token provider.
	endpointURL := "https://insert.tharsis.endpoint.here"
	serviceAccountPath := "==insert-service-account-path-here=="
	serviceAccountToken := "==insert-service-account-token-here=="
	serviceAccountTokenProvider, err := auth.NewServiceAccountTokenProvider(
		endpointURL, serviceAccountPath, serviceAccountToken)
	if err != nil {
		return err
	}

	cfg, err := config.Load(config.WithTokenProvider(serviceAccountTokenProvider))
	if err != nil {
		return err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	var hundred int32 = 100
	sortBy := types.WorkspaceSortableFieldFullPathAsc
	getWorkspacesOutput, err := client.Workspaces.GetWorkspaces(ctx,
		&types.GetWorkspacesInput{
			Sort: &sortBy,
			PaginationOptions: &types.PaginationOptions{
				Limit: &hundred,
			},
		})
	if err != nil {
		return err
	}
	_ = getWorkspacesOutput

	return nil
}

// ExampleGetWorkspacesFiltered demonstrates GetWorkspaces with a groupPath filter:
func ExampleGetWorkspacesFiltered() error {

	// For authentication, this example relies on environment variables being
	// automatically imported into the configuration object.
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	var hundred int32 = 100
	sortBy := types.WorkspaceSortableFieldFullPathAsc
	groupPathFilter := "wild-space"
	getWorkspacesOutput, err := client.Workspaces.GetWorkspaces(ctx,
		&types.GetWorkspacesInput{
			Sort: &sortBy,
			PaginationOptions: &types.PaginationOptions{
				Limit: &hundred,
			},
			Filter: &types.WorkspaceFilter{
				GroupPath: &groupPathFilter,
			},
		})
	if err != nil {
		return err
	}
	_ = getWorkspacesOutput

	return nil
}

// ExampleGetWorkspacesPaged demonstrates GetWorkspacesPaged, 3 workspaces per page:
func ExampleGetWorkspacesPaged() error {

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	var three int32 = 3
	sortBy := types.WorkspaceSortableFieldFullPathDesc
	paginator, err := client.Workspaces.GetWorkspacePaginator(ctx,
		&types.GetWorkspacesInput{
			Sort: &sortBy,
			PaginationOptions: &types.PaginationOptions{
				Limit: &three,
			},
		})
	if err != nil {
		return err
	}

	for paginator.HasMore() {
		getWorkspacesOutput, err := paginator.Next(ctx)
		if err != nil {
			return err
		}

		_ = getWorkspacesOutput.Workspaces
	}

	return nil
}

// ExampleCreateWorkspace demonstrates CreateWorkspace:
func ExampleCreateWorkspace() (*types.Workspace, error) {

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// First, must find the group ID.
	ctx := context.Background()

	created, err := client.Workspaces.CreateWorkspace(ctx,
		&types.CreateWorkspaceInput{
			Name:        "new-workspace-01",
			GroupPath:   "wild-space",
			Description: "This is a newly created workspace.",
		},
	)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// ExampleUpdateWorkspace demonstrates UpdateWorkspace:
func ExampleUpdateWorkspace(workspacePath string) error {

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	updated, err := client.Workspaces.UpdateWorkspace(ctx,
		&types.UpdateWorkspaceInput{
			WorkspacePath: &workspacePath,
			Description:   "This is the updated workspace.",
		},
	)
	if err != nil {
		return err
	}
	_ = updated

	return nil
}

// ExampleDeleteWorkspace demonstrates DeleteWorkspace:
func ExampleDeleteWorkspace(workspacePath string) error {

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = client.Workspaces.DeleteWorkspace(ctx,
		&types.DeleteWorkspaceInput{
			WorkspacePath: &workspacePath,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// The End.
