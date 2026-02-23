package query

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/core/errors"
	"context"
	"log"

	"github.com/sirupsen/logrus"
)

type UserQuery struct {
	UserUUID string
}

type UserHandler decorator.QueryHandler[UserQuery, UserDTO]

type userHandler struct {
	users UsersReadModel
}

func NewUserHandler(usersReadModel UsersReadModel, logger *logrus.Entry) UserHandler {
	if usersReadModel == nil {
		log.Fatalln("nil usersReadModel")
	}

	handler := userHandler{
		users: usersReadModel,
	}

	return decorator.ApplyQueryDecorators(handler, logger)
}

func (h userHandler) Handle(ctx context.Context, query UserQuery) (UserDTO, error) {
	op := errors.Op("query.User")

	domainUser, err := h.users.FindById(ctx, query.UserUUID)
	if err != nil {
		return UserDTO{}, errors.E(op, err)
	}

	return UserDTO{
		UUID:  domainUser.UUID(),
		Role:  domainUser.Role().String(),
		Name:  domainUser.Name(),
		Email: domainUser.Email(),
	}, nil
}
