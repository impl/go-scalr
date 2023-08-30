package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunsRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	runTest, runTestCleanup := createRun(t, client, nil, nil)
	defer runTestCleanup()

	t.Run("when the run exists", func(t *testing.T) {
		_, err := client.Runs.Read(ctx, runTest.ID)
		assert.NoError(t, err)
	})

	t.Run("when the run does not exist", func(t *testing.T) {
		var runId = "nonexisting"
		r, err := client.Runs.Read(ctx, runId)
		assert.Nil(t, r)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Run with ID '%s' not found or user unauthorized.", runId),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("with invalid run ID", func(t *testing.T) {
		r, err := client.Runs.Read(ctx, badIdentifier)
		assert.Nil(t, r)
		assert.EqualError(t, err, "invalid value for run ID")
	})
}

func TestRunsCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	wsTest, wsTestCleanup := createWorkspace(t, client, nil)
	defer wsTestCleanup()

	cvTest, _ := createConfigurationVersion(t, client, wsTest)

	t.Run("without a configuration version", func(t *testing.T) {
		options := RunCreateOptions{
			Workspace: wsTest,
		}

		_, err := client.Runs.Create(ctx, options)
		assert.EqualError(t, err, "configuration-version is required")
	})

	t.Run("with invalid configuration-version ID", func(t *testing.T) {
		options := RunCreateOptions{
			ConfigurationVersion: &ConfigurationVersion{ID: badIdentifier},
			Workspace:            wsTest,
		}

		r, err := client.Runs.Create(ctx, options)
		assert.Nil(t, r)
		assert.EqualError(t, err, "invalid value for configuration-version ID")
	})

	t.Run("without a workspace", func(t *testing.T) {
		r, err := client.Runs.Create(ctx, RunCreateOptions{})
		assert.Nil(t, r)
		assert.EqualError(t, err, "workspace is required")
	})

	t.Run("with invalid workspace ID", func(t *testing.T) {
		options := RunCreateOptions{
			ConfigurationVersion: cvTest,
			Workspace:            &Workspace{ID: badIdentifier},
		}

		r, err := client.Runs.Create(ctx, options)
		assert.Nil(t, r)
		assert.EqualError(t, err, "invalid value for workspace ID")
	})

	t.Run("with valid options", func(t *testing.T) {
		options := RunCreateOptions{
			ConfigurationVersion: cvTest,
			Workspace:            wsTest,
		}

		r, err := client.Runs.Create(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, cvTest.ID, r.ConfigurationVersion.ID)
	})
}
