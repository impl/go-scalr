package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurationVersionsCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	wsTest, wsTestCleanup := createWorkspace(t, client, nil)
	defer wsTestCleanup()

	t.Run("with valid options", func(t *testing.T) {
		cv, err := client.ConfigurationVersions.Create(ctx,
			ConfigurationVersionCreateOptions{Workspace: wsTest},
		)
		require.NoError(t, err)

		// Get a refreshed view of the configuration version.
		refreshed, err := client.ConfigurationVersions.Read(ctx, cv.ID)
		require.NoError(t, err)
		assert.Equal(t, cv, refreshed)
	})
	t.Run("when no workspace is provided", func(t *testing.T) {
		_, err := client.ConfigurationVersions.Create(ctx, ConfigurationVersionCreateOptions{})
		assert.EqualError(t, err, "workspace is required")
	})

	t.Run("with invalid workspace id", func(t *testing.T) {
		cv, err := client.ConfigurationVersions.Create(
			ctx,
			ConfigurationVersionCreateOptions{Workspace: &Workspace{ID: badIdentifier}},
		)
		assert.Nil(t, cv)
		assert.EqualError(t, err, "invalid value for workspace ID")
	})
}

func TestConfigurationVersionsRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	cvTest, cvCleanup := createConfigurationVersion(t, client, nil)
	defer cvCleanup()

	t.Run("when the configuration version exists", func(t *testing.T) {
		cv, err := client.ConfigurationVersions.Read(ctx, cvTest.ID)
		require.NoError(t, err)
		assert.Equal(t, cvTest, cv)
	})

	t.Run("when the configuration version does not exist", func(t *testing.T) {
		var cvName = "nonexisting"
		cv, err := client.ConfigurationVersions.Read(ctx, cvName)
		assert.Nil(t, cv)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("ConfigurationVersion with ID '%s' not found or user unauthorized.", cvName),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("with invalid configuration version id", func(t *testing.T) {
		cv, err := client.ConfigurationVersions.Read(ctx, badIdentifier)
		assert.Nil(t, cv)
		assert.EqualError(t, err, "invalid value for configuration version ID")
	})
}
