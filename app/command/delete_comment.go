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

type DeleteComment struct {
	CommentID string
	DeleterID string
	TaskID    string
}

type DeleteCommentHandler decorator.CommandHandler[DeleteComment]

type deleteCommentHandler struct {
	comments   CommentRepository
	activities ActivityRepository
}

func NewDeleteCommentHandler(comments CommentRepository, activities ActivityRepository, logger *logrus.Entry) DeleteCommentHandler {
	if comments == nil {
		log.Fatalln("nil comment repository")
	}
	if activities == nil {
		log.Fatalln("nil activity repository")
	}
	return decorator.ApplyCommandDecorators(deleteCommentHandler{comments: comments, activities: activities}, logger)
}

func (h deleteCommentHandler) Handle(ctx context.Context, cmd DeleteComment) error {
	op := errors.Op("cmd.DeleteComment")

	if err := h.comments.UpdateByID(ctx, cmd.CommentID, func(ctx context.Context, c comment.Comment) (comment.Comment, error) {
		return c, c.Delete(cmd.DeleterID)
	}); err != nil {
		return errors.E(op, err)
	}

	a, err := activity.New(cmd.TaskID, cmd.DeleterID, activity.TypeCommentDeleted,
		map[string]any{"comment_id": cmd.CommentID})
	if err != nil {
		return errors.E(op, err)
	}

	return errors.E(op, h.activities.Add(ctx, a))
}
