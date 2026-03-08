package user

import "errors"

var (
	RoleUnknown  = Role{}
	RoleAdmin    = Role{"admin"}
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
	return r == RoleUnknown
}

func UserRoleFromString(s string) (Role, error) {
	switch s {
	case RoleAdmin.slug:
		return RoleAdmin, nil
	case RoleEmployer.slug:
		return RoleEmployer, nil
	case RoleEmployee.slug:
		return RoleEmployee, nil
	default:
		return RoleUnknown, errors.New("unknown role: " + s)
	}
}
