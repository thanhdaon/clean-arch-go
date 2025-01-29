package main

import (
	"clean-arch-go/adapters"
	"clean-arch-go/app"
	"clean-arch-go/app/query"
	"clean-arch-go/common/logs"
	"clean-arch-go/ports"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func main() {
	logs.Init()

	logger := logrus.NewEntry(logrus.StandardLogger())

	mysqlDB, err := adapters.NewMySQLConnection()
	if err != nil {
		logger.Fatalln("Can not connect to mysql", err)
	}

	app := newApplication(mysqlDB, logger)

	ports.RunHTTPServer(func(router chi.Router) http.Handler {
		return ports.HandlerFromMux(ports.NewHttpHandler(app), router)
	})
}

func newApplication(db *sqlx.DB, logger *logrus.Entry) app.Application {

	taskRepository := adapters.NewMysqlTaskRepository(db)

	application := app.Application{
		Queries: app.Queries{
			Tasks: query.NewTaskHandler(taskRepository, logger),
		},
	}

	return application
}
