package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunTriggersCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	env1Test, env1TestCleanup := createEnvironment(t, client)
	defer env1TestCleanup()

	env2Test, envTest2Cleanup := createEnvironment(t, client)
	defer envTest2Cleanup()

	wsEnv1Test1, wsEnv1Test1Cleanup := createWorkspace(t, client, env1Test)
	defer wsEnv1Test1Cleanup()
	wsEnv1Test2, wsEnv1Test2Cleanup := createWorkspace(t, client, env1Test)
	defer wsEnv1Test2Cleanup()
	wsEnv2Test1, wsEnv2Test1Cleanup := createWorkspace(t, client, env2Test)
	defer wsEnv2Test1Cleanup()

	t.Run("missing downstream workspace", func(t *testing.T) {
		options := RunTriggerCreateOptions{
			Upstream: &Upstream{ID: wsEnv1Test1.ID},
		}
		trigger, err := client.RunTriggers.Create(ctx, options)
		require.Error(t, err)
		assert.EqualError(t, err, "downstream ID is required")
		assert.Nil(t, trigger)
	})

	t.Run("missing upstream workspace", func(t *testing.T) {
		options := RunTriggerCreateOptions{
			Downstream: &Downstream{ID: wsEnv1Test1.ID},
		}
		trigger, err := client.RunTriggers.Create(ctx, options)
		require.Error(t, err)
		assert.EqualError(t, err, "upstream ID is required")
		assert.Nil(t, trigger)
	})

	t.Run("from different environments", func(t *testing.T) {
		options := RunTriggerCreateOptions{
			Downstream: &Downstream{ID: wsEnv1Test1.ID},
			Upstream:   &Upstream{ID: wsEnv2Test1.ID},
		}
		trigger, err := client.RunTriggers.Create(ctx, options)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "The downstream and upstream workspaces must be within the same Scalr environment")
		assert.Nil(t, trigger)
	})

	t.Run("create trigger", func(t *testing.T) {
		options := RunTriggerCreateOptions{
			Downstream: &Downstream{ID: wsEnv1Test1.ID},
			Upstream:   &Upstream{ID: wsEnv1Test2.ID},
		}
		trigger, err := client.RunTriggers.Create(ctx, options)
		require.NoError(t, err)
		assert.NotEmpty(t, trigger.ID)
		assert.NotEmpty(t, trigger.CreatedAt)
		assert.Equal(t, wsEnv1Test1.ID, trigger.Downstream.ID)
		assert.Equal(t, wsEnv1Test2.ID, trigger.Upstream.ID)
	})

}

func TestRunTriggersRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	wsTest1, wsEnv1Test1Cleanup := createWorkspace(t, client, envTest)
	defer wsEnv1Test1Cleanup()
	wsTest2, wsTest2Cleanup := createWorkspace(t, client, envTest)
	defer wsTest2Cleanup()

	options := RunTriggerCreateOptions{
		Downstream: &Downstream{ID: wsTest1.ID},
		Upstream:   &Upstream{ID: wsTest2.ID},
	}
	createdTrigger, err := client.RunTriggers.Create(ctx, options)
	require.NoError(t, err)
	assert.NotEmpty(t, createdTrigger.ID)

	t.Run("get run trigger by id", func(t *testing.T) {
		trigger, err := client.RunTriggers.Read(ctx, createdTrigger.ID)
		require.NoError(t, err)
		assert.Equal(t, createdTrigger.ID, trigger.ID)
		assert.Equal(t, createdTrigger.Upstream.ID, trigger.Upstream.ID)
		assert.Equal(t, createdTrigger.Downstream.ID, trigger.Downstream.ID)
	})

	t.Run("try to get run trigger with not valid ID", func(t *testing.T) {
		_, err := client.RunTriggers.Read(ctx, badIdentifier)
		require.Error(t, err)
		assert.EqualError(t, err, "invalid value for RunTrigger ID")
	})

}

func TestRunTriggersDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	wsTest1, wsEnv1Test1Cleanup := createWorkspace(t, client, envTest)
	defer wsEnv1Test1Cleanup()
	wsTest2, wsTest2Cleanup := createWorkspace(t, client, envTest)
	defer wsTest2Cleanup()

	options := RunTriggerCreateOptions{
		Downstream: &Downstream{ID: wsTest1.ID},
		Upstream:   &Upstream{ID: wsTest2.ID},
	}
	createdTrigger, err := client.RunTriggers.Create(ctx, options)
	require.NoError(t, err)
	assert.NotEmpty(t, createdTrigger.ID)

	t.Run("delete run trigger by id", func(t *testing.T) {
		err := client.RunTriggers.Delete(ctx, createdTrigger.ID)
		require.NoError(t, err)

		// read RunTrigger by ID should fail
		trigger, err := client.RunTriggers.Read(ctx, createdTrigger.ID)
		require.Error(t, err)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("RunTrigger with ID '%s' not found or user unauthorized.", createdTrigger.ID),
			}.Error(),
			err.Error(),
		)
		assert.Nil(t, trigger)

	})

	t.Run("try to delete run trigger with not valid ID", func(t *testing.T) {
		err := client.RunTriggers.Delete(ctx, badIdentifier)
		require.Error(t, err)
		assert.EqualError(t, err, "invalid value for RunTrigger ID")
	})

}
