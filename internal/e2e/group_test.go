//go:build integration
// +build integration

package main

import (
	"context"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/likexian/gokit/assert"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TODO: This module has e2e tests only for newer method(s) added in December, 2022.
// The other methods should also have e2e tests added, including a TestGetGroupByPath.

func TestGetGroupByID(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	// Test only a nested group here, because this test's focus is on get, not create.
	getGroupName := "get-group-01"
	getGroupFullPath := topGroupName + "/" + getGroupName

	// Create the group.
	newDescription := "This is a test group, " + getGroupName
	toCreate := &types.CreateGroupInput{
		Name:        getGroupName,
		ParentPath:  ptr.String(topGroupName), // must use ptr.String, because it's an untyped constant
		Description: newDescription,
	}
	createdGroup, err := client.Group.CreateGroup(ctx, toCreate)
	assert.Nil(t, err)
	assert.NotNil(t, createdGroup)
	assert.Equal(t, createdGroup.Name, getGroupName)
	assert.Equal(t, createdGroup.FullPath, getGroupFullPath)
	assert.Equal(t, createdGroup.Description, newDescription)

	// Get the group.
	gotGroup, err := client.Group.GetGroup(ctx, &types.GetGroupInput{
		ID: &createdGroup.Metadata.ID,
	})
	assert.Nil(t, err)

	// Verify the returned contents are what they should be.
	assert.Equal(t, gotGroup.Name, toCreate.Name)
	assert.Equal(t, gotGroup.FullPath, getGroupFullPath)
	assert.Equal(t, gotGroup.Description, toCreate.Description)

	err = client.Group.DeleteGroup(ctx, &types.DeleteGroupInput{
		ID: &gotGroup.Metadata.ID,
	})
	assert.Nil(t, err)
}

// The End.
