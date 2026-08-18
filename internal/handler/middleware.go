package handler

import (
	"errors"
	"mime"
	"net/http"
	"time"

	"github.com/PrometheRus/metricscollector/internal/model"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// contentTypeMiddleware enforces application/json Content-Type and non-empty body
// for endpoints that expect JSON payloads.
func requireJSONContent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mt != "application/json" {
			http.Error(w, "Type is required", http.StatusBadRequest)
			return
		}
		if r.ContentLength == 0 {
			writeJSONError(w, http.StatusBadRequest, errors.New("Request body is required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// typeMiddleware rejects requests with an unknown metric type.
func requireMetricType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := chi.URLParam(r, "TYPE")
		if t != model.Gauge && t != model.Counter {
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// loggerMiddleware logs incoming HTTP requests
func loggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		start := time.Now()

		ri := &responseInfo{
			status: 0,
			size:   0,
		}

		lwr := loggingResponseWriter{
			ResponseWriter: rw,
			responseInfo:   ri,
		}

		h.ServeHTTP(&lwr, req)

		Logger.Info("got incoming HTTP request",
			zap.String("URI", req.RequestURI),
			zap.String("method", req.Method),
			zap.Duration("duration", time.Since(start)),
		)

		Logger.Info("respose for incoming HTTP request",
			zap.Int("status", ri.status),
			zap.Int("size", ri.size),
		)
	})
}
