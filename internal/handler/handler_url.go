package handler

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/nikitaw13/metricscollector/internal/model"
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

// MetricsHandler serves HTTP requests for reading and updating metrics.
type MetricsHandler struct {
	hashKey string
	service Service
	logger  *zap.Logger
}

// NewMetricsHandler creates a MetricsHandler with the provided signing key, service, and logger.
func NewMetricsHandler(hashKey string, service Service, logger *zap.Logger) *MetricsHandler {
	return &MetricsHandler{
		hashKey: hashKey,
		service: service,
		logger:  logger,
	}
}

// handleListMetrics returns an HTML page listing all known metrics.
func (h *MetricsHandler) handleListMetrics(w http.ResponseWriter, r *http.Request) {
	gauges, counters, err := h.service.ListAllMetrics()
	if err != nil {
		h.logger.Error("failed to list all metrics", zap.Error(err))
		http.Error(w, internalErrorMessage, http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		Counters map[string]int64
		Gauges   map[string]float64
	}{counters, gauges})
	if err != nil {
		h.logger.Error("failed to apply a parsed template to the specified data object", zap.Error(err))
		http.Error(w, internalErrorMessage, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	buf.WriteTo(w)
}

// handleURLRead returns the current value of a single metric.
func (h *MetricsHandler) handleURLRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	var (
		metricType = chi.URLParam(r, "type")
		metricID   = chi.URLParam(r, "metric")
		value      float64
		delta      int64
		err        error
	)

	switch metricType {
	case model.Gauge:
		value, err = h.service.GetGauge(metricID)
		if errors.Is(err, model.ErrMetricNotFound) {
			http.Error(w, fmt.Sprintf("%s %s %s", model.Gauge, metricID, err.Error()), http.StatusNotFound)
			return
		}
	case model.Counter:
		delta, err = h.service.GetCounter(metricID)
		if errors.Is(err, model.ErrMetricNotFound) {
			http.Error(w, fmt.Sprintf("%s %s %s", model.Counter, metricID, err.Error()), http.StatusNotFound)
			return
		}
	}
	if err != nil {
		h.logger.Error("failed to get a metric", zap.Error(err))
		http.Error(w, internalErrorMessage, http.StatusInternalServerError)
		return
	}
	switch metricType {
	case model.Gauge:
		fmt.Fprintf(w, "%v\n", value)
	case model.Counter:
		fmt.Fprintf(w, "%v\n", delta)
	}
}

// handleMissingMetric returns 404 when the metric name segment is absent.
func (h *MetricsHandler) handleMissingMetric(w http.ResponseWriter, r *http.Request) {
	http.Error(w, model.ErrMissingID.Error(), http.StatusNotFound)
}

// handleMissingValue returns 400 when the metric value segment is absent.
func (h *MetricsHandler) handleMissingValue(w http.ResponseWriter, r *http.Request) {
	http.Error(w, model.ErrMissingValue.Error(), http.StatusBadRequest)
}

// handleURLUpdate sets a new value for the given metric.
func (h *MetricsHandler) handleURLUpdate(w http.ResponseWriter, r *http.Request) {
	var (
		metricType  = chi.URLParam(r, "type")
		metricID    = chi.URLParam(r, "metric")
		metricValue = chi.URLParam(r, "value")
	)
	switch metricType {
	case model.Gauge:
		parsedValue, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, model.ErrInvalidValue.Error(), http.StatusBadRequest)
			return
		}
		if err := h.service.SetGauge(metricID, parsedValue); err != nil {
			h.logger.Error("failed to update a gauge", zap.Error(err))
			http.Error(w, internalErrorMessage, http.StatusInternalServerError)
			return
		}

	case model.Counter:
		parsedValue, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			h.logger.Debug("invalid metric value", zap.Error(err))
			http.Error(w, model.ErrInvalidValue.Error(), http.StatusBadRequest)
			return
		}
		if _, err := h.service.AddCounter(metricID, parsedValue); err != nil {
			h.logger.Error("failed to update a counter", zap.Error(err))
			http.Error(w, internalErrorMessage, http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Metric '%s' updated\n", metricID)
}
