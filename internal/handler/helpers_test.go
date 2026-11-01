package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nikitaw13/metricscollector/internal/repository"
	"github.com/nikitaw13/metricscollector/internal/service"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	// defaultGaugeValue is the pre-seeded gauge value used across handler tests.
	defaultGaugeValue = 123.00

	// defaultCounterValue is the pre-seeded counter value used across handler tests.
	defaultCounterValue = 234

	expectedJSONContentType = "application/json; charset=utf-8"

	expectedHTMLResponse = "<html><body>\n<h1>List of names and results of all currently known metrics</h1>\n<p><b>___test___</b>: <code>234</code></p>\n<p><b>___test___</b>: <code>123</code></p>\n<hr></body></html>"
)

type testResponse struct {
	statusCode int
	header     http.Header
	body       string
}

// newTestServer returns a test HTTP server backed by an in-memory storage
// seeded with the default gauge and counter, with no signing key configured.
func newTestServer() (server *httptest.Server) {
	storage := repository.NewMemStorage()
	var dbPinger service.DBPinger
	svc := service.New(storage, dbPinger, *zap.NewNop())

	storage.SetGauge("___test___", defaultGaugeValue)
	storage.AddCounter("___test___", defaultCounterValue)

	metricsHandler := NewMetricsHandler("", svc, zap.NewNop())
	router := metricsHandler.Router()
	server = httptest.NewServer(router)
	return
}

// newTestServerWithKey returns a test HTTP server like newTestServer, but
// configured with the given HMAC signing key.
func newTestServerWithKey(key string) (server *httptest.Server) {
	storage := repository.NewMemStorage()
	var dbPinger service.DBPinger
	svc := service.New(storage, dbPinger, *zap.NewNop())

	storage.SetGauge("___test___", defaultGaugeValue)
	storage.AddCounter("___test___", defaultCounterValue)
	metricsHandler := NewMetricsHandler(key, svc, zap.NewNop())
	router := metricsHandler.Router()
	server = httptest.NewServer(router)
	return
}

// doJSONRequest sends an HTTP request with a JSON body to the test server
// and returns a testResponse with the status code, headers, and body.
func doJSONRequest(t *testing.T, ts *httptest.Server, method, path, body string, contentType string) testResponse {
	t.Helper()
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, ts.URL+path, bodyReader)
	require.NoError(t, err)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return testResponse{statusCode: resp.StatusCode, header: resp.Header, body: string(bodyBytes)}
}

// ptrFloat64 returns a pointer to value, used to build *float64 for test expectations.
func ptrFloat64(value float64) *float64 { return &value }

// ptrInt64 returns a pointer to value, used to build *int64 for test expectations.
func ptrInt64(value int64) *int64 { return &value }

// doRequest sends an HTTP request to the test server and returns a
// testResponse with the status code, headers, and body.
func doRequest(t *testing.T, ts *httptest.Server, method, path string) testResponse {
	t.Helper()
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return testResponse{statusCode: resp.StatusCode, header: resp.Header, body: string(respBody)}
}

// expectedResponse describes the anticipated HTTP response; body is left
// empty for cases that do not assert the response body.
type expectedResponse struct {
	statusCode  int
	body        string
	contentType string
}

// testHashKey is the HMAC key shared by the hash middleware tests.
const testHashKey = "TestKey"

// fixedResponseHandler returns a handler replying with the given Content-Type and body.
func fixedResponseHandler(contentType, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write([]byte(body))
	})
}
