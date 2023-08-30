package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceAccountTokenList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	sa, saCleanup := createServiceAccount(
		t, client, &Account{ID: defaultAccountID}, ServiceAccountStatusPtr(ServiceAccountStatusActive),
	)
	defer saCleanup()

	at1, at1Cleanup := createServiceAccountToken(t, client, sa.ID)
	defer at1Cleanup()

	at2, at2Cleanup := createServiceAccountToken(t, client, sa.ID)
	defer at2Cleanup()

	t.Run("with valid service account", func(t *testing.T) {
		atl, err := client.ServiceAccountTokens.List(ctx, sa.ID, AccessTokenListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 2, atl.TotalCount)

		atIDs := make([]string, len(atl.Items))
		for i, at := range atl.Items {
			atIDs[i] = at.ID
		}
		assert.Contains(t, atIDs, at1.ID)
		assert.Contains(t, atIDs, at2.ID)
	})
	t.Run("with nonexistent service account", func(t *testing.T) {
		var saId = "notexisting"
		_, err := client.ServiceAccountTokens.List(ctx, saId, AccessTokenListOptions{})
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("ServiceAccount with ID '%s' not found or user unauthorized.", saId),
			}.Error(),
			err.Error(),
		)
	})
}

func TestServiceAccountTokenCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	sa, saCleanup := createServiceAccount(
		t, client, &Account{ID: defaultAccountID}, ServiceAccountStatusPtr(ServiceAccountStatusActive),
	)
	defer saCleanup()

	t.Run("when description is provided", func(t *testing.T) {
		options := AccessTokenCreateOptions{
			Description: String("tst-description-" + randomString(t)),
		}

		at, err := client.ServiceAccountTokens.Create(ctx, sa.ID, options)
		require.NoError(t, err)

		defer func() { _ = client.AccessTokens.Delete(ctx, at.ID) }()

		// Get a refreshed view from the API.
		atl, err := client.ServiceAccountTokens.List(ctx, sa.ID, AccessTokenListOptions{})
		require.NoError(t, err)

		refreshed := atl.Items[0]

		assert.NotEmpty(t, refreshed.ID)
		assert.Equal(t, *options.Description, refreshed.Description)
	})

	t.Run("when description is not provided", func(t *testing.T) {
		options := AccessTokenCreateOptions{}

		at, err := client.ServiceAccountTokens.Create(ctx, sa.ID, options)
		require.NoError(t, err)

		defer func() { _ = client.AccessTokens.Delete(ctx, at.ID) }()

		// Get a refreshed view from the API.
		atl, err := client.ServiceAccountTokens.List(ctx, sa.ID, AccessTokenListOptions{})
		require.NoError(t, err)

		refreshed := atl.Items[0]

		assert.NotEmpty(t, refreshed.ID)
		assert.Equal(t, refreshed.Description, "")
	})

	t.Run("with nonexistent service account id", func(t *testing.T) {
		var saID = "notexisting"
		_, err := client.ServiceAccountTokens.Create(ctx, saID, AccessTokenCreateOptions{})
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("ServiceAccount with ID '%s' not found or user unauthorized.", saID),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("with invalid service account id", func(t *testing.T) {
		_, err := client.ServiceAccountTokens.Create(ctx, badIdentifier, AccessTokenCreateOptions{})
		assert.EqualError(t, err, "invalid value for service account ID")
	})
}
