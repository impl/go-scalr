package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModulesList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	m, err := client.Modules.Read(ctx, defaultModuleID)
	require.NoError(t, err)
	assert.Equal(t, m.ID, defaultModuleID)

	t.Run("without list options", func(t *testing.T) {
		ml, err := client.Modules.List(ctx, ModuleListOptions{})
		require.NoError(t, err)
		mlIDs := make([]string, len(ml.Items))
		for _, ws := range ml.Items {
			mlIDs = append(mlIDs, ws.ID)
		}
		assert.Contains(t, mlIDs, defaultModuleID)
	})

	t.Run("with list options", func(t *testing.T) {
		ml, err := client.Modules.List(ctx, ModuleListOptions{
			ListOptions: ListOptions{
				PageNumber: 999,
				PageSize:   100,
			},
		})
		require.NoError(t, err)
		assert.Empty(t, ml.Items)
		assert.Equal(t, 999, ml.CurrentPage)
	})
	filters := []ModuleListOptions{
		{Name: &m.Name},
		{Status: &m.Status, Name: &m.Name},
		{Provider: &m.Provider, Name: &m.Name},
		{Account: &m.Account.ID, Environment: &m.Environment.ID},
		{Environment: &m.Environment.ID},
	}
	for _, mlo := range filters {
		t.Run("with a valid option", func(t *testing.T) {
			wl, err := client.Modules.List(ctx, mlo)
			assert.Len(t, wl.Items, 1, mlo)
			assert.NoError(t, err)
			assert.Equal(t, m.ID, wl.Items[0].ID)
		})
	}
}

func TestModulesCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("when empty options", func(t *testing.T) {
		w, err := client.Modules.Create(ctx, ModuleCreateOptions{})
		assert.Nil(t, w)
		assert.EqualError(t, err, "vcs repo is required")
	})

	t.Run("when options has invalid vcs repo identifier", func(t *testing.T) {
		w, err := client.Modules.Create(ctx, ModuleCreateOptions{
			VCSRepo:     &ModuleVCSRepo{Identifier: "foo/bar"},
			VcsProvider: &VcsProvider{ID: *String(badIdentifier)},
		})
		assert.Nil(t, w)
		assert.EqualError(t, err, ResourceNotFoundError{
			Message: fmt.Sprintf("Invalid Relationship\n\nVcsProvider with ID '%s' not found or user unauthorized.", badIdentifier),
		}.Error())
	})

	t.Run("when an error is returned from the api", func(t *testing.T) {
		ws, err := client.Modules.Create(ctx, ModuleCreateOptions{
			Environment: &Environment{ID: *String(badIdentifier)},
			VCSRepo:     &ModuleVCSRepo{Identifier: "foo/bar"},
			VcsProvider: &VcsProvider{ID: "vcs-test"},
		})
		assert.Nil(t, ws)
		assert.Error(t, err)
	})
}

func TestModulesRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("when the module exists", func(t *testing.T) {
		m, err := client.Modules.Read(ctx, defaultModuleID)
		require.NoError(t, err)
		assert.Equal(t, defaultModuleID, m.ID)
	})

	t.Run("when the module does not exist", func(t *testing.T) {
		_, err := client.Modules.Read(ctx, "nonexisting")
		assert.Error(t, err)
	})

	t.Run("without a valid identifier", func(t *testing.T) {
		_, err := client.Modules.Read(ctx, badIdentifier)
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value for module ID")
	})
}

func TestModulesReadBySource(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	module, err := client.Modules.Read(ctx, defaultModuleID)
	require.NoError(t, err)

	t.Run("with valid source", func(t *testing.T) {
		m, err := client.Modules.ReadBySource(ctx, module.Source)
		require.NoError(t, err)
		assert.NotEmpty(t, m.Source)
		assert.Equal(t, module.ID, m.ID)
		assert.Equal(t, module.Source, m.Source)
	})

	t.Run("Invalid source", func(t *testing.T) {
		ms := "invalidSource"
		_, err := client.Modules.ReadBySource(ctx, "invalidSource")
		require.Error(t, err)
		assert.EqualError(t, err, ResourceNotFoundError{
			Message: fmt.Sprintf("Module with source '%s' not found.", ms),
		}.Error())
	})
}

func TestModulesDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("without a valid module ID", func(t *testing.T) {
		err := client.Modules.Delete(ctx, badIdentifier)
		assert.EqualError(t, err, "invalid value for module ID")
	})
}
