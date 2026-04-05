package query

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/errkind"
	"clean-arch-go/domain/user"
	"context"
	"log"

	"github.com/sirupsen/logrus"
)

type Login struct {
	Email    string
	Password string
}

type LoginResponse struct {
	Token string
	User  UserDTO
}

type LoginHandler decorator.QueryHandler[Login, LoginResponse]

type loginHandler struct {
	users UsersReadModel
	auth  AuthService
}

type UsersReadModel interface {
	FindByEmail(ctx context.Context, email string) (user.User, error)
	FindById(ctx context.Context, uuid string) (user.User, error)
	FindAll(ctx context.Context) ([]user.User, error)
}

type AuthService interface {
	CreateIDToken(data map[string]any) (string, error)
}

func NewLoginHandler(usersReadModel UsersReadModel, authService AuthService, logger *logrus.Entry) LoginHandler {
	if usersReadModel == nil {
		log.Fatalln("nil usersReadModel")
	}

	if authService == nil {
		log.Fatalln("nil authService")
	}

	handler := loginHandler{
		users: usersReadModel,
		auth:  authService,
	}

	return decorator.ApplyQueryDecorators(handler, logger)
}

func (h loginHandler) Handle(ctx context.Context, query Login) (LoginResponse, error) {
	op := errors.Op("query.Login")

	domainUser, err := h.users.FindByEmail(ctx, query.Email)
	if err != nil {
		return LoginResponse{}, errors.E(op, errkind.Authorization, errors.Str("invalid email or password"))
	}

	if !user.VerifyPassword(query.Password, domainUser.PasswordHash()) {
		return LoginResponse{}, errors.E(op, errkind.Authorization, errors.Str("invalid email or password"))
	}

	token, err := h.auth.CreateIDToken(map[string]any{
		"user_uuid": domainUser.UUID(),
		"user_role": domainUser.Role().String(),
	})
	if err != nil {
		return LoginResponse{}, errors.E(op, err)
	}

	return LoginResponse{
		Token: token,
		User: UserDTO{
			UUID:  domainUser.UUID(),
			Role:  domainUser.Role().String(),
			Name:  domainUser.Name(),
			Email: domainUser.Email(),
		},
	}, nil
}
