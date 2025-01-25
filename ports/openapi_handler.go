package ports

import (
	"clean-arch-go/app"
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
	render.Respond(w, r, []Task{})
}

func (h HttpHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	render.Respond(w, r, []Task{})
}

func (h HttpHandler) AssignTask(w http.ResponseWriter, r *http.Request, taskId openapi_types.UUID) {
	render.Respond(w, r, []Task{})
}
