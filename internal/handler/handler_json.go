package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/nikitaw13/metricscollector/internal/model"
	"go.uber.org/zap"
)

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

// handleJSONUpdate handles POST /update: validates the JSON payload, stores the metric,
// and returns the updated metric value in the response body.
func (h *MetricsHandler) handleJSONUpdate(w http.ResponseWriter, r *http.Request) {
	var metric model.Metric

	if err := decodeJSON(r, &metric); err != nil {
		status := decodeErrorStatus(err)
		logErrorByStatus(h.logger, status, err)
		writeJSONError(w, status, err)
		return
	}

	if err := metric.ValidateForUpdate(); err != nil {
		logErrorByStatus(h.logger, http.StatusBadRequest, err)
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	switch metric.Type {
	case model.Gauge:
		if err := h.service.SetGauge(metric.ID, *metric.Value); err != nil {
			logErrorByStatus(h.logger, http.StatusInternalServerError, err)
			writeJSONError(w, http.StatusInternalServerError, err)
			return
		}
	case model.Counter:
		newDelta, err := h.service.AddCounter(metric.ID, *metric.Delta)
		if err != nil {
			logErrorByStatus(h.logger, http.StatusInternalServerError, err)
			writeJSONError(w, http.StatusInternalServerError, err)
			return
		}
		metric.Delta = &newDelta
	default:
		logErrorByStatus(h.logger, http.StatusBadRequest, model.ErrInvalidType)
		writeJSONError(w, http.StatusBadRequest, model.ErrInvalidType)
		return
	}

	if err := writeJSON(w, http.StatusOK, metric); err != nil {
		h.logger.Error("failed to write json", zap.Error(err))
		return
	}
}

// handleJSONUpdates handles POST /updates: validates and stores a batch of metrics.
func (h *MetricsHandler) handleJSONUpdates(w http.ResponseWriter, r *http.Request) {
	var metrics []model.Metric

	if err := decodeJSON(r, &metrics); err != nil {
		status := decodeErrorStatus(err)
		logErrorByStatus(h.logger, status, err)
		writeJSONError(w, status, err)
		return
	}

	for _, metric := range metrics {
		if err := metric.ValidateForUpdate(); err != nil {
			logErrorByStatus(h.logger, http.StatusBadRequest, err)
			writeJSONError(w, http.StatusBadRequest, err)
			return
		}
	}

	err := h.service.UpdateMetricsBatch(r.Context(), metrics)
	if err != nil {
		logErrorByStatus(h.logger, http.StatusInternalServerError, err)
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// handleJSONRead handles POST /value: validates the JSON payload, looks up the stored
// metric and returns its current value in the response body.
func (h *MetricsHandler) handleJSONRead(w http.ResponseWriter, r *http.Request) {
	var metric model.Metric

	if err := decodeJSON(r, &metric); err != nil {
		status := decodeErrorStatus(err)
		logErrorByStatus(h.logger, status, err)
		writeJSONError(w, status, err)
		return
	}

	if err := metric.ValidateForRead(); err != nil {
		logErrorByStatus(h.logger, http.StatusBadRequest, err)
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	switch metric.Type {
	case model.Gauge:
		value, err := h.service.GetGauge(metric.ID)
		if errors.Is(err, model.ErrMetricNotFound) {
			logErrorByStatus(h.logger, http.StatusNotFound, err)
			writeJSONError(w, http.StatusNotFound, err)
			return
		}
		if err != nil {
			logErrorByStatus(h.logger, http.StatusInternalServerError, err)
			writeJSONError(w, http.StatusInternalServerError, err)
			return
		}
		metric.Value = &value

	case model.Counter:
		delta, err := h.service.GetCounter(metric.ID)
		if errors.Is(err, model.ErrMetricNotFound) {
			logErrorByStatus(h.logger, http.StatusNotFound, err)
			writeJSONError(w, http.StatusNotFound, err)
			return
		}
		if err != nil {
			logErrorByStatus(h.logger, http.StatusInternalServerError, err)
			writeJSONError(w, http.StatusInternalServerError, err)
			return
		}
		metric.Delta = &delta
	default:
		logErrorByStatus(h.logger, http.StatusBadRequest, model.ErrInvalidType)
		writeJSONError(w, http.StatusBadRequest, model.ErrInvalidType)
		return
	}
	if err := writeJSON(w, http.StatusOK, metric); err != nil {
		h.logger.Error("failed to write json", zap.Error(err))
		return
	}
}
