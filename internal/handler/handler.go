package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/PrometheRus/metricscollector/internal/model"
	"github.com/go-chi/chi"
)

// MetricsHandler serves HTTP requests for reading and updating metrics.
type MetricsHandler struct {
	Storage Repository
}

// TypeMiddleware rejects requests with an unknown metric type.
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

// HandleIndex returns an HTML page listing all known metrics.
func (h *MetricsHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
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

// HandleGetMetric returns the current value of a single metric.
func (h *MetricsHandler) HandleGetMetric(w http.ResponseWriter, r *http.Request) {
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

// HandleMissingType returns 400 when the metric type segment is absent.
func (h *MetricsHandler) HandleMissingType(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Type is required", http.StatusBadRequest)
}

// HandleMissingMetric returns 404 when the metric name segment is absent.
func (h *MetricsHandler) HandleMissingMetric(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Metric is required", http.StatusNotFound)
}

// HandleMissingValue returns 400 when the metric value segment is absent.
func (h *MetricsHandler) HandleMissingValue(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Metric value is required", http.StatusBadRequest)
}

// HandleUpdate sets a new value for the given metric.
func (h *MetricsHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	switch chi.URLParam(r, "TYPE") {
	case model.Gauge:
		gaugeValue, err := strconv.ParseFloat(chi.URLParam(r, "VALUE"), 64)
		if err != nil {
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}
		if err := h.Storage.SetGauge(chi.URLParam(r, "METRIC"), gaugeValue); err != nil {
			http.Error(w, "Failed to update Gauge", http.StatusInternalServerError)
			return
		}

	case model.Counter:
		counterValue, err := strconv.ParseInt(chi.URLParam(r, "VALUE"), 10, 64)
		if err != nil {
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}
		if err := h.Storage.AddCounter(chi.URLParam(r, "METRIC"), counterValue); err != nil {
			http.Error(w, "Failed to update Counter", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Metric '%s' updated\n", chi.URLParam(r, "METRIC"))
}
