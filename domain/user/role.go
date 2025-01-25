package user

import "errors"

var (
	RoleUnknow   = UserRole{}
	RoleEmployer = UserRole{"employer"}
	RoleEmployee = UserRole{"employee"}
)

type UserRole struct {
	slug string
}

func (r UserRole) String() string {
	return r.slug
}

func (r UserRole) IsZero() bool {
	return r == RoleUnknow
}

func UserRoleFromString(s string) (UserRole, error) {
	switch s {
	case RoleEmployer.slug:
		return RoleEmployer, nil
	case RoleEmployee.slug:
		return RoleEmployee, nil
	default:
		return RoleUnknow, errors.New("unknow role: " + s)
	}
}
