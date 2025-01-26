package errors

// Kind defines the kind of error this is, mostly for use by systems
// such as FUSE that must act differently depending on the error.
type Kind uint8

// Kinds of errors.
//
// The values of the error kinds are common between both
// clients and servers. Do not reorder this list or remove
// any items since that will change their values.
// New items must be added only to the end.
const (
	Other      Kind = iota // Unclassified error. This value is not printed in the error message.
	Permission             // Permission denied.
	Exist                  // Item already exists.
	NotExist               // Item does not exist.
	Internal               // Internal error or inconsistency.
)

func (k Kind) String() string {
	switch k {
	case Other:
		return "other error"
	case Permission:
		return "permission denied"
	case Exist:
		return "item already exists"
	case NotExist:
		return "item does not exist"
	case Internal:
		return "internal error"
	default:
		return "unknown error kind"
	}
}
