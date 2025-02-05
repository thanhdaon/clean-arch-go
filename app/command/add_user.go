package command

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/user"
	"context"
	"log"

	"github.com/sirupsen/logrus"
)

type AddUser struct {
	Role string
}

type AddUserHandler decorator.CommandHandler[AddUser]

type addUserHandler struct {
	id     ID
	users  UserRepository
	videos VideoService
}

func NewAddUserHandler(id ID, userRepository UserRepository, videoService VideoService, logger *logrus.Entry) AddUserHandler {
	if id == nil {
		log.Fatalln("nil id")
	}

	if userRepository == nil {
		log.Fatalln("nil userRepository")
	}

	if videoService == nil {
		log.Fatalln("nil videoService")
	}

	handler := addUserHandler{
		id:     id,
		users:  userRepository,
		videos: videoService,
	}

	return decorator.ApplyCommandDecorators(handler, logger)
}

func (h addUserHandler) Handle(ctx context.Context, cmd AddUser) error {
	op := errors.Op("cmd.AddUser")

	domainUser, err := user.From(h.id.New(), cmd.Role)
	if err != nil {
		return errors.E(op, err)
	}

	if err := h.users.Add(ctx, domainUser); err != nil {
		return errors.E(op, err)
	}

	if err := h.videos.GetAll(ctx); err != nil {
		return errors.E(op, err)
	}

	return nil
}
