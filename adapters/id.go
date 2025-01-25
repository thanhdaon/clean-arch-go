package adapters

import (
	"github.com/oklog/ulid/v2"
)

type ID struct {
}

func NewID() ID {
	return ID{}
}

func (id ID) New() string {
	return ulid.Make().String()
}

func (id ID) IsValid(s string) bool {
	_, err := ulid.Parse(s)
	return err == nil
}
