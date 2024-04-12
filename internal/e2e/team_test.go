//go:build integration
// +build integration

package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestTeamOperations(t *testing.T) {
	ctx := context.Background()
	client, err := createClient()
	require.Nil(t, err)
	assert.NotNil(t, client)

	teamName := "tharsis-e2e-group-memberships-test-team"

	// Create a team.
	team, err := client.Team.CreateTeam(ctx, &types.CreateTeamInput{
		Name: teamName,
	})
	require.Nil(t, err)
	assert.NotNil(t, team)

	// Get the team by name.
	gotTeam, err := client.Team.GetTeam(ctx, &types.GetTeamInput{Name: &teamName})
	require.Nil(t, err)
	assert.NotNil(t, gotTeam)

	// Cannot test AddTeamMember, because that requires a user.

	// Delete the team.
	err = client.Team.DeleteTeam(ctx, &types.DeleteTeamInput{
		Name: team.Name,
	})
	assert.Nil(t, err)
}
