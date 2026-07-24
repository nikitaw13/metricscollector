package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/PrometheRus/metricscollector/internal/model"
	"github.com/PrometheRus/metricscollector/internal/repository"
	"github.com/go-chi/chi"
)

type MetricsHandler struct {
	Storage repository.Repository
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

func (h *MetricsHandler) GetRootHandler(rw http.ResponseWriter, r *http.Request) {
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

func (h *MetricsHandler) GetMetricHandler(res http.ResponseWriter, r *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	switch chi.URLParam(r, "TYPE") {
	case model.Gauge:
		result, err := h.Storage.GetGauge(chi.URLParam(r, "METRIC"))
		if err != nil {
			http.Error(res, err.Error(), http.StatusNotFound)
			return
		}
		fmt.Fprintf(res, "%g\n", result)

	case model.Counter:
		result, err := h.Storage.GetCounter(chi.URLParam(r, "METRIC"))
		if err != nil {
			http.Error(res, err.Error(), http.StatusNotFound)
			return
		}
		fmt.Fprintf(res, "%d\n", result)
	default:
		http.Error(res, "Invalid metric type", http.StatusBadRequest)
		return
	}
}

func (h *MetricsHandler) PostNoTypeHandler(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric type is required", http.StatusBadRequest)
}

func (h *MetricsHandler) PostNoMetricHandler(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric is required", http.StatusNotFound)
}

func (h *MetricsHandler) PostNoValueHandler(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric value is required", http.StatusBadRequest)
}

func (h *MetricsHandler) PostFullHandler(res http.ResponseWriter, req *http.Request) {
	switch chi.URLParam(req, "TYPE") {
	case "gauge":
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

	case "counter":
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
	// log.Printf("The handler received request: %s %s %s and the response is %d", req.RemoteAddr, req.Method, req.URL.Path, http.StatusOK)
}
