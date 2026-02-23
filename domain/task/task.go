package task

import (
	"clean-arch-go/domain/user"
	"database/sql"
	"errors"
	"fmt"
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
	Unassign(remover user.User) error
	Reopen(opener user.User) error
	IsDeleted() bool
	DeletedAt() time.Time
	Delete(deleter user.User) error
	IsArchived() bool
	ArchivedAt() time.Time
	Archive(archiver user.User) error
}

type task struct {
	uuid string

	title string

	status Status

	createdBy  string
	assignedTo string

	createdAt  time.Time
	updatedAt  time.Time
	deletedAt  sql.NullTime
	archivedAt sql.NullTime
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
	if !t.status.CanTransitionTo(s) {
		return fmt.Errorf("cannot transition from %s to %s", t.status, s)
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

func (t *task) Unassign(remover user.User) error {
	if remover.Role() != user.RoleEmployer {
		return errors.New("only employer can unassign task")
	}

	t.assignedTo = ""
	t.updatedAt = time.Now()

	return nil
}

func (t *task) Reopen(opener user.User) error {
	if t.status != StatusCompleted {
		return errors.New("only completed tasks can be reopened")
	}
	if !t.allowToChangeStatus(opener) {
		return errors.New("user is not allowed to reopen this task")
	}
	t.status = StatusInProgress
	t.updatedAt = time.Now()
	return nil
}

func (t *task) IsDeleted() bool {
	return t.deletedAt.Valid
}

func (t *task) DeletedAt() time.Time {
	if t.deletedAt.Valid {
		return t.deletedAt.Time
	}
	return time.Time{}
}

func (t *task) Delete(deleter user.User) error {
	if deleter.Role() != user.RoleEmployer {
		return errors.New("only employer can delete task")
	}
	if deleter.UUID() != t.createdBy {
		return errors.New("only task creator can delete task")
	}
	t.deletedAt = sql.NullTime{Time: time.Now(), Valid: true}
	t.updatedAt = time.Now()
	return nil
}

func (t *task) IsArchived() bool {
	return t.archivedAt.Valid
}

func (t *task) ArchivedAt() time.Time {
	if t.archivedAt.Valid {
		return t.archivedAt.Time
	}
	return time.Time{}
}

func (t *task) Archive(archiver user.User) error {
	if archiver.Role() != user.RoleEmployer {
		return errors.New("only employer can archive task")
	}
	if t.status != StatusCompleted {
		return errors.New("only completed tasks can be archived")
	}
	t.archivedAt = sql.NullTime{Time: time.Now(), Valid: true}
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

func From(id, title, statusString, createdBy, assignedTo string, createdAt, updatedAt time.Time, deletedAt, archivedAt sql.NullTime) (Task, error) {
	status, err := StatusFromString(statusString)
	if err != nil {
		return nil, err
	}

	return &task{
		uuid:       id,
		title:      title,
		status:     status,
		createdBy:  createdBy,
		assignedTo: assignedTo,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
		deletedAt:  deletedAt,
		archivedAt: archivedAt,
	}, nil
}
