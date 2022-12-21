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

// This module tests getting a list of groups with_OUT_ use of a paginator.
// Three groups are created as direct children of the top-level group.
// Prefix 'gg' stands for 'get groups' to avoid collision with other tests.

type ggGroupInfo struct {
	name        string
	description string
	path        string
}

const (
	ggGroupCount = 3
)

// TestGetGroups gets multiple groups.
func TestGetGroups(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	// Build the group names, descriptions, and paths.
	groupsInfo := buildInfoForGetGroups(ggGroupCount)

	// Create the groups.
	groupPaths, err := setupForGetGroups(ctx, client, groupsInfo)
	assert.Nil(t, err)
	assert.Equal(t, ggGroupCount, len(groupPaths))

	// Tear down the groups when the test has finished.
	defer teardownFromGetGroups(ctx, client, t, groupPaths)

	// Get the list of groups.
	toSort := types.GroupSortableFieldFullPathAsc
	pageLimit := pageLimit100
	wantParent := topGroupName
	foundGroupsOutput, err := client.Group.GetGroups(ctx,
		&types.GetGroupsInput{
			Sort: &toSort,
			PaginationOptions: &types.PaginationOptions{
				Limit: &pageLimit,
			},
			Filter: &types.GroupFilter{ParentPath: &wantParent},
		})
	assert.Nil(t, err)
	assert.NotNil(t, foundGroupsOutput)
	foundGroups := foundGroupsOutput.Groups

	// Check the paths.
	foundPaths := []string{}
	for _, foundGroup := range foundGroups {
		foundPaths = append(foundPaths, foundGroup.FullPath)
	}

	// Expect gg group IDs.
	expectPaths := []string{}
	for _, info := range groupsInfo {
		expectPaths = append(expectPaths, info.path)
	}

	assert.True(t, reflect.DeepEqual(expectPaths, foundPaths))
}

func buildInfoForGetGroups(count int) []ggGroupInfo {
	result := []ggGroupInfo{}

	for ix := 1; ix <= count; ix++ {
		name := fmt.Sprintf("gg-group-%d", ix)
		info := ggGroupInfo{
			name:        name,
			description: fmt.Sprintf("This is gg test group %d.", ix),
			path:        fmt.Sprintf("%s/%s", topGroupName, name),
		}
		result = append(result, info)
	}

	return result
}

func setupForGetGroups(ctx context.Context, client *tharsis.Client, groupsInfo []ggGroupInfo) ([]string, error) {
	result := []string{}

	for _, info := range groupsInfo {
		group, err := ggCreateOneGroup(ctx, client, info)
		if err != nil {
			return nil, err
		}
		result = append(result, group.FullPath)
	}

	return result, nil
}

func ggCreateOneGroup(ctx context.Context, client *tharsis.Client, info ggGroupInfo) (*types.Group, error) {
	topGroupName := topGroupName // make it possible to make a pointer
	return client.Group.CreateGroup(ctx, &types.CreateGroupInput{
		Name:        info.name,
		ParentPath:  &topGroupName,
		Description: info.description,
	})
}

func teardownFromGetGroups(ctx context.Context, client *tharsis.Client, t *testing.T, paths []string) {
	for _, path := range paths {
		err := client.Group.DeleteGroup(ctx, &types.DeleteGroupInput{
			GroupPath: &path,
		})
		assert.Nil(t, err)
	}
}

// The End.
