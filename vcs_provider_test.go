package scalr

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVCSProvidersList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	vcsTest1, vcsTest1Cleanup := createVcsProvider(t, client, nil)
	defer vcsTest1Cleanup()

	env, envCleanup := createEnvironment(t, client)
	defer envCleanup()
	envs := []*Environment{env}
	vcsTest2, vcsTest2Cleanup := createVcsProvider(t, client, envs)
	defer vcsTest2Cleanup()

	t.Run("without list options", func(t *testing.T) {
		response, err := client.VcsProviders.List(ctx, VcsProvidersListOptions{})
		require.NoError(t, err)
		vcsIDs := make([]string, len(response.Items))
		for _, vcs := range response.Items {
			vcsIDs = append(vcsIDs, vcs.ID)
		}
		assert.Contains(t, vcsIDs, vcsTest1.ID)
		assert.Contains(t, vcsIDs, vcsTest2.ID)
		assert.Equal(t, 1, response.CurrentPage)
	})

	t.Run("with list options", func(t *testing.T) {
		response, err := client.VcsProviders.List(
			ctx,
			VcsProvidersListOptions{Query: &vcsTest2.Name, VcsType: &vcsTest2.VcsType, Environment: &env.ID},
		)
		require.NoError(t, err)
		vcsIDs := make([]string, len(response.Items))
		for _, vcs := range response.Items {
			vcsIDs = append(vcsIDs, vcs.ID)
		}
		assert.Contains(t, vcsIDs, vcsTest2.ID)
		assert.Equal(t, 1, response.CurrentPage)
		assert.Equal(t, 1, response.TotalCount)
	})

	t.Run("with invalid environment filter", func(t *testing.T) {
		response, err := client.VcsProviders.List(ctx, VcsProvidersListOptions{Environment: String(badIdentifier)})
		assert.Len(t, response.Items, 0)
		assert.NoError(t, err)
	})
}

func TestVcsProvidersCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	t.Run("with valid options", func(t *testing.T) {
		options := VcsProviderCreateOptions{
			Name:     String("vcs-" + randomString(t)),
			VcsType:  Github,
			AuthType: PersonalToken,
			Token:    os.Getenv("GITHUB_TOKEN"),

			Environments: []*Environment{envTest},
			Account:      &Account{ID: defaultAccountID},
		}

		vcs, err := client.VcsProviders.Create(ctx, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.VcsProviders.Read(ctx, vcs.ID)
		require.NoError(t, err)

		for _, item := range []*VcsProvider{
			vcs,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, options.VcsType, item.VcsType)
			assert.Equal(t, options.AuthType, item.AuthType)
			assert.Equal(t, false, item.IsShared)
		}
	})

	t.Run("shared provider", func(t *testing.T) {
		options := VcsProviderCreateOptions{
			Name:     String("vcs-" + randomString(t)),
			VcsType:  Github,
			AuthType: PersonalToken,
			Token:    os.Getenv("GITHUB_TOKEN"),
			IsShared: Bool(true),

			Account: &Account{ID: defaultAccountID},
		}

		vcs, err := client.VcsProviders.Create(ctx, options)
		require.NoError(t, err)

		assert.NotEmpty(t, vcs.ID)
		assert.Equal(t, *options.Name, vcs.Name)
		assert.Equal(t, options.VcsType, vcs.VcsType)
		assert.Equal(t, options.AuthType, vcs.AuthType)
		assert.Equal(t, *options.IsShared, vcs.IsShared)

	})

	t.Run("with agent-pool attr vcs-enabled: false", func(t *testing.T) {
		ap, apCleanup := createAgentPool(t, client, false)
		defer apCleanup()

		options := VcsProviderCreateOptions{
			Name:     String("vcs-" + randomString(t)),
			VcsType:  Github,
			AuthType: PersonalToken,
			Token:    os.Getenv("GITHUB_TOKEN"),

			Environments: []*Environment{envTest},
			Account:      &Account{ID: defaultAccountID},
			AgentPool:    ap,
		}

		_, err := client.VcsProviders.Create(ctx, options)
		require.Error(t, err)
	})

	t.Run("when options has an invalid environment", func(t *testing.T) {
		_, err := client.VcsProviders.Create(ctx, VcsProviderCreateOptions{
			Name:         String("test-vcs"),
			VcsType:      Gitlab,
			AuthType:     PersonalToken,
			Environments: []*Environment{{ID: badIdentifier}},
		})
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf(
					"Invalid Relationship\n\nRelationship 'environments' with ID '%s' not found or user unauthorized.",
					badIdentifier,
				),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("when an error is returned from the api", func(t *testing.T) {
		ws, err := client.VcsProviders.Create(ctx, VcsProviderCreateOptions{
			Name:     String("foo"),
			VcsType:  Github,
			AuthType: PersonalToken,
			Token:    "invalid-token",

			Environments: []*Environment{envTest},
			Account:      &Account{ID: defaultAccountID},
		})
		assert.Nil(t, ws)
		assert.Error(t, err)
	})
}

func TestVcsProvidersRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	vcsTest, vcsTestCleanup := createVcsProvider(t, client, []*Environment{envTest})
	defer vcsTestCleanup()

	t.Run("when the vcs provider exists", func(t *testing.T) {
		vcs, err := client.VcsProviders.Read(ctx, vcsTest.ID)
		require.NoError(t, err)
		assert.Equal(t, vcsTest.ID, vcs.ID)
		assert.Len(t, vcs.Environments, 1)

		t.Run("relationships are properly decoded", func(t *testing.T) {
			assert.Equal(t, envTest.ID, vcs.Environments[0].ID)
		})
	})

	t.Run("when the vcs provider does not exist", func(t *testing.T) {
		_, err := client.VcsProviders.Read(ctx, "nonexisting")
		assert.Error(t, err)
	})
}

func TestVcsProvidersUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	vcsTest, _ := createVcsProvider(t, client, []*Environment{envTest})

	t.Run("when updating a subset of values", func(t *testing.T) {
		options := VcsProviderUpdateOptions{
			Name:     String(randomString(t)),
			IsShared: Bool(true),
		}

		vcsAfter, err := client.VcsProviders.Update(ctx, vcsTest.ID, options)
		require.NoError(t, err)

		assert.Equal(t, vcsTest.AuthType, vcsAfter.AuthType)
		assert.Equal(t, vcsTest.VcsType, vcsAfter.VcsType)
		assert.True(t, vcsTest.IsShared != vcsAfter.IsShared)
	})

	t.Run("with valid options", func(t *testing.T) {
		options := VcsProviderUpdateOptions{
			Name: String(randomString(t)),
		}

		vcs, err := client.VcsProviders.Update(ctx, vcsTest.ID, options)
		require.NoError(t, err)

		// Get a refreshed view of the vcs provider from the API
		refreshed, err := client.VcsProviders.Read(ctx, vcsTest.ID)
		require.NoError(t, err)

		for _, item := range []*VcsProvider{
			vcs,
			refreshed,
		} {
			assert.Equal(t, *options.Name, item.Name)
		}
	})

	t.Run("when an error is returned from the api", func(t *testing.T) {
		vcs, err := client.VcsProviders.Update(ctx, vcsTest.ID, VcsProviderUpdateOptions{
			Name: String("~invalid_name!"),
		})
		assert.Nil(t, vcs)
		assert.Error(t, err)
	})

	t.Run("when options has an invalid name", func(t *testing.T) {
		vcs, err := client.VcsProviders.Update(ctx, badIdentifier, VcsProviderUpdateOptions{})
		assert.Nil(t, vcs)
		assert.EqualError(t, err, "invalid value for vcs provider ID")
	})

}

func TestVcsProvidersDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	vcsTest, _ := createVcsProvider(t, client, []*Environment{envTest})

	t.Run("with valid options", func(t *testing.T) {
		err := client.VcsProviders.Delete(ctx, vcsTest.ID)
		require.NoError(t, err)

		// Try loading the vcs provider - it should fail.
		_, err = client.VcsProviders.Read(ctx, vcsTest.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("VcsProvider with ID '%s' not found or user unauthorized.", vcsTest.ID),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("without a valid vcs provider ID", func(t *testing.T) {
		err := client.VcsProviders.Delete(ctx, badIdentifier)
		assert.EqualError(t, err, "invalid value for vcs provider ID")
	})
}
