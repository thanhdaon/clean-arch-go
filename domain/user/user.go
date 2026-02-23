package user

import "errors"

type User interface {
	UUID() string
	Role() Role
	Name() string
	Email() string
	PasswordHash() string
	ChangeRole(Role)
	UpdateProfile(name, email string)
	SetPasswordHash(hash string)
	CanChangeRoleOf(target User) bool
}

type user struct {
	uuid         string
	role         Role
	name         string
	email        string
	passwordHash string
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

func (u *user) Name() string {
	return u.name
}

func (u *user) Email() string {
	return u.email
}

func (u *user) PasswordHash() string {
	return u.passwordHash
}

func (u *user) UpdateProfile(name, email string) {
	u.name = name
	u.email = email
}

func (u *user) SetPasswordHash(hash string) {
	u.passwordHash = hash
}

func (u *user) CanChangeRoleOf(target User) bool {
	return u.role == RoleAdmin && u.uuid != target.UUID()
}

func NewUser(uuid string, role Role, name, email string) (User, error) {
	if uuid == "" {
		return nil, errors.New("missing user uuid")
	}

	if role.IsZero() {
		return nil, errors.New("missing user role")
	}

	if name == "" {
		return nil, errors.New("missing user name")
	}

	if email == "" {
		return nil, errors.New("missing user email")
	}

	return &user{uuid: uuid, role: role, name: name, email: email}, nil
}

func From(uuid, roleString, name, email, passwordHash string) (User, error) {
	if uuid == "" {
		return nil, errors.New("missing user uuid")
	}

	role, err := UserRoleFromString(roleString)
	if err != nil {
		return nil, errors.New("invalid role")
	}

	return &user{uuid: uuid, role: role, name: name, email: email, passwordHash: passwordHash}, nil
}
