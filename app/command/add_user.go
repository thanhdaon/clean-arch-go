package command

import (
	"clean-arch-go/common/errors"
	"clean-arch-go/domain/user"
	"context"
	"log"
)

type AddUser struct {
	Role string
}

type AddUserHandler struct {
	id    ID
	users UserRepository
}

func NewAddUserHandler(id ID, userRepository UserRepository) AddUserHandler {
	if id == nil {
		log.Fatalln("nil id")
	}

	if userRepository == nil {
		log.Fatalln("nil userRepository")
	}

	return AddUserHandler{
		id:    id,
		users: userRepository,
	}
}

func (h AddUserHandler) Handle(ctx context.Context, cmd AddUser) error {
	op := errors.Op("cmd.AddUser")

	domainUser, err := user.From(h.id.New(), cmd.Role)
	if err != nil {
		return errors.E(op, err)
	}

	if err := h.users.Add(ctx, domainUser); err != nil {
		return errors.E(op, err)
	}

	return nil
}
