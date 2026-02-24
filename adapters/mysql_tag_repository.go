package adapters

import (
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/errkind"
	"clean-arch-go/domain/tag"
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type MysqlTag struct {
	ID        string    `db:"id"`
	TaskID    string    `db:"task_id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

type MysqlTagRepository struct {
	db *sqlx.DB
}

func NewMysqlTagRepository(db *sqlx.DB) MysqlTagRepository {
	return MysqlTagRepository{db: db}
}

func (r MysqlTagRepository) Add(ctx context.Context, t tag.Tag) error {
	op := errors.Op("MysqlTagRepository.Add")

	row := MysqlTag{
		ID:        t.UUID(),
		TaskID:    t.TaskID(),
		Name:      t.Name(),
		CreatedAt: time.Now(),
	}

	query := `
		INSERT INTO task_tags (id, task_id, name, created_at)
		VALUES (:id, :task_id, :name, :created_at)
	`

	if _, err := r.db.NamedExecContext(ctx, query, row); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (r MysqlTagRepository) Remove(ctx context.Context, tagID string) error {
	op := errors.Op("MysqlTagRepository.Remove")

	query := `DELETE FROM task_tags WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, tagID)
	if err != nil {
		return errors.E(op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.E(op, err)
	}

	if rowsAffected == 0 {
		return errors.E(op, errkind.NotExist, errors.Str("tag not found"))
	}

	return nil
}

func (r MysqlTagRepository) FindById(ctx context.Context, uuid string) (tag.Tag, error) {
	op := errors.Op("MysqlTagRepository.FindById")

	data := MysqlTag{}
	query := `SELECT * FROM task_tags WHERE id = ?`

	if err := r.db.GetContext(ctx, &data, query, uuid); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(op, errkind.NotExist, err)
		}
		return nil, errors.E(op, err)
	}

	return tag.From(data.ID, data.TaskID, data.Name)
}

func (r MysqlTagRepository) FindByTaskId(ctx context.Context, taskID string) ([]tag.Tag, error) {
	op := errors.Op("MysqlTagRepository.FindByTaskId")

	data := []MysqlTag{}
	query := `SELECT * FROM task_tags WHERE task_id = ?`

	if err := r.db.SelectContext(ctx, &data, query, taskID); err != nil {
		return nil, errors.E(op, err)
	}

	tags := make([]tag.Tag, 0, len(data))
	for _, row := range data {
		t, err := tag.From(row.ID, row.TaskID, row.Name)
		if err != nil {
			return nil, errors.E(op, err)
		}
		tags = append(tags, t)
	}

	return tags, nil
}
