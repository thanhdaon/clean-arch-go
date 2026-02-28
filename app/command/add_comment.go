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

type AddComment struct {
	TaskID   string
	AuthorID string
	Content  string
}

type AddCommentHandler decorator.CommandHandler[AddComment]

type addCommentHandler struct {
	comments   CommentRepository
	activities ActivityRepository
}

func NewAddCommentHandler(comments CommentRepository, activities ActivityRepository, logger *logrus.Entry) AddCommentHandler {
	if comments == nil {
		log.Fatalln("nil comment repository")
	}
	if activities == nil {
		log.Fatalln("nil activity repository")
	}
	return decorator.ApplyCommandDecorators(addCommentHandler{comments: comments, activities: activities}, logger)
}

func (h addCommentHandler) Handle(ctx context.Context, cmd AddComment) error {
	op := errors.Op("cmd.AddComment")

	c, err := comment.New(cmd.TaskID, cmd.AuthorID, cmd.Content)
	if err != nil {
		return errors.E(op, err)
	}

	if err = h.comments.Add(ctx, c); err != nil {
		return errors.E(op, err)
	}

	payload := map[string]any{"comment_id": c.UUID()}
	a, err := activity.New(cmd.TaskID, cmd.AuthorID, activity.TypeCommentAdded, payload)
	if err != nil {
		return errors.E(op, err)
	}

	if err := h.activities.Add(ctx, a); err != nil {
		return errors.E(op, err)
	}

	return nil
}
