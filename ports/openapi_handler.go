package ports

import (
	"clean-arch-go/app"
	"clean-arch-go/app/command"
	"clean-arch-go/app/query"
	"clean-arch-go/core/errors"
	"net/http"

	"github.com/go-chi/render"
	"go.opentelemetry.io/otel/trace"
)

type HttpHandler struct {
	app app.Application
}

func NewHttpHandler(app app.Application) HttpHandler {
	return HttpHandler{app: app}
}

func (h HttpHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	op := errors.Op("http.GetTasks")
	ctx := r.Context()

	user, err := userFromCtx(ctx)
	if err != nil {
		unauthorised(ctx, errors.E(op, err), w, r)
		return
	}

	tasks, err := h.app.Queries.Tasks.Handle(ctx, query.Tasks{
		User: user,
	})

	if err != nil {
		internalError(ctx, errors.E(op, err), w, r)
		return
	}

	render.Respond(w, r, map[string]any{
		"status":   http.StatusOK,
		"tasks":    tasks,
		"trace_id": trace.SpanContextFromContext(ctx).TraceID(),
	})
}

func (h HttpHandler) ChangeTaskStatus(w http.ResponseWriter, r *http.Request, taskId string) {
	op := errors.Op("http.ChangeTaskStatus")

	changer, err := userFromCtx(r.Context())
	if err != nil {
		unauthorised(r.Context(), errors.E(op, err), w, r)
		return
	}

	body := PutTaskStatus{}
	if err := render.Decode(r, &body); err != nil {
		badRequest(r.Context(), err, w, r)
		return
	}

	err = h.app.Commands.ChangeTaskStatus.Handle(r.Context(), command.ChangeTaskStatus{
		TaskId:  taskId,
		Status:  string(body.Status),
		Changer: changer,
	})

	if err != nil {
		badRequest(r.Context(), errors.E(op, err), w, r)
		return
	}

	responseSuccess(r.Context(), w, r)
}

func (h HttpHandler) AddUser(w http.ResponseWriter, r *http.Request) {
	op := errors.Op("http.AddUser")

	body := PostUser{}
	if err := render.Decode(r, &body); err != nil {
		badRequest(r.Context(), err, w, r)
		return
	}

	err := h.app.Commands.AddUser.Handle(r.Context(), command.AddUser{
		Role:     string(body.Role),
		Name:     body.Name,
		Email:    string(body.Email),
		Password: body.Password,
	})

	if err != nil {
		badRequest(r.Context(), errors.E(op, err), w, r)
		return
	}

	responseSuccess(r.Context(), w, r)
}

func (h HttpHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	op := errors.Op("http.ListUsers")
	ctx := r.Context()

	users, err := h.app.Queries.Users.Handle(ctx, query.UsersQuery{})
	if err != nil {
		internalError(ctx, errors.E(op, err), w, r)
		return
	}

	render.Respond(w, r, map[string]any{
		"status":   http.StatusOK,
		"users":    users,
		"trace_id": trace.SpanContextFromContext(ctx).TraceID(),
	})
}

func (h HttpHandler) GetUser(w http.ResponseWriter, r *http.Request, userId string) {
	op := errors.Op("http.GetUser")
	ctx := r.Context()

	user, err := h.app.Queries.User.Handle(ctx, query.UserQuery{
		UserUUID: userId,
	})

	if err != nil {
		internalError(ctx, errors.E(op, err), w, r)
		return
	}

	render.Respond(w, r, map[string]any{
		"status":   http.StatusOK,
		"user":     user,
		"trace_id": trace.SpanContextFromContext(ctx).TraceID(),
	})
}

func (h HttpHandler) DeleteUser(w http.ResponseWriter, r *http.Request, userId string) {
	op := errors.Op("http.DeleteUser")

	err := h.app.Commands.DeleteUser.Handle(r.Context(), command.DeleteUser{
		UserUUID: userId,
	})

	if err != nil {
		badRequest(r.Context(), errors.E(op, err), w, r)
		return
	}

	responseSuccess(r.Context(), w, r)
}

func (h HttpHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request, userId string) {
	op := errors.Op("http.UpdateUserProfile")

	body := PatchUser{}
	if err := render.Decode(r, &body); err != nil {
		badRequest(r.Context(), err, w, r)
		return
	}

	name := ""
	email := ""
	if body.Name != nil {
		name = *body.Name
	}
	if body.Email != nil {
		email = string(*body.Email)
	}

	err := h.app.Commands.UpdateUserProfile.Handle(r.Context(), command.UpdateUserProfile{
		UserUUID: userId,
		Name:     name,
		Email:    email,
	})

	if err != nil {
		badRequest(r.Context(), errors.E(op, err), w, r)
		return
	}

	responseSuccess(r.Context(), w, r)
}

func (h HttpHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request, userId string) {
	op := errors.Op("http.UpdateUserRole")

	currentUser, err := userFromCtx(r.Context())
	if err != nil {
		unauthorised(r.Context(), errors.E(op, err), w, r)
		return
	}

	body := PutUserRole{}
	if err := render.Decode(r, &body); err != nil {
		badRequest(r.Context(), err, w, r)
		return
	}

	err = h.app.Commands.UpdateUserRole.Handle(r.Context(), command.UpdateUserRole{
		CurrentUserUUID: currentUser.UUID(),
		TargetUserUUID:  userId,
		NewRole:         string(body.Role),
	})

	if err != nil {
		badRequest(r.Context(), errors.E(op, err), w, r)
		return
	}

	responseSuccess(r.Context(), w, r)
}

