package activity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Type string

const (
	TypeStatusChanged   Type = "status_changed"
	TypeAssigned        Type = "assigned"
	TypeUnassigned      Type = "unassigned"
	TypeTitleUpdated    Type = "title_updated"
	TypePriorityChanged Type = "priority_changed"
	TypeDueDateSet      Type = "due_date_set"
	TypeDescriptionSet  Type = "description_set"
	TypeArchived        Type = "archived"
	TypeReopened        Type = "reopened"
	TypeCommentAdded    Type = "comment_added"
	TypeCommentUpdated  Type = "comment_updated"
	TypeCommentDeleted  Type = "comment_deleted"
)

type Activity interface {
	UUID() string
	TaskID() string
	ActorID() string
	Type() Type
	Payload() map[string]any
	CreatedAt() time.Time
}

type activityRecord struct {
	uuid      string
	taskID    string
	actorID   string
	actType   Type
	payload   map[string]any
	createdAt time.Time
}

func (a *activityRecord) UUID() string            { return a.uuid }
func (a *activityRecord) TaskID() string          { return a.taskID }
func (a *activityRecord) ActorID() string         { return a.actorID }
func (a *activityRecord) Type() Type              { return a.actType }
func (a *activityRecord) Payload() map[string]any { return a.payload }
func (a *activityRecord) CreatedAt() time.Time    { return a.createdAt }

func New(taskID, actorID string, t Type, payload map[string]any) (Activity, error) {
	if taskID == "" {
		return nil, errors.New("empty task id")
	}
	if actorID == "" {
		return nil, errors.New("empty actor id")
	}
	return &activityRecord{
		uuid:      uuid.New().String(),
		taskID:    taskID,
		actorID:   actorID,
		actType:   t,
		payload:   payload,
		createdAt: time.Now(),
	}, nil
}

func From(id, taskID, actorID string, t Type, payload map[string]any, createdAt time.Time) (Activity, error) {
	if id == "" {
		return nil, errors.New("empty id")
	}
	if taskID == "" {
		return nil, errors.New("empty task id")
	}
	if actorID == "" {
		return nil, errors.New("empty actor id")
	}
	return &activityRecord{
		uuid:      id,
		taskID:    taskID,
		actorID:   actorID,
		actType:   t,
		payload:   payload,
		createdAt: createdAt,
	}, nil
}
