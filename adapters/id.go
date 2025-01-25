package adapters

type ID struct {
}

func NewID() ID {
	return ID{}
}

func (id ID) New() string {
	return ""
}

func (id ID) IsValid(s string) bool {
	return true
}
