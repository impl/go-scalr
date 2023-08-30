package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentPoolsList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	agentPoolTest1, agentPoolTest1Cleanup := createAgentPool(t, client, false)
	defer agentPoolTest1Cleanup()
	agentPoolTest2, agentPoolTest2Cleanup := createAgentPool(t, client, true)
	defer agentPoolTest2Cleanup()
	agentPoolTest3, agentPoolTest3Cleanup := createAgentPool(t, client, true)
	defer agentPoolTest3Cleanup()

	t.Run("without options", func(t *testing.T) {
		apList, err := client.AgentPools.List(ctx, AgentPoolListOptions{})
		require.NoError(t, err)
		apListIDs := make([]string, len(apList.Items))
		for _, agentPool := range apList.Items {
			apListIDs = append(apListIDs, agentPool.ID)
		}
		assert.Contains(t, apListIDs, agentPoolTest1.ID)
		assert.Contains(t, apListIDs, agentPoolTest2.ID)
		assert.Contains(t, apListIDs, agentPoolTest3.ID)
	})
	t.Run("with account filter", func(t *testing.T) {
		apList, err := client.AgentPools.List(ctx, AgentPoolListOptions{Account: String(defaultAccountID)})
		require.NoError(t, err)
		apListIDs := make([]string, len(apList.Items))
		for _, agentPool := range apList.Items {
			apListIDs = append(apListIDs, agentPool.ID)
		}
		assert.Contains(t, apListIDs, agentPoolTest1.ID)
		assert.Contains(t, apListIDs, agentPoolTest2.ID)
		assert.Contains(t, apListIDs, agentPoolTest3.ID)
	})
	t.Run("with account and name filter", func(t *testing.T) {
		apList, err := client.AgentPools.List(ctx, AgentPoolListOptions{Account: String(defaultAccountID), Name: agentPoolTest1.Name})
		require.NoError(t, err)
		assert.Len(t, apList.Items, 1)
		assert.Equal(t, apList.Items[0].ID, agentPoolTest1.ID)
	})
	t.Run("with id filter", func(t *testing.T) {
		apList, err := client.AgentPools.List(ctx, AgentPoolListOptions{AgentPool: agentPoolTest2.ID})
		require.NoError(t, err)
		assert.Len(t, apList.Items, 1)
		assert.Equal(t, apList.Items[0].ID, agentPoolTest2.ID)
	})
	t.Run("with vcs-enabled filter", func(t *testing.T) {
		var vcsEnabledFiler = true
		apList, err := client.AgentPools.List(ctx, AgentPoolListOptions{VcsEnabled: &vcsEnabledFiler})
		require.NoError(t, err)
		apListIDs := make([]string, 0)
		for _, agentPool := range apList.Items {
			apListIDs = append(apListIDs, agentPool.ID)
		}
		assert.Len(t, apListIDs, 2)
		assert.Contains(t, apListIDs, agentPoolTest2.ID)
		assert.Contains(t, apListIDs, agentPoolTest3.ID)
	})
}

func TestAgentPoolsCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("when account and name are provided", func(t *testing.T) {
		options := AgentPoolCreateOptions{
			Account:    &Account{ID: defaultAccountID},
			Name:       String("test-provider-pool-" + randomString(t)),
			VcsEnabled: Bool(true),
		}

		agentPool, err := client.AgentPools.Create(ctx, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.AgentPools.Read(ctx, agentPool.ID)
		require.NoError(t, err)

		for _, item := range []*AgentPool{
			agentPool,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, options.Account, item.Account)
			assert.Equal(t, *options.VcsEnabled, item.VcsEnabled)
		}
		err = client.AgentPools.Delete(ctx, agentPool.ID)
		require.NoError(t, err)
	})

	t.Run("when create without vcs_enabled", func(t *testing.T) {
		options := AgentPoolCreateOptions{
			Account: &Account{ID: defaultAccountID},
			Name:    String("test-provider-pool-" + randomString(t)),
		}

		agentPool, err := client.AgentPools.Create(ctx, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.AgentPools.Read(ctx, agentPool.ID)
		require.NoError(t, err)

		for _, item := range []*AgentPool{
			agentPool,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, item.VcsEnabled, false)
		}
		err = client.AgentPools.Delete(ctx, agentPool.ID)
		require.NoError(t, err)
	})

	t.Run("when environment is provided", func(t *testing.T) {
		client := testClient(t)
		env, envCleanup := createEnvironment(t, client)
		defer envCleanup()

		options := AgentPoolCreateOptions{
			Account:     &Account{ID: defaultAccountID},
			Environment: &Environment{ID: env.ID},
			Name:        String("test-provider-pool-" + randomString(t)),
			VcsEnabled:  Bool(false),
		}

		agentPool, err := client.AgentPools.Create(ctx, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.AgentPools.Read(ctx, agentPool.ID)
		require.NoError(t, err)

		for _, item := range []*AgentPool{
			agentPool,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, options.Account.ID, item.Account.ID)
			assert.Equal(t, options.Environment.ID, item.Environment.ID)
		}
		err = client.AgentPools.Delete(ctx, agentPool.ID)
		require.NoError(t, err)
	})

	t.Run("when workspace is provided", func(t *testing.T) {
		client := testClient(t)
		env, envCleanup := createEnvironment(t, client)
		defer envCleanup()
		ws, wsCleanup := createWorkspace(t, client, env)

		options := AgentPoolCreateOptions{
			Account:     &Account{ID: defaultAccountID},
			Environment: &Environment{ID: env.ID},
			Workspaces:  []*Workspace{{ID: ws.ID}},
			Name:        String("test-provider-pool-" + randomString(t)),
			VcsEnabled:  Bool(false),
		}

		agentPool, err := client.AgentPools.Create(ctx, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.AgentPools.Read(ctx, agentPool.ID)
		require.NoError(t, err)

		for _, item := range []*AgentPool{
			agentPool,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, options.Account.ID, item.Account.ID)
			assert.Equal(t, options.Environment.ID, item.Environment.ID)
			assert.Equal(t, options.Workspaces[0].ID, item.Workspaces[0].ID)
		}
		wsCleanup()
		err = client.AgentPools.Delete(ctx, agentPool.ID)
		require.NoError(t, err)
	})

	t.Run("when options has name missing", func(t *testing.T) {
		r, err := client.AgentPools.Create(ctx, AgentPoolCreateOptions{
			Account: &Account{ID: defaultAccountID},
		})
		assert.Nil(t, r)
		assert.EqualError(t, err, "name is required")
	})

	t.Run("when options has an empty name", func(t *testing.T) {
		ap, err := client.AgentPools.Create(ctx, AgentPoolCreateOptions{
			Name:    String("  "),
			Account: &Account{ID: defaultAccountID},
		})
		assert.Nil(t, ap)
		assert.EqualError(t, err, "invalid value for agent pool name: '  '")
	})

	t.Run("when options has an invalid account", func(t *testing.T) {
		var accountId = "acc-234"
		_, err := client.AgentPools.Create(ctx, AgentPoolCreateOptions{
			Name:    String("test-provider-pool-" + randomString(t)),
			Account: &Account{ID: accountId},
		})
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Invalid Relationship\n\nAccount with ID '%s' not found or user unauthorized.", accountId),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("when options has an nonexistent environment", func(t *testing.T) {
		envID := "env-1234"
		_, err := client.AgentPools.Create(ctx, AgentPoolCreateOptions{
			Name:        String("test-provider-pool-" + randomString(t)),
			Account:     &Account{ID: defaultAccountID},
			Environment: &Environment{ID: envID},
		})
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Invalid Relationship\n\nEnvironment with ID '%s' not found or user unauthorized.", envID),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("when options has invalid environment", func(t *testing.T) {
		envID := badIdentifier
		ap, err := client.AgentPools.Create(ctx, AgentPoolCreateOptions{
			Name:        String("test-provider-pool-" + randomString(t)),
			Account:     &Account{ID: defaultAccountID},
			Environment: &Environment{ID: envID},
		})
		assert.Nil(t, ap)
		assert.EqualError(t, err, fmt.Sprintf("invalid value for environment ID: '%s'", envID))

	})

	t.Run("when options has invalid workpace", func(t *testing.T) {
		wsID := badIdentifier
		ap, err := client.AgentPools.Create(ctx, AgentPoolCreateOptions{
			Name:       String("test-provider-pool-" + randomString(t)),
			Account:    &Account{ID: defaultAccountID},
			Workspaces: []*Workspace{{ID: wsID}},
		})
		assert.Nil(t, ap)
		assert.EqualError(t, err, fmt.Sprintf("0: invalid value for workspace ID: '%s'", wsID))

	})

	t.Run("when options has nonexistent workpace", func(t *testing.T) {
		wsID := "ws-323"
		ap, err := client.AgentPools.Create(ctx, AgentPoolCreateOptions{
			Name:       String("test-provider-pool-" + randomString(t)),
			Account:    &Account{ID: defaultAccountID},
			Workspaces: []*Workspace{{ID: wsID}},
		})
		assert.Nil(t, ap)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Invalid Relationship\n\nRelationship 'workspaces' with ID '%s' not found or user unauthorized.", wsID),
			}.Error(),
			err.Error(),
		)
	})
}

func TestAgentPoolsRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	agentPoolTest, agentPoolTestCleanup := createAgentPool(t, client, false)
	defer agentPoolTestCleanup()

	t.Run("when the agentPool exists", func(t *testing.T) {
		agentPool, err := client.AgentPools.Read(ctx, agentPoolTest.ID)
		require.NoError(t, err)
		assert.Equal(t, agentPoolTest.ID, agentPool.ID)

		t.Run("relationships are properly decoded", func(t *testing.T) {
			assert.Equal(t, agentPool.Account.ID, agentPoolTest.Account.ID)
		})
	})

	t.Run("when the agentPool does not exist", func(t *testing.T) {
		apID := "ap-123"
		agentPool, err := client.AgentPools.Read(ctx, apID)
		assert.Nil(t, agentPool)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("AgentPool with ID '%s' not found or user unauthorized.", apID),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("with invalid agentPool ID", func(t *testing.T) {
		agentPool, err := client.AgentPools.Read(ctx, badIdentifier)
		assert.Nil(t, agentPool)
		assert.EqualError(t, err, fmt.Sprintf("invalid value for agent pool ID: '%s'", badIdentifier))
	})
}

func TestAgentPoolsUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	agentPoolTest, agentPoolTestCleanup := createAgentPool(t, client, false)
	defer agentPoolTestCleanup()

	t.Run("when updating a name", func(t *testing.T) {
		newName := "updated"
		options := AgentPoolUpdateOptions{
			Name: String(newName),
		}

		agentPoolAfter, err := client.AgentPools.Update(ctx, agentPoolTest.ID, options)
		require.NoError(t, err)

		assert.Equal(t, *options.Name, agentPoolAfter.Name)
	})

	t.Run("when updating the workspaces", func(t *testing.T) {
		client := testClient(t)
		env, envCleanup := createEnvironment(t, client)
		defer envCleanup()
		ws1, ws1Cleanup := createWorkspace(t, client, env)
		defer ws1Cleanup()

		ws2, ws2Cleanup := createWorkspace(t, client, env)
		defer ws2Cleanup()

		options := AgentPoolUpdateOptions{
			Workspaces: []*Workspace{{ID: ws1.ID}, {ID: ws2.ID}},
		}

		ap, err := client.AgentPools.Update(ctx, agentPoolTest.ID, options)
		require.NoError(t, err)

		// Get a refreshed view of the agentPool from the API
		refreshed, err := client.AgentPools.Read(ctx, agentPoolTest.ID)
		require.NoError(t, err)
		wsIds := []string{ws1.ID, ws2.ID}

		for _, item := range []*AgentPool{
			ap,
			refreshed,
		} {
			assert.Contains(t, wsIds, item.Workspaces[0].ID)
			assert.Contains(t, wsIds, item.Workspaces[1].ID)
		}
	})

	t.Run("when an error is returned from the api", func(t *testing.T) {
		r, err := client.AgentPools.Update(ctx, agentPoolTest.ID, AgentPoolUpdateOptions{
			Workspaces: []*Workspace{{ID: "ws-asdf"}},
		})
		assert.Nil(t, r)
		assert.Error(t, err)
	})
}

func TestAgentPoolsDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	pool, _ := createAgentPool(t, client, false)

	t.Run("with valid agent pool id", func(t *testing.T) {
		err := client.AgentPools.Delete(ctx, pool.ID)
		require.NoError(t, err)

		// Try loading the agentPool - it should fail.
		_, err = client.AgentPools.Read(ctx, pool.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("AgentPool with ID '%s' not found or user unauthorized.", pool.ID),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("without a valid agent pool ID", func(t *testing.T) {
		err := client.AgentPools.Delete(ctx, badIdentifier)
		assert.EqualError(t, err, "invalid value for agent pool ID")
	})
}
