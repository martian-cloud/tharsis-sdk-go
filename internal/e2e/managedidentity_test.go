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

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, getManagedIdentityIAMRole, d.Role)

	// Get the managed identity by ID.
	gotIdentity, err := client.ManagedIdentity.GetManagedIdentity(ctx, &types.GetManagedIdentityInput{
		ID: &createdIdentity.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.NotNil(t, gotIdentity)

	// Verify the returned contents are what they should be.
	assert.Equal(t, toCreate.Name, gotIdentity.Name)
	assert.Equal(t, toCreate.Type, gotIdentity.Type)
	assert.Equal(t, toCreate.Description, gotIdentity.Description)
	assert.Equal(t, toCreate.GroupPath+"/"+toCreate.Name, gotIdentity.ResourcePath)
	// The Data field has a subject added to it, so it's not practical to verify it here.

	// Get the managed identity by path.
	gotIdentity, err = client.ManagedIdentity.GetManagedIdentity(ctx, &types.GetManagedIdentityInput{
		Path: &createdIdentity.ResourcePath,
	})
	assert.Nil(t, err)
	assert.NotNil(t, gotIdentity)

	// Verify the returned contents are what they should be.
	assert.Equal(t, toCreate.Name, gotIdentity.Name)
	assert.Equal(t, toCreate.Type, gotIdentity.Type)
	assert.Equal(t, toCreate.Description, gotIdentity.Description)
	assert.Equal(t, toCreate.GroupPath+"/"+toCreate.Name, gotIdentity.ResourcePath)
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
			Type:                   types.ManagedIdentityAccessRuleEligiblePrincipals,
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
	assert.Equal(t, updateManagedIdentityIAMRole, d.Role)

	// Get the access rules.
	accessRules, err := client.ManagedIdentity.GetManagedIdentityAccessRules(ctx, &types.GetManagedIdentityInput{
		ID: &createdIdentity.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(accessRules))

	// Verify accessRules.
	for _, accessRule := range accessRules {
		assert.Equal(t, types.ManagedIdentityAccessRuleEligiblePrincipals, accessRule.Type)
		assert.Equal(t, types.JobPlanType, accessRule.RunStage)
		assert.Empty(t, accessRule.ModuleAttestationPolicies)
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
	assert.Equal(t, 2, len(parts))
	username := parts[0]

	// Create an access rule.
	ruleInput := &types.CreateManagedIdentityAccessRuleInput{
		Type:                   types.ManagedIdentityAccessRuleEligiblePrincipals,
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
		ID: &createdIdentity.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(accessRules))
	readRule := accessRules[0]

	// Verify accessRule.
	assert.Equal(t, types.JobPlanType, readRule.RunStage)
	assert.NotNil(t, readRule.AllowedUsers)
	assert.Equal(t, 1, len(readRule.AllowedUsers))
	assert.Equal(t, types.ManagedIdentityAccessRuleEligiblePrincipals, readRule.Type)
	assert.Equal(t, email, readRule.AllowedUsers[0].Email)
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
	assert.Equal(t, types.JobApplyType, updatedRule.RunStage)

	// Retrieve the updated access rules to make sure they persisted.
	accessRules, err = client.ManagedIdentity.GetManagedIdentityAccessRules(ctx, &types.GetManagedIdentityInput{
		ID: &createdIdentity.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(accessRules))
	read2Rule := accessRules[0]

	// Verify the retrieved rule.
	assert.Equal(t, types.JobApplyType, read2Rule.RunStage)

	// Delete the access rule.
	err = client.ManagedIdentity.DeleteManagedIdentityAccessRule(ctx,
		&types.DeleteManagedIdentityAccessRuleInput{
			ID: readRule.Metadata.ID,
		})
	assert.Nil(t, err)

	// Verify the access rule is gone.
	accessRules, err = client.ManagedIdentity.GetManagedIdentityAccessRules(ctx, &types.GetManagedIdentityInput{
		ID: &createdIdentity.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(accessRules))

	// Delete the new managed identity.
	err = client.ManagedIdentity.DeleteManagedIdentity(ctx, &types.DeleteManagedIdentityInput{
		ID:    createdIdentity.Metadata.ID,
		Force: true,
	})
	assert.Nil(t, err)
}

func TestCreateDeleteManagedIdentityAlias(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	managedIdentityName := "managed-identity-02"
	managedIdentityIAMRole := "managed-identity-02-iam-role"
	aliasName := managedIdentityName + "-alias"

	// Prepare managed identity data.
	bytes, err := json.Marshal(map[string]string{
		"role": managedIdentityIAMRole,
	})
	assert.Nil(t, err)

	// Base64 encode the data.
	managedIdentityData := base64.StdEncoding.EncodeToString(bytes)

	// Create a child group beneath the root group, so that an alias can be shared up
	// to the parent group. Reverse isn't possible.
	identityGroup, err := client.Group.CreateGroup(ctx, &types.CreateGroupInput{
		Name:        "group-for-managed-identity",
		ParentPath:  ptr.String(topGroupName),
		Description: "This is a group created to test managed identity aliases",
	})
	assert.Nil(t, err)
	assert.NotNil(t, identityGroup)

	// Create the managed identity first so it can be aliased.
	createdIdentity, err := client.ManagedIdentity.CreateManagedIdentity(ctx, &types.CreateManagedIdentityInput{
		Name:        managedIdentityName,
		Type:        types.ManagedIdentityAWSFederated,
		Description: "This is a test managed identity",
		GroupPath:   identityGroup.FullPath,
		Data:        managedIdentityData,
	})
	assert.Nil(t, err)
	assert.NotNil(t, createdIdentity)

	// Create the managed identity alias in the root group.
	createdAlias, err := client.ManagedIdentity.CreateManagedIdentityAlias(ctx, &types.CreateManagedIdentityAliasInput{
		AliasSourceID: &createdIdentity.Metadata.ID,
		Name:          aliasName,
		GroupPath:     topGroupName,
	})
	assert.Nil(t, err)
	assert.NotNil(t, createdAlias)
	assert.NotNil(t, createdAlias.AliasSourceID)
	assert.Equal(t, aliasName, createdAlias.Name)
	assert.Equal(t, createdIdentity.Description, createdAlias.Description)
	assert.Equal(t, createdIdentity.Type, createdAlias.Type)
	assert.Equal(t, strings.Join([]string{topGroupName, aliasName}, "/"), createdAlias.ResourcePath)
	assert.Equal(t, createdIdentity.Data, createdAlias.Data)
	assert.Equal(t, createdIdentity.Metadata.ID, *createdAlias.AliasSourceID)
	assert.True(t, createdAlias.IsAlias)

	err = client.ManagedIdentity.DeleteManagedIdentityAlias(ctx, &types.DeleteManagedIdentityAliasInput{
		ID:    createdAlias.Metadata.ID,
		Force: true,
	})
	assert.Nil(t, err)

	managedIdentity, err := client.ManagedIdentity.GetManagedIdentity(ctx, &types.GetManagedIdentityInput{
		ID: &createdAlias.Metadata.ID,
	})
	assert.Nil(t, managedIdentity)
	assert.NotNil(t, err) // Expect an error here.

	// Deleting the group will delete the nested managed identity as well.
	err = client.Group.DeleteGroup(ctx, &types.DeleteGroupInput{
		ID:    &identityGroup.Metadata.ID,
		Force: ptr.Bool(true),
	})
	assert.Nil(t, err)
}
