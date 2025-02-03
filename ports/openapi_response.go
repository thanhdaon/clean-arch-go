package ports

import (
	"net/http"

	"github.com/go-chi/render"
)

func internalError(err error, w http.ResponseWriter, r *http.Request) {
	responseError(err, w, r, http.StatusInternalServerError)
}

func unauthorised(err error, w http.ResponseWriter, r *http.Request) {
	responseError(err, w, r, http.StatusUnauthorized)
}

func badRequest(err error, w http.ResponseWriter, r *http.Request) {
	responseError(err, w, r, http.StatusBadRequest)
}

func responseError(err error, w http.ResponseWriter, r *http.Request, status int) {
	render.Respond(w, r, map[string]any{
		"error":  err.Error(),
		"status": status,
	})
}

func responseSuccess(w http.ResponseWriter, r *http.Request) {
	render.Respond(w, r, map[string]any{"status": http.StatusOK})
}