func (h HttpHandler) Login(w http.ResponseWriter, r *http.Request) {
	op := errors.Op("http.Login")

	body := LoginRequest{}
	if err := render.Decode(r, &body); err != nil {
		badRequest(r.Context(), err, w, r)
		return
	}

	result, err := h.app.Queries.Login.Handle(r.Context(), query.Login{
		Email:    string(body.Email),
		Password: body.Password,
	})

	if err != nil {
		badRequest(r.Context(), errors.E(op, err), w, r)
		return
	}

	render.Respond(w, r, map[string]any{
		"status":   http.StatusOK,
		"token":    result.Token,
		"user":     result.User,
		"trace_id": trace.SpanContextFromContext(r.Context()).TraceID(),
	})
}

func (h HttpHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	op := errors.Op("http.CreateTask")
	ctx := r.Context()

	user, err := userFromCtx(ctx)
	if err != nil {
		unauthorised(ctx, errors.E(op, err), w, r)
		return
	}

	err = h.app.Commands.CreateTask.Handle(ctx, command.CreateTask{
		Title:   "",
		Creator: user,
	})

	if err != nil {
		badRequest(ctx, errors.E(op, err), w, r)
		return
	}

	responseSuccess(r.Context(), w, r)
}

func (h HttpHandler) AssignTask(w http.ResponseWriter, r *http.Request, taskId, assigneeId string) {
	op := errors.Op("http.AssignTask")
	ctx := r.Context()

	assigner, err := userFromCtx(ctx)
	if err != nil {
		unauthorised(ctx, errors.E(op, err), w, r)
		return
	}

	err = h.app.Commands.AssignTask.Handle(ctx, command.AssignTask{
		TaskId:     taskId,
		AssigneeId: assigneeId,
		Assigner:   assigner,
	})

	responseSuccess(r.Context(), w, r)
}

func (h HttpHandler) UnassignTask(w http.ResponseWriter, r *http.Request, taskId string) {
	op := errors.Op("http.UnassignTask")

	remover, err := userFromCtx(r.Context())
	if err != nil {
		unauthorised(r.Context(), errors.E(op, err), w, r)
		return
	}

	err = h.app.Commands.UnassignTask.Handle(r.Context(), command.UnassignTask{
		TaskId: taskId,
		Caller: remover,
	})

	if err != nil {
		badRequest(r.Context(), errors.E(op, err), w, r)
		return
	}

	responseSuccess(r.Context(), w, r)
}

func (h HttpHandler) UpdateTaskTitle(w http.ResponseWriter, r *http.Request, taskId string) {
	op := errors.Op("http.UpdateTaskTitle")

	updater, err := userFromCtx(r.Context())
	if err != nil {
		unauthorised(r.Context(), errors.E(op, err), w, r)
		return
	}

	body := PatchTaskTitle{}
	if err := render.Decode(r, &body); err != nil {
		badRequest(r.Context(), err, w, r)
		return
	}

	err = h.app.Commands.UpdateTaskTitle.Handle(r.Context(), command.UpdateTaskTitle{
		TaskId:  taskId,
		Title:   body.Title,
		Updater: updater,
	})

	if err != nil {
		badRequest(r.Context(), errors.E(op, err), w, r)
		return
	}

	responseSuccess(r.Context(), w, r)
}

func (h HttpHandler) ReopenTask(w http.ResponseWriter, r *http.Request, taskId string) {
	op := errors.Op("http.ReopenTask")

	caller, err := userFromCtx(r.Context())
	if err != nil {
		unauthorised(r.Context(), errors.E(op, err), w, r)
		return
	}

	err = h.app.Commands.ReopenTask.Handle(r.Context(), command.ReopenTask{
		TaskId: taskId,
		Caller: caller,
	})

	if err != nil {
		badRequest(r.Context(), errors.E(op, err), w, r)
		return
	}

	responseSuccess(r.Context(), w, r)
}

func (h HttpHandler) DeleteTask(w http.ResponseWriter, r *http.Request, taskId string) {
	op := errors.Op("http.DeleteTask")

	caller, err := userFromCtx(r.Context())
	if err != nil {
		unauthorised(r.Context(), errors.E(op, err), w, r)
		return
	}

	err = h.app.Commands.DeleteTask.Handle(r.Context(), command.DeleteTask{
		TaskId: taskId,
		Caller: caller,
	})

	if err != nil {
		badRequest(r.Context(), errors.E(op, err), w, r)
		return
	}

	responseSuccess(r.Context(), w, r)
}

func (h HttpHandler) ArchiveTask(w http.ResponseWriter, r *http.Request, taskId string) {
	op := errors.Op("http.ArchiveTask")

	caller, err := userFromCtx(r.Context())
	if err != nil {
		unauthorised(r.Context(), errors.E(op, err), w, r)
		return
	}

	err = h.app.Commands.ArchiveTask.Handle(r.Context(), command.ArchiveTask{
		TaskId: taskId,
		Caller: caller,
	})

	if err != nil {
		badRequest(r.Context(), errors.E(op, err), w, r)
		return
	}

	responseSuccess(r.Context(), w, r)
}
