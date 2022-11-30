//go:build integration
// +build integration

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/likexian/gokit/assert"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// data represents the managed identity data.
type data struct {
	Subject string `json:"subject"`
	Role    string `json:"role"`
}

func TestGetManagedIdentity(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	getManagedIdentityName := "get-managed-identity-01"
	getManagedIdentityIAMRole := "get-managed-identity-iam-role"

	// Prepare managed identity data.
	bytes, err := json.Marshal(map[string]string{
		"role": getManagedIdentityIAMRole,
	})
	assert.Nil(t, err)

	// Base64 encode the data.
	getManagedIdentityData := base64.StdEncoding.EncodeToString(bytes)

	// Create the managed identity.
	newDescription := "This is a test managed identity, " + getManagedIdentityName
	toCreate := &types.CreateManagedIdentityInput{
		Name:        getManagedIdentityName,
		Type:        types.ManagedIdentityAWSFederated,
		Description: newDescription,
		GroupPath:   topGroupName,
		Data:        getManagedIdentityData,
		AccessRules: []types.ManagedIdentityAccessRuleInput{},
	}
	createdIdentity, err := client.ManagedIdentity.CreateManagedIdentity(ctx, toCreate)
	assert.Nil(t, err)
	assert.NotNil(t, createdIdentity)
	assert.Equal(t, getManagedIdentityName, createdIdentity.Name)
	assert.Equal(t, newDescription, createdIdentity.Description)
	assert.Equal(t, types.ManagedIdentityAWSFederated, createdIdentity.Type)

	createdIdentityData, err := base64.StdEncoding.DecodeString(createdIdentity.Data)
	assert.Nil(t, err)

	var d data
	err = json.Unmarshal(createdIdentityData, &d)
	assert.Nil(t, err)

	// Verify data IAM role matches.
	assert.Equal(t, d.Role, getManagedIdentityIAMRole)

	// Get the managed identity.
	gotIdentity, err := client.ManagedIdentity.GetManagedIdentity(ctx, &types.GetManagedIdentityInput{
		ID: createdIdentity.Metadata.ID,
	})
	assert.Nil(t, err)

	// Verify the returned contents are what they should be.
	assert.Equal(t, gotIdentity.Name, toCreate.Name)
	assert.Equal(t, gotIdentity.Type, toCreate.Type)
	assert.Equal(t, gotIdentity.Description, toCreate.Description)
	assert.Equal(t, gotIdentity.ResourcePath, toCreate.GroupPath+"/"+toCreate.Name)
	// The Data field has a subject added to it, so it's not practical to verify it here.

	// Delete the new managed identity.
	err = client.ManagedIdentity.DeleteManagedIdentity(ctx, &types.DeleteManagedIdentityInput{
		ID:    gotIdentity.Metadata.ID,
		Force: true,
	})
	assert.Nil(t, err)
}

