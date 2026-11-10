package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"

	"go.uber.org/zap"
)

const tmplHTML = `<html><body>
<h1>List of names and results of all currently known metrics</h1>
{{- range $name, $value := .Counters}}
<p><b>{{$name}}</b>: <code>{{$value}}</code></p>
{{- end}}
{{- range $name, $value := .Gauges}}
<p><b>{{$name}}</b>: <code>{{$value}}</code></p>
{{- end}}
<hr></body></html>`

const internalErrorMessage = "Internal Server Error"

var tmpl = template.Must(template.New("metrics").Parse(tmplHTML))

// apiError holds the JSON body for error responses.
type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// decodeError carries a client-safe message while preserving the original
// error chain for errors.Is/As and logging.
type decodeError struct {
	status  int
	message string
	err     error
}

func (e *decodeError) Error() string { return e.message }

func (e *decodeError) Unwrap() error { return e.err }

// writeJSON marshals payload and writes it as the response body with the given HTTP status code.
func writeJSON(w http.ResponseWriter, status int, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal error"))
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(data)
	return nil
}

// writeJSONError writes an apiError with the given HTTP status code: server errors (5xx)
// report only the generic status text so internal details are not leaked, client errors
// (4xx) keep the original message.
func writeJSONError(w http.ResponseWriter, status int, err error) {
	var message string
	if status >= http.StatusInternalServerError {
		message = http.StatusText(status)
	} else {
		message = err.Error()
	}

	_ = writeJSON(w, status, apiError{
		Code:    status,
		Message: message,
	})
}

// logErrorByStatus logs err at Error level for 5xx statuses and Debug for 4xx.
func logErrorByStatus(logger *zap.Logger, status int, err error) {
	if status >= http.StatusInternalServerError {
		logger.Error("server error", zap.Int("status", status), zap.Error(err))
	} else {
		logger.Debug("client error", zap.Int("status", status), zap.Error(err))
	}
}

// decodeJSON unmarshals the request body into target and classifies decode errors.
func decodeJSON(r *http.Request, target any) error {
	err := json.NewDecoder(r.Body).Decode(target)

	if syntaxErr, ok := errors.AsType[*json.SyntaxError](err); ok {
		return &decodeError{
			status:  http.StatusBadRequest,
			message: fmt.Sprintf("invalid JSON syntax at offset %d", syntaxErr.Offset),
			err:     err,
		}
	}

	if unmarshalTypeErr, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
		return &decodeError{
			status:  http.StatusBadRequest,
			message: fmt.Sprintf("invalid type %q for field %q", unmarshalTypeErr.Value, unmarshalTypeErr.Field),
			err:     err,
		}
	}

	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return &decodeError{
			status:  http.StatusBadRequest,
			message: "request body is empty",
			err:     err,
		}
	}

	if err != nil {
		return &decodeError{
			status:  http.StatusInternalServerError,
			message: "failed to decode request body",
			err:     err,
		}
	}
	return nil
}

// decodeErrorStatus extracts the HTTP status code from a decodeError, defaulting to 500.
func decodeErrorStatus(err error) int {
	if decodeErr, ok := errors.AsType[*decodeError](err); ok {
		return decodeErr.status
	}
	return http.StatusInternalServerError
}

func clientIP(remoteAddress string) string {
	host, _, err := net.SplitHostPort(remoteAddress)
	if err != nil {
		host = remoteAddress
	}
	return host
}
