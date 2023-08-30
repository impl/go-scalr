package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultTeamID = "team-t67mjtnokncjpd8"
)

func TestTeamsList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	teamTest1, teamTest1Cleanup := createTeam(t, client, nil)
	defer teamTest1Cleanup()
	teamTest2, teamTest2Cleanup := createTeam(t, client, nil)
	defer teamTest2Cleanup()

	t.Run("with empty options", func(t *testing.T) {
		tl, err := client.Teams.List(ctx, TeamListOptions{})
		require.NoError(t, err)

		var tIDs []string
		for _, t := range tl.Items {
			tIDs = append(tIDs, t.ID)
		}
		assert.Equal(t, 1, tl.CurrentPage)
		assert.Contains(t, tIDs, teamTest1.ID)
		assert.Contains(t, tIDs, teamTest2.ID)
	})

	t.Run("with team filter", func(t *testing.T) {
		tl, err := client.Teams.List(ctx, TeamListOptions{
			Team: String(teamTest1.ID),
		})
		require.NoError(t, err)
		assert.Equal(t, 1, tl.CurrentPage)
		assert.Equal(t, 1, tl.TotalCount)
		assert.Equal(t, teamTest1.ID, tl.Items[0].ID)
	})

	t.Run("with name filter", func(t *testing.T) {
		tl, err := client.Teams.List(ctx, TeamListOptions{
			Name: String(teamTest1.Name),
		})
		require.NoError(t, err)

		var tIDs []string
		// Set of team names
		tNames := make(map[string]bool)
		for _, t := range tl.Items {
			tIDs = append(tIDs, t.ID)
			if !tNames[t.Name] {
				tNames[t.Name] = true
			}
		}

		assert.Equal(t, 1, tl.CurrentPage)
		assert.True(t, tl.TotalCount >= 1)
		assert.Contains(t, tIDs, teamTest1.ID)
		assert.Equal(t, 1, len(tNames))
		assert.Contains(t, tNames, teamTest1.Name)
	})

	t.Run("with identity provider filter", func(t *testing.T) {
		tl, err := client.Teams.List(ctx, TeamListOptions{
			IdentityProvider: String(defaultIdentityProviderScalrID),
		})
		require.NoError(t, err)

		var tIDs []string
		// Set of IDP IDs
		idpIDs := make(map[string]bool)
		for _, team := range tl.Items {
			tIDs = append(tIDs, team.ID)
			if !idpIDs[team.IdentityProvider.ID] {
				idpIDs[team.IdentityProvider.ID] = true
			}
		}

		assert.Equal(t, 1, tl.CurrentPage)
		assert.True(t, tl.TotalCount >= 1)
		assert.Contains(t, tIDs, teamTest1.ID)
		assert.Contains(t, tIDs, teamTest2.ID)
		assert.Equal(t, 1, len(idpIDs))
		assert.Contains(t, idpIDs, defaultIdentityProviderScalrID)
	})

	t.Run("with account filter", func(t *testing.T) {
		tl, err := client.Teams.List(ctx, TeamListOptions{
			Account: String(defaultAccountID),
		})
		require.NoError(t, err)

		var tIDs []string
		accIDs := make(map[string]bool)
		for _, t := range tl.Items {
			tIDs = append(tIDs, t.ID)
			if !accIDs[t.Account.ID] {
				accIDs[t.Account.ID] = true
			}
		}
		assert.Equal(t, 1, tl.CurrentPage)
		assert.True(t, tl.TotalCount >= 1)
		assert.Contains(t, tIDs, teamTest1.ID)
		assert.Contains(t, tIDs, teamTest2.ID)
		assert.Equal(t, 1, len(accIDs))
		assert.Contains(t, accIDs, defaultAccountID)
	})

	t.Run("without a valid account", func(t *testing.T) {
		tl, err := client.Teams.List(ctx, TeamListOptions{
			Account: String(badIdentifier),
		})
		assert.NoError(t, err)
		assert.Len(t, tl.Items, 0)
	})

	t.Run("without a valid identity provider", func(t *testing.T) {
		tl, err := client.Teams.List(ctx, TeamListOptions{
			IdentityProvider: String(badIdentifier),
		})
		assert.NoError(t, err)
		assert.Len(t, tl.Items, 0)
	})

	t.Run("with list options", func(t *testing.T) {
		tl, err := client.Teams.List(ctx, TeamListOptions{
			ListOptions: ListOptions{
				PageNumber: 999,
				PageSize:   100,
			},
		})
		require.NoError(t, err)
		assert.Empty(t, tl.Items)
		assert.Equal(t, 999, tl.CurrentPage)
		assert.True(t, tl.TotalCount >= 1)
	})
}

func TestTeamsCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("with valid options", func(t *testing.T) {
		options := TeamCreateOptions{
			Account:     &Account{ID: defaultAccountID},
			Name:        String("foo" + randomString(t)),
			Description: String("bar"),
			Users:       []*User{{ID: defaultUserID}},
		}

		team, err := client.Teams.Create(ctx, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.Teams.Read(ctx, team.ID)
		require.NoError(t, err)

		for _, item := range []*Team{
			team,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, *options.Description, item.Description)
			assert.Equal(t, options.Account, item.Account)
			assert.Equal(t, options.Users, item.Users)
		}
		err = client.Teams.Delete(ctx, team.ID)
		require.NoError(t, err)
	})

	t.Run("with empty options", func(t *testing.T) {
		team, err := client.Teams.Create(ctx, TeamCreateOptions{})
		assert.Nil(t, team)
		assert.EqualError(t, err, "name is required")
	})

	t.Run("when options has an invalid account", func(t *testing.T) {
		team, err := client.Teams.Create(ctx, TeamCreateOptions{
			Name:    String("foo"),
			Account: &Account{ID: badIdentifier},
		})
		assert.Nil(t, team)
		assert.EqualError(t, err, "invalid value for account ID")
	})

	t.Run("when options has an invalid identity provider", func(t *testing.T) {
		team, err := client.Teams.Create(ctx, TeamCreateOptions{
			Name:             String("foo"),
			IdentityProvider: &IdentityProvider{ID: badIdentifier},
		})
		assert.Nil(t, team)
		assert.EqualError(t, err, "invalid value for identity provider ID")
	})
}

func TestTeamsRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	testTeam, testTeamCleanup := createTeam(t, client, []*User{{ID: defaultUserID}})
	defer testTeamCleanup()

	t.Run("when the team exists", func(t *testing.T) {
		team, err := client.Teams.Read(ctx, testTeam.ID)
		require.NoError(t, err)
		assert.Equal(t, testTeam.ID, team.ID)

		t.Run("relationships are properly decoded", func(t *testing.T) {
			assert.Equal(t, team.Account.ID, defaultAccountID)
			assert.Equal(t, team.IdentityProvider.ID, defaultIdentityProviderScalrID)
			assert.Equal(t, team.Users[0].ID, defaultUserID)
		})
	})

	t.Run("when the team does not exist", func(t *testing.T) {
		team, err := client.Teams.Read(ctx, "team-nonexisting")
		assert.Nil(t, team)
		assert.Error(t, err)
	})

	t.Run("without a valid team ID", func(t *testing.T) {
		team, err := client.Teams.Read(ctx, badIdentifier)
		assert.Nil(t, team)
		assert.EqualError(t, err, "invalid value for team ID")
	})
}

func TestTeamsUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	testTeam, testTeamCleanup := createTeam(t, client, nil)
	defer testTeamCleanup()

	t.Run("when updating a subset of values", func(t *testing.T) {
		options := TeamUpdateOptions{
			Name: String("tst-" + randomString(t)),
		}

		teamAfter, err := client.Teams.Update(ctx, testTeam.ID, options)
		require.NoError(t, err)

		assert.Equal(t, *options.Name, teamAfter.Name)
		assert.Equal(t, testTeam.Description, teamAfter.Description)
	})

	t.Run("when updating a users list", func(t *testing.T) {
		users := []*User{{ID: defaultUserID}}

		options := TeamUpdateOptions{
			Name:  String(testTeam.Name),
			Users: users,
		}

		teamAfter, err := client.Teams.Update(ctx, testTeam.ID, options)
		require.NoError(t, err)

		assert.Equal(t, testTeam.Name, teamAfter.Name)
		assert.Equal(t, users, teamAfter.Users)
	})

	t.Run("without a valid team ID", func(t *testing.T) {
		team, err := client.Teams.Update(ctx, badIdentifier, TeamUpdateOptions{})
		assert.Nil(t, team)
		assert.EqualError(t, err, "invalid value for team ID")
	})
}

func TestTeamsDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	testTeam, _ := createTeam(t, client, nil)

	t.Run("with valid options", func(t *testing.T) {
		err := client.Teams.Delete(ctx, testTeam.ID)
		require.NoError(t, err)

		// Try loading team - it should fail.
		_, err = client.Teams.Read(ctx, testTeam.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Team with ID '%s' not found or user unauthorized.", testTeam.ID),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("without a valid team ID", func(t *testing.T) {
		err := client.Teams.Delete(ctx, badIdentifier)
		assert.EqualError(t, err, "invalid value for team ID")
	})
}
