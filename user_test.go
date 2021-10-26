package scalr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultUserID = "user-suh84u6vuvidtbg"
const defaultUserLdapID = "user-suh84u72qsmbuvg"

func TestUsersList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("with empty options", func(t *testing.T) {
		ul, err := client.Users.List(ctx, UserListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 1, ul.CurrentPage)
		assert.True(t, ul.TotalCount >= 1)
	})

	t.Run("with page size option", func(t *testing.T) {
		ul, err := client.Users.List(ctx, UserListOptions{
			ListOptions: ListOptions{
				PageSize: 99,
			},
		})
		require.NoError(t, err)

		uIDs := make([]string, len(ul.Items))
		for _, u := range ul.Items {
			uIDs = append(uIDs, u.ID)
		}
		assert.Equal(t, 1, ul.CurrentPage)
		assert.Contains(t, uIDs, defaultUserID)
	})

	t.Run("with user filter", func(t *testing.T) {
		ul, err := client.Users.List(ctx, UserListOptions{
			User: String(defaultUserID),
		})
		require.NoError(t, err)
		assert.Equal(t, 1, ul.CurrentPage)
		assert.Equal(t, 1, ul.TotalCount)
		assert.Equal(t, defaultUserID, ul.Items[0].ID)
	})

	t.Run("with identity provider filter", func(t *testing.T) {
		ul, err := client.Users.List(ctx, UserListOptions{
			IdentityProvider: String(defaultIdentityProviderLdapID),
		})
		require.NoError(t, err)

		uIDs := make([]string, len(ul.Items))
		// Set of IDP IDs
		idpIDs := make(map[string]bool)
		for _, u := range ul.Items {
			uIDs = append(uIDs, u.ID)
			for _, idp := range u.IdentityProviders {
				if !idpIDs[idp.ID] {
					idpIDs[idp.ID] = true
				}
			}
		}

		assert.Equal(t, 1, ul.CurrentPage)
		assert.True(t, ul.TotalCount >= 1)
		assert.Contains(t, uIDs, defaultUserLdapID)
		assert.Equal(t, 1, len(idpIDs))
		assert.Contains(t, idpIDs, defaultIdentityProviderLdapID)
	})

	t.Run("without a valid user", func(t *testing.T) {
		ul, err := client.Users.List(ctx, UserListOptions{
			User: String(badIdentifier),
		})
		assert.NoError(t, err)
		assert.Len(t, ul.Items, 0)
	})

	t.Run("without a valid identity provider", func(t *testing.T) {
		ul, err := client.Users.List(ctx, UserListOptions{
			IdentityProvider: String(badIdentifier),
		})
		assert.NoError(t, err)
		assert.Len(t, ul.Items, 0)
	})
}
