package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderConfigurationLinkCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("workspace link", func(t *testing.T) {
		environment, removeEnvironment := createEnvironment(t, client)
		defer removeEnvironment()

		configuration, deleteConfiguration := createProviderConfiguration(
			t, client, "kubernetes", "kubernetes_dev",
		)
		defer deleteConfiguration()

		workspace, deleteWorkspace := createWorkspace(t, client, environment)
		defer deleteWorkspace()

		createLinkOptions := ProviderConfigurationLinkCreateOptions{
			ProviderConfiguration: configuration,
			Alias:                 String("dev"),
		}

		workspaceLink, err := client.ProviderConfigurationLinks.Create(
			ctx, workspace.ID, createLinkOptions,
		)
		require.NoError(t, err)

		workspaceLink, err = client.ProviderConfigurationLinks.Read(ctx, workspaceLink.ID)
		require.NoError(t, err)

		assert.Equal(t, *createLinkOptions.Alias, workspaceLink.Alias)
	})
}

func TestProviderConfigurationLinkUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	configuration, deleteConfiguration := createProviderConfiguration(
		t, client, "kubernetes", "kubernetes_dev",
	)
	defer deleteConfiguration()

	environment, removeEnvironment := createEnvironment(t, client)
	defer removeEnvironment()

	workspace, deleteWorkspace := createWorkspace(t, client, environment)
	defer deleteWorkspace()

	t.Run("workspace link", func(t *testing.T) {
		createOptions := ProviderConfigurationLinkCreateOptions{
			Alias: String("dev"), ProviderConfiguration: configuration,
		}

		workspacelink, err := client.ProviderConfigurationLinks.Create(
			ctx, workspace.ID, createOptions,
		)
		require.NoError(t, err)

		updateOptions := ProviderConfigurationLinkUpdateOptions{Alias: String("demo")}

		workspacelink, err = client.ProviderConfigurationLinks.Update(ctx, workspacelink.ID, updateOptions)
		require.NoError(t, err)

		assert.Equal(t, *updateOptions.Alias, workspacelink.Alias)
	})
}

func TestProviderConfigurationLinkDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	environment, removeEnvironment := createEnvironment(t, client)
	defer removeEnvironment()

	workspace, deleteWorkspace := createWorkspace(t, client, environment)
	defer deleteWorkspace()

	configuration, deleteConfiguration := createProviderConfiguration(
		t, client, "kubernetes", "kubernetes_dev",
	)
	defer deleteConfiguration()

	t.Run("success", func(t *testing.T) {
		options := ProviderConfigurationLinkCreateOptions{
			Alias: String("dev"), ProviderConfiguration: configuration,
		}

		link, err := client.ProviderConfigurationLinks.Create(
			ctx, workspace.ID, options,
		)
		require.NoError(t, err)

		err = client.ProviderConfigurationLinks.Delete(ctx, link.ID)
		require.NoError(t, err)

		// Try loading the configuration - it should fail.
		_, err = client.ProviderConfigurationLinks.Read(ctx, link.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("ProviderConfigurationLink with ID '%s' not found or user unauthorized.", link.ID),
			}.Error(),
			err.Error(),
		)
	})
}
