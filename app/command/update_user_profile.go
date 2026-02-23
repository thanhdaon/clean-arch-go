package command

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/user"
	"context"
	"log"

	"github.com/sirupsen/logrus"
)

type UpdateUserProfile struct {
	UserUUID string
	Name     string
	Email    string
}

type UpdateUserProfileHandler decorator.CommandHandler[UpdateUserProfile]

type updateUserProfileHandler struct {
	users UserRepository
}

func NewUpdateUserProfileHandler(userRepository UserRepository, logger *logrus.Entry) UpdateUserProfileHandler {
	if userRepository == nil {
		log.Fatalln("nil userRepository")
	}

	handler := updateUserProfileHandler{
		users: userRepository,
	}

	return decorator.ApplyCommandDecorators(handler, logger)
}

func (h updateUserProfileHandler) Handle(ctx context.Context, cmd UpdateUserProfile) error {
	op := errors.Op("cmd.UpdateUserProfile")

	err := h.users.UpdateByID(ctx, cmd.UserUUID, func(_ context.Context, u user.User) (user.User, error) {
		u.UpdateProfile(cmd.Name, cmd.Email)
		return u, nil
	})
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}
