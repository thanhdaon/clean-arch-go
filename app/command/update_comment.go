package command

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/activity"
	"clean-arch-go/domain/comment"
	"context"
	"log"

	"github.com/sirupsen/logrus"
)

type UpdateComment struct {
	CommentID string
	EditorID  string
	Content   string
}

type UpdateCommentHandler decorator.CommandHandler[UpdateComment]

type updateCommentHandler struct {
	comments   CommentRepository
	activities ActivityRepository
}

func NewUpdateCommentHandler(comments CommentRepository, activities ActivityRepository, logger *logrus.Entry) UpdateCommentHandler {
	if comments == nil {
		log.Fatalln("nil comment repository")
	}
	if activities == nil {
		log.Fatalln("nil activity repository")
	}
	return decorator.ApplyCommandDecorators(updateCommentHandler{comments: comments, activities: activities}, logger)
}

func (h updateCommentHandler) Handle(ctx context.Context, cmd UpdateComment) error {
	op := errors.Op("cmd.UpdateComment")

	var taskID string
	if err := h.comments.UpdateByID(ctx, cmd.CommentID, func(ctx context.Context, c comment.Comment) (comment.Comment, error) {
		taskID = c.TaskID()
		return c, c.Update(cmd.EditorID, cmd.Content)
	}); err != nil {
		return errors.E(op, err)
	}

	a, err := activity.New(taskID, cmd.EditorID, activity.TypeCommentUpdated,
		map[string]any{"comment_id": cmd.CommentID})
	if err != nil {
		return errors.E(op, err)
	}

	return errors.E(op, h.activities.Add(ctx, a))
}
