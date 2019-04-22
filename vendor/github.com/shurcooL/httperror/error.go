package httperror

import (
	"fmt"
	"strings"
)

// Method is an error type used for methods that aren't allowed.
type Method struct {
	Allowed []string // Allowed methods.
}

func (m Method) Error() string {
	return fmt.Sprintf("method should be %v", strings.Join(m.Allowed, " or "))
}

// IsMethod reports if err is considered a method error, returning it if so.
func IsMethod(err error) (Method, bool) {
	e, ok := err.(Method)
	return e, ok
}

// Redirect is an error type used for representing a simple HTTP redirection.
type Redirect struct {
	URL string
}

func (r Redirect) Error() string { return fmt.Sprintf("redirecting to %s", r.URL) }

// IsRedirect reports if err is considered a redirect, returning it if so.
func IsRedirect(err error) (Redirect, bool) {
	e, ok := err.(Redirect)
	return e, ok
}

// BadRequest is an error type used for representing a bad request error.
type BadRequest struct {
	Err error // Not nil.
}

// Error returns BadRequest.Err.Error().
func (b BadRequest) Error() string { return b.Err.Error() }

// IsBadRequest reports if err is considered a bad request error, returning it if so.
func IsBadRequest(err error) (BadRequest, bool) {
	e, ok := err.(BadRequest)
	return e, ok
}

// HTTP is an error type used for representing a non-nil error with a status code.
type HTTP struct {
	Code int
	Err  error // Not nil.
}

// Error returns HTTP.Err.Error().
func (h HTTP) Error() string { return h.Err.Error() }

// IsHTTP reports if err is considered an HTTP error, returning it if so.
func IsHTTP(err error) (HTTP, bool) {
	e, ok := err.(HTTP)
	return e, ok
}

// JSONResponse is an error type used for representing a JSON response.
type JSONResponse struct {
	V interface{}
}

func (JSONResponse) Error() string { return "JSONResponse" }

// IsJSONResponse reports if err is considered a JSON response, returning it if so.
func IsJSONResponse(err error) (JSONResponse, bool) {
	e, ok := err.(JSONResponse)
	return e, ok
}
