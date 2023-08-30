package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessPoliciesList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	roleReadTest, roleReadTestCleanup := createRole(t, client, readPermissions)
	defer roleReadTestCleanup()

	roleWriteTest, roleWriteTestCleanup := createRole(t, client, updatePermissions)
	defer roleWriteTestCleanup()

	apTest1, apTest1Cleanup := createAccessPolicy(t, client, []*Role{roleReadTest}, &User{ID: defaultUserID})
	defer apTest1Cleanup()

	apTest2, apTest2Cleanup := createAccessPolicy(t, client, []*Role{roleWriteTest}, &Team{ID: defaultTeamID})
	defer apTest2Cleanup()

	t.Run("without list options", func(t *testing.T) {
		apl, err := client.AccessPolicies.List(ctx, AccessPolicyListOptions{})
		require.NoError(t, err)
		aplIDs := make([]string, len(apl.Items))
		for _, ap := range apl.Items {
			aplIDs = append(aplIDs, ap.ID)
		}
		assert.Contains(t, aplIDs, apTest1.ID)
		assert.Contains(t, aplIDs, apTest2.ID)
	})

	t.Run("with list options", func(t *testing.T) {
		// Request a page number which is out of range. The result should
		// be successful, but return no results if the paging options are
		// properly passed along.
		wl, err := client.AccessPolicies.List(ctx, AccessPolicyListOptions{
			ListOptions: ListOptions{
				PageNumber: 999,
				PageSize:   100,
			},
			Account: String(defaultAccountID),
		})
		require.NoError(t, err)
		assert.Empty(t, wl.Items)
		assert.Equal(t, 999, wl.CurrentPage)
	})
	t.Run("without a valid account", func(t *testing.T) {
		wl, err := client.AccessPolicies.List(ctx, AccessPolicyListOptions{Account: String(badIdentifier)})
		assert.Len(t, wl.Items, 0)
		assert.NoError(t, err)
	})
}

func TestAccessPoliciesCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	envTest, envTestCleanup := createEnvironment(t, client)
	defer envTestCleanup()

	roleReadTest, roleReadTestCleanup := createRole(t, client, readPermissions)
	defer roleReadTestCleanup()

	t.Run("with valid options", func(t *testing.T) {
		options := AccessPolicyCreateOptions{
			Environment: envTest,
			Roles:       []*Role{roleReadTest},
			User:        &User{ID: defaultUserID},
		}

		ap, err := client.AccessPolicies.Create(ctx, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.AccessPolicies.Read(ctx, ap.ID)
		require.NoError(t, err)

		for _, item := range []*AccessPolicy{
			ap,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, options.Environment.ID, item.Environment.ID)
			assert.Equal(t, options.User.ID, item.User.ID)
			assert.Equal(t, options.Roles[0].ID, item.Roles[0].ID)
			assert.Equal(t, false, item.IsSystem)
		}
		client.AccessPolicies.Delete(ctx, ap.ID)
	})

	t.Run("when options is missing scope", func(t *testing.T) {
		w, err := client.AccessPolicies.Create(ctx, AccessPolicyCreateOptions{
			Roles: []*Role{roleReadTest},
			User:  &User{ID: defaultUserID},
		})
		assert.Nil(t, w)
		assert.EqualError(t, err, "one of: account,environment,workspace must be provided")
	})

	t.Run("when options is missing object", func(t *testing.T) {
		w, err := client.AccessPolicies.Create(ctx, AccessPolicyCreateOptions{
			Roles:       []*Role{roleReadTest},
			Environment: envTest,
		})
		assert.Nil(t, w)
		assert.EqualError(t, err, "one of: user,team,service_account must be provided")
	})

	t.Run("when options is missing roles", func(t *testing.T) {
		w, err := client.AccessPolicies.Create(ctx, AccessPolicyCreateOptions{
			Environment: envTest,
			User:        &User{ID: defaultUserID},
			Roles:       []*Role{},
		})
		assert.Nil(t, w)
		assert.EqualError(t, err, "at least one role must be provided")
	})

	t.Run("when options has an invalid environment", func(t *testing.T) {
		_, err := client.AccessPolicies.Create(ctx, AccessPolicyCreateOptions{
			Roles:       []*Role{roleReadTest},
			User:        &User{ID: defaultUserID},
			Environment: &Environment{ID: badIdentifier},
		})
		assert.EqualError(t, err, fmt.Sprintf("invalid value for environment ID: %v", badIdentifier))
	})

	t.Run("when an error is returned from the api", func(t *testing.T) {
		ap, err := client.AccessPolicies.Create(ctx, AccessPolicyCreateOptions{
			Roles:       []*Role{roleReadTest},
			User:        &User{ID: defaultUserID},
			Environment: &Environment{ID: "env-123"},
		})
		assert.Nil(t, ap)
		assert.Error(t, err)
	})
}

func TestAccessPoliciesRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	roleReadTest, roleReadTestCleanup := createRole(t, client, readPermissions)
	defer roleReadTestCleanup()

	apTest, apTestCleanup := createAccessPolicy(t, client, []*Role{roleReadTest}, &User{ID: defaultUserID})
	defer apTestCleanup()

	t.Run("when the accessPolicy exists", func(t *testing.T) {
		ap, err := client.AccessPolicies.Read(ctx, apTest.ID)
		require.NoError(t, err)
		assert.Equal(t, apTest.ID, ap.ID)

		t.Run("account relationships are properly decoded", func(t *testing.T) {
			assert.Equal(t, ap.Account.ID, defaultAccountID)
		})

		t.Run("user relationships are properly decoded", func(t *testing.T) {
			assert.Equal(t, ap.User.ID, defaultUserID)
		})
		t.Run("roles relationships are properly decoded", func(t *testing.T) {
			assert.Equal(t, ap.Roles[0].ID, roleReadTest.ID)
		})

	})

	t.Run("when the accessPolicy does not exist", func(t *testing.T) {
		_, err := client.AccessPolicies.Read(ctx, "role-0000000")
		assert.Error(t, err)
	})

	t.Run("with invalid accessPolicy id", func(t *testing.T) {
		ap, err := client.AccessPolicies.Read(ctx, badIdentifier)
		assert.Nil(t, ap)
		assert.EqualError(t, err, "invalid value for access policy ID")
	})
}

func TestAccessPoliciesUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	roleReadTest, roleReadTestCleanup := createRole(t, client, readPermissions)
	defer roleReadTestCleanup()

	roleWriteTest, roleWriteTestCleanup := createRole(t, client, updatePermissions)
	defer roleWriteTestCleanup()

	apTest, apTestCleanup := createAccessPolicy(t, client, []*Role{roleReadTest}, &User{ID: defaultUserID})
	defer apTestCleanup()

	t.Run("with valid options", func(t *testing.T) {
		options := AccessPolicyUpdateOptions{
			Roles: []*Role{roleWriteTest},
		}

		w, err := client.AccessPolicies.Update(ctx, apTest.ID, options)
		require.NoError(t, err)

		// Get a refreshed view of the accessPolicy from the API
		refreshed, err := client.AccessPolicies.Read(ctx, apTest.ID)
		require.NoError(t, err)

		for _, item := range []*AccessPolicy{
			w,
			refreshed,
		} {
			assert.Equal(t, len(options.Roles), len(item.Roles))
			assert.Equal(t, options.Roles[0].ID, item.Roles[0].ID)
		}
	})

	t.Run("when an error is returned from the api", func(t *testing.T) {
		w, err := client.AccessPolicies.Update(ctx, apTest.ID, AccessPolicyUpdateOptions{
			Roles: []*Role{{ID: "role-0000000"}},
		})
		assert.Nil(t, w)
		assert.Error(t, err)
	})
}

func TestAccessPoliciesDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	roleReadTest, roleReadTestCleanup := createRole(t, client, readPermissions)
	defer roleReadTestCleanup()

	apTest, _ := createAccessPolicy(t, client, []*Role{roleReadTest}, &User{ID: defaultUserID})

	t.Run("with valid options", func(t *testing.T) {
		err := client.AccessPolicies.Delete(ctx, apTest.ID)
		require.NoError(t, err)

		// Try loading the accessPolicy - it should fail.
		_, err = client.AccessPolicies.Read(ctx, apTest.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("IamAccessPolicy with ID '%v' not found or user unauthorized.", apTest.ID),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("without a valid accessPolicy ID", func(t *testing.T) {
		err := client.AccessPolicies.Delete(ctx, badIdentifier)
		assert.EqualError(t, err, "invalid value for access policy ID")
	})
}
