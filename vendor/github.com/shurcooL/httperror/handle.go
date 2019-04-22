package httperror

import (
	"net/http"
	"strings"
)

// HandleMethod handles a method error.
func HandleMethod(w http.ResponseWriter, err Method) {
	w.Header().Set("Allow", strings.Join(err.Allowed, ", "))
	error := "405 Method Not Allowed\n\n" + err.Error()
	http.Error(w, error, http.StatusMethodNotAllowed)
}

// HandleBadRequest handles a bad request error.
// The contents of err.Err are displayed to user, so you shouldn't include
// any sensitive information there, only information about the bad request.
func HandleBadRequest(w http.ResponseWriter, err BadRequest) {
	error := "400 Bad Request\n\n" + err.Error()
	http.Error(w, error, http.StatusBadRequest)
}
