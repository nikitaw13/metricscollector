package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/PrometheRus/metricscollector/internal/logger"
	"github.com/PrometheRus/metricscollector/internal/model"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

type MetricsHandler struct {
	Storage Repository
}

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

// RequestLogger — middleware-логер для входящих HTTP-запросов
func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		logger.Log.Debug("got incoming HTTP request",
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path),
		)
		h.ServeHTTP(rw, req)
	})
}

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

func (h *MetricsHandler) PostNoType(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Type is required", http.StatusBadRequest)
}

func (h *MetricsHandler) PostNoMetric(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric is required", http.StatusNotFound)
}

func (h *MetricsHandler) PostNoValue(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric value is required", http.StatusBadRequest)
}

func (h *MetricsHandler) PostFull(res http.ResponseWriter, req *http.Request) {
	switch chi.URLParam(req, "TYPE") {
	case model.Gauge:
		mGaugeValue, err := strconv.ParseFloat(chi.URLParam(req, "VALUE"), 64)
		// При попытке передать запрос с некорректным значением метрики возвращать http.StatusBadRequest.
		if err != nil {
			http.Error(res, "Invalid metric value", http.StatusBadRequest)
			return
		}

		// При ошибке обработки запроса обновления Gauge возвращать http.StatusInternalServerError
		if err := h.Storage.UpdateGauge(chi.URLParam(req, "METRIC"), mGaugeValue); err != nil {
			http.Error(res, "Failed to update Gauge", http.StatusInternalServerError)
			return
		}

	case model.Counter:
		mCounterValue, err := strconv.ParseInt(chi.URLParam(req, "VALUE"), 10, 64)
		// При попытке передать запрос с некорректным значением метрики возвращать http.StatusBadRequest.
		if err != nil {
			http.Error(res, "Invalid metric value", http.StatusBadRequest)
			return
		}
		// При ошибке обработки запроса обновления Counter возвращать http.StatusInternalServerError
		if err := h.Storage.UpdateCounter(chi.URLParam(req, "METRIC"), mCounterValue); err != nil {
			http.Error(res, "Failed to update Counter", http.StatusInternalServerError)
			return
		}
	}

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	fmt.Fprintf(res, "Metric '%s' updated✅\n", chi.URLParam(req, "METRIC"))
}
