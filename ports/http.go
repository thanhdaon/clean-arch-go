package ports

import (
	"clean-arch-go/app"
	"net/http"

	"github.com/go-chi/render"
)

type HttpServer struct {
	app app.Application
}

func NewHttpServer(app app.Application) HttpServer {
	return HttpServer{app: app}
}

func (h HttpServer) GetTasks(w http.ResponseWriter, r *http.Request) {
	render.Respond(w, r, []Task{})
}
