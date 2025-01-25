package ports

import (
	"clean-arch-go/app"
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
	render.Respond(w, r, []Task{})
}
