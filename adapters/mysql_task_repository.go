package adapters

import (
	"clean-arch-go/app/query"
	"clean-arch-go/common/errors"
	"clean-arch-go/domain/errkind"
	"clean-arch-go/domain/task"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type MysqlTask struct {
	ID         string       `db:"id"`
	Title      string       `db:"title"`
	Status     string       `db:"status"`
	CreatedBy  string       `db:"created_by"`
	AssignedTo string       `db:"assigned_to"`
	CreatedAt  time.Time    `db:"created_at"`
	UpdatedAt  sql.NullTime `db:"updated_at"`
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
		ID:         t.UUID(),
		Title:      t.Title(),
		Status:     t.Status().String(),
		CreatedBy:  t.CreatedBy(),
		AssignedTo: t.AssignedTo(),
		CreatedAt:  t.CreatedAt(),
		UpdatedAt:  timeToNullTime(t.UpdatedAt()),
	}

	query := `
		INSERT INTO tasks
			(id, title, status, created_by, assigned_to, created_at, updated_at) 
		VALUES 
			(:id, :title, :status, :created_by, :assigned_to, :created_at, :updated_at)
	`

	if _, err := r.db.NamedExec(query, added); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (r MysqlTaskRepository) UpdateByID(ctx context.Context, uuid string, updateFn func(context.Context, task.Task) (task.Task, error)) error {
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
		ID:         updatedTask.UUID(),
		Title:      updatedTask.Title(),
		Status:     updatedTask.Status().String(),
		CreatedBy:  updatedTask.CreatedBy(),
		AssignedTo: updatedTask.AssignedTo(),
		CreatedAt:  updatedTask.CreatedAt(),
		UpdatedAt:  timeToNullTime(updatedTask.UpdatedAt()),
	}

	query := `
		UPDATE tasks 
		SET 
			title = :title,
			status = :status,
			created_by = :created_by,
			assigned_to = :assigned_to,
			created_at = :created_at,
			updated_at = :updated_at
		WHERE id = :id;
	`
	result, err := r.db.NamedExec(query, updated)
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
	query := "SELECT * FROM `tasks`"

	if err := r.db.SelectContext(ctx, &data, query); err != nil {
		return nil, errors.E(op, err)
	}

	return fromMysqlTasksToQueryTasks(data), nil
}

func (r MysqlTaskRepository) FindById(ctx context.Context, uuid string) (task.Task, error) {
	op := errors.Op("MysqlTaskRepository.FindById")

	data := MysqlTask{}
	query := "SELECT * FROM `tasks` WHERE `id` = ?"

	if err := r.db.GetContext(ctx, &data, query, uuid); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(op, errkind.NotExist, err)
		}

		return nil, errors.E(op, err)
	}

	domainTask, err := task.From(
		data.ID, data.Title, data.Status, data.CreatedBy, data.AssignedTo,
		data.CreatedAt, nullTimeToTime(data.UpdatedAt),
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

	query := `TRUNCATE TABLE tasks`

	if _, err := r.db.ExecContext(ctx, query); err != nil {
		return errors.E(op, err)
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
