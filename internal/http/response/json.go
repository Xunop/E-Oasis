package response // import "github.com/Xunop/e-learn/http/response"

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Xunop/e-oasis/http/request"
	"github.com/Xunop/e-oasis/log"
	"go.uber.org/zap"
)

const contentTypeHeader = `application/json`

// OK creates a new JSON response with a 200 status code.
func OK(w http.ResponseWriter, r *http.Request, body interface{}) {
	builder := New(w, r)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSON(body))
	builder.Write()
}

// Created sends a created response to the client.
func Created(w http.ResponseWriter, r *http.Request, body interface{}) {
	builder := New(w, r)
	builder.WithStatus(http.StatusCreated)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSON(body))
	builder.Write()
}

// NoContent sends a no content response to the client.
func NoContent(w http.ResponseWriter, r *http.Request) {
	builder := New(w, r)
	builder.WithStatus(http.StatusNoContent)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.Write()
}

func Accepted(w http.ResponseWriter, r *http.Request) {
	builder := New(w, r)
	builder.WithStatus(http.StatusAccepted)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.Write()
}

// ServerError sends an internal error to the client.
func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Error(http.StatusText(http.StatusInternalServerError),
		zap.Error(err),
		zap.String("client_ip", request.FindClientIP(r)),
		zap.String("request.method", r.Method),
		zap.String("request.uri", r.RequestURI),
		zap.String("request.user_agent", r.UserAgent()),
		zap.Int("response.status_code", http.StatusBadRequest),
	)

	builder := New(w, r)
	builder.WithStatus(http.StatusInternalServerError)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSONError(err))
	builder.Write()
}

// BadRequest sends a bad request error to the client.
func BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	log.Warn(http.StatusText(http.StatusBadRequest),
		zap.Any("error", err),
		zap.String("client_ip", request.FindClientIP(r)),
		zap.String("request.method", r.Method),
		zap.String("request.uri", r.RequestURI),
		zap.String("request.user_agent", r.UserAgent()),
		zap.Int("response.status_code", http.StatusBadRequest),
	)

	builder := New(w, r)
	builder.WithStatus(http.StatusBadRequest)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSONError(err))
	builder.Write()
}

// Unauthorized sends a not authorized error to the client.
func Unauthorized(w http.ResponseWriter, r *http.Request) {
	log.Warn(http.StatusText(http.StatusUnauthorized),
		zap.String("client_ip", request.FindClientIP(r)),
		zap.String("request.method", r.Method),
		zap.String("request.uri", r.RequestURI),
		zap.String("request.user_agent", r.UserAgent()),
		zap.Int("response.status_code", http.StatusUnauthorized),
	)

	builder := New(w, r)
	builder.WithStatus(http.StatusUnauthorized)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSONError(errors.New("access unauthorized")))
	builder.Write()
}

// Forbidden sends a forbidden error to the client.
func Forbidden(w http.ResponseWriter, r *http.Request) {
	log.Warn(http.StatusText(http.StatusForbidden),
		zap.String("client_ip", request.FindClientIP(r)),
		zap.String("request.method", r.Method),
		zap.String("request.uri", r.RequestURI),
		zap.String("request.user_agent", r.UserAgent()),
		zap.Int("response.status_code", http.StatusForbidden),
	)

	builder := New(w, r)
	builder.WithStatus(http.StatusForbidden)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSONError(errors.New("access forbidden")))
	builder.Write()
}

// NotFound sends a page not found error to the client.
func NotFound(w http.ResponseWriter, r *http.Request) {
	log.Warn(http.StatusText(http.StatusNotFound),
		zap.String("client_ip", request.FindClientIP(r)),
		zap.String("request.method", r.Method),
		zap.String("request.uri", r.RequestURI),
		zap.String("request.user_agent", r.UserAgent()),
		zap.Int("response.status_code", http.StatusNotFound),
	)

	builder := New(w, r)
	builder.WithStatus(http.StatusNotFound)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSONError(errors.New("resource not found")))
	builder.Write()
}

func toJSONError(err error) []byte {
	type errorMsg struct {
		ErrorMessage string `json:"error_message"`
	}

	return toJSON(errorMsg{ErrorMessage: err.Error()})
}

func toJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		log.Error("Unable to marshal JSON response", zap.Any("error", err))
		return []byte("")
	}

	return b
}
