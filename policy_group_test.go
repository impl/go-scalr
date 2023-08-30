package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyGroupsList(t *testing.T) {
	// TODO: delete skip after SCALRCORE-19891
	t.Skip("Works with personal token but does not work with github action token.")

	client := testClient(t)
	ctx := context.Background()

	vcsProvider, vcsProviderCleanup := createVcsProvider(t, client, nil)
	defer vcsProviderCleanup()

	pg1, pg1Cleanup := createPolicyGroup(t, client, vcsProvider)
	defer pg1Cleanup()

	pg2, pg2Cleanup := createPolicyGroup(t, client, vcsProvider)
	defer pg2Cleanup()

	t.Run("without list options", func(t *testing.T) {
		pgl, err := client.PolicyGroups.List(ctx, PolicyGroupListOptions{})
		require.NoError(t, err)
		pgIDs := make([]string, len(pgl.Items))
		for _, pg := range pgl.Items {
			pgIDs = append(pgIDs, pg.ID)
		}
		assert.Contains(t, pgIDs, pg1.ID)
		assert.Contains(t, pgIDs, pg2.ID)
		assert.Equal(t, 1, pgl.CurrentPage)
		assert.True(t, pgl.TotalCount >= 2)
	})

	t.Run("with name and account options", func(t *testing.T) {
		pgl, err := client.PolicyGroups.List(ctx, PolicyGroupListOptions{
			Account: defaultAccountID,
			Name:    pg1.Name,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, pgl.CurrentPage)
		assert.Equal(t, 1, pgl.TotalCount)
		assert.Equal(t, pg1.ID, pgl.Items[0].ID)
	})

	t.Run("with list options", func(t *testing.T) {
		pgl, err := client.PolicyGroups.List(ctx, PolicyGroupListOptions{
			ListOptions: ListOptions{
				PageNumber: 999,
				PageSize:   100,
			},
		})
		require.NoError(t, err)
		assert.Empty(t, pgl.Items)
		assert.Equal(t, 999, pgl.CurrentPage)
		assert.True(t, pgl.TotalCount >= 2)
	})

	t.Run("without a valid account", func(t *testing.T) {
		pgl, err := client.PolicyGroups.List(ctx, PolicyGroupListOptions{Account: badIdentifier})
		assert.Len(t, pgl.Items, 0)
		assert.NoError(t, err)
	})
}

func TestPolicyGroupsCreate(t *testing.T) {
	// TODO: delete skip after SCALRCORE-19891
	t.Skip("Works with personal token but does not work with github action token.")

	client := testClient(t)
	ctx := context.Background()

	vcsProvider, vcsProviderCleanup := createVcsProvider(t, client, nil)
	defer vcsProviderCleanup()

	t.Run("with valid options", func(t *testing.T) {
		options := PolicyGroupCreateOptions{
			Name:        String("foo"),
			Account:     &Account{ID: defaultAccountID},
			VcsProvider: vcsProvider,
			VCSRepo: &PolicyGroupVCSRepoOptions{
				Identifier: String(policyGroupVcsRepoID),
				Path:       String(policyGroupVcsRepoPath),
			},
		}

		pg, err := client.PolicyGroups.Create(ctx, options)
		defer func() { client.PolicyGroups.Delete(ctx, pg.ID) }()

		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.PolicyGroups.Read(ctx, pg.ID)
		require.NoError(t, err)

		for _, item := range []*PolicyGroup{
			pg,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
			assert.NotEmpty(t, item.OpaVersion)
			assert.Equal(t, options.Account.ID, item.Account.ID)
			assert.Equal(t, options.VcsProvider.ID, item.VcsProvider.ID)
			assert.Equal(t, *options.VCSRepo.Identifier, item.VCSRepo.Identifier)
			assert.Equal(t, *options.VCSRepo.Path, item.VCSRepo.Path)
			assert.NotEmpty(t, item.VCSRepo.Branch)
		}
	})

	t.Run("with empty options", func(t *testing.T) {
		pg, err := client.PolicyGroups.Create(ctx, PolicyGroupCreateOptions{})
		assert.Nil(t, pg)
		assert.EqualError(t, err, "name is required")
	})

	t.Run("without vcs repo options", func(t *testing.T) {
		pg, err := client.PolicyGroups.Create(ctx, PolicyGroupCreateOptions{
			Name:        String("foo"),
			Account:     &Account{ID: defaultAccountID},
			VcsProvider: vcsProvider,
		})
		assert.Nil(t, pg)
		assert.EqualError(t, err, "vcs repo is required")
	})

	t.Run("when options has an invalid account", func(t *testing.T) {
		var accID = "acc-123"
		pg, err := client.PolicyGroups.Create(ctx, PolicyGroupCreateOptions{
			Name:        String("foo"),
			Account:     &Account{ID: accID},
			VcsProvider: vcsProvider,
			VCSRepo: &PolicyGroupVCSRepoOptions{
				Identifier: String(policyGroupVcsRepoID),
			},
		})
		assert.Nil(t, pg)
		assert.EqualError(
			t,
			err,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Clients with ID '%s' not found or user unauthorized.", accID),
			}.Error(),
		)
	})

	t.Run("when options has an invalid vcs provider", func(t *testing.T) {
		var vcsID = "vcs-123"
		pg, err := client.PolicyGroups.Create(ctx, PolicyGroupCreateOptions{
			Name:        String("foo"),
			Account:     &Account{ID: defaultAccountID},
			VcsProvider: &VcsProvider{ID: vcsID},
			VCSRepo: &PolicyGroupVCSRepoOptions{
				Identifier: String(policyGroupVcsRepoID),
			},
		})
		assert.Nil(t, pg)
		assert.EqualError(
			t,
			err,
			ResourceNotFoundError{
				Message: fmt.Sprintf("VcsProvider with ID '%s' not found or user unauthorized.", vcsID),
			}.Error(),
		)
	})
}

