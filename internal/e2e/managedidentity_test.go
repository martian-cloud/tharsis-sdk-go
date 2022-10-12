//go:build integration
// +build integration

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
	accessRules, err := client.ManagedIdentity.GetManagedIdentityAccessRules(ctx, &types.GetManagedIdentityAccessRuleInput{
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
