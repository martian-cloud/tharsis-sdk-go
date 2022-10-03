// example/group/main.go contains examples for using group-related functions.

// Package main contains group examples
package main

import (
	"context"

	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/auth"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/config"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// ExampleGetGroup demonstrates GetGroup:
func ExampleGetGroup() error {

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
	theGroup, err := client.Group.GetGroup(ctx,
		&types.GetGroupInput{Path: "wild-space"})
	if err != nil {
		return err
	}
	_ = theGroup

	return nil
}

// ExampleGetGroups demonstrates GetGroups:
func ExampleGetGroups() error {

	// For authentication, this example uses an explicitly specified service account token provicer.
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
	sortBy := types.GroupSortableFieldFullPathAsc
	getGroupsOutput, err := client.Group.GetGroups(ctx,
		&types.GetGroupsInput{
			Sort: &sortBy,
			PaginationOptions: &types.PaginationOptions{
				Limit: &hundred,
			},
		})
	if err != nil {
		return err
	}
	_ = getGroupsOutput

	return nil
}

// ExampleGetGroupsFiltered demonstrates GetGroups with a parentPath filter:
func ExampleGetGroupsFiltered() error {

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
	sortBy := types.GroupSortableFieldFullPathAsc
	parentPathFilter := "wild-space"
	getGroupsOutput, err := client.Group.GetGroups(ctx,
		&types.GetGroupsInput{
			Sort: &sortBy,
			PaginationOptions: &types.PaginationOptions{
				Limit: &hundred,
			},
			Filter: &types.GroupFilter{
				ParentPath: &parentPathFilter,
			},
		})
	if err != nil {
		return err
	}
	_ = getGroupsOutput

	return nil
}

// ExampleGetGroupsPaged demonstrates GetGroupsPaged, 2 groups per page:
func ExampleGetGroupsPaged() error {

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	var four int32 = 4
	sortBy := types.GroupSortableFieldFullPathDesc
	paginator, err := client.Group.GetGroupPaginator(ctx,
		&types.GetGroupsInput{
			Sort: &sortBy,
			PaginationOptions: &types.PaginationOptions{
				Limit: &four,
			},
		})
	if err != nil {
		return err
	}

	for paginator.HasMore() {
		getGroupsOutput, err := paginator.Next(ctx)
		if err != nil {
			return err
		}

		_ = getGroupsOutput.Groups
	}

	return nil
}

// ExampleCreateGroup demonstrates CreateGroup:
func ExampleCreateGroup() (*types.Group, error) {

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// First, must find the parent group ID.
	ctx := context.Background()
	parentPath := "wild-space"

	created, err := client.Group.CreateGroup(ctx,
		&types.CreateGroupInput{
			Name:        "new-group-01",
			ParentPath:  &parentPath,
			Description: "This is a newly created group.",
		},
	)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// ExampleUpdateGroup demonstrates UpdateGroup:
func ExampleUpdateGroup(groupPath string) error {

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	updated, err := client.Group.UpdateGroup(ctx,
		&types.UpdateGroupInput{
			GroupPath:   groupPath,
			Description: "This is the updated group.",
		},
	)
	if err != nil {
		return err
	}
	_ = updated

	return nil
}

// ExampleDeleteGroup demonstrates DeleteGroup:
//
// GraphiQL documentation does not yet show a Delete Group capability, which would be the input for deleting a group.
// This is speculative as to what would be done.
func ExampleDeleteGroup(groupPath string) error {

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = client.Group.DeleteGroup(ctx,
		&types.DeleteGroupInput{
			GroupPath: groupPath,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// The End.
