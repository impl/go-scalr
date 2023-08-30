package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessTokenRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	ap, apCleanup := createAgentPool(t, client, false)
	defer apCleanup()

	atTest, atTestCleanup := createAgentPoolToken(t, client, ap.ID)
	defer atTestCleanup()

	t.Run("when the token exists", func(t *testing.T) {
		at, err := client.AccessTokens.Read(ctx, atTest.ID)
		require.NoError(t, err)
		assert.Equal(t, atTest.ID, at.ID)
	})

	t.Run("when the token does not exist", func(t *testing.T) {
		_, err := client.AccessTokens.Read(ctx, "at-nonexisting")
		assert.Error(t, err)
	})

	t.Run("with invalid token ID", func(t *testing.T) {
		_, err := client.AccessTokens.Read(ctx, badIdentifier)
		assert.EqualError(t, err, "invalid value for access token ID")
	})
}

func TestAccessTokenUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	ap, apCleanup := createAgentPool(t, client, false)
	defer apCleanup()

	apt, aptCleanup := createAgentPoolToken(t, client, ap.ID)
	defer aptCleanup()

	t.Run("when updating a agent pool's token description", func(t *testing.T) {
		newDescr := "updated"
		options := AccessTokenUpdateOptions{
			Description: String(newDescr),
		}

		accessTokenAfter, err := client.AccessTokens.Update(ctx, apt.ID, options)
		require.NoError(t, err)

		assert.Equal(t, *options.Description, accessTokenAfter.Description)
	})

	t.Run("when updating nonexistent access token", func(t *testing.T) {
		r, err := client.AccessTokens.Update(ctx, "at-123", AccessTokenUpdateOptions{Description: String("asdf")})
		assert.Nil(t, r)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("AccessToken with ID '%s' not found or user unauthorized.", "at-123"),
			}.Error(),
			err.Error(),
		)
	})
}

func TestAccessTokenDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	ap, apCleanup := createAgentPool(t, client, false)
	defer apCleanup()

	apt, _ := createAgentPoolToken(t, client, ap.ID)

	t.Run("with valid agent pool token id", func(t *testing.T) {
		err := client.AccessTokens.Delete(ctx, apt.ID)
		require.NoError(t, err)

		l, err := client.AgentPoolTokens.List(ctx, ap.ID, AccessTokenListOptions{})
		assert.Len(t, l.Items, 0)
	})

	t.Run("without a valid agent pool ID", func(t *testing.T) {
		err := client.AccessTokens.Delete(ctx, badIdentifier)
		assert.EqualError(t, err, fmt.Sprintf("invalid value for access token ID: '%s'", badIdentifier))
	})
}
