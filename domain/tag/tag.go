package tag

import (
	"errors"

	"github.com/google/uuid"
)

type Tag interface {
	UUID() string
	Name() string
	TaskID() string
}

type tag struct {
	uuid   string
	name   string
	taskID string
}

func (t *tag) UUID() string   { return t.uuid }
func (t *tag) Name() string   { return t.name }
func (t *tag) TaskID() string { return t.taskID }

func NewTag(taskID, name string) (Tag, error) {
	if taskID == "" {
		return nil, errors.New("empty task id")
	}
	if name == "" {
		return nil, errors.New("empty tag name")
	}
	return &tag{
		uuid:   uuid.New().String(),
		name:   name,
		taskID: taskID,
	}, nil
}

func From(id, taskID, name string) (Tag, error) {
	if id == "" {
		return nil, errors.New("empty tag id")
	}
	if taskID == "" {
		return nil, errors.New("empty task id")
	}
	if name == "" {
		return nil, errors.New("empty tag name")
	}
	return &tag{
		uuid:   id,
		name:   name,
		taskID: taskID,
	}, nil
}
