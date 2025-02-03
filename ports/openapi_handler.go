package ports

import (
	"clean-arch-go/app"
	"clean-arch-go/app/command"
	"clean-arch-go/app/query"
	"clean-arch-go/common/errors"
	"net/http"

	"github.com/go-chi/render"
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
		unauthorised(errors.E(op, err), w, r)
		return
	}

	tasks, err := h.app.Queries.Tasks.Handle(ctx, query.Tasks{
		User: user,
	})

	render.Respond(w, r, tasks)
}

func (h HttpHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	op := errors.Op("http.CreateTask")
	ctx := r.Context()

	user, err := userFromCtx(ctx)
	if err != nil {
		unauthorised(errors.E(op, err), w, r)
		return
	}

	err = h.app.Commands.CreateTask.Handle(ctx, command.CreateTask{
		Title:   "",
		Creator: user,
	})

	if err != nil {
		badRequest(errors.E(op, err), w, r)
		return
	}

	responseSuccess(w, r)
}

func (h HttpHandler) AssignTask(w http.ResponseWriter, r *http.Request, taskId, assigneeId string) {
	op := errors.Op("http.AssignTask")
	ctx := r.Context()

	assigner, err := userFromCtx(ctx)
	if err != nil {
		unauthorised(errors.E(op, err), w, r)
		return
	}

	err = h.app.Commands.AssignTask.Handle(ctx, command.AssignTask{
		TaskId:     taskId,
		AssigneeId: assigneeId,
		Assigner:   assigner,
	})

	responseSuccess(w, r)
}
