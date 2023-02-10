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

func TestGetGPGKeyCRUD(t *testing.T) {

	ctx := context.Background()
	client, err := createClient()
	require.Nil(t, err)
	assert.NotNil(t, client)

	// a made-up GPG key for testing and its resulting fields
	armor := `-----BEGIN PGP PUBLIC KEY BLOCK-----

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
=HJT6
-----END PGP PUBLIC KEY BLOCK-----
`
	gpgKeyID := "41F57AB6741F92F1"
	fingerprint := "C48F7F19B14E59433DF9402341F57AB6741F92F1"

	// Create the GPG key.
	toCreate := &types.CreateGPGKeyInput{
		ASCIIArmor: armor,
		GroupPath:  topGroupName,
	}
	createdGPGKey, err := client.GPGKey.CreateGPGKey(ctx, toCreate)

	require.Nil(t, err)
	assert.NotNil(t, createdGPGKey)
	assert.Equal(t, armor, createdGPGKey.ASCIIArmor)
	assert.Equal(t, fingerprint, createdGPGKey.Fingerprint)
	assert.Equal(t, gpgKeyID, createdGPGKey.GPGKeyID)

	// Get the GPG key to make sure it persisted.
	gotGPGKey, err := client.GPGKey.GetGPGKey(ctx, &types.GetGPGKeyInput{
		ID: createdGPGKey.Metadata.ID,
	})
	require.Nil(t, err)
	assert.NotNil(t, gotGPGKey)

	// Verify the returned contents are what they should be.
	assert.Equal(t, armor, gotGPGKey.ASCIIArmor)
	assert.Equal(t, fingerprint, gotGPGKey.Fingerprint)
	assert.Equal(t, gpgKeyID, gotGPGKey.GPGKeyID)

	// Delete the GPG key.
	deletedGPGKey, err := client.GPGKey.DeleteGPGKey(ctx, &types.DeleteGPGKeyInput{
		ID: gotGPGKey.Metadata.ID,
	})
	require.Nil(t, err)
	assert.NotNil(t, deletedGPGKey)
	assert.Equal(t, armor, deletedGPGKey.ASCIIArmor)
	assert.Equal(t, fingerprint, deletedGPGKey.Fingerprint)
	assert.Equal(t, gpgKeyID, deletedGPGKey.GPGKeyID)
}

// The End.
