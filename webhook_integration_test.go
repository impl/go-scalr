package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookIntegrationsList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	env1, deleteEnv1 := createEnvironment(t, client)
	defer deleteEnv1()
	env2, deleteEnv2 := createEnvironment(t, client)
	defer deleteEnv2()

	whTest1, whTest1Cleanup := createWebhookIntegration(t, client, true, nil)
	defer whTest1Cleanup()
	_, whTest2Cleanup := createWebhookIntegration(t, client, false, []*Environment{env1})
	defer whTest2Cleanup()
	whTest3, whTest3Cleanup := createWebhookIntegration(t, client, false, []*Environment{env2})
	defer whTest3Cleanup()

	t.Run("with options", func(t *testing.T) {
		whl, err := client.WebhookIntegrations.List(
			ctx, WebhookIntegrationListOptions{
				Account:     String(defaultAccountID),
				Environment: &env2.ID,
			},
		)
		require.NoError(t, err)
		assert.Equal(t, 2, whl.TotalCount)

		expectedIDs := []string{whTest1.ID, whTest3.ID}
		actualIDs := make([]string, len(whl.Items))
		for i, wh := range whl.Items {
			actualIDs[i] = wh.ID
		}
		assert.ElementsMatch(t, expectedIDs, actualIDs)
	})
}

func TestWebhookIntegrationsRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	whTest, whTestCleanup := createWebhookIntegration(t, client, true, nil)
	defer whTestCleanup()

	t.Run("by ID when the webhook exists", func(t *testing.T) {
		wh, err := client.WebhookIntegrations.Read(ctx, whTest.ID)
		require.NoError(t, err)
		assert.Equal(t, whTest.ID, wh.ID)
	})

	t.Run("by ID when the webhook does not exist", func(t *testing.T) {
		wh, err := client.WebhookIntegrations.Read(ctx, "wh-nonexisting")
		assert.Nil(t, wh)
		assert.Error(t, err)
	})

	t.Run("by ID without a valid webhook ID", func(t *testing.T) {
		wh, err := client.WebhookIntegrations.Read(ctx, badIdentifier)
		assert.Nil(t, wh)
		assert.EqualError(t, err, "invalid value for webhook ID")
	})
}

func TestWebhookIntegrationsCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("with valid options", func(t *testing.T) {
		options := WebhookIntegrationCreateOptions{
			Name:    String("tst-" + randomString(t)),
			Account: &Account{ID: defaultAccountID},
			Url:     String("https://example.com"),
			Events:  []*EventDefinition{{ID: "run:completed"}},
		}

		wh, err := client.WebhookIntegrations.Create(ctx, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.WebhookIntegrations.Read(ctx, wh.ID)
		require.NoError(t, err)

		for _, item := range []*WebhookIntegration{
			wh,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, options.Account, item.Account)
			assert.Equal(t, *options.Url, item.Url)
			assert.Equal(t, options.Events, item.Events)
		}
		err = client.WebhookIntegrations.Delete(ctx, wh.ID)
		require.NoError(t, err)
	})
}

func TestWebhookIntegrationsUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	env1, deleteEnv1 := createEnvironment(t, client)
	defer deleteEnv1()
	env2, deleteEnv2 := createEnvironment(t, client)
	defer deleteEnv2()

	whTest, whTestCleanup := createWebhookIntegration(t, client, true, nil)
	defer whTestCleanup()

	t.Run("with valid options", func(t *testing.T) {
		options := WebhookIntegrationUpdateOptions{
			Name:         String(randomString(t)),
			IsShared:     Bool(false),
			Environments: []*Environment{env1, env2},
		}

		wh, err := client.WebhookIntegrations.Update(ctx, whTest.ID, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.WebhookIntegrations.Read(ctx, whTest.ID)
		require.NoError(t, err)

		for _, item := range []*WebhookIntegration{
			wh,
			refreshed,
		} {
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, *options.IsShared, item.IsShared)
			assert.Len(t, item.Environments, 2)
		}
	})
}

func TestWebhookIntegrationsDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	whTest, _ := createWebhookIntegration(t, client, true, nil)

	t.Run("with valid options", func(t *testing.T) {
		err := client.WebhookIntegrations.Delete(ctx, whTest.ID)
		require.NoError(t, err)

		_, err = client.WebhookIntegrations.Read(ctx, whTest.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Webhook with ID '%s' not found or user unauthorized.", whTest.ID),
			}.Error(),
			err.Error(),
		)
	})
}
