package handler

import (
	"net/http/httptest"

	"github.com/PrometheRus/metricscollector/internal/repository"
)

// Function creates a test server with a fresh storage instance and configured routes.
func GetTestRouter() (server *httptest.Server) {
	var s = repository.New()
	s.UpdateGauge("___test___", 123)
	s.UpdateCounter("___test___", 234)
	var h = MetricsHandler{Storage: s}
	router := h.New()
	server = httptest.NewServer(router)
	return
}
