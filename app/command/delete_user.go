package command

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/core/errors"
	"context"
	"log"

	"github.com/sirupsen/logrus"
)

type DeleteUser struct {
	UserUUID string
}

type DeleteUserHandler decorator.CommandHandler[DeleteUser]

type deleteUserHandler struct {
	users UserRepository
}

func NewDeleteUserHandler(userRepository UserRepository, logger *logrus.Entry) DeleteUserHandler {
	if userRepository == nil {
		log.Fatalln("nil userRepository")
	}

	handler := deleteUserHandler{
		users: userRepository,
	}

	return decorator.ApplyCommandDecorators(handler, logger)
}

func (h deleteUserHandler) Handle(ctx context.Context, cmd DeleteUser) error {
	op := errors.Op("cmd.DeleteUser")

	if err := h.users.DeleteByID(ctx, cmd.UserUUID); err != nil {
		return errors.E(op, err)
	}

	return nil
}
