package main

import (
	"clean-arch-go/adapters"
	"clean-arch-go/app"
	"clean-arch-go/app/query"
	"clean-arch-go/common/logs"
	"clean-arch-go/ports"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

func main() {
	logs.Init()

	app, cleanup := newApplication()
	defer cleanup()

	ports.RunHTTPServer(func(router chi.Router) http.Handler {
		return ports.HandlerFromMux(ports.NewHttpHandler(app), router)
	})
}

func newApplication() (app.Application, func()) {
	logger := logrus.NewEntry(logrus.StandardLogger())

	taskRepository := adapters.NewTaskPgRepository()

	application := app.Application{
		Queries: app.Queries{
			Tasks: query.NewTaskHandler(taskRepository, logger),
		},
	}

	return application, func() {}
}
