// Package main contains managed identity examples
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"

	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/config"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// ExampleCreateManagedIdentity demonstrates CreateManagedIdentity.
func ExampleCreateManagedIdentity() (*types.ManagedIdentity, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Access rules for the managed identity.
	accessRules := []types.ManagedIdentityAccessRuleInput{
		{
			RunStage: types.JobPlanType,
			AllowedUsers: []string{
				"user-1-username",
				"user-2-username",
			},
			AllowedServiceAccounts: []string{
				"wild-space/service-account-1-path",
			},
			AllowedTeams: []string{
				"team-1-name",
			},
		},
	}

	// Prepare managed identity data.
	bytes, err := json.Marshal(map[string]string{
		"role": "some-iam-role",
		/* For ManagedIdentityAzureFederated...
		"subject": "..."
		"clientId": "..."
		"tenantId": "..."
		*/
	})
	if err != nil {
		return nil, err
	}

	// Base64 encode the data.
	payload := base64.StdEncoding.EncodeToString(bytes)

	// Create an AWS federated managed identity.
	created, err := client.ManagedIdentity.CreateManagedIdentity(ctx,
		&types.CreateManagedIdentityInput{
			Type:        types.ManagedIdentityAWSFederated,
			Name:        "new-managed-identity-01",
			GroupPath:   "wild-space", // Group where this managed identity will be created.
			Description: "This is a newly created managed identity.",
			Data:        payload,
			AccessRules: accessRules,
		},
	)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// ExampleUpdateManagedIdentity demonstrates UpdateManagedIdentity.
func ExampleUpdateManagedIdentity(id, data string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	updated, err := client.ManagedIdentity.UpdateManagedIdentity(ctx,
		&types.UpdateManagedIdentityInput{
			ID:          id,
			Data:        data,
			Description: "This is the updated managed identity",
		},
	)
	if err != nil {
		return err
	}

	_ = updated

	return nil
}

// ExampleDeleteManagedIdentity demonstrates DeleteManagedIdentity.
func ExampleDeleteManagedIdentity(id string, force bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = client.ManagedIdentity.DeleteManagedIdentity(ctx,
		&types.DeleteManagedIdentityInput{
			ID:    id,
			Force: force,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
