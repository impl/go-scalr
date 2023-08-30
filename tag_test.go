package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagsList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	tagTest1, tagTest1Cleanup := createTag(t, client)
	defer tagTest1Cleanup()
	tagTest2, tagTest2Cleanup := createTag(t, client)
	defer tagTest2Cleanup()

	t.Run("without options", func(t *testing.T) {
		tagl, err := client.Tags.List(ctx, TagListOptions{})
		require.NoError(t, err)
		assert.Equal(t, 2, tagl.TotalCount)

		tagIDs := make([]string, len(tagl.Items))
		for i, tag := range tagl.Items {
			tagIDs[i] = tag.ID
		}
		assert.Contains(t, tagIDs, tagTest1.ID)
		assert.Contains(t, tagIDs, tagTest2.ID)
	})

	t.Run("with options", func(t *testing.T) {
		tagl, err := client.Tags.List(ctx,
			TagListOptions{Account: String(defaultAccountID), Name: String(tagTest1.Name)},
		)
		require.NoError(t, err)
		assert.Equal(t, 1, tagl.TotalCount)
		assert.Equal(t, tagTest1.ID, tagl.Items[0].ID)
	})

	t.Run("with ID in options", func(t *testing.T) {
		tagl, err := client.Tags.List(ctx, TagListOptions{
			Account: String(defaultAccountID),
			Tag:     String(tagTest1.ID)},
		)
		require.NoError(t, err)
		assert.Equal(t, 1, tagl.TotalCount)
		assert.Equal(t, tagTest1.ID, tagl.Items[0].ID)
	})
}

func TestTagsCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("with valid options", func(t *testing.T) {
		options := TagCreateOptions{
			Name:    String("tst-" + randomString(t)),
			Account: &Account{ID: defaultAccountID},
		}

		tag, err := client.Tags.Create(ctx, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.Tags.Read(ctx, tag.ID)
		require.NoError(t, err)

		for _, item := range []*Tag{
			tag,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
			assert.Equal(t, options.Account, item.Account)
		}
		err = client.Tags.Delete(ctx, tag.ID)
		require.NoError(t, err)
	})

	t.Run("when options has name missing", func(t *testing.T) {
		tag, err := client.Tags.Create(ctx, TagCreateOptions{
			Account: &Account{ID: defaultAccountID},
		})
		assert.Nil(t, tag)
		assert.EqualError(t, err, "name is required")
	})

	t.Run("when options has an empty name", func(t *testing.T) {
		tag, err := client.Tags.Create(ctx, TagCreateOptions{
			Name:    String(" "),
			Account: &Account{ID: defaultAccountID},
		})
		assert.Nil(t, tag)
		assert.EqualError(t, err, "Invalid Attribute\n\nName cannot be empty.")
	})

	t.Run("when options has an invalid account", func(t *testing.T) {
		var accountId = "acc-123"
		_, err := client.Tags.Create(ctx, TagCreateOptions{
			Name:    String(" "),
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

func TestTagsRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	tagTest, tagTestCleanup := createTag(t, client)
	defer tagTestCleanup()

	t.Run("by ID when the tag exists", func(t *testing.T) {
		tag, err := client.Tags.Read(ctx, tagTest.ID)
		require.NoError(t, err)
		assert.Equal(t, tagTest.ID, tag.ID)
	})

	t.Run("by ID when the tag does not exist", func(t *testing.T) {
		tag, err := client.Tags.Read(ctx, "tag-nonexisting")
		assert.Nil(t, tag)
		assert.Error(t, err)
	})

	t.Run("by ID without a valid tag ID", func(t *testing.T) {
		tag, err := client.Tags.Read(ctx, badIdentifier)
		assert.Nil(t, tag)
		assert.EqualError(t, err, "invalid value for tag ID")
	})
}

func TestTagsUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	tagTest, tagTestCleanup := createTag(t, client)
	defer tagTestCleanup()

	t.Run("with valid options", func(t *testing.T) {
		options := TagUpdateOptions{
			Name: String(randomString(t)),
		}

		tag, err := client.Tags.Update(ctx, tagTest.ID, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.Tags.Read(ctx, tagTest.ID)
		require.NoError(t, err)

		for _, item := range []*Tag{
			tag,
			refreshed,
		} {
			assert.Equal(t, *options.Name, item.Name)
		}
	})

	t.Run("with invalid name", func(t *testing.T) {
		tag, err := client.Tags.Update(ctx, tagTest.ID, TagUpdateOptions{
			Name: String(badIdentifier),
		})
		assert.Nil(t, tag)
		assert.Error(t, err)
	})
}

func TestTagsDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	tagTest, _ := createTag(t, client)

	t.Run("with valid options", func(t *testing.T) {
		err := client.Tags.Delete(ctx, tagTest.ID)
		require.NoError(t, err)

		_, err = client.Tags.Read(ctx, tagTest.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("AccountTag with ID '%s' not found or user unauthorized.", tagTest.ID),
			}.Error(),
			err.Error(),
		)
	})
}
