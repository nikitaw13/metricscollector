package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/PrometheRus/metricscollector/internal/model"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// MetricsHandler ties the HTTP route handlers to a metric Repository.
type MetricsHandler struct {
	Storage Repository
}

// TypeMiddleware validates that the URL param TYPE is "gauge" or "counter".
func TypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := chi.URLParam(r, "TYPE")
		if t != model.Gauge && t != model.Counter {
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// LoggerMiddleware logs request details (method, URI, duration) and response status/size.
func LoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rd := &responseData{
			status: 0,
			size:   0,
		}

		lwr := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   rd,
		}

		h.ServeHTTP(&lwr, r)

		Logger.Info("got incoming HTTP request",
			zap.String("URI", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", time.Since(start)),
		)

		Logger.Info("response for incoming HTTP request",
			zap.Int("status", rd.status),
			zap.Int("size", rd.size),
		)
	})
}

// GetRoot renders an HTML page listing all known counters and gauges.
func (h *MetricsHandler) GetRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprintln(w, "<html><body>")
	fmt.Fprintln(w, "<h1>List of names and results of all currently known metrics</h1>")
	for k, v := range h.Storage.GetAllCounters() {
		fmt.Fprintf(w, "<b>%v</b>:   <code>%v</code><br>", k, v)
	}
	for k, v := range h.Storage.GetAllGauges() {
		fmt.Fprintf(w, "<b>%v</b>:   <code>%v</code><br>", k, v)
	}
	fmt.Fprintln(w, "<hr></body></html>")
}

// GetMetric returns the current value of a single metric as plain text.
func (h *MetricsHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	switch chi.URLParam(r, "TYPE") {
	case model.Gauge:
		result, err := h.Storage.GetGauge(chi.URLParam(r, "METRIC"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "%g\n", result)

	case model.Counter:
		result, err := h.Storage.GetCounter(chi.URLParam(r, "METRIC"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "%d\n", result)
	}
}

// HandleUpdateMissingName returns 404 when a POST /update request is missing the metric name.
func (h *MetricsHandler) HandleUpdateMissingName(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Metric ID is required", http.StatusNotFound)
}

// HandleUpdateMissingValue returns 400 when a POST /update request is missing the metric value.
func (h *MetricsHandler) HandleUpdateMissingValue(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Metric value is required", http.StatusBadRequest)
}

// HandleUpdateParam processes a full URL-encoded update: parses the value, updates storage,
// and returns a plain-text confirmation on success.
func (h *MetricsHandler) HandleUpdateParam(w http.ResponseWriter, r *http.Request) {
	switch chi.URLParam(r, "TYPE") {
	case model.Gauge:
		gaugeValue, err := strconv.ParseFloat(chi.URLParam(r, "VALUE"), 64)
		if err != nil {
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}

		if err := h.Storage.UpdateGauge(chi.URLParam(r, "METRIC"), gaugeValue); err != nil {
			http.Error(w, "Failed to update Gauge", http.StatusInternalServerError)
			return
		}

	case model.Counter:
		counterValue, err := strconv.ParseInt(chi.URLParam(r, "VALUE"), 10, 64)
		if err != nil || counterValue < 0 {
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}
		if err := h.Storage.UpdateCounter(chi.URLParam(r, "METRIC"), counterValue); err != nil {
			http.Error(w, "Failed to update Counter", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Metric '%s' updated\n", chi.URLParam(r, "METRIC"))
}
