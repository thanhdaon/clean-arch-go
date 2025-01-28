package adapters

import (
	"clean-arch-go/app/query"
	"clean-arch-go/domain/errors"
	"clean-arch-go/domain/task"
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
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
		return errors.E(op, errors.Internal, err)
	}

	return nil
}

func (r MysqlTaskRepository) UpdateByID(ctx context.Context, uuid string, updateFn TaskUpdater) error {

	return nil
}

func (r MysqlTaskRepository) FindTasks(ctx context.Context) ([]query.Task, error) {
	return []query.Task{}, nil
}

func (r MysqlTaskRepository) FindById(ctx context.Context, uuid string) (task.Task, error) {
	op := errors.Op("MysqlTaskRepository.FindById")

	data := MysqlTask{}
	query := "SELECT * FROM `tasks` WHERE `id` = ?"

	if err := r.db.GetContext(ctx, &data, query, uuid); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(op, errors.NotExist, err)
		}

		return nil, errors.E(op, errors.Internal, err)
	}

	domainTask, err := task.From(
		data.ID, data.Title, data.Status, data.CreatedBy, data.AssignedTo,
		data.CreatedAt, nullTimeToTime(data.UpdatedAt),
	)
	if err != nil {
		return nil, errors.E(op, errors.Internal, err)
	}

	return domainTask, nil
}

func (r MysqlTaskRepository) FindTasksForUser(ctx context.Context, userUUID string) ([]query.Task, error) {
	return []query.Task{}, errors.E(errors.Op("task.FindTasksForUser"), errors.Internal, fmt.Errorf("dump"))
}

type TaskUpdater func(context.Context, task.Task) (task.Task, error)

func NewMySQLConnection() (*sqlx.DB, error) {
	config := mysql.NewConfig()

	config.Net = "tcp"
	config.Addr = os.Getenv("MYSQL_ADDR")
	config.User = os.Getenv("MYSQL_USER")
	config.Passwd = os.Getenv("MYSQL_PASSWORD")
	config.DBName = os.Getenv("MYSQL_DATABASE")
	config.ParseTime = true // with that parameter, we can use time.Time in mysqlHour.Hour

	db, err := sqlx.Connect("mysql", config.FormatDSN())
	if err != nil {
		return nil, errors.E(errors.Op("connect-mysql"), errors.Connection, err)
	}

	return db, nil
}

func timeToNullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func nullTimeToTime(nt sql.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}
