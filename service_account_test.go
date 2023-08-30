package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceAccountsList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	acc := Account{ID: defaultAccountID}
	saTest1, saTest1Cleanup := createServiceAccount(
		t, client, &acc, ServiceAccountStatusPtr(ServiceAccountStatusActive),
	)
	defer saTest1Cleanup()
	saTest2, saTest2Cleanup := createServiceAccount(
		t, client, &acc, ServiceAccountStatusPtr(ServiceAccountStatusInactive),
	)
	defer saTest2Cleanup()

	t.Run("without options", func(t *testing.T) {
		sal, err := client.ServiceAccounts.List(ctx, ServiceAccountListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 3, sal.TotalCount) // one Service Account is created by default on test environment

		saIDs := make([]string, len(sal.Items))
		for i, sa := range sal.Items {
			saIDs[i] = sa.ID
		}
		assert.Contains(t, saIDs, saTest1.ID)
		assert.Contains(t, saIDs, saTest2.ID)
	})

	t.Run("with filters", func(t *testing.T) {
		sal, err := client.ServiceAccounts.List(ctx,
			ServiceAccountListOptions{Account: String(defaultAccountID), Email: String(saTest1.Email)},
		)
		require.NoError(t, err)
		assert.Equal(t, 1, sal.TotalCount)
		assert.Equal(t, saTest1.ID, sal.Items[0].ID)
	})

	t.Run("with ID filter", func(t *testing.T) {
		sal, err := client.ServiceAccounts.List(ctx, ServiceAccountListOptions{
			Account:        String(defaultAccountID),
			ServiceAccount: String(saTest1.ID)},
		)
		require.NoError(t, err)
		assert.Equal(t, 1, sal.TotalCount)
		assert.Equal(t, saTest1.ID, sal.Items[0].ID)
	})

	t.Run("with query", func(t *testing.T) {
		sal, err := client.ServiceAccounts.List(ctx,
			ServiceAccountListOptions{Query: String(saTest2.Description)},
		)
		require.NoError(t, err)
		assert.Equal(t, 1, sal.TotalCount)
		assert.Equal(t, saTest2.Description, sal.Items[0].Description)
	})
}

func TestServiceAccountsCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("with valid options", func(t *testing.T) {
		options := ServiceAccountCreateOptions{
			Name:        String("tst-" + randomString(t)),
			Description: String("tst-description-" + randomString(t)),
			Status:      ServiceAccountStatusPtr(ServiceAccountStatusActive),
			Account:     &Account{ID: defaultAccountID},
		}

		sa, err := client.ServiceAccounts.Create(ctx, options)
		require.NoError(t, err)

		defer func() { _ = client.ServiceAccounts.Delete(ctx, sa.ID) }()

		// Get a refreshed view from the API.
		refreshed, err := client.ServiceAccounts.Read(ctx, sa.ID)
		require.NoError(t, err)

		for _, item := range []*ServiceAccount{
			sa,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, options.Name, &item.Name)
			assert.Equal(t, options.Description, &item.Description)
			assert.Equal(t, options.Status, &item.Status)
			assert.Equal(t, &options.Account, &item.Account)
		}
	})

	t.Run("when options has name missing", func(t *testing.T) {
		_, err := client.ServiceAccounts.Create(ctx, ServiceAccountCreateOptions{
			Account: &Account{ID: defaultAccountID},
		})
		assert.EqualError(t, err, "name is required")
	})

	t.Run("when options has an empty name", func(t *testing.T) {
		_, err := client.ServiceAccounts.Create(ctx, ServiceAccountCreateOptions{
			Name:    String(" "),
			Account: &Account{ID: defaultAccountID},
		})
		assert.EqualError(t, err, "Invalid Attribute\n\nService account name can not be empty.")
	})

	t.Run("when options has account missing", func(t *testing.T) {
		_, err := client.ServiceAccounts.Create(ctx, ServiceAccountCreateOptions{
			Name: String("tst-" + randomString(t)),
		})
		assert.EqualError(t, err, "account is required")
	})

	t.Run("when options has invalid account id", func(t *testing.T) {
		_, err := client.ServiceAccounts.Create(ctx, ServiceAccountCreateOptions{
			Name:    String("tst-" + randomString(t)),
			Account: &Account{ID: badIdentifier},
		})
		assert.EqualError(t, err, "invalid value for account ID")
	})

	t.Run("when options has an invalid account", func(t *testing.T) {
		var accountId = "acc-123"
		_, err := client.ServiceAccounts.Create(ctx, ServiceAccountCreateOptions{
			Name:    String("tst-" + randomString(t)),
			Account: &Account{ID: accountId},
		})
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Invalid Relationship\n\nAccount with ID '%s' not found or user unauthorized.", accountId),
			}.Error(),
			err.Error(),
		)
	})
}

func TestServiceAccountsRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	saTest, saTestCleanup := createServiceAccount(
		t, client, &Account{ID: defaultAccountID}, ServiceAccountStatusPtr(ServiceAccountStatusActive),
	)
	defer saTestCleanup()

	t.Run("by ID when the service account exists", func(t *testing.T) {
		sa, err := client.ServiceAccounts.Read(ctx, saTest.ID)
		require.NoError(t, err)
		assert.Equal(t, saTest.ID, sa.ID)
	})

	t.Run("by ID when the service account does not exist", func(t *testing.T) {
		saID := "sa-nonexisting"
		_, err := client.ServiceAccounts.Read(ctx, saID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("ServiceAccount with ID '%s' not found or user unauthorized.", saID),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("by ID with invalid service account ID", func(t *testing.T) {
		_, err := client.ServiceAccounts.Read(ctx, badIdentifier)
		assert.EqualError(t, err, "invalid value for service account ID")
	})
}

func TestServiceAccountsUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	saTest, saTestCleanup := createServiceAccount(
		t, client, &Account{ID: defaultAccountID}, ServiceAccountStatusPtr(ServiceAccountStatusActive),
	)
	defer saTestCleanup()

	t.Run("with valid options", func(t *testing.T) {
		options := ServiceAccountUpdateOptions{
			Description: String("tst-description-" + randomString(t)),
			Status:      ServiceAccountStatusPtr(ServiceAccountStatusInactive),
		}

		sa, err := client.ServiceAccounts.Update(ctx, saTest.ID, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.ServiceAccounts.Read(ctx, saTest.ID)
		require.NoError(t, err)

		for _, item := range []*ServiceAccount{
			sa,
			refreshed,
		} {
			assert.Equal(t, options.Description, &item.Description)
			assert.Equal(t, options.Status, &item.Status)
		}
	})
}

func TestServiceAccountsDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	saTest, _ := createServiceAccount(
		t, client, &Account{ID: defaultAccountID}, ServiceAccountStatusPtr(ServiceAccountStatusActive),
	)

	t.Run("with valid options", func(t *testing.T) {
		err := client.ServiceAccounts.Delete(ctx, saTest.ID)
		require.NoError(t, err)

		_, err = client.ServiceAccounts.Read(ctx, saTest.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("ServiceAccount with ID '%s' not found or user unauthorized.", saTest.ID),
			}.Error(),
			err.Error(),
		)
	})
}
