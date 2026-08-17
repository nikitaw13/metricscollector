package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/PrometheRus/metricscollector/internal/model"
)

// apiError holds the JSON body for error responses.
type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// writeJSONError marshals an apiError and writes it with the given HTTP status code.
func writeJSONError(rw http.ResponseWriter, status int, err error) {
	resp, err := json.Marshal(apiError{Code: status, Message: err.Error()})

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Internal error"))
		return
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(status)
	rw.Write(resp)
}

// decodeJSON unmarshals the request body into m and classifies decode errors.
func decodeJSON(res http.ResponseWriter, req *http.Request, m *model.Metrics) error {
	var (
		syntaxErr        *json.SyntaxError
		unmarshalTypeErr *json.UnmarshalTypeError
	)
	err := json.NewDecoder(req.Body).Decode(m)
	if errors.As(err, &syntaxErr) {
		writeJSONError(res, http.StatusBadRequest, fmt.Errorf("Invalid JSON syntax at offset %d", syntaxErr.Offset))
		return err
	}
	if errors.As(err, &unmarshalTypeErr) {
		writeJSONError(res, http.StatusBadRequest, fmt.Errorf("Invalid type %q for field %q", unmarshalTypeErr.Value, unmarshalTypeErr.Field))
		return err
	}

	if err != nil {
		writeJSONError(res, http.StatusInternalServerError, err)
		return err
	}
	return nil
}

// PostUpdate handles POST /update: validates the JSON payload, stores the metric,
// and returns the updated metric value in the response body.
func (h *MetricsHandler) PostUpdate(res http.ResponseWriter, req *http.Request) {
	var m model.Metrics

	decodeErr := decodeJSON(res, req, &m)

	if decodeErr != nil {
		return
	}

	if err := m.ValidateForUpdate(); err != nil {
		writeJSONError(res, http.StatusBadRequest, err)
		return
	}

	switch m.MType {
	case model.Gauge:
		if err := h.Storage.UpdateGauge(m.ID, *m.Value); err != nil {
			writeJSONError(res, http.StatusInternalServerError, err)
			return
		}
		newVal, err := h.Storage.GetGauge(m.ID)
		if err != nil {
			writeJSONError(res, http.StatusInternalServerError, err)
			return
		}
		m.Value = &newVal
	case model.Counter:
		if err := h.Storage.UpdateCounter(m.ID, *m.Delta); err != nil {
			writeJSONError(res, http.StatusInternalServerError, err)
			return
		}
		newVal, err := h.Storage.GetCounter(m.ID)
		if err != nil {
			writeJSONError(res, http.StatusInternalServerError, err)
			return
		}
		m.Delta = &newVal
	}

	resp, err := json.Marshal(m)
	if err != nil {
		writeJSONError(res, http.StatusInternalServerError, err)
		return
	}
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}

// PostValue handles POST /value: validates the JSON payload, looks up the stored
// metric and returns its current value in the response body.
func (h *MetricsHandler) PostValue(res http.ResponseWriter, req *http.Request) {
	var m model.Metrics

	decodeErr := decodeJSON(res, req, &m)

	if decodeErr != nil {
		return
	}

	if err := m.ValidateForRead(); err != nil {
		writeJSONError(res, http.StatusBadRequest, err)
		return
	}

	switch m.MType {
	case model.Gauge:
		result, err := h.Storage.GetGauge(m.ID)
		if err != nil {
			writeJSONError(res, http.StatusNotFound, err)
			return
		}
		m.Value = &result

	case model.Counter:
		result, err := h.Storage.GetCounter(m.ID)
		if err != nil {
			writeJSONError(res, http.StatusNotFound, err)
			return
		}
		m.Delta = &result
	}

	resp, err := json.Marshal(m)
	if err != nil {
		writeJSONError(res, http.StatusInternalServerError, err)
		return
	}

	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}
