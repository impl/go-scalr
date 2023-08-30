package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("account exists", func(t *testing.T) {
		account, err := client.Accounts.Read(ctx, defaultAccountID)
		require.NoError(t, err)

		assert.Equal(t, defaultAccountID, account.ID)
		assert.Equal(t, defaultAccountName, account.Name)

		assert.Equal(t, []string{}, account.AllowedIPs)
	})

	t.Run("account does not exist", func(t *testing.T) {
		var accId = "notexisting"
		_, err := client.Accounts.Read(ctx, accId)
		assert.Equal(
			t,
			fmt.Sprintf("Clients with ID '%s' not found or user unauthorized.", accId),
			err.Error(),
		)
	})

	t.Run("with invalid acc ID", func(t *testing.T) {
		r, err := client.Accounts.Read(ctx, badIdentifier)
		assert.Nil(t, r)
		assert.EqualError(t, err, "invalid value for account ID")
	})
}

func TestAccountUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	defer func() {
		options := AccountUpdateOptions{
			AllowedIPs: &[]string{},
		}
		if _, err := client.Accounts.Update(ctx, defaultAccountID, options); err != nil {
			t.Errorf("Error resetting allowed ips for account! "+
				"The full error is shown below.\n\n"+
				"Account: %s\nError: %s", defaultAccountID, err)
		}
	}()

	t.Run("valid allowed ips", func(t *testing.T) {
		options := AccountUpdateOptions{
			AllowedIPs: &[]string{"0.0.0.0/0", "192.168.0.0/24"},
		}
		account, err := client.Accounts.Update(ctx, defaultAccountID, options)
		require.NoError(t, err)
		for i, ip := range account.AllowedIPs {
			assert.Equal(t, (*options.AllowedIPs)[i], ip)
		}

		account, err = client.Accounts.Read(ctx, defaultAccountID)
		require.NoError(t, err)
		for i, ip := range account.AllowedIPs {
			assert.Equal(t, (*options.AllowedIPs)[i], ip)
		}
	})

	t.Run("invalid allowed ips", func(t *testing.T) {
		options := AccountUpdateOptions{
			AllowedIPs: &[]string{"127.0.00"},
		}
		account, err := client.Accounts.Update(ctx, defaultAccountID, options)
		assert.Nil(t, account)
		assert.EqualError(t, err, "Invalid Attribute\n\nvalue is not a valid IPv4 network")
	})

	t.Run("invalid allowed ips ipv6", func(t *testing.T) {
		options := AccountUpdateOptions{
			AllowedIPs: &[]string{"FE80:CD00:0000:0CDE:1257:0000:211E:729C"},
		}
		account, err := client.Accounts.Update(ctx, defaultAccountID, options)
		assert.Nil(t, account)
		assert.EqualError(t, err, "Invalid Attribute\n\nvalue is not a valid IPv4 network")
	})

	t.Run("reset allowed ips", func(t *testing.T) {
		options := AccountUpdateOptions{
			AllowedIPs: &[]string{},
		}
		account, err := client.Accounts.Update(ctx, defaultAccountID, options)
		require.NoError(t, err)
		assert.Equal(t, []string{}, account.AllowedIPs)
	})
}
