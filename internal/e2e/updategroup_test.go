//go:build integration
// +build integration

package main

import (
	"context"
	"testing"
	"time"

	"github.com/likexian/gokit/assert"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// This module tests creating, getting, updating, and (eventually) deleting a group.

func TestUpdateGroup(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	// Name and path for the group.
	updateGroupName := "update-group-01"
	updateGroupPath := topGroupName + "/" + updateGroupName

	// Create the new group.
	newDescription := "This is a test group not yet updated, " + updateGroupName
	topGroupName := topGroupName // enable a pointer to be made
	createdGroup, err := client.Group.CreateGroup(ctx, &types.CreateGroupInput{
		Name:        updateGroupName,
		ParentPath:  &topGroupName,
		Description: newDescription,
	})
	assert.Nil(t, err)
	assert.NotNil(t, createdGroup)
	assert.Equal(t, updateGroupName, createdGroup.Name)
	assert.Equal(t, newDescription, createdGroup.Description)
	assert.Equal(t, updateGroupPath, createdGroup.FullPath)

	// Get the newly-created group.
	toUpdateGroup, err := client.Group.GetGroup(ctx, &types.GetGroupInput{Path: updateGroupPath})
	assert.Nil(t, err)
	assert.NotNil(t, toUpdateGroup)

	// Update the group's description.
	newDescription = "This is a test group updated at " + time.Now().String()
	updatedGroup, err := client.Group.UpdateGroup(ctx,
		&types.UpdateGroupInput{
			GroupPath:   toUpdateGroup.FullPath,
			Description: newDescription,
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, newDescription, updatedGroup.Description)

	// Delete the new group.
	err = client.Group.DeleteGroup(ctx, &types.DeleteGroupInput{
		GroupPath: updatedGroup.FullPath,
	})
	assert.Nil(t, err)

}

// The End.
