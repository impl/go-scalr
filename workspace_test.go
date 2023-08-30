package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspacesList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	wsTest1, wsTest1Cleanup := createWorkspace(t, client, envTest)
	defer wsTest1Cleanup()
	wsTest2, wsTest2Cleanup := createWorkspace(t, client, envTest)
	defer wsTest2Cleanup()

	t.Run("without list options", func(t *testing.T) {
		wsl, err := client.Workspaces.List(ctx,
			WorkspaceListOptions{
				Filter: &WorkspaceFilter{
					Environment: &envTest.ID,
				},
			},
		)
		require.NoError(t, err)
		wslIDs := make([]string, len(wsl.Items))
		for _, ws := range wsl.Items {
			wslIDs = append(wslIDs, ws.ID)
		}
		assert.Contains(t, wslIDs, wsTest1.ID)
		assert.Contains(t, wslIDs, wsTest2.ID)
		assert.Equal(t, 1, wsl.CurrentPage)
		assert.Equal(t, 2, wsl.TotalCount)
	})

	t.Run("with ID in list options", func(t *testing.T) {
		wsl, err := client.Workspaces.List(ctx, WorkspaceListOptions{
			Filter: &WorkspaceFilter{
				Environment: &envTest.ID,
				Id:          &wsTest1.ID,
			},
		})
		require.NoError(t, err)
		assert.Equal(t, 1, wsl.TotalCount)
		assert.Equal(t, wsTest1.ID, wsl.Items[0].ID)
	})

	t.Run("with list options", func(t *testing.T) {
		// Request a page number which is out of range. The result should
		// be successful, but return no results if the paging options are
		// properly passed along.
		wl, err := client.Workspaces.List(ctx, WorkspaceListOptions{
			ListOptions: ListOptions{
				PageNumber: 999,
				PageSize:   100,
			},
			Filter: &WorkspaceFilter{
				Environment: &envTest.ID,
			},
		})
		require.NoError(t, err)
		assert.Empty(t, wl.Items)
		assert.Equal(t, 999, wl.CurrentPage)
		assert.Equal(t, 2, wl.TotalCount)
	})
	t.Run("without a valid environment", func(t *testing.T) {
		wl, err := client.Workspaces.List(ctx,
			WorkspaceListOptions{
				Filter: &WorkspaceFilter{
					Environment: String(badIdentifier),
				},
			},
		)
		assert.Len(t, wl.Items, 0)
		assert.NoError(t, err)
	})
}

func TestWorkspacesCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	pool, poolCleanup := createAgentPool(t, client, false)
	defer poolCleanup()

	t.Run("with valid options", func(t *testing.T) {
		options := WorkspaceCreateOptions{
			Environment:               envTest,
			Name:                      String(randomString(t)),
			AutoApply:                 Bool(true),
			ForceLatestRun:            Bool(true),
			DeletionProtectionEnabled: Bool(false),
			ExecutionMode:             WorkspaceExecutionModePtr(WorkspaceExecutionModeRemote),
			TerraformVersion:          String("1.1.9"),
			WorkingDirectory:          String("bar/"),
			RunOperationTimeout:       Int(15),
			AutoQueueRuns:             AutoQueueRunsModePtr(AutoQueueRunsModeNever),
		}

		ws, err := client.Workspaces.Create(ctx, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.Workspaces.ReadByID(ctx, ws.ID)
		require.NoError(t, err)

		for _, item := range []*Workspace{
			ws,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, *options.AutoApply, item.AutoApply)
			assert.Equal(t, *options.ForceLatestRun, item.ForceLatestRun)
			assert.Equal(t, *options.DeletionProtectionEnabled, item.DeletionProtectionEnabled)
			assert.Equal(t, false, item.HasResources)
			assert.Equal(t, *options.ExecutionMode, item.ExecutionMode)
			assert.Equal(t, *options.TerraformVersion, item.TerraformVersion)
			assert.Equal(t, *options.WorkingDirectory, item.WorkingDirectory)
			assert.Equal(t, options.RunOperationTimeout, item.RunOperationTimeout)
			assert.Equal(t, *options.AutoQueueRuns, item.AutoQueueRuns)
		}
	})

	t.Run("with agent pool", func(t *testing.T) {
		options := WorkspaceCreateOptions{
			Environment:      envTest,
			AgentPool:        pool,
			Name:             String(randomString(t)),
			AutoApply:        Bool(true),
			ExecutionMode:    WorkspaceExecutionModePtr(WorkspaceExecutionModeRemote),
			TerraformVersion: String("1.1.9"),
			WorkingDirectory: String("bar/"),
		}

		ws, err := client.Workspaces.Create(ctx, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.Workspaces.ReadByID(ctx, ws.ID)
		require.NoError(t, err)

		assert.NotEmpty(t, ws.ID)
		assert.NotEmpty(t, ws.AgentPool)
		if refreshed.AgentPool != nil {
			assert.Equal(t, pool.ID, refreshed.AgentPool.ID)
		}
		assert.Equal(t, pool.ID, ws.AgentPool.ID)
		defer client.Workspaces.Delete(ctx, ws.ID)
	})

	t.Run("when options is missing name", func(t *testing.T) {
		w, err := client.Workspaces.Create(ctx, WorkspaceCreateOptions{Environment: envTest})
		assert.Nil(t, w)
		assert.EqualError(t, err, "name is required")
	})

	t.Run("when options has an invalid name", func(t *testing.T) {
		w, err := client.Workspaces.Create(ctx, WorkspaceCreateOptions{
			Name:        String(badIdentifier),
			Environment: envTest,
		})
		assert.Nil(t, w)
		assert.EqualError(t, err, "invalid value for name")
	})

	t.Run("when options has an invalid environment", func(t *testing.T) {
		_, err := client.Workspaces.Create(ctx, WorkspaceCreateOptions{
			Name:        String("foo"),
			Environment: &Environment{ID: badIdentifier},
		})
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Invalid Relationship\n\nEnvironment with ID '%s' not found or user unauthorized.", badIdentifier),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("when an error is returned from the api", func(t *testing.T) {
		ws, err := client.Workspaces.Create(ctx, WorkspaceCreateOptions{
			Name:             String("bar"),
			TerraformVersion: String("nonexisting"),
			Environment:      envTest,
		})
		assert.Nil(t, ws)
		assert.Error(t, err)
	})
}

func TestWorkspacesRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	wsTest, wsTestCleanup := createWorkspace(t, client, envTest)
	defer wsTestCleanup()

	t.Run("when the workspace exists", func(t *testing.T) {
		ws, err := client.Workspaces.Read(ctx, envTest.ID, wsTest.Name)
		require.NoError(t, err)
		assert.Equal(t, wsTest.ID, ws.ID)

		t.Run("relationships are properly decoded", func(t *testing.T) {
			assert.Equal(t, envTest.ID, ws.Environment.ID)
		})

		t.Run("timestamps are properly decoded", func(t *testing.T) {
			assert.NotEmpty(t, ws.CreatedAt)
		})
	})

	t.Run("when the workspace does not exist", func(t *testing.T) {
		_, err := client.Workspaces.Read(ctx, envTest.ID, "nonexisting")
		assert.Error(t, err)
	})

	t.Run("when the environment does not exist", func(t *testing.T) {
		_, err := client.Workspaces.Read(ctx, "nonexisting", "nonexisting")
		assert.Error(t, err)
	})

	t.Run("without a valid environment", func(t *testing.T) {
		_, err := client.Workspaces.Read(ctx, badIdentifier, wsTest.Name)
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value for environment")
	})

	t.Run("without a valid workspace", func(t *testing.T) {
		ws, err := client.Workspaces.Read(ctx, envTest.Name, badIdentifier)
		assert.Nil(t, ws)
		assert.EqualError(t, err, "invalid value for workspace")
	})
}

func TestWorkspacesReadByID(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	wsTest, wsTestCleanup := createWorkspace(t, client, envTest)
	defer wsTestCleanup()

	t.Run("when the workspace exists", func(t *testing.T) {
		ws, err := client.Workspaces.ReadByID(ctx, wsTest.ID)
		require.NoError(t, err)
		assert.Equal(t, wsTest.ID, ws.ID)

		t.Run("relationships are properly decoded", func(t *testing.T) {
			assert.Equal(t, envTest.ID, ws.Environment.ID)
		})

		t.Run("timestamps are properly decoded", func(t *testing.T) {
			assert.NotEmpty(t, ws.CreatedAt)
		})
	})

	t.Run("when the workspace does not exist", func(t *testing.T) {
		ws, err := client.Workspaces.ReadByID(ctx, "nonexisting")
		assert.Nil(t, ws)
		assert.Error(t, err)
	})

	t.Run("without a valid workspace ID", func(t *testing.T) {
		ws, err := client.Workspaces.ReadByID(ctx, badIdentifier)
		assert.Nil(t, ws)
		assert.EqualError(t, err, "invalid value for workspace ID")
	})
}

func TestWorkspacesUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	pool, poolCleanup := createAgentPool(t, client, false)
	defer poolCleanup()

	wsTest, _ := createWorkspace(t, client, envTest)

	t.Run("when updating a subset of values", func(t *testing.T) {
		options := WorkspaceUpdateOptions{
			Name:                      String(wsTest.Name),
			AutoApply:                 Bool(true),
			ForceLatestRun:            Bool(true),
			DeletionProtectionEnabled: Bool(false),
			ExecutionMode:             WorkspaceExecutionModePtr(WorkspaceExecutionModeRemote),
			TerraformVersion:          String("1.2.9"),
			RunOperationTimeout:       Int(20),
			AutoQueueRuns:             AutoQueueRunsModePtr(AutoQueueRunsModeAlways),
		}

		wsAfter, err := client.Workspaces.Update(ctx, wsTest.ID, options)
		require.NoError(t, err)

		assert.Equal(t, wsTest.Name, wsAfter.Name)
		assert.Equal(t, AutoQueueRunsModeSkipFirst, wsTest.AutoQueueRuns)
		assert.Equal(t, *options.AutoQueueRuns, wsAfter.AutoQueueRuns)
		assert.NotEqual(t, wsTest.AutoApply, wsAfter.AutoApply)
		assert.NotEqual(t, wsTest.ForceLatestRun, wsAfter.ForceLatestRun)
		assert.NotEqual(t, wsTest.DeletionProtectionEnabled, wsAfter.DeletionProtectionEnabled)
		assert.NotEqual(t, wsTest.TerraformVersion, wsAfter.TerraformVersion)
		assert.Equal(t, wsTest.WorkingDirectory, wsAfter.WorkingDirectory)
		assert.Equal(t, int(20), *wsAfter.RunOperationTimeout)
	})

	t.Run("when attaching/detaching an agent pool", func(t *testing.T) {

		options := WorkspaceUpdateOptions{
			AgentPool: pool,
		}

		wsAfter, err := client.Workspaces.Update(ctx, wsTest.ID, options)
		require.NoError(t, err)

		assert.Equal(t, pool.ID, wsAfter.AgentPool.ID)

		options = WorkspaceUpdateOptions{
			AgentPool: nil,
		}

		wsAfter, err = client.Workspaces.Update(ctx, wsTest.ID, options)
		require.NoError(t, err)

		assert.Nil(t, wsAfter.AgentPool)
	})

	t.Run("with valid options", func(t *testing.T) {
		options := WorkspaceUpdateOptions{
			Name:                      String(randomString(t)),
			AutoApply:                 Bool(false),
			ForceLatestRun:            Bool(false),
			DeletionProtectionEnabled: Bool(false),
			ExecutionMode:             WorkspaceExecutionModePtr(WorkspaceExecutionModeLocal),
			TerraformVersion:          String("1.1.9"),
			WorkingDirectory:          String("baz/"),
		}

		w, err := client.Workspaces.Update(ctx, wsTest.ID, options)
		require.NoError(t, err)

		// Get a refreshed view of the workspace from the API
		refreshed, err := client.Workspaces.Read(ctx, envTest.ID, *options.Name)
		require.NoError(t, err)

		for _, item := range []*Workspace{
			w,
			refreshed,
		} {
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, *options.AutoApply, item.AutoApply)
			assert.Equal(t, *options.ForceLatestRun, item.ForceLatestRun)
			assert.Equal(t, *options.DeletionProtectionEnabled, item.DeletionProtectionEnabled)
			assert.Equal(t, *options.ExecutionMode, item.ExecutionMode)
			assert.Equal(t, *options.TerraformVersion, item.TerraformVersion)
			assert.Equal(t, *options.WorkingDirectory, item.WorkingDirectory)
		}
	})

	t.Run("when an error is returned from the api", func(t *testing.T) {
		w, err := client.Workspaces.Update(ctx, wsTest.ID, WorkspaceUpdateOptions{
			TerraformVersion: String("nonexisting"),
		})
		assert.Nil(t, w)
		assert.Error(t, err)
	})

	t.Run("when options has an invalid name", func(t *testing.T) {
		w, err := client.Workspaces.Update(ctx, badIdentifier, WorkspaceUpdateOptions{})
		assert.Nil(t, w)
		assert.EqualError(t, err, "invalid value for workspace ID")
	})

}

