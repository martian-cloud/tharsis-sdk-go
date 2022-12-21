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

// This module tests getting a list of groups via a paginator.
// Three groups are created as direct children of the top-level group.
// Prefix 'ggp' stands for 'get groups paginator' to avoid collision with other tests.

type ggpGroupInfo struct {
	name        string
	description string
	path        string
}

const (
	ggpGroupCount = 3
)

// TestGetGroupsPaginator gets a list of groups.
func TestGetGroupsPaginator(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	// Build the group names, descriptions, and paths.
	groupsInfo := buildInfoForGetGroupsPaginator(ggpGroupCount)

	// Create the groups.
	groupPaths, err := setupForGetGroupsPaginator(ctx, client, groupsInfo)
	assert.Nil(t, err)
	assert.Equal(t, ggpGroupCount, len(groupPaths))

	// Tear down the groups when the test has finished.
	defer teardownFromGetGroupsPaginator(ctx, client, t, groupPaths)

	// Get the paginator.
	toSort := types.GroupSortableFieldFullPathAsc
	pageLimit := pageLimit2
	wantParent := topGroupName
	groupsPaginator, err := client.Group.GetGroupPaginator(ctx,
		&types.GetGroupsInput{
			Sort: &toSort,
			PaginationOptions: &types.PaginationOptions{
				Limit: &pageLimit,
			},
			Filter: &types.GroupFilter{ParentPath: &wantParent},
		})
	assert.Nil(t, err)
	assert.NotNil(t, groupsPaginator)

	// Scan the pages.
	expectLengths := []int{2, 1, 99999} // should never see the 99999
	foundPaths := []string{}
	for groupsPaginator.HasMore() {
		getGroupsOutput, err := groupsPaginator.Next(ctx)
		assert.Nil(t, err)

		// Make sure we're getting the correct number of groups on each page.
		var expectLength int
		expectLength, expectLengths = expectLengths[0], expectLengths[1:]

		assert.Equal(t, expectLength, len(getGroupsOutput.Groups))

		// Prepare to make sure we eventually get all the groups.
		for _, group := range getGroupsOutput.Groups {
			assert.NotNil(t, group)
			foundPaths = append(foundPaths, group.FullPath)
		}
	}

	// Expect ggp group IDs.
	expectPaths := []string{}
	for _, info := range groupsInfo {
		expectPaths = append(expectPaths, info.path)
	}

	assert.True(t, reflect.DeepEqual(expectPaths, foundPaths))
}

func buildInfoForGetGroupsPaginator(count int) []ggpGroupInfo {
	result := []ggpGroupInfo{}

	for ix := 1; ix <= count; ix++ {
		name := fmt.Sprintf("ggp-group-%d", ix)
		info := ggpGroupInfo{
			name:        name,
			description: fmt.Sprintf("This is ggp test group %d.", ix),
			path:        fmt.Sprintf("%s/%s", topGroupName, name),
		}
		result = append(result, info)
	}

	return result
}

func setupForGetGroupsPaginator(ctx context.Context, client *tharsis.Client,
	groupsInfo []ggpGroupInfo) ([]string, error) {
	result := []string{}

	for _, info := range groupsInfo {
		group, err := ggpCreateOneGroup(ctx, client, info)
		if err != nil {
			return nil, err
		}
		result = append(result, group.FullPath)
	}

	return result, nil
}

func ggpCreateOneGroup(ctx context.Context, client *tharsis.Client, info ggpGroupInfo) (*types.Group, error) {
	topGroupName := topGroupName // make it possible to make a pointer
	return client.Group.CreateGroup(ctx, &types.CreateGroupInput{
		Name:        info.name,
		ParentPath:  &topGroupName,
		Description: info.description,
	})
}

func teardownFromGetGroupsPaginator(ctx context.Context, client *tharsis.Client, t *testing.T, paths []string) {
	for _, path := range paths {
		err := client.Group.DeleteGroup(ctx, &types.DeleteGroupInput{
			GroupPath: &path,
		})
		assert.Nil(t, err)
	}
}

// The End.
