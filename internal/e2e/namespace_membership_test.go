//go:build integration
// +build integration

package main

import (
	"context"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestGroupMembershipOperations(t *testing.T) {
	ctx := context.Background()
	client, err := createClient()
	require.Nil(t, err)
	assert.NotNil(t, client)

	testGroupName := "membership-test-group-name"
	testGroupPath := topGroupName + "/" + testGroupName
	serviceAccountName := "membership-test-service-account"
	teamName := "tharsis-e2e-group-memberships-test-team"

	// Create a nested group to host this test.
	newDescription := "This is a test group, " + testGroupName
	toCreate := &types.CreateGroupInput{
		Name:        testGroupName,
		ParentPath:  ptr.String(topGroupName), // must use ptr.String, because it's an untyped constant
		Description: newDescription,
	}
	hostGroup, err := client.Group.CreateGroup(ctx, toCreate)
	require.Nil(t, err)
	assert.NotNil(t, hostGroup)

	// Capture the number of pre-existing memberships for later use.
	listed, err := client.NamespaceMembership.GetMemberships(ctx, &types.GetNamespaceMembershipsInput{
		NamespacePath: testGroupPath,
	})
	assert.Nil(t, err)
	assert.NotNil(t, listed)
	preExistingMembershipLen := len(listed)

	// Creating a service account is a necessary prerequisite for creating a membership.
	serviceAccount, err := client.ServiceAccount.CreateServiceAccount(ctx, &types.CreateServiceAccountInput{
		Name:      serviceAccountName,
		GroupPath: testGroupPath,
		OIDCTrustPolicies: []types.OIDCTrustPolicy{
			{
				Issuer:      "https://example.invalid",
				BoundClaims: map[string]string{"butter": "fly"},
			},
		},
	})
	require.Nil(t, err)
	assert.NotNil(t, serviceAccount)

	// Creating a team is a necessary prerequisite for creating a membership.
	team, err := client.Team.CreateTeam(ctx, &types.CreateTeamInput{
		Name: teamName,
	})
	require.Nil(t, err)
	assert.NotNil(t, team)

	// Create two memberships to cover service account and team members with deployer and owner roles.
	membershipsToCreate := 2
	membership2, err := client.NamespaceMembership.AddMembership(ctx, &types.CreateNamespaceMembershipInput{
		NamespacePath:    testGroupPath,
		ServiceAccountID: &serviceAccount.Metadata.ID,
		Role:             "deployer",
	})
	assert.Nil(t, err)
	assert.NotNil(t, membership2)

	membership3, err := client.NamespaceMembership.AddMembership(ctx, &types.CreateNamespaceMembershipInput{
		NamespacePath: testGroupPath,
		TeamName:      &teamName,
		Role:          "owner",
	})
	assert.Nil(t, err)
	assert.NotNil(t, membership3)

	// Update the memberships' roles.
	updated2, err := client.NamespaceMembership.UpdateMembership(ctx, &types.UpdateNamespaceMembershipInput{
		ID:   membership2.Metadata.ID,
		Role: "owner",
	})
	assert.Nil(t, err)
	assert.NotNil(t, updated2)

	updated3, err := client.NamespaceMembership.UpdateMembership(ctx, &types.UpdateNamespaceMembershipInput{
		ID:   membership3.Metadata.ID,
		Role: "viewer",
	})
	assert.Nil(t, err)
	assert.NotNil(t, updated3)

	// List the memberships.
	listed, err = client.NamespaceMembership.GetMemberships(ctx, &types.GetNamespaceMembershipsInput{
		NamespacePath: testGroupPath,
	})
	assert.Nil(t, err)
	assert.NotNil(t, listed)
	assert.Equal(t, preExistingMembershipLen+membershipsToCreate, len(listed))

	// Delete the memberships.
	deleted2, err := client.NamespaceMembership.DeleteMembership(ctx, &types.DeleteNamespaceMembershipInput{
		ID: updated2.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.NotNil(t, deleted2)

	deleted3, err := client.NamespaceMembership.DeleteMembership(ctx, &types.DeleteNamespaceMembershipInput{
		ID: updated3.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.NotNil(t, deleted3)

	// Verify the memberships are gone.
	_, err = client.NamespaceMembership.DeleteMembership(ctx, &types.DeleteNamespaceMembershipInput{
		ID: updated2.Metadata.ID,
	})
	assert.NotNil(t, err)
	assert.True(t, tharsis.IsNotFoundError(err))

	_, err = client.NamespaceMembership.DeleteMembership(ctx, &types.DeleteNamespaceMembershipInput{
		ID: updated3.Metadata.ID,
	})
	assert.NotNil(t, err)
	assert.True(t, tharsis.IsNotFoundError(err))

	// Delete the team.
	err = client.Team.DeleteTeam(ctx, &types.DeleteTeamInput{
		Name: team.Name,
	})
	assert.Nil(t, err)

	// Delete the service account.
	err = client.ServiceAccount.DeleteServiceAccount(ctx, &types.DeleteServiceAccountInput{
		ID: serviceAccount.Metadata.ID,
	})
	assert.Nil(t, err)

	// Delete the host group.
	err = client.Group.DeleteGroup(ctx, &types.DeleteGroupInput{
		ID: &hostGroup.Metadata.ID,
	})
	assert.Nil(t, err)
}

func TestWorkspaceMembershipOperations(t *testing.T) {
	ctx := context.Background()
	client, err := createClient()
	require.Nil(t, err)
	assert.NotNil(t, client)

	testWorkspaceName := "membership-test-workspace-name"
	testWorkspacePath := topGroupName + "/" + testWorkspaceName
	serviceAccountName := "membership-test-service-account"
	teamName := "tharsis-e2e-workspace-memberships-test-team"

	// Create a workspace to host this test.
	newDescription := "This is a test workspace, " + testWorkspaceName
	toCreate := &types.CreateWorkspaceInput{
		Name:        testWorkspaceName,
		GroupPath:   topGroupName, // must use ptr.String, because it's an untyped constant
		Description: newDescription,
	}
	hostWorkspace, err := client.Workspaces.CreateWorkspace(ctx, toCreate)
	require.Nil(t, err)
	assert.NotNil(t, hostWorkspace)

	// Capture the number of pre-existing memberships for later use.
	listed, err := client.NamespaceMembership.GetMemberships(ctx, &types.GetNamespaceMembershipsInput{
		NamespacePath: testWorkspacePath,
	})
	assert.Nil(t, err)
	assert.NotNil(t, listed)
	preExistingMembershipLen := len(listed)

	// Creating a service account in the parent group is a necessary prerequisite for creating a membership.
	serviceAccount, err := client.ServiceAccount.CreateServiceAccount(ctx, &types.CreateServiceAccountInput{
		Name:      serviceAccountName,
		GroupPath: topGroupName,
		OIDCTrustPolicies: []types.OIDCTrustPolicy{
			{
				Issuer:      "https://example.invalid",
				BoundClaims: map[string]string{"butter": "fly"},
			},
		},
	})
	require.Nil(t, err)
	assert.NotNil(t, serviceAccount)

	// Creating a team is a necessary prerequisite for creating a membership.
	team, err := client.Team.CreateTeam(ctx, &types.CreateTeamInput{
		Name: teamName,
	})
	require.Nil(t, err)
	assert.NotNil(t, team)

	// Create two memberships to cover service account and team members with deployer and owner roles.
	membershipsToCreate := 2
	membership2, err := client.NamespaceMembership.AddMembership(ctx, &types.CreateNamespaceMembershipInput{
		NamespacePath:    testWorkspacePath,
		ServiceAccountID: &serviceAccount.Metadata.ID,
		Role:             "deployer",
	})
	assert.Nil(t, err)
	assert.NotNil(t, membership2)

	membership3, err := client.NamespaceMembership.AddMembership(ctx, &types.CreateNamespaceMembershipInput{
		NamespacePath: testWorkspacePath,
		TeamName:      &teamName,
		Role:          "owner",
	})
	assert.Nil(t, err)
	assert.NotNil(t, membership3)

	// Update the memberships' roles.
	updated2, err := client.NamespaceMembership.UpdateMembership(ctx, &types.UpdateNamespaceMembershipInput{
		ID:   membership2.Metadata.ID,
		Role: "owner",
	})
	assert.Nil(t, err)
	assert.NotNil(t, updated2)

	updated3, err := client.NamespaceMembership.UpdateMembership(ctx, &types.UpdateNamespaceMembershipInput{
		ID:   membership3.Metadata.ID,
		Role: "viewer",
	})
	assert.Nil(t, err)
	assert.NotNil(t, updated3)

	// List the memberships.
	listed, err = client.NamespaceMembership.GetMemberships(ctx, &types.GetNamespaceMembershipsInput{
		NamespacePath: testWorkspacePath,
	})
	assert.Nil(t, err)
	assert.NotNil(t, listed)
	assert.Equal(t, preExistingMembershipLen+membershipsToCreate, len(listed))

	// Delete the memberships.
	deleted2, err := client.NamespaceMembership.DeleteMembership(ctx, &types.DeleteNamespaceMembershipInput{
		ID: updated2.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.NotNil(t, deleted2)

	deleted3, err := client.NamespaceMembership.DeleteMembership(ctx, &types.DeleteNamespaceMembershipInput{
		ID: updated3.Metadata.ID,
	})
	assert.Nil(t, err)
	assert.NotNil(t, deleted3)

	// Verify the memberships are gone.
	_, err = client.NamespaceMembership.DeleteMembership(ctx, &types.DeleteNamespaceMembershipInput{
		ID: updated2.Metadata.ID,
	})
	assert.NotNil(t, err)
	assert.True(t, tharsis.IsNotFoundError(err))

	_, err = client.NamespaceMembership.DeleteMembership(ctx, &types.DeleteNamespaceMembershipInput{
		ID: updated3.Metadata.ID,
	})
	assert.NotNil(t, err)
	assert.True(t, tharsis.IsNotFoundError(err))

	// Delete the team.
	err = client.Team.DeleteTeam(ctx, &types.DeleteTeamInput{
		Name: team.Name,
	})
	assert.Nil(t, err)

	// Delete the service account.
	err = client.ServiceAccount.DeleteServiceAccount(ctx, &types.DeleteServiceAccountInput{
		ID: serviceAccount.Metadata.ID,
	})
	assert.Nil(t, err)

	// Delete the host workspace.
	err = client.Workspaces.DeleteWorkspace(ctx, &types.DeleteWorkspaceInput{
		ID: &hostWorkspace.Metadata.ID,
	})
	assert.Nil(t, err)
}
