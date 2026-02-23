package query

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/core/errors"
	"context"
	"log"

	"github.com/sirupsen/logrus"
)

type UsersQuery struct{}

type UsersHandler decorator.QueryHandler[UsersQuery, []UserDTO]

type usersHandler struct {
	users UsersReadModel
}

func NewUsersHandler(usersReadModel UsersReadModel, logger *logrus.Entry) UsersHandler {
	if usersReadModel == nil {
		log.Fatalln("nil usersReadModel")
	}

	handler := usersHandler{
		users: usersReadModel,
	}

	return decorator.ApplyQueryDecorators(handler, logger)
}

func (h usersHandler) Handle(ctx context.Context, _ UsersQuery) ([]UserDTO, error) {
	op := errors.Op("query.Users")

	domainUsers, err := h.users.FindAll(ctx)
	if err != nil {
		return nil, errors.E(op, err)
	}

	usersDTO := make([]UserDTO, len(domainUsers))
	for i, u := range domainUsers {
		usersDTO[i] = UserDTO{
			UUID:  u.UUID(),
			Role:  u.Role().String(),
			Name:  u.Name(),
			Email: u.Email(),
		}
	}

	return usersDTO, nil
}
