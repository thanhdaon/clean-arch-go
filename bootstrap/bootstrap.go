package bootstrap

import (
	"context"
	"errors"
	"fmt"
	stdHTTP "net/http"
	"os"

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

func New(db *sqlx.DB, tracerProvider *trace.TracerProvider, logger *logrus.Entry, videoSvc command.VideoService) (Service, error) {
	if videoSvc == nil {
		httpclient := adapters.NewHttpClient()
		videoSvc = adapters.NewVideoService(httpclient)
	}
	application := newApplication(db, videoSvc, logger)

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

func newApplication(db *sqlx.DB, videoSvc command.VideoService, logger *logrus.Entry) app.Application {
	id := adapters.NewID()
	taskRepository := adapters.NewMysqlTaskRepository(db)
	userRepository := adapters.NewMysqlUserRepository(db)
	commentRepository := adapters.NewMysqlCommentRepository(db)
	activityRepository := adapters.NewMysqlActivityRepository(db)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secret-key-for-development"
		logrus.Warn("JWT_SECRET not set, using default development key")
	}
	authService := auth.NewAuth(jwtSecret)

	application := app.Application{
		Commands: app.Commands{
			AddUser:            command.NewAddUserHandler(id, userRepository, videoSvc, logger),
			UpdateUserRole:     command.NewUpdateUserRoleHandler(userRepository, logger),
			DeleteUser:         command.NewDeleteUserHandler(userRepository, logger),
			UpdateUserProfile:  command.NewUpdateUserProfileHandler(userRepository, logger),
			CreateTask:         command.NewCreateTaskHandler(id, taskRepository, logger),
			ChangeTaskStatus:   command.NewChangeTaskStatusHandler(taskRepository, activityRepository, logger),
			AssignTask:         command.NewAssignTaskHandler(taskRepository, userRepository, activityRepository, logger),
			UpdateTaskTitle:    command.NewUpdateTaskTitleHandler(taskRepository, activityRepository, logger),
			UnassignTask:       command.NewUnassignTaskHandler(taskRepository, activityRepository, logger),
			ReopenTask:         command.NewReopenTaskHandler(taskRepository, activityRepository, logger),
			DeleteTask:         command.NewDeleteTaskHandler(taskRepository, activityRepository, logger),
			ArchiveTask:        command.NewArchiveTaskHandler(taskRepository, activityRepository, logger),
			SetTaskPriority:    command.NewSetTaskPriorityHandler(taskRepository, activityRepository, logger),
			SetTaskDueDate:     command.NewSetTaskDueDateHandler(taskRepository, activityRepository, logger),
			SetTaskDescription: command.NewSetTaskDescriptionHandler(taskRepository, activityRepository, logger),
			AddTaskTag:         command.NewAddTaskTagHandler(taskRepository, logger),
			RemoveTaskTag:      command.NewRemoveTaskTagHandler(taskRepository, logger),
			AddComment:         command.NewAddCommentHandler(commentRepository, activityRepository, logger),
			UpdateComment:      command.NewUpdateCommentHandler(commentRepository, activityRepository, logger),
			DeleteComment:      command.NewDeleteCommentHandler(commentRepository, activityRepository, logger),
		},
		Queries: app.Queries{
			Tasks:          query.NewTaskHandler(taskRepository, logger),
			User:           query.NewUserHandler(userRepository, logger),
			Users:          query.NewUsersHandler(userRepository, logger),
			Login:          query.NewLoginHandler(userRepository, authService, logger),
			TaskActivities: query.NewTaskActivitiesHandler(activityRepository, logger),
		},
	}

	return application
}
