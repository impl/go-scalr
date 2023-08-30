package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentPoolTokenList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	ap, apCleanup := createAgentPool(t, client, false)
	defer apCleanup()

	apt, aptCleanup := createAgentPoolToken(t, client, ap.ID)
	defer aptCleanup()

	t.Run("with valid agent pool", func(t *testing.T) {
		tList, err := client.AgentPoolTokens.List(ctx, ap.ID, AccessTokenListOptions{})
		require.NoError(t, err)
		assert.Len(t, tList.Items, 1)
		assert.Equal(t, tList.Items[0].ID, apt.ID)
	})
	t.Run("with nonexistent agent pool", func(t *testing.T) {
		_, err := client.AgentPoolTokens.List(ctx, "ap-123", AccessTokenListOptions{})
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("AgentPool with ID '%s' not found or user unauthorized.", "ap-123"),
			}.Error(),
			err.Error(),
		)
	})
}

func TestAgentPoolTokenCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	ap, apCleanup := createAgentPool(t, client, false)
	defer apCleanup()

	t.Run("when description is provided", func(t *testing.T) {
		options := AccessTokenCreateOptions{
			Description: String("provider tests token"),
		}

		apToken, err := client.AgentPoolTokens.Create(ctx, ap.ID, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		aptList, err := client.AgentPoolTokens.List(ctx, ap.ID, AccessTokenListOptions{})
		require.NoError(t, err)

		refreshed := aptList.Items[0]

		assert.NotEmpty(t, refreshed.ID)
		assert.Equal(t, *options.Description, refreshed.Description)

		err = client.AccessTokens.Delete(ctx, apToken.ID)
		require.NoError(t, err)
	})

	t.Run("when description is not provided", func(t *testing.T) {
		options := AccessTokenCreateOptions{}
		apToken, err := client.AgentPoolTokens.Create(ctx, ap.ID, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		aptList, err := client.AgentPoolTokens.List(ctx, ap.ID, AccessTokenListOptions{})
		require.NoError(t, err)

		refreshed := aptList.Items[0]

		assert.NotEmpty(t, refreshed.ID)
		assert.Equal(t, refreshed.Description, "")

		err = client.AccessTokens.Delete(ctx, apToken.ID)
		require.NoError(t, err)
	})

	t.Run("with nonexistent pool id", func(t *testing.T) {
		var apID = "ap-234"
		_, err := client.AgentPoolTokens.Create(ctx, apID, AccessTokenCreateOptions{})
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("AgentPool with ID '%s' not found or user unauthorized.", apID),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("with invalid pool id", func(t *testing.T) {
		apID := badIdentifier
		ap, err := client.AgentPoolTokens.Create(ctx, apID, AccessTokenCreateOptions{})
		assert.Nil(t, ap)
		assert.EqualError(t, err, fmt.Sprintf("invalid value for agent pool ID: '%s'", apID))

	})

}
