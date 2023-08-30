package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVCSRevisionRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("when the vcs revision does not exist", func(t *testing.T) {
		var vcsId = "nonexisting"
		cv, err := client.VcsRevisions.Read(ctx, vcsId)
		assert.Nil(t, cv)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("VcsRevisionBinding with ID '%s' not found or user unauthorized.", vcsId),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("with invalid vcs revision id", func(t *testing.T) {
		cv, err := client.VcsRevisions.Read(ctx, badIdentifier)
		assert.Nil(t, cv)
		assert.EqualError(t, err, "invalid value for vcs revision ID")
	})
}
