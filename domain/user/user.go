package user

import "errors"

type User interface {
	UUID() string
	Role() Role
	ChangeRole(Role)
}

type user struct {
	uuid string
	role Role
}

func (u *user) UUID() string {
	return u.uuid
}

func (u *user) Role() Role {
	return u.role
}

func (u *user) ChangeRole(role Role) {
	u.role = role
}

func NewUser(uuid string, role Role) (User, error) {
	if uuid == "" {
		return nil, errors.New("missing user uuid")
	}

	if role.IsZero() {
		return nil, errors.New("missing user role")
	}

	return &user{uuid: uuid, role: role}, nil
}

func From(uuid, roleString string) (User, error) {
	if uuid == "" {
		return nil, errors.New("missing user uuid")
	}

	role, err := UserRoleFromString(roleString)
	if err != nil {
		return nil, errors.New("invalid role")
	}

	return &user{uuid: uuid, role: role}, nil
}
