package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/PrometheRus/metricscollector/internal/model"
	"github.com/PrometheRus/metricscollector/internal/repository"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// MetricsHandler serves HTTP requests for reading and updating metrics.
type MetricsHandler struct {
	storage  Repository
	database DBPinger
}

// NewMetricsHandler creates a MetricsHandler with the provided storage and database pinger.
func NewMetricsHandler(storage Repository, db DBPinger) *MetricsHandler {
	return &MetricsHandler{
		storage:  storage,
		database: db,
	}
}

// handleListMetrics returns an HTML page listing all known metrics.
func (h *MetricsHandler) handleListMetrics(w http.ResponseWriter, r *http.Request) {
	countersMap, err := h.storage.GetAllCounters()
	if err != nil {
		Logger.Error("failed to get all counters", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	gaugesMap, err := h.storage.GetAllGauges()
	if err != nil {
		Logger.Error("failed to get all gauges", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprintln(w, "<html><body>")
	fmt.Fprintln(w, "<h1>List of names and results of all currently known metrics</h1>")

	for k, v := range countersMap {
		fmt.Fprintf(w, "<b>%v</b>:   <code>%v</code><br>", k, v)
	}

	for k, v := range gaugesMap {
		fmt.Fprintf(w, "<b>%v</b>:   <code>%v</code><br>", k, v)
	}
	fmt.Fprintln(w, "<hr></body></html>")
}

// handleURLRead returns the current value of a single metric.
func (h *MetricsHandler) handleURLRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	metricName := chi.URLParam(r, "METRIC")
	switch chi.URLParam(r, "TYPE") {
	case model.Gauge:
		result, err := h.storage.GetGauge(metricName)
		if errors.Is(err, repository.ErrMetricNotFound) {
			Logger.Info("gauge not found", zap.String("metric", metricName))
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err != nil {
			Logger.Error("failed to get gauge", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%g\n", result)

	case model.Counter:
		result, err := h.storage.GetCounter(metricName)
		if errors.Is(err, repository.ErrMetricNotFound) {
			Logger.Info("counter not found", zap.String("metric", metricName))
			http.Error(w, err.Error(), http.StatusNotFound)
			return

		}
		if err != nil {
			Logger.Error("failed to get counter", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%d\n", result)
	}
}

// handleMissingMetric returns 404 when the metric name segment is absent.
func (h *MetricsHandler) handleMissingMetric(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Metric is required", http.StatusNotFound)
}

// handleMissingValue returns 400 when the metric value segment is absent.
func (h *MetricsHandler) handleMissingValue(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Metric value is required", http.StatusBadRequest)
}

// handleURLUpdate sets a new value for the given metric.
func (h *MetricsHandler) handleURLUpdate(w http.ResponseWriter, r *http.Request) {
	switch chi.URLParam(r, "TYPE") {
	case model.Gauge:
		gaugeValue, err := strconv.ParseFloat(chi.URLParam(r, "VALUE"), 64)
		if err != nil {
			Logger.Error("invalid metric value", zap.Error(err))
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}
		if err := h.storage.SetGauge(chi.URLParam(r, "METRIC"), gaugeValue); err != nil {
			Logger.Error("failed to update gauge", zap.Error(err))
			http.Error(w, "Failed to update gauge", http.StatusInternalServerError)
			return
		}

	case model.Counter:
		counterValue, err := strconv.ParseInt(chi.URLParam(r, "VALUE"), 10, 64)
		if err != nil {
			Logger.Error("invalid metric value", zap.Error(err))
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}
		if err := h.storage.AddCounter(chi.URLParam(r, "METRIC"), counterValue); err != nil {
			Logger.Error("failed to update counter", zap.Error(err))
			http.Error(w, "Failed to update counter", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Metric '%s' updated\n", chi.URLParam(r, "METRIC"))
}

// handleDatabasePing responds to GET /ping by checking the database connectivity.
func (h *MetricsHandler) handleDatabasePing(w http.ResponseWriter, r *http.Request) {
	if h.database == nil {
		http.Error(w, "Database not configured", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := h.database.PingContext(ctx); err != nil {
		Logger.Error("failed to ping database", zap.Error(err))
		http.Error(w, "Failed to ping database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Pong\n")
}
