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
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		t := chi.URLParam(req, "TYPE")
		if t != model.Gauge && t != model.Counter {
			http.Error(res, "Invalid metric type", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(res, req)
	})
}

// LoggerMiddleware logs request details (method, URI, duration) and response status/size.
func LoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		start := time.Now()

		rd := &responseData{
			status: 0,
			size:   0,
		}

		lwr := loggingResponseWriter{
			ResponseWriter: rw,
			responseData:   rd,
		}

		h.ServeHTTP(&lwr, req)

		Log.Info("got incoming HTTP request",
			zap.String("URI", req.RequestURI),
			zap.String("method", req.Method),
			zap.Duration("duration", time.Since(start)),
		)

		Log.Info("response for incoming HTTP request",
			zap.Int("status", rd.status),
			zap.Int("size", rd.size),
		)
	})
}

// GetRoot renders an HTML page listing all known counters and gauges.
func (h *MetricsHandler) GetRoot(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=UTF-8")
	// Start of the HTML page
	fmt.Fprintln(rw, "<html><body>")
	fmt.Fprintln(rw, "<h1>List of names and results of all currently known metrics</h1>")
	for k, v := range h.Storage.GetAllCounters() {
		fmt.Fprintf(rw, "<b>%v</b>:   <code>%v</code><br>", k, v)
	}
	for k, v := range h.Storage.GetAllGauges() {
		fmt.Fprintf(rw, "<b>%v</b>:   <code>%v</code><br>", k, v)
	}
	// End of the HTML page
	fmt.Fprintln(rw, "<hr></body></html>")
}

// GetMetric returns the current value of a single metric as plain text.
func (h *MetricsHandler) GetMetric(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	switch chi.URLParam(req, "TYPE") {
	case model.Gauge:
		result, err := h.Storage.GetGauge(chi.URLParam(req, "METRIC"))
		if err != nil {
			http.Error(res, err.Error(), http.StatusNotFound)
			return
		}
		fmt.Fprintf(res, "%g\n", result)

	case model.Counter:
		result, err := h.Storage.GetCounter(chi.URLParam(req, "METRIC"))
		if err != nil {
			http.Error(res, err.Error(), http.StatusNotFound)
			return
		}
		fmt.Fprintf(res, "%d\n", result)
	}
}

// PostNoMetric returns 404 when a POST /update request is missing the metric name.
func (h *MetricsHandler) PostNoMetric(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric ID is required", http.StatusNotFound)
}

// PostNoValue returns 400 when a POST /update request is missing the metric value.
func (h *MetricsHandler) PostNoValue(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric value is required", http.StatusBadRequest)
}

// PostFull processes a full URL-encoded update: parses the value, updates storage,
// and returns a plain-text confirmation on success.
func (h *MetricsHandler) PostFull(res http.ResponseWriter, req *http.Request) {
	switch chi.URLParam(req, "TYPE") {
	case model.Gauge:
		mGaugeValue, err := strconv.ParseFloat(chi.URLParam(req, "VALUE"), 64)
		if err != nil {
			http.Error(res, "Invalid metric value", http.StatusBadRequest)
			return
		}

		if err := h.Storage.UpdateGauge(chi.URLParam(req, "METRIC"), mGaugeValue); err != nil {
			http.Error(res, "Failed to update Gauge", http.StatusInternalServerError)
			return
		}

	case model.Counter:
		mCounterValue, err := strconv.ParseInt(chi.URLParam(req, "VALUE"), 10, 64)
		if err != nil || mCounterValue < 0 {
			http.Error(res, "Invalid metric value", http.StatusBadRequest)
			return
		}
		if err := h.Storage.UpdateCounter(chi.URLParam(req, "METRIC"), mCounterValue); err != nil {
			http.Error(res, "Failed to update Counter", http.StatusInternalServerError)
			return
		}
	}

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	fmt.Fprintf(res, "Metric '%s' updated✅\n", chi.URLParam(req, "METRIC"))
}
