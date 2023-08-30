package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var readPermissions []*Permission = []*Permission{{ID: "*:read"}}
var updatePermissions []*Permission = []*Permission{{ID: "*:update"}}

func TestRolesList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	permissions := []*Permission{{ID: "*:*"}}
	roleTest1, roleTest1Cleanup := createRole(t, client, permissions)
	defer roleTest1Cleanup()
	roleTest2, roleTest2Cleanup := createRole(t, client, permissions)
	defer roleTest2Cleanup()

	t.Run("without options", func(t *testing.T) {
		rolel, err := client.Roles.List(ctx, RoleListOptions{})
		require.NoError(t, err)
		rolelIDs := make([]string, len(rolel.Items))
		for _, role := range rolel.Items {
			rolelIDs = append(rolelIDs, role.ID)
		}
		assert.Contains(t, rolelIDs, roleTest1.ID)
		assert.Contains(t, rolelIDs, roleTest2.ID)
	})
	t.Run("with options", func(t *testing.T) {
		rolel, err := client.Roles.List(ctx, RoleListOptions{Account: String(defaultAccountID)})
		require.NoError(t, err)
		rolelIDs := make([]string, len(rolel.Items))
		for _, role := range rolel.Items {
			rolelIDs = append(rolelIDs, role.ID)
		}
		assert.Contains(t, rolelIDs, roleTest1.ID)
		assert.Contains(t, rolelIDs, roleTest2.ID)
	})
}

func TestRolesCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("with empty permissions", func(t *testing.T) {
		options := RoleCreateOptions{
			Account:     &Account{ID: defaultAccountID},
			Name:        String("foo" + randomString(t)),
			Description: String("bar"),
		}

		role, err := client.Roles.Create(ctx, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.Roles.Read(ctx, role.ID)
		require.NoError(t, err)

		for _, item := range []*Role{
			role,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, *options.Description, item.Description)
			assert.Equal(t, item.IsSystem, false)
			assert.Equal(t, options.Account, item.Account)
			assert.Equal(t, options.Permissions, item.Permissions)
		}
		err = client.Roles.Delete(ctx, role.ID)
		require.NoError(t, err)
	})

	t.Run("with permissions", func(t *testing.T) {
		options := RoleCreateOptions{
			Account:     &Account{ID: defaultAccountID},
			Permissions: readPermissions,
			Name:        String("foo" + randomString(t)),
			Description: String("bar"),
		}

		role, err := client.Roles.Create(ctx, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.Roles.Read(ctx, role.ID)
		require.NoError(t, err)

		for _, item := range []*Role{
			role,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, *options.Description, item.Description)
			assert.Equal(t, item.IsSystem, false)
			assert.Equal(t, options.Account, item.Account)
			assert.Equal(t, options.Permissions, item.Permissions)
		}
		err = client.Roles.Delete(ctx, role.ID)
		require.NoError(t, err)
	})

	t.Run("when options has name missing", func(t *testing.T) {
		r, err := client.Roles.Create(ctx, RoleCreateOptions{
			Account:     &Account{ID: defaultAccountID},
			Permissions: readPermissions,
			Description: String("bar"),
		})
		assert.Nil(t, r)
		assert.EqualError(t, err, "name is required")
	})

	t.Run("when options has an empty name", func(t *testing.T) {
		w, err := client.Roles.Create(ctx, RoleCreateOptions{
			Name:        String("  "),
			Permissions: readPermissions,
			Description: String("bar"),
			Account:     &Account{ID: defaultAccountID},
		})
		assert.Nil(t, w)
		assert.EqualError(t, err, "invalid value for name")
	})

	t.Run("when options has an invalid account", func(t *testing.T) {
		var accountId = "acc-234"
		_, err := client.Roles.Create(ctx, RoleCreateOptions{
			Name:        String("foo" + randomString(t)),
			Permissions: readPermissions,
			Description: String("bar"),
			Account:     &Account{ID: accountId},
		})
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Invalid Relationship\n\nAccount with ID '%s' not found or user unauthorized.", accountId),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("bad permissions", func(t *testing.T) {
		role, err := client.Roles.Create(ctx, RoleCreateOptions{
			Name:        String("foo"),
			Permissions: []*Permission{{ID: "something:create"}, {ID: "*:read"}},
			Description: String("bar"),
			Account:     &Account{ID: defaultAccountID},
		})
		assert.Nil(t, role)
		assert.Error(t, err)
	})
}

func TestRolesRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	roleTest, roleTestCleanup := createRole(t, client, readPermissions)
	defer roleTestCleanup()

	t.Run("when the role exists", func(t *testing.T) {
		role, err := client.Roles.Read(ctx, roleTest.ID)
		require.NoError(t, err)
		assert.Equal(t, roleTest.ID, role.ID)

		t.Run("relationships are properly decoded", func(t *testing.T) {
			assert.Equal(t, role.Account.ID, roleTest.Account.ID)
		})

		t.Run("permissions are properly decoded", func(t *testing.T) {
			assert.Equal(t, *role.Permissions[0], *roleTest.Permissions[0])
		})
	})

	t.Run("when the role does not exist", func(t *testing.T) {
		role, err := client.Roles.Read(ctx, "role-nonexisting")
		assert.Nil(t, role)
		assert.Error(t, err)
	})

	t.Run("without a valid role ID", func(t *testing.T) {
		role, err := client.Roles.Read(ctx, badIdentifier)
		assert.Nil(t, role)
		assert.EqualError(t, err, "invalid value for role ID")
	})
}

func TestRolesUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	roleTest, roleTestCleanup := createRole(t, client, readPermissions)
	defer roleTestCleanup()

	t.Run("when updating a subset of values", func(t *testing.T) {
		options := RoleUpdateOptions{
			Name:        String(roleTest.Name),
			Description: String("Updated"),
		}

		roleAfter, err := client.Roles.Update(ctx, roleTest.ID, options)
		require.NoError(t, err)

		assert.Equal(t, *options.Name, roleAfter.Name)
		assert.Equal(t, *options.Description, roleAfter.Description)
	})

	t.Run("with valid options", func(t *testing.T) {
		options := RoleUpdateOptions{
			Name:        String(randomString(t)),
			Description: String(randomString(t)),
			Permissions: updatePermissions,
		}

		r, err := client.Roles.Update(ctx, roleTest.ID, options)
		require.NoError(t, err)

		// Get a refreshed view of the role from the API
		refreshed, err := client.Roles.Read(ctx, roleTest.ID)
		require.NoError(t, err)

		for _, item := range []*Role{
			r,
			refreshed,
		} {
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, *options.Description, item.Description)
			assert.Equal(t, options.Permissions[0].ID, item.Permissions[0].ID)
		}
	})

	t.Run("when an error is returned from the api", func(t *testing.T) {
		r, err := client.Roles.Update(ctx, roleTest.ID, RoleUpdateOptions{
			Permissions: []*Permission{{ID: "non-existent:read"}},
		})
		assert.Nil(t, r)
		assert.Error(t, err)
	})
}

func TestRolesDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	rTest, _ := createRole(t, client, readPermissions)

	t.Run("with valid options", func(t *testing.T) {
		err := client.Roles.Delete(ctx, rTest.ID)
		require.NoError(t, err)

		// Try loading the role - it should fail.
		_, err = client.Roles.Read(ctx, rTest.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("IamRole with ID '%s' not found or user unauthorized.", rTest.ID),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("without a valid role ID", func(t *testing.T) {
		err := client.Roles.Delete(ctx, badIdentifier)
		assert.EqualError(t, err, "invalid value for role ID")
	})
}
