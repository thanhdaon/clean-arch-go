package ports

import (
	"net/http"

	"github.com/go-chi/render"
)

func internalError(err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, w, r, "Internal server error", http.StatusInternalServerError)
}

func unauthorised(err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, w, r, "Unauthorised", http.StatusUnauthorized)
}

func badRequest(err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, w, r, "Bad request", http.StatusBadRequest)
}

func httpRespondWithError(err error, w http.ResponseWriter, r *http.Request, logMSg string, status int) {
	render.Respond(w, r, map[string]any{
		"message": logMSg,
		"status":  status,
		"error":   err.Error(),
	})
}
