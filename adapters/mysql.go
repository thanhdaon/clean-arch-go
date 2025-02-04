package adapters

import (
	"clean-arch-go/app/query"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/errkind"
	"database/sql"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func NewMySQLConnection() (*sqlx.DB, error) {
	config := mysql.NewConfig()

	config.Net = "tcp"
	config.Addr = os.Getenv("MYSQL_ADDR")
	config.User = os.Getenv("MYSQL_USER")
	config.Passwd = os.Getenv("MYSQL_PASSWORD")
	config.DBName = os.Getenv("MYSQL_DATABASE")
	config.ParseTime = true // with that parameter, we can use time.Time in mysqlHour.Hour
	config.Loc = time.UTC

	traceDB, err := otelsql.Open("mysql", config.FormatDSN(), otelsql.WithAttributes(semconv.DBSystemMySQL))
	if err != nil {
		return nil, errors.E(errors.Op("connect-mysql"), errkind.Connection, err)
	}

	return sqlx.NewDb(traceDB, "mysql"), nil
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

func fromMysqlTasksToQueryTasks(items []MysqlTask) []query.Task {
	ret := []query.Task{}

	for _, item := range items {
		ret = append(ret, query.Task{
			UUID:       item.ID,
			Title:      item.Title,
			Status:     item.Status,
			CreatedBy:  item.CreatedBy,
			AssignedTo: item.AssignedTo,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  nullTimeToTime(item.UpdatedAt),
		})
	}

	return ret
}