func TestUpdateManagedIdentity(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	updateManagedIdentityName := "update-managed-identity-01"
	updateManagedIdentityIAMRole := "update-managed-identity-iam-role"

	// Access rules for the managed identity.
	updateManagedIdentityAccessRules := []types.ManagedIdentityAccessRuleInput{
		{
			RunStage:               types.JobPlanType,
			AllowedUsers:           []string{},
			AllowedServiceAccounts: []string{},
			AllowedTeams:           []string{},
		},
	}

	// Prepare managed identity data.
	bytes, err := json.Marshal(map[string]string{
		"role": updateManagedIdentityIAMRole,
	})
	assert.Nil(t, err)

	// Base64 encode the data.
	updateManagedIdentityData := base64.StdEncoding.EncodeToString(bytes)

	// Create the managed identity.
	newDescription := "This is a test managed identity not yet updated, " + updateManagedIdentityName
	createdIdentity, err := client.ManagedIdentity.CreateManagedIdentity(ctx, &types.CreateManagedIdentityInput{
		Name:        updateManagedIdentityName,
		Type:        types.ManagedIdentityAWSFederated,
		Description: newDescription,
		GroupPath:   topGroupName,
		Data:        updateManagedIdentityData,
		AccessRules: updateManagedIdentityAccessRules,
	})
	assert.Nil(t, err)
	assert.NotNil(t, createdIdentity)
	assert.Equal(t, updateManagedIdentityName, createdIdentity.Name)
	assert.Equal(t, newDescription, createdIdentity.Description)
	assert.Equal(t, types.ManagedIdentityAWSFederated, createdIdentity.Type)

	createdIdentityData, err := base64.StdEncoding.DecodeString(createdIdentity.Data)
	assert.Nil(t, err)

	var d data
	err = json.Unmarshal(createdIdentityData, &d)
	assert.Nil(t, err)

	// Verify data IAM role matches.
	assert.Equal(t, d.Role, updateManagedIdentityIAMRole)

	// Get the access rules.
	accessRules, err := client.ManagedIdentity.GetManagedIdentityAccessRules(ctx, &types.GetManagedIdentityInput{
		ID: createdIdentity.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, len(accessRules), 1)

	// Verify accessRules.
	for _, accessRule := range accessRules {
		assert.Equal(t, accessRule.RunStage, types.JobPlanType)
		assert.NotNil(t, accessRule.AllowedUsers)
		assert.NotNil(t, accessRule.AllowedServiceAccounts)
		assert.NotNil(t, accessRule.AllowedTeams)
	}

	// Prepare managed identity data.
	bytes, err = json.Marshal(map[string]string{
		"role": "update-managed-identity-iam-role",
	})
	assert.Nil(t, err)

	// Base64 encode the data.
	updatedManagedIdentityData := base64.StdEncoding.EncodeToString(bytes)

	// Update the data and description of the managed identity.
	newDescription = "This is a test managed identity updated at " + time.Now().String()
	updatedIdentity, err := client.ManagedIdentity.UpdateManagedIdentity(ctx, &types.UpdateManagedIdentityInput{
		ID:          createdIdentity.Metadata.ID,
		Data:        updatedManagedIdentityData,
		Description: newDescription,
	})
	assert.Nil(t, err)
	assert.Equal(t, newDescription, updatedIdentity.Description)

	// Delete the new managed identity.
	err = client.ManagedIdentity.DeleteManagedIdentity(ctx, &types.DeleteManagedIdentityInput{
		ID:    updatedIdentity.Metadata.ID,
		Force: true,
	})
	assert.Nil(t, err)
}

// TestCRUDManagedIdentityAccessRule tests managed identity access rule create, get, update, and delete.
func TestCRUDManagedIdentityAccessRule(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	managedIdentityName := "managed-identity-02"
	managedIdentityIAMRole := "managed-identity-02-iam-role"

	// Prepare managed identity data.
	bytes, err := json.Marshal(map[string]string{
		"role": managedIdentityIAMRole,
	})
	assert.Nil(t, err)

	// Base64 encode the data.
	managedIdentityData := base64.StdEncoding.EncodeToString(bytes)

	// Create the managed identity.
	newDescription := "This is a test managed identity, " + managedIdentityName
	createdIdentity, err := client.ManagedIdentity.CreateManagedIdentity(ctx, &types.CreateManagedIdentityInput{
		Name:        managedIdentityName,
		Type:        types.ManagedIdentityAWSFederated,
		Description: newDescription,
		GroupPath:   topGroupName,
		Data:        managedIdentityData,
	})
	assert.Nil(t, err)
	assert.NotNil(t, createdIdentity)
	assert.Equal(t, managedIdentityName, createdIdentity.Name)
	assert.Equal(t, newDescription, createdIdentity.Description)
	assert.Equal(t, types.ManagedIdentityAWSFederated, createdIdentity.Type)

	// Rather than creating users, teams, and/or service accounts just for this test,
	// take advantage of the CreatedBy field as a user.  That might not work in CI,
	// but it will work for local manual testing.
	email := createdIdentity.CreatedBy
	parts := strings.Split(email, "@")
	assert.Equal(t, len(parts), 2)
	username := parts[0]

	// Create an access rule.
	ruleInput := &types.CreateManagedIdentityAccessRuleInput{
		ManagedIdentityID:      createdIdentity.Metadata.ID,
		RunStage:               types.JobPlanType,
		AllowedUsers:           []string{username},
		AllowedServiceAccounts: []string{},
		AllowedTeams:           []string{},
	}
	createdRule, err := client.ManagedIdentity.CreateManagedIdentityAccessRule(ctx, ruleInput)
	assert.Nil(t, err)
	assert.NotNil(t, createdRule)

	// Get/read the access rules.
	accessRules, err := client.ManagedIdentity.GetManagedIdentityAccessRules(ctx, &types.GetManagedIdentityInput{
		ID: createdIdentity.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, len(accessRules), 1)
	readRule := accessRules[0]

	// Verify accessRule.
	assert.Equal(t, readRule.RunStage, types.JobPlanType)
	assert.NotNil(t, readRule.AllowedUsers)
	assert.Equal(t, len(readRule.AllowedUsers), 1)
	assert.Equal(t, readRule.AllowedUsers[0].Email, email)
	assert.NotNil(t, readRule.AllowedServiceAccounts)
	assert.NotNil(t, readRule.AllowedTeams)

	// Update the access rule.
	updatedRule, err := client.ManagedIdentity.UpdateManagedIdentityAccessRule(ctx,
		&types.UpdateManagedIdentityAccessRuleInput{
			ID:                     readRule.Metadata.ID,
			RunStage:               types.JobApplyType,
			AllowedUsers:           []string{username},
			AllowedServiceAccounts: []string{},
			AllowedTeams:           []string{},
		})
	assert.Nil(t, err)

	// Verify the claimed update.
	assert.Equal(t, updatedRule.RunStage, types.JobApplyType)

	// Retrieve the updated access rules to make sure they persisted.
	accessRules, err = client.ManagedIdentity.GetManagedIdentityAccessRules(ctx, &types.GetManagedIdentityInput{
		ID: createdIdentity.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, len(accessRules), 1)
	read2Rule := accessRules[0]

	// Verify the retrieved rule.
	assert.Equal(t, read2Rule.RunStage, types.JobApplyType)

	// Delete the access rule.
	err = client.ManagedIdentity.DeleteManagedIdentityAccessRule(ctx,
		&types.DeleteManagedIdentityAccessRuleInput{
			ID: readRule.Metadata.ID,
		})
	assert.Nil(t, err)

	// Verify the access rule is gone.
	accessRules, err = client.ManagedIdentity.GetManagedIdentityAccessRules(ctx, &types.GetManagedIdentityInput{
		ID: createdIdentity.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, len(accessRules), 0)

	// Delete the new managed identity.
	err = client.ManagedIdentity.DeleteManagedIdentity(ctx, &types.DeleteManagedIdentityInput{
		ID:    createdIdentity.Metadata.ID,
		Force: true,
	})
	assert.Nil(t, err)
}
