//go:build integration
// +build integration

package main

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TestCRUDNamespaceVariable tests namespace variable create, get, update, and delete.
func TestCRUDNamespaceVariable(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	assert.Nil(t, err)
	assert.NotNil(t, client)

	namespaceVariableCategory := types.TerraformVariableCategory
	namespaceVariableHCL := true
	namespaceVariableKey := "variable-key"
	namespaceVariableValue := "variable-value"

	updatedNamespaceVariableHCL := false
	updatedNamespaceVariableKey := "updated-variable-key"
	updatedNamespaceVariableValue := "updated-variable-value"

	// Create the namespace variable.
	createdNamespaceVariable, err := client.Variable.CreateVariable(ctx, &types.CreateNamespaceVariableInput{
		NamespacePath: topGroupName,
		Category:      namespaceVariableCategory,
		HCL:           namespaceVariableHCL,
		Key:           namespaceVariableKey,
		Value:         namespaceVariableValue,
	})
	assert.Nil(t, err)
	assert.NotNil(t, createdNamespaceVariable)

	// Verify all the fields except metadata.
	assert.Equal(t, topGroupName, createdNamespaceVariable.NamespacePath)
	assert.Equal(t, namespaceVariableCategory, createdNamespaceVariable.Category)
	assert.Equal(t, namespaceVariableHCL, createdNamespaceVariable.HCL)
	assert.Equal(t, namespaceVariableKey, createdNamespaceVariable.Key)
	assert.Equal(t, namespaceVariableValue, *createdNamespaceVariable.Value)

	// Get/read and verify the namespace variable.
	readNamespaceVariable, err := client.Variable.GetVariable(ctx, &types.GetNamespaceVariableInput{
		ID: createdNamespaceVariable.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, createdNamespaceVariable, readNamespaceVariable)

	// Update the namespace variable.
	updatedNamespaceVariable, err := client.Variable.UpdateVariable(ctx,
		&types.UpdateNamespaceVariableInput{
			ID:    createdNamespaceVariable.Metadata.ID,
			HCL:   updatedNamespaceVariableHCL,
			Key:   updatedNamespaceVariableKey,
			Value: updatedNamespaceVariableValue,
		})
	assert.Nil(t, err)

	// Verify the claimed update.
	assert.Equal(t, updatedNamespaceVariableHCL, updatedNamespaceVariable.HCL)
	assert.Equal(t, updatedNamespaceVariableKey, updatedNamespaceVariable.Key)
	assert.Equal(t, updatedNamespaceVariableValue, *updatedNamespaceVariable.Value)

	// Retrieve and verify the updated namespace variable to make sure it persisted.
	read2NamespaceVariable, err := client.Variable.GetVariable(ctx, &types.GetNamespaceVariableInput{
		ID: createdNamespaceVariable.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, updatedNamespaceVariable, read2NamespaceVariable)

	// Delete the namespace variable.
	err = client.Variable.DeleteVariable(ctx,
		&types.DeleteNamespaceVariableInput{
			ID: read2NamespaceVariable.Metadata.ID,
		})
	assert.Nil(t, err)

	// Verify the namespace variable is gone.
	read3NamespaceVariable, err := client.Variable.GetVariable(ctx, &types.GetNamespaceVariableInput{
		ID: read2NamespaceVariable.Metadata.ID,
	})

	assert.True(t, strings.Contains(err.Error(), "not found"))
	assert.Equal(t, (*types.NamespaceVariable)(nil), read3NamespaceVariable)
}
