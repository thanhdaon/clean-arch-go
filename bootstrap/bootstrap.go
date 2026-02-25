package bootstrap

import (
	"context"
	"errors"
	"fmt"
	stdHTTP "net/http"

	"clean-arch-go/adapters"
	"clean-arch-go/app"
	"clean-arch-go/app/command"
	"clean-arch-go/app/query"
	"clean-arch-go/core/auth"
	"clean-arch-go/ports"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	db            *sqlx.DB
	httpServer    *stdHTTP.Server
	traceProvider *trace.TracerProvider
}

func New(db *sqlx.DB, tracerProvider *trace.TracerProvider, logger *logrus.Entry) (Service, error) {
	httpclient := adapters.NewHttpClient()
	application := newApplication(db, httpclient, logger)

	httpServer := ports.NewHTTPServer(func(router chi.Router) stdHTTP.Handler {
		return ports.HandlerFromMux(ports.NewHttpHandler(application), router)
	})

	return Service{
		db:            db,
		httpServer:    httpServer,
		traceProvider: tracerProvider,
	}, nil
}

func (s Service) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logrus.Info("Starting HTTP server on ", s.httpServer.Addr)
		logrus.Info("API Documentation http://localhost", s.httpServer.Addr, "/doc/index.html")

		err := s.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, stdHTTP.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		return s.httpServer.Shutdown(context.Background())
	})

	g.Go(func() error {
		<-ctx.Done()
		if s.traceProvider != nil {
			return s.traceProvider.Shutdown(context.Background())
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		if s.db != nil {
			return s.db.Close()
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("service run error: %w", err)
	}
	return nil
}

func newApplication(db *sqlx.DB, httpclient *stdHTTP.Client, logger *logrus.Entry) app.Application {
	id := adapters.NewID()
	taskRepository := adapters.NewMysqlTaskRepository(db)
	userRepository := adapters.NewMysqlUserRepository(db)
	videoService := adapters.NewVideoService(httpclient)
	authService := auth.NewAuth("secret-key-for-development")

	application := app.Application{
		Commands: app.Commands{
			AddUser:            command.NewAddUserHandler(id, userRepository, videoService, logger),
			UpdateUserRole:     command.NewUpdateUserRoleHandler(userRepository, logger),
			DeleteUser:         command.NewDeleteUserHandler(userRepository, logger),
			UpdateUserProfile:  command.NewUpdateUserProfileHandler(userRepository, logger),
			CreateTask:         command.NewCreateTaskHandler(id, taskRepository, logger),
			ChangeTaskStatus:   command.NewChangeTaskStatusHandler(taskRepository, logger),
			AssignTask:         command.NewAssignTaskHandler(taskRepository, userRepository, logger),
			UpdateTaskTitle:    command.NewUpdateTaskTitleHandler(taskRepository, logger),
			UnassignTask:       command.NewUnassignTaskHandler(taskRepository, logger),
			ReopenTask:         command.NewReopenTaskHandler(taskRepository, logger),
			DeleteTask:         command.NewDeleteTaskHandler(taskRepository, logger),
			ArchiveTask:        command.NewArchiveTaskHandler(taskRepository, logger),
			SetTaskPriority:    command.NewSetTaskPriorityHandler(taskRepository, logger),
			SetTaskDueDate:     command.NewSetTaskDueDateHandler(taskRepository, logger),
			SetTaskDescription: command.NewSetTaskDescriptionHandler(taskRepository, logger),
			AddTaskTag:         command.NewAddTaskTagHandler(taskRepository, logger),
			RemoveTaskTag:      command.NewRemoveTaskTagHandler(taskRepository, logger),
		},
		Queries: app.Queries{
			Tasks: query.NewTaskHandler(taskRepository, logger),
			User:  query.NewUserHandler(userRepository, logger),
			Users: query.NewUsersHandler(userRepository, logger),
			Login: query.NewLoginHandler(userRepository, authService, logger),
		},
	}

	return application
}
