package task

import (
	"clean-arch-go/domain/user"
	"errors"
	"time"
)

type Task interface {
	UUID() string
	Title() string
	Status() Status
	CreatedBy() string
	AssignedTo() string
	CreatedAt() time.Time
	UpdatedAt() time.Time
	UpdateTitle(user.User, string) error
	ChangeStatus(user.User, Status) error
	AssignTo(assigner user.User, assignee user.User) error
}

type task struct {
	uuid string

	title string

	status Status

	createdBy  string
	assignedTo string

	createdAt time.Time
	updatedAt time.Time
}

func (t *task) UUID() string {
	return t.uuid
}

func (t *task) Title() string {
	return t.title
}

func (t *task) Status() Status {
	return t.status
}

func (t *task) CreatedBy() string {
	return t.createdBy
}

func (t *task) AssignedTo() string {
	return t.assignedTo
}

func (t *task) CreatedAt() time.Time {
	return t.createdAt
}

func (t *task) UpdatedAt() time.Time {
	return t.updatedAt
}

func (t *task) ChangeStatus(updater user.User, s Status) error {
	if s.IsZero() {
		return errors.New("cannot update status of task to empty")
	}

	if allow := t.allowToChangeStatus(updater); !allow {
		return errors.New("user is not allow to update status of this task")
	}

	t.status = s
	t.updatedAt = time.Now()

	return nil
}

func (t *task) allowToChangeStatus(updater user.User) bool {
	if updater.Role() == user.RoleEmployer {
		return true
	}

	if updater.UUID() == t.assignedTo {
		return true
	}

	return false
}

func (t *task) UpdateTitle(updater user.User, title string) error {
	if title == "" {
		return errors.New("Cannot update task title to empty")
	}

	if allow := t.allowToUpdateTitle(updater); !allow {
		return errors.New("user is not allow to update this task title")
	}

	t.title = title
	t.updatedAt = time.Now()

	return nil
}

func (t *task) allowToUpdateTitle(updater user.User) bool {
	if updater.Role() == user.RoleEmployer {
		return true
	}

	if updater.UUID() == t.assignedTo {
		return true
	}

	return false
}

func (t *task) AssignTo(assigner user.User, assignee user.User) error {
	if assignee.Role() != user.RoleEmployer {
		return errors.New("only employer role can assign task")
	}

	t.assignedTo = assignee.UUID()
	t.updatedAt = time.Now()

	return nil
}

func NewTask(creator user.User, uuid, title string) (Task, error) {
	if role := creator.Role(); role != user.RoleEmployer {
		return nil, errors.New("only employ can create task")
	}

	if uuid == "" {
		return nil, errors.New("empty task uuid")
	}

	if title == "" {
		return nil, errors.New("empty task title")
	}

	return &task{
		uuid:      uuid,
		title:     title,
		status:    StatusTodo,
		createdBy: creator.UUID(),
		createdAt: time.Now(),
	}, nil
}
