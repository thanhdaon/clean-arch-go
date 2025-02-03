package ports

import (
	"clean-arch-go/app"
	"clean-arch-go/app/query"
	"net/http"

	"github.com/go-chi/render"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type HttpHandler struct {
	app app.Application
}

func NewHttpHandler(app app.Application) HttpHandler {
	return HttpHandler{app: app}
}

func (h HttpHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	user, err := userFromCtx(r.Context())
	if err != nil {
		unauthorised(err, w, r)
		return
	}

	tasks, err := h.app.Queries.Tasks.Handle(r.Context(), query.Tasks{
		User: user,
	})

	render.Respond(w, r, tasks)
}

func (h HttpHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	render.Respond(w, r, []Task{})
}

func (h HttpHandler) AssignTask(w http.ResponseWriter, r *http.Request, taskId openapi_types.UUID) {
	render.Respond(w, r, []Task{})
}
