package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentsList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	emptyOptions := EnvironmentListOptions{}

	envl, err := client.Environments.List(ctx, emptyOptions)
	if err != nil {
		t.Fatal(err)
	}
	totalCount := envl.TotalCount
	envTest1, envTest1Cleanup := createEnvironment(t, client)
	defer envTest1Cleanup()

	t.Run("with no list options", func(t *testing.T) {
		envl, err := client.Environments.List(ctx, emptyOptions)
		envlIDs := make([]string, len(envl.Items))
		for _, env := range envl.Items {
			envlIDs = append(envlIDs, env.ID)
		}
		require.NoError(t, err)
		assert.Contains(t, envlIDs, envTest1.ID)

		assert.Equal(t, 1, envl.CurrentPage)
		assert.Equal(t, 1+totalCount, envl.TotalCount)
	})

	includeOptions := EnvironmentListOptions{
		Include: String("created-by"),
	}
	t.Run("with include option", func(t *testing.T) {
		envl, err := client.Environments.List(ctx, includeOptions)
		require.NoError(t, err)

		for _, env := range envl.Items {
			assert.NotEqual(t, env.CreatedBy, nil)
		}
	})

	filterByNameOptions := EnvironmentListOptions{
		Filter: &EnvironmentFilter{Name: &envTest1.Name},
	}
	t.Run("with filter by name option", func(t *testing.T) {
		envl, err := client.Environments.List(ctx, filterByNameOptions)
		require.NoError(t, err)
		assert.Equal(t, 1, len(envl.Items))
		env := envl.Items[0]
		assert.Equal(t, env.ID, envTest1.ID)
	})

	filterByIdOptions := EnvironmentListOptions{
		Filter: &EnvironmentFilter{Id: &envTest1.ID},
	}
	t.Run("with filter by ID option", func(t *testing.T) {
		envl, err := client.Environments.List(ctx, filterByIdOptions)
		require.NoError(t, err)
		assert.Equal(t, 1, len(envl.Items))
		env := envl.Items[0]
		assert.Equal(t, env.ID, envTest1.ID)
	})

	filterByAccountIdOptions := EnvironmentListOptions{
		Filter: &EnvironmentFilter{Account: String(defaultAccountID)},
	}
	t.Run("with filter by account option", func(t *testing.T) {
		envl, err := client.Environments.List(ctx, filterByAccountIdOptions)

		require.NoError(t, err)
		for _, env := range envl.Items {
			assert.Equal(t, defaultAccountID, env.Account.ID)
		}
	})

}

func TestEnvironmentsCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	t.Run("when no name is provided", func(t *testing.T) {
		_, err := client.Environments.Create(ctx, EnvironmentCreateOptions{
			Account: &Account{ID: defaultAccountID},
		})
		assert.EqualError(t, err, "name is required")
	})
	t.Run("when no account is provided", func(t *testing.T) {
		_, err := client.Environments.Create(ctx, EnvironmentCreateOptions{
			Name: String(randomString(t)),
		})
		assert.EqualError(t, err, "account is required")
	})
	t.Run("with invalid accountID", func(t *testing.T) {
		env, err := client.Environments.Create(ctx, EnvironmentCreateOptions{
			Account: &Account{ID: badIdentifier},
			Name:    String(randomString(t)),
		})
		assert.Nil(t, env)
		assert.EqualError(t, err, "invalid value for account ID")
	})
	t.Run("with valid options", func(t *testing.T) {
		provider_configuration, remove_provider_configuration := createProviderConfiguration(t, client, "consul", "consul")
		defer remove_provider_configuration()

		options := EnvironmentCreateOptions{
			Name:                          String("tst-" + randomString(t)),
			Account:                       &Account{ID: defaultAccountID},
			DefaultProviderConfigurations: []*ProviderConfiguration{provider_configuration},
		}

		env, err := client.Environments.Create(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		// Get a refreshed view of the environment
		_, err = client.Environments.Read(ctx, env.ID)
		require.NoError(t, err)

		defer client.Environments.Delete(ctx, env.ID)

		assert.Equal(t, *options.Name, env.Name)
		assert.Equal(t, options.Account.ID, env.Account.ID)
		assert.Len(t, env.DefaultProviderConfigurations, 1)
		assert.Equal(t, provider_configuration.ID, env.DefaultProviderConfigurations[0].ID)
	})

}

func TestEnvironmentsRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()
	t.Run("when the env exists", func(t *testing.T) {
		_, err := client.Environments.Read(ctx, envTest.ID)
		require.NoError(t, err)
	})

	t.Run("when the env does not exist", func(t *testing.T) {
		var envId = "notexisting"
		_, err := client.Environments.Read(ctx, envId)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Environment with ID '%s' not found or user unauthorized.", envId),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("with invalid env ID", func(t *testing.T) {
		r, err := client.Environments.Read(ctx, badIdentifier)
		assert.Nil(t, r)
		assert.EqualError(t, err, "invalid value for environment ID")
	})
}

func TestEnvironmentsUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("with valid options", func(t *testing.T) {
		envTest, envTestCleanup := createEnvironment(t, client)
		defer envTestCleanup()

		options := EnvironmentUpdateOptions{
			Name:                  String("tst-" + randomString(t)),
			CostEstimationEnabled: Bool(false),
		}

		env, err := client.Environments.Update(ctx, envTest.ID, options)
		require.NoError(t, err)

		// Also get a fresh result from the API to ensure we get the
		// expected values back.
		refreshed, err := client.Environments.Read(ctx, env.ID)
		require.NoError(t, err)

		for _, item := range []*Environment{
			env,
			refreshed,
		} {
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, *options.CostEstimationEnabled, item.CostEstimationEnabled)
		}
	})

	t.Run("update default provider configuration", func(t *testing.T) {
		envTest, envTestCleanup := createEnvironment(t, client)
		defer envTestCleanup()
		providerConfiguration, removeProviderConfiguration := createProviderConfiguration(t, client, "consul", "consul")
		defer removeProviderConfiguration()

		options := EnvironmentUpdateOptions{
			DefaultProviderConfigurations: []*ProviderConfiguration{providerConfiguration},
		}

		env, err := client.Environments.Update(ctx, envTest.ID, options)
		require.NoError(t, err)

		refreshed, err := client.Environments.Read(ctx, env.ID)
		require.NoError(t, err)

		for _, item := range []*Environment{
			env,
			refreshed,
		} {
			assert.Len(t, item.DefaultProviderConfigurations, 1)
			assert.Equal(t, providerConfiguration.ID, item.DefaultProviderConfigurations[0].ID)
		}

		options = EnvironmentUpdateOptions{
			DefaultProviderConfigurations: make([]*ProviderConfiguration, 0),
		}

		env, err = client.Environments.Update(ctx, envTest.ID, options)
		require.NoError(t, err)

		refreshed, err = client.Environments.Read(ctx, env.ID)
		require.NoError(t, err)

		for _, item := range []*Environment{
			env,
			refreshed,
		} {
			assert.Len(t, item.DefaultProviderConfigurations, 0)
		}
	})

	t.Run("when only updating a subset of fields", func(t *testing.T) {
		envTest, envTestCleanup := createEnvironment(t, client)
		defer envTestCleanup()

		env, err := client.Environments.Update(ctx, envTest.ID, EnvironmentUpdateOptions{})
		require.NoError(t, err)
		assert.Equal(t, envTest.Name, env.Name)
	})
}

func TestEnvironmentsDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("with valid options", func(t *testing.T) {
		envTest, _ := createEnvironment(t, client)

		err := client.Environments.Delete(ctx, envTest.ID)
		require.NoError(t, err)

		// Try fetching the env again - it should error.
		_, err = client.Environments.Read(ctx, envTest.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Environment with ID '%s' not found or user unauthorized.", envTest.ID),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("when the env does not exist", func(t *testing.T) {
		var envId = randomString(t)
		err := client.Environments.Delete(ctx, envId)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Environment with ID '%s' not found or user unauthorized.", envId),
			}.Error(),
			err.Error(),
		)
	})
}
