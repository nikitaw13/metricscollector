package handler

import (
	"net/http/httptest"

	"github.com/PrometheRus/metricscollector/internal/repository"
)

const defaultGaugeValue = 123
const defaultCounterValue = 234

// Function creates a test server with a fresh storage instance and configured routes.
func GetTestRouter() (server *httptest.Server) {
	var s = repository.New()
	s.UpdateGauge("___test___", defaultGaugeValue)
	s.UpdateCounter("___test___", defaultCounterValue)
	var h = MetricsHandler{Storage: s}
	router := h.New()
	server = httptest.NewServer(router)
	return
}

var htmlResponse = `<html><body>
<h1>List of names and results of all currently known metrics</h1>
<b>___test___</b>:   <code>234</code><br><b>___test___</b>:   <code>123</code><br><hr></body></html>
`
