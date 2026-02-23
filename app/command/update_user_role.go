package command

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/errkind"
	"clean-arch-go/domain/user"
	"context"
	"log"

	"github.com/sirupsen/logrus"
)

type UpdateUserRole struct {
	CurrentUserUUID string
	TargetUserUUID  string
	NewRole         string
}

type UpdateUserRoleHandler decorator.CommandHandler[UpdateUserRole]

type updateUserRoleHandler struct {
	users UserRepository
}

func NewUpdateUserRoleHandler(userRepository UserRepository, logger *logrus.Entry) UpdateUserRoleHandler {
	if userRepository == nil {
		log.Fatalln("nil userRepository")
	}

	handler := updateUserRoleHandler{
		users: userRepository,
	}

	return decorator.ApplyCommandDecorators(handler, logger)
}

func (h updateUserRoleHandler) Handle(ctx context.Context, cmd UpdateUserRole) error {
	op := errors.Op("cmd.UpdateUserRole")

	currentUser, err := h.users.FindById(ctx, cmd.CurrentUserUUID)
	if err != nil {
		return errors.E(op, err)
	}

	targetUser, err := h.users.FindById(ctx, cmd.TargetUserUUID)
	if err != nil {
		return errors.E(op, err)
	}

	if !currentUser.CanChangeRoleOf(targetUser) {
		return errors.E(op, errkind.Permission, errors.Str("not authorized to change role"))
	}

	newRole, err := user.UserRoleFromString(cmd.NewRole)
	if err != nil {
		return errors.E(op, errkind.Other, err)
	}

	err = h.users.UpdateByID(ctx, cmd.TargetUserUUID, func(_ context.Context, u user.User) (user.User, error) {
		u.ChangeRole(newRole)
		return u, nil
	})
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}
