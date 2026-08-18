package handler

import (
	"net/http/httptest"

	"github.com/PrometheRus/metricscollector/internal/repository"
)

// defaultGaugeValue is the pre-seeded gauge value used across handler tests.
const defaultGaugeValue = 123.00

// defaultCounterValue is the pre-seeded counter value used across handler tests.
const defaultCounterValue = 234

// GetTestServer returns an httptest.Server backed by a fresh MemStorage
// pre-populated with one gauge and one counter metric named "___test___".
func GetTestServer() (server *httptest.Server) {
	var s = repository.New()
	s.SetGauge("___test___", defaultGaugeValue)
	s.AddCounter("___test___", defaultCounterValue)
	var h = MetricsHandler{Storage: s}
	router := h.New()
	server = httptest.NewServer(router)
	return
}

// expectedHTMLResponse is the expected HTML page returned by GET / for the seeded storage.
var expectedHTMLResponse = `<html><body>
<h1>List of names and results of all currently known metrics</h1>
<b>___test___</b>:   <code>234</code><br><b>___test___</b>:   <code>123</code><br><hr></body></html>
`
