package main

import (
	"clean-arch-go/adapters"
	"clean-arch-go/app"
	"clean-arch-go/app/command"
	"clean-arch-go/app/query"
	"clean-arch-go/core/logs"
	"clean-arch-go/core/tracer"
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

	shutdownTracer := tracer.SetupTracer()
	defer shutdownTracer()

	app := newApplication(mysqlDB, logger)

	ports.RunHTTPServer(func(router chi.Router) http.Handler {
		return ports.HandlerFromMux(ports.NewHttpHandler(app), router)
	})
}

func newApplication(db *sqlx.DB, logger *logrus.Entry) app.Application {
	id := adapters.NewID()
	taskRepository := adapters.NewMysqlTaskRepository(db)
	userRepository := adapters.NewMysqlUserRepository(db)

	application := app.Application{
		Commands: app.Commands{
			AddUser:          command.NewAddUserHandler(id, userRepository, logger),
			CreateTask:       command.NewCreateTaskHandler(id, taskRepository),
			ChangeTaskStatus: command.NewChangeTaskStatusHandler(taskRepository),
			AssignTask:       command.NewAssignTaskHandler(taskRepository, userRepository),
		},
		Queries: app.Queries{
			Tasks: query.NewTaskHandler(taskRepository, logger),
		},
	}

	return application
}
