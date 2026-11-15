package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/nikitaw13/metricscollector/internal/model"
	"go.uber.org/zap"
)

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

	host := clientIP(r.RemoteAddr)
	err := h.auditor.Notify(time.Now().Unix(), []model.Metric{metric}, host)
	if err != nil {
		logErrorByStatus(h.logger, http.StatusInternalServerError, err)
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

	host := clientIP(r.RemoteAddr)
	err = h.auditor.Notify(time.Now().Unix(), metrics, host)
	if err != nil {
		logErrorByStatus(h.logger, http.StatusInternalServerError, err)
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
