//go:build integration
// +build integration

package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// This module tests migrating a group.

func TestMigrateGroup(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	// Groups to create in order to test migration.
	topGroupName := topGroupName // enable a pointer to be made
	oldParentName := "old-parent"
	newParentName := "new-parent"
	oldParentPath := topGroupName + "/" + oldParentName
	newParentPath := topGroupName + "/" + newParentName
	toMigrateName := "group-to-migrate"
	groupsToCreate := []types.CreateGroupInput{
		{
			Name:        oldParentName,
			ParentPath:  &topGroupName,
			Description: "This is the old parent of the group to migrate.",
		},
		{
			Name:        newParentName,
			ParentPath:  &topGroupName,
			Description: "This is the new parent of the group to migrate.",
		},
		{
			Name:        toMigrateName,
			ParentPath:  &oldParentPath,
			Description: "This is the group to migrate.",
		},
	}

	// Create the groups.
	groups := []types.Group{}
	for _, toCreate := range groupsToCreate {
		createdGroup, cErr := client.Group.CreateGroup(ctx, &toCreate)
		assert.Nil(t, cErr)
		checkGroup(t, toCreate.Name, toCreate.Description, buildPath(toCreate.ParentPath, toCreate.Name), createdGroup)
		groups = append(groups, *createdGroup)
	}
	parents := groups[0:2]
	toMove := groups[2]

	// Move the target group from the old parent to a new parent.
	toMove = *migrateGroupAndCheck(t, ctx, client, parents, toMove, &newParentPath)

	// Move the target group from the new parent to top-level.
	toMove = *migrateGroupAndCheck(t, ctx, client, parents, toMove, nil)

	// Move the target group from top level back to the old parent.
	toMove = *migrateGroupAndCheck(t, ctx, client, parents, toMove, &oldParentPath)

	// Delete the groups, the child first, then the parents.
	toDeleteGroups := []types.Group{toMove}
	toDeleteGroups = append(toDeleteGroups, parents...)
	for _, toDelete := range toDeleteGroups {
		dErr := client.Group.DeleteGroup(ctx, &types.DeleteGroupInput{
			GroupPath: &toDelete.FullPath,
		})
		assert.Nil(t, dErr)
	}
}

// Utility functions.

// migrateGroupAndCheck migrates a group and returns the migrated group.
func migrateGroupAndCheck(t *testing.T, ctx context.Context, client *tharsis.Client,
	parents []types.Group, toMove types.Group, moveToParent *string) *types.Group {
	newPath := buildPath(moveToParent, toMove.Name)

	// Do the migration and check the claimed migrated group.
	migratedGroup, err := client.Group.MigrateGroup(ctx, &types.MigrateGroupInput{
		GroupPath:     toMove.FullPath,
		NewParentPath: moveToParent,
	})
	assert.Nil(t, err)
	checkGroup(t, toMove.Name, toMove.Description, newPath, migratedGroup)

	// Get the group by ID and check that it was persisted correctly.
	gotGroup, err := client.Group.GetGroup(ctx, &types.GetGroupInput{
		ID: &toMove.Metadata.ID,
	})
	assert.Nil(t, err)
	checkGroup(t, toMove.Name, toMove.Description, newPath, gotGroup)

	// Check the parents to make sure they weren't modified.
	for _, parent := range parents {
		gotParent, err := client.Group.GetGroup(ctx, &types.GetGroupInput{
			ID: &parent.Metadata.ID,
		})
		assert.Nil(t, err)
		checkGroup(t, parent.Name, parent.Description, parent.FullPath, gotParent)
	}

	return migratedGroup
}

func checkGroup(t *testing.T, expectName, expectDescription, expectPath string, actual *types.Group) {
	assert.NotNil(t, actual)
	assert.Equal(t, expectName, actual.Name)
	assert.Equal(t, expectDescription, actual.Description)
	assert.Equal(t, expectPath, actual.FullPath)
}

func buildPath(parentPath *string, childName string) string {
	if parentPath == nil {
		return childName
	}

	return *parentPath + "/" + childName
}

// The End.