func TestWorkspacesUpdateByID(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	wTest, _ := createWorkspace(t, client, envTest)

	t.Run("when updating a subset of values", func(t *testing.T) {
		options := WorkspaceUpdateOptions{
			Name:             String(wTest.Name),
			AutoApply:        Bool(true),
			ForceLatestRun:   Bool(true),
			ExecutionMode:    WorkspaceExecutionModePtr(WorkspaceExecutionModeRemote),
			TerraformVersion: String("1.2.9"),
		}

		wAfter, err := client.Workspaces.Update(ctx, wTest.ID, options)
		require.NoError(t, err)

		assert.Equal(t, wTest.Name, wAfter.Name)
		assert.NotEqual(t, wTest.AutoApply, wAfter.AutoApply)
		assert.NotEqual(t, wTest.ForceLatestRun, wAfter.ForceLatestRun)
		assert.NotEqual(t, wTest.TerraformVersion, wAfter.TerraformVersion)
		assert.Equal(t, wTest.WorkingDirectory, wAfter.WorkingDirectory)
	})

	t.Run("with valid options", func(t *testing.T) {
		options := WorkspaceUpdateOptions{
			Name:             String(randomString(t)),
			AutoApply:        Bool(false),
			ForceLatestRun:   Bool(false),
			ExecutionMode:    WorkspaceExecutionModePtr(WorkspaceExecutionModeLocal),
			TerraformVersion: String("1.1.9"),
			WorkingDirectory: String("baz/"),
		}

		w, err := client.Workspaces.Update(ctx, wTest.ID, options)
		require.NoError(t, err)

		// Get a refreshed view of the workspace from the API
		refreshed, err := client.Workspaces.Read(ctx, envTest.ID, *options.Name)
		require.NoError(t, err)

		for _, item := range []*Workspace{
			w,
			refreshed,
		} {
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, *options.AutoApply, item.AutoApply)
			assert.Equal(t, *options.ForceLatestRun, item.ForceLatestRun)
			assert.Equal(t, *options.ExecutionMode, item.ExecutionMode)
			assert.Equal(t, *options.TerraformVersion, item.TerraformVersion)
			assert.Equal(t, *options.WorkingDirectory, item.WorkingDirectory)
		}
	})

	t.Run("when an error is returned from the api", func(t *testing.T) {
		w, err := client.Workspaces.Update(ctx, wTest.ID, WorkspaceUpdateOptions{
			TerraformVersion: String("nonexisting"),
		})
		assert.Nil(t, w)
		assert.Error(t, err)
	})

	t.Run("without a valid workspace ID", func(t *testing.T) {
		w, err := client.Workspaces.Update(ctx, badIdentifier, WorkspaceUpdateOptions{})
		assert.Nil(t, w)
		assert.EqualError(t, err, "invalid value for workspace ID")
	})
}

func TestWorkspacesDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	wTest, _ := createWorkspace(t, client, envTest)

	t.Run("with valid options", func(t *testing.T) {
		err := client.Workspaces.Delete(ctx, wTest.ID)
		require.NoError(t, err)

		// Try loading the workspace - it should fail.
		_, err = client.Workspaces.ReadByID(ctx, wTest.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Workspace with ID '%s' not found or user unauthorized.", wTest.ID),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("without a valid workspace ID", func(t *testing.T) {
		err := client.Workspaces.Delete(ctx, badIdentifier)
		assert.EqualError(t, err, "invalid value for workspace ID")
	})
}

func TestWorkspacesSetSchedule(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	wTest, _ := createWorkspace(t, client, envTest)

	t.Run("with valid options", func(t *testing.T) {
		applySchedule := "30 3 5 3-5 2"
		destroySchedule := "31 5 5 3-5 2"
		options := WorkspaceRunScheduleOptions{
			ApplySchedule:   &applySchedule,
			DestroySchedule: &destroySchedule,
		}

		w, err := client.Workspaces.SetSchedule(ctx, wTest.ID, options)
		require.NoError(t, err)

		// Get a refreshed view of the workspace from the API
		refreshed, err := client.Workspaces.ReadByID(ctx, wTest.ID)
		require.NoError(t, err)

		for _, item := range []*Workspace{
			w,
			refreshed,
		} {
			assert.Equal(t, applySchedule, item.ApplySchedule)
			assert.Equal(t, destroySchedule, item.DestroySchedule)
		}
	})

	t.Run("when an error is returned from the api", func(t *testing.T) {
		applySchedule := "bla-bla-bla"
		w, err := client.Workspaces.SetSchedule(ctx, wTest.ID, WorkspaceRunScheduleOptions{
			ApplySchedule: &applySchedule,
		})
		assert.Nil(t, w)
		assert.Error(t, err)
	})

	t.Run("without a valid workspace ID", func(t *testing.T) {
		w, err := client.Workspaces.SetSchedule(ctx, badIdentifier, WorkspaceRunScheduleOptions{})
		assert.Nil(t, w)
		assert.EqualError(t, err, "invalid value for workspace ID")
	})
}
