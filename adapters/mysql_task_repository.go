package adapters

import (
	"clean-arch-go/app/command"
	"clean-arch-go/app/query"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/errkind"
	"clean-arch-go/domain/tag"
	"clean-arch-go/domain/task"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type MysqlTask struct {
	ID          string         `db:"id"`
	Title       string         `db:"title"`
	Status      string         `db:"status"`
	CreatedBy   string         `db:"created_by"`
	AssignedTo  string         `db:"assigned_to"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
	DeletedAt   sql.NullTime   `db:"deleted_at"`
	ArchivedAt  sql.NullTime   `db:"archived_at"`
	Priority    int            `db:"priority"`
	DueDate     sql.NullTime   `db:"due_date"`
	Description sql.NullString `db:"description"`
}

type MysqlTaskRepository struct {
	db *sqlx.DB
}

func NewMysqlTaskRepository(db *sqlx.DB) MysqlTaskRepository {
	return MysqlTaskRepository{db: db}
}

func (r MysqlTaskRepository) Add(ctx context.Context, t task.Task) error {
	op := errors.Op("MysqlTaskRepository.Add")

	added := MysqlTask{
		ID:          t.UUID(),
		Title:       t.Title(),
		Status:      t.Status().String(),
		CreatedBy:   t.CreatedBy(),
		AssignedTo:  t.AssignedTo(),
		CreatedAt:   t.CreatedAt(),
		UpdatedAt:   timeToNullTime(t.UpdatedAt()),
		DeletedAt:   timeToNullTime(t.DeletedAt()),
		ArchivedAt:  timeToNullTime(t.ArchivedAt()),
		Priority:    int(t.Priority()),
		DueDate:     timeToNullTime(t.DueDate()),
		Description: stringToNullString(t.Description()),
	}

	query := `
		INSERT INTO tasks
			(id, title, status, created_by, assigned_to, created_at, updated_at, deleted_at, archived_at, priority, due_date, description)
		VALUES
			(:id, :title, :status, :created_by, :assigned_to, :created_at, :updated_at, :deleted_at, :archived_at, :priority, :due_date, :description)
	`

	if _, err := r.db.NamedExecContext(ctx, query, added); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (r MysqlTaskRepository) UpdateByID(ctx context.Context, uuid string, updateFn command.TaskUpdater) error {
	op := errors.Op("MysqlTaskRepository.UpdateByID")

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.E(op, err)
	}

	defer func() {
		err = r.finishTransaction(err, tx)
	}()

	existingTask, err := r.FindById(ctx, uuid)
	if err != nil {
		if errors.Is(errkind.NotExist, err) {
			return errors.E(op, errkind.NotExist, err)
		}

		return errors.E(op, err)
	}

	updatedTask, err := updateFn(ctx, existingTask)
	if err != nil {
		return errors.E(op, err)
	}

	updated := MysqlTask{
		ID:          updatedTask.UUID(),
		Title:       updatedTask.Title(),
		Status:      updatedTask.Status().String(),
		CreatedBy:   updatedTask.CreatedBy(),
		AssignedTo:  updatedTask.AssignedTo(),
		CreatedAt:   updatedTask.CreatedAt(),
		UpdatedAt:   timeToNullTime(updatedTask.UpdatedAt()),
		Priority:    int(updatedTask.Priority()),
		DueDate:     timeToNullTime(updatedTask.DueDate()),
		Description: stringToNullString(updatedTask.Description()),
		DeletedAt:   timeToNullTime(updatedTask.DeletedAt()),
		ArchivedAt:  timeToNullTime(updatedTask.ArchivedAt()),
	}

	query := `
		UPDATE tasks
		SET
			title = :title,
			status = :status,
			created_by = :created_by,
			assigned_to = :assigned_to,
			created_at = :created_at,
			updated_at = :updated_at,
			priority = :priority,
			due_date = :due_date,
			description = :description,
			deleted_at = :deleted_at,
			archived_at = :archived_at
		WHERE id = :id;
	`
	result, err := r.db.NamedExecContext(ctx, query, updated)
	if err != nil {
		return errors.E(op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.E(op, err)
	}

	if rowsAffected == 0 {
		return errors.E(op, errkind.NotExist, fmt.Errorf("task with id %s not found", uuid))
	}

	return nil
}

func (r MysqlTaskRepository) AllTasks(ctx context.Context) ([]query.Task, error) {
	op := errors.Op("MysqlTaskRepository.AllTasks")

	data := []MysqlTask{}
	query := "SELECT * FROM `tasks` WHERE `deleted_at` IS NULL AND `archived_at` IS NULL"

	if err := r.db.SelectContext(ctx, &data, query); err != nil {
		return nil, errors.E(op, err)
	}

	return fromMysqlTasksToQueryTasks(data), nil
}

func (r MysqlTaskRepository) FindById(ctx context.Context, uuid string) (task.Task, error) {
	op := errors.Op("MysqlTaskRepository.FindById")

	data := MysqlTask{}
	query := "SELECT * FROM `tasks` WHERE `id` = ? AND `deleted_at` IS NULL AND `archived_at` IS NULL"

	if err := r.db.GetContext(ctx, &data, query, uuid); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(op, errkind.NotExist, err)
		}

		return nil, errors.E(op, err)
	}

	domainTask, err := task.From(
		data.ID, data.Title, data.Status, data.CreatedBy, data.AssignedTo,
		data.CreatedAt, nullTimeToTime(data.UpdatedAt), data.DeletedAt, data.ArchivedAt,
		data.Priority, data.DueDate, data.Description,
	)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return domainTask, nil
}

func (r MysqlTaskRepository) FindTasksForUser(ctx context.Context, userUUID string) ([]query.Task, error) {
	return []query.Task{}, errors.E(errors.Op("task.FindTasksForUser"), fmt.Errorf("dump"))
}

func (r MysqlTaskRepository) RemoveAllTasks(ctx context.Context) error {
	op := errors.Op("MysqlTaskRepository.RemoveAllTasks")

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.E(op, err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return errors.E(op, err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM task_tags"); err != nil {
		return errors.E(op, err)
	}

	query := `TRUNCATE TABLE tasks`

	if _, err := tx.ExecContext(ctx, query); err != nil {
		return errors.E(op, err)
	}

	if _, err := tx.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		return errors.E(op, err)
	}

	if err := tx.Commit(); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (r MysqlTaskRepository) AddTag(ctx context.Context, t tag.Tag) error {
	op := errors.Op("MysqlTaskRepository.AddTag")

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

func (r MysqlTaskRepository) RemoveTag(ctx context.Context, tagID string) error {
	op := errors.Op("MysqlTaskRepository.RemoveTag")

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
		return errors.E(op, errkind.NotExist, fmt.Errorf("tag with id %s not found", tagID))
	}

	return nil
}

func (r MysqlTaskRepository) finishTransaction(err error, tx *sqlx.Tx) error {
	op := errors.Op("MysqlTaskRepository.finishTransaction")

	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.E(op, rollbackErr)
		}

		return errors.E(op, err)
	} else {
		if commitErr := tx.Commit(); commitErr != nil {
			return errors.E(op, "failed to commit transaction")
		}

		return nil
	}
}
