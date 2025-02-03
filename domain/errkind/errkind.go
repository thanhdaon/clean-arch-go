package errkind

import "clean-arch-go/common/errors"

var (
	Other         = errors.NewKind("")
	Authorization = errors.NewKind("authorization error")
	Permission    = errors.NewKind("permission denied")
	Exist         = errors.NewKind("item already exists")
	NotExist      = errors.NewKind("item does not exist")
	Internal      = errors.NewKind("internal error")
	Connection    = errors.NewKind("pconnection error") // connection related error (ex. mysql, rabbitmq, ...)
)
