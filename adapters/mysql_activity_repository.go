package adapters

import (
	"clean-arch-go/app/query"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/activity"
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
)

type MysqlActivity struct {
	ID        string    `db:"id"`
	TaskID    string    `db:"task_id"`
	ActorID   string    `db:"actor_id"`
	Type      string    `db:"type"`
	Payload   []byte    `db:"payload"`
	CreatedAt time.Time `db:"created_at"`
}

type MysqlActivityRepository struct {
	db *sqlx.DB
}

func NewMysqlActivityRepository(db *sqlx.DB) MysqlActivityRepository {
	return MysqlActivityRepository{db: db}
}

func (r MysqlActivityRepository) Add(ctx context.Context, a activity.Activity) error {
	op := errors.Op("MysqlActivityRepository.Add")

	payload, err := json.Marshal(a.Payload())
	if err != nil {
		return errors.E(op, err)
	}

	row := MysqlActivity{
		ID:        a.UUID(),
		TaskID:    a.TaskID(),
		ActorID:   a.ActorID(),
		Type:      string(a.Type()),
		Payload:   payload,
		CreatedAt: a.CreatedAt(),
	}

	if _, err = r.db.NamedExecContext(ctx, `
		INSERT INTO task_activities (id, task_id, actor_id, type, payload, created_at)
		VALUES (:id, :task_id, :actor_id, :type, :payload, :created_at)
	`, row); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (r MysqlActivityRepository) ActivityForTask(ctx context.Context, taskID string, limit, offset int) ([]query.ActivityDTO, error) {
	op := errors.Op("MysqlActivityRepository.ActivityForTask")

	type row struct {
		MysqlActivity
		ActorName string `db:"actor_name"`
	}

	rows := []row{}
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT a.id, a.task_id, a.actor_id, a.type, a.payload, a.created_at, COALESCE(u.name, '') as actor_name
		FROM task_activities a
		LEFT JOIN users u ON u.id = a.actor_id
		WHERE a.task_id = ?
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?
	`, taskID, limit, offset); err != nil {
		return nil, errors.E(op, err)
	}

	result := make([]query.ActivityDTO, 0, len(rows))
	for _, r := range rows {
		var payload map[string]any
		_ = json.Unmarshal(r.Payload, &payload)
		result = append(result, query.ActivityDTO{
			UUID:      r.ID,
			TaskID:    r.TaskID,
			ActorID:   r.ActorID,
			ActorName: r.ActorName,
			Type:      r.Type,
			Payload:   payload,
			CreatedAt: r.CreatedAt,
		})
	}
	return result, nil
}
