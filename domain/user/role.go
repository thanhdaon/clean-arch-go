package user

import "errors"

var (
	RoleUnknow   = Role{}
	RoleEmployer = Role{"employer"}
	RoleEmployee = Role{"employee"}
)

type Role struct {
	slug string
}

func (r Role) String() string {
	return r.slug
}

func (r Role) IsZero() bool {
	return r == RoleUnknow
}

func UserRoleFromString(s string) (Role, error) {
	switch s {
	case RoleEmployer.slug:
		return RoleEmployer, nil
	case RoleEmployee.slug:
		return RoleEmployee, nil
	default:
		return RoleUnknow, errors.New("unknow role: " + s)
	}
}