func TestPolicyGroupsRead(t *testing.T) {
	// TODO: delete skip after SCALRCORE-19891
	t.Skip("Works with personal token but does not work with github action token.")

	client := testClient(t)
	ctx := context.Background()

	policyGroup, policyGroupCleanup := createPolicyGroup(t, client, nil)
	defer policyGroupCleanup()

	t.Run("when the policy group exists", func(t *testing.T) {
		pg, err := client.PolicyGroups.Read(ctx, policyGroup.ID)
		require.NoError(t, err)
		assert.Equal(t, policyGroup.ID, pg.ID)

		t.Run("relationships are properly decoded", func(t *testing.T) {
			assert.Equal(t, pg.Account.ID, defaultAccountID)
		})

		t.Run("vcs repo is properly decoded", func(t *testing.T) {
			assert.Equal(t, policyGroupVcsRepoID, pg.VCSRepo.Identifier)
			assert.NotEmpty(t, pg.VCSRepo.Branch)
			assert.Equal(t, policyGroupVcsRepoPath, pg.VCSRepo.Path)
		})
	})

	t.Run("when the policy group does not exist", func(t *testing.T) {
		pg, err := client.PolicyGroups.Read(ctx, "pgrp-nonexisting")
		assert.Nil(t, pg)
		assert.Error(t, err)
	})

	t.Run("without a valid policy group ID", func(t *testing.T) {
		pg, err := client.PolicyGroups.Read(ctx, badIdentifier)
		assert.Nil(t, pg)
		assert.EqualError(t, err, "invalid value for policy group ID")
	})
}

func TestPolicyGroupsUpdate(t *testing.T) {
	// TODO: delete skip after SCALRCORE-19891
	t.Skip("Works with personal token but does not work with github action token.")

	client := testClient(t)
	ctx := context.Background()

	policyGroup, policyGroupCleanup := createPolicyGroup(t, client, nil)
	defer policyGroupCleanup()

	t.Run("when updating a subset of values", func(t *testing.T) {
		options := PolicyGroupUpdateOptions{
			Name: String("tst-" + randomString(t)),
		}

		pgAfter, err := client.PolicyGroups.Update(ctx, policyGroup.ID, options)
		require.NoError(t, err)

		assert.Equal(t, *options.Name, pgAfter.Name)
		assert.Equal(t, policyGroup.OpaVersion, pgAfter.OpaVersion)
	})

	t.Run("when an error is returned from the api", func(t *testing.T) {
		pg, err := client.PolicyGroups.Update(ctx, policyGroup.ID, PolicyGroupUpdateOptions{
			OpaVersion: String("nonexisting"),
		})
		assert.Nil(t, pg)
		assert.Error(t, err)
	})

	t.Run("without a valid policy group ID", func(t *testing.T) {
		pg, err := client.PolicyGroups.Update(ctx, badIdentifier, PolicyGroupUpdateOptions{})
		assert.Nil(t, pg)
		assert.EqualError(t, err, "invalid value for policy group ID")
	})
}

func TestPolicyGroupsDelete(t *testing.T) {
	// TODO: delete skip after SCALRCORE-19891
	t.Skip("Works with personal token but does not work with github action token.")

	client := testClient(t)
	ctx := context.Background()

	vcsProvider, vcsProviderCleanup := createVcsProvider(t, client, nil)
	defer vcsProviderCleanup()

	policyGroup, _ := createPolicyGroup(t, client, vcsProvider)

	t.Run("with valid options", func(t *testing.T) {
		err := client.PolicyGroups.Delete(ctx, policyGroup.ID)
		require.NoError(t, err)

		// Try loading policy group - it should fail.
		_, err = client.PolicyGroups.Read(ctx, policyGroup.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("PolicyGroups with ID '%s' not found or user unauthorized.", policyGroup.ID),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("without a valid policy group ID", func(t *testing.T) {
		err := client.PolicyGroups.Delete(ctx, badIdentifier)
		assert.EqualError(t, err, "invalid value for policy group ID")
	})
}
