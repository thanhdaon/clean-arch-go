package comment_test

import (
	"clean-arch-go/domain/comment"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewComment(t *testing.T) {
	t.Parallel()
	c, err := comment.New("task-id-1", "author-id-1", "Hello world")
	require.NoError(t, err)
	require.NotEmpty(t, c.UUID())
	require.Equal(t, "task-id-1", c.TaskID())
	require.Equal(t, "author-id-1", c.AuthorID())
	require.Equal(t, "Hello world", c.Content())
	require.Empty(t, c.References())
	require.False(t, c.IsDeleted())
}

func TestNewComment_ParsesMentions(t *testing.T) {
	t.Parallel()
	c, err := comment.New("task-id-1", "author-id-1", "Hey @[user:uuid-123] see @[task:uuid-456]")
	require.NoError(t, err)
	refs := c.References()
	require.Len(t, refs, 2)
	require.Equal(t, comment.ReferenceTypeUser, refs[0].Type)
	require.Equal(t, "uuid-123", refs[0].ID)
	require.Equal(t, comment.ReferenceTypeTask, refs[1].Type)
	require.Equal(t, "uuid-456", refs[1].ID)
}

func TestNewComment_EmptyTaskID(t *testing.T) {
	t.Parallel()
	_, err := comment.New("", "author-id-1", "text")
	require.EqualError(t, err, "empty task id")
}

func TestNewComment_EmptyAuthorID(t *testing.T) {
	t.Parallel()
	_, err := comment.New("task-id-1", "", "text")
	require.EqualError(t, err, "empty author id")
}

func TestNewComment_EmptyContent(t *testing.T) {
	t.Parallel()
	_, err := comment.New("task-id-1", "author-id-1", "")
	require.EqualError(t, err, "empty content")
}

func TestComment_Update(t *testing.T) {
	t.Parallel()
	c, _ := comment.New("task-id-1", "author-id-1", "original")
	err := c.Update("author-id-1", "updated content")
	require.NoError(t, err)
	require.Equal(t, "updated content", c.Content())
}

func TestComment_Update_WrongAuthor(t *testing.T) {
	t.Parallel()
	c, _ := comment.New("task-id-1", "author-id-1", "original")
	err := c.Update("other-user-id", "updated content")
	require.EqualError(t, err, "only comment author can edit this comment")
}

func TestComment_Delete(t *testing.T) {
	t.Parallel()
	c, _ := comment.New("task-id-1", "author-id-1", "text")
	err := c.Delete("author-id-1")
	require.NoError(t, err)
	require.True(t, c.IsDeleted())
}

func TestComment_Delete_WrongAuthor(t *testing.T) {
	t.Parallel()
	c, _ := comment.New("task-id-1", "author-id-1", "text")
	err := c.Delete("other-user-id")
	require.EqualError(t, err, "only comment author can delete this comment")
}
