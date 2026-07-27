package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PrometheRus/metricscollector/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Function sends an HTTP request to the test server and returns the response and body.
func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

// Template defines a single table-driven test case.
type tableTestTemplate struct {
	name   string
	method string
	path   string
	want   tableWantTemplate
}

// Template defines expected response for a table-driven test case.
type tableWantTemplate struct {
	code        int
	response    string
	contenttype string
}

// HTTP method validation cases: only POST is allowed on /update/*.
var tableTestMethods = []tableTestTemplate{
	// Test full /update endpoint
	{
		name:   "200 proper POST",
		method: http.MethodPost,
		path:   "/update/counter/test/9999999",
		want: tableWantTemplate{
			code:        http.StatusOK,
			response:    "Metric 'test' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "405 GET",
		method: http.MethodGet,
		path:   "/update/counter/test/9999999",
		want: tableWantTemplate{
			code:        http.StatusMethodNotAllowed,
			response:    "",
			contenttype: "",
		},
	},
	// Test / root endpoint
	{
		name:   "200 GET /",
		method: http.MethodGet,
		path:   "/",
		want: tableWantTemplate{
			code: http.StatusOK,
			response: `<html><body>
<h1>List of names and results of all currently known metrics</h1>
<b>test</b>:   <code>9999999</code><br><hr></body></html>
`,
			contenttype: "text/html; charset=UTF-8",
		},
	},
	{
		name:   "405 POST /",
		method: http.MethodPost,
		path:   "/",
		want: tableWantTemplate{
			code:        http.StatusMethodNotAllowed,
			response:    "",
			contenttype: "",
		},
	},
}

// Metric type validation cases: only "gauge" and "counter" are accepted.
var tableTestTypes = []tableTestTemplate{
	{
		name:   "400 if POST without metric type",
		method: http.MethodPost,
		path:   "/update",
		want: tableWantTemplate{
			code:        http.StatusBadRequest,
			response:    "Metric type is required\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "400 if POST with 'random' type",
		method: http.MethodPost,
		path:   "/update/random/",
		want: tableWantTemplate{
			code:        http.StatusBadRequest,
			response:    "Invalid metric type\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "200 if POST with 'gauge' type",
		method: http.MethodPost,
		path:   "/update/gauge/test/1.00",
		want: tableWantTemplate{
			code:        http.StatusOK,
			response:    "Metric 'test' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "200 if POST with 'counter' type",
		method: http.MethodPost,
		path:   "/update/counter/test/1",
		want: tableWantTemplate{
			code:        http.StatusOK,
			response:    "Metric 'test' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
}

// Metric name validation cases: missing name returns 404.
var tableTestMetrics = []tableTestTemplate{
	{
		name:   "404 POST without gauge metric name",
		method: http.MethodPost,
		path:   "/update/gauge",
		want: tableWantTemplate{
			code:        http.StatusNotFound,
			response:    "Metric is required\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "404 POST without counter metric name",
		method: http.MethodPost,
		path:   "/update/counter",
		want: tableWantTemplate{
			code:        http.StatusNotFound,
			response:    "Metric is required\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "404 GET with unknown metric",
		method: http.MethodGet,
		path:   "/value/counter/unknown",
		want: tableWantTemplate{
			code:        http.StatusNotFound,
			response:    "Counter unknown not found\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
}

// Metric value validation cases: empty, valid, and invalid values for both types.
var tableTestValues = []tableTestTemplate{
	// Empty value
	{
		name:   "400 if POST gauge value empty",
		method: http.MethodPost,
		path:   "/update/gauge/name",
		want: tableWantTemplate{
			code:        http.StatusBadRequest,
			response:    "Metric value is required\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "400 if counter gauge value empty",
		method: http.MethodPost,
		path:   "/update/counter/name",
		want: tableWantTemplate{
			code:        http.StatusBadRequest,
			response:    "Metric value is required\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	// Parse Gauge values
	{
		name:   "200 if POST gauge value is positive float",
		method: http.MethodPost,
		path:   "/update/gauge/name/1.00",
		want: tableWantTemplate{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "200 if POST gauge value is negative float",
		method: http.MethodPost,
		path:   "/update/gauge/name/-1000.00",
		want: tableWantTemplate{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "200 if POST gauge value is positive int",
		method: http.MethodPost,
		path:   "/update/gauge/name/1000",
		want: tableWantTemplate{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "200 if POST gauge value is negative int",
		method: http.MethodPost,
		path:   "/update/gauge/name/-1000",
		want: tableWantTemplate{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "400 if POST gauge value is latin",
		method: http.MethodPost,
		path:   "/update/gauge/name/value",
		want: tableWantTemplate{
			code:        http.StatusBadRequest,
			response:    "Invalid metric value\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},

	// Parse Counter values
	{
		name:   "200 if POST counter value is positive int",
		method: http.MethodPost,
		path:   "/update/counter/name/1",
		want: tableWantTemplate{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "200 if POST counter value is negative int",
		method: http.MethodPost,
		path:   "/update/counter/name/-1001",
		want: tableWantTemplate{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "400 if POST counter value is positive float",
		method: http.MethodPost,
		path:   "/update/counter/name/1000.00",
		want: tableWantTemplate{
			code:        http.StatusBadRequest,
			response:    "Invalid metric value\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "400 if POST counter value is negative float",
		method: http.MethodPost,
		path:   "/update/counter/name/-1001,00",
		want: tableWantTemplate{
			code:        http.StatusBadRequest,
			response:    "Invalid metric value\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "400 if POST counter value is latin",
		method: http.MethodPost,
		path:   "/update/counter/name/value",
		want: tableWantTemplate{
			code:        http.StatusBadRequest,
			response:    "Invalid metric value\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
}

// Function creates a test server with a fresh storage instance and configured routes.
func GetTestRouter() (server *httptest.Server) {
	var s = repository.New()
	var h = MetricsHandler{Storage: s}
	router := h.New()
	server = httptest.NewServer(router)
	return
}

// Test verifies that only POST is allowed on /update/* routes.
func TestMethods(t *testing.T) {
	ts := GetTestRouter()
	defer ts.Close()

	for _, v := range tableTestMethods {
		t.Run(v.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, v.method, v.path)
			// Check Status code
			assert.Equal(t, v.want.code, resp.StatusCode)
			// t.Logf("%v %v\n", v.want.code, resp.StatusCode)

			// Check response
			assert.Equal(t, v.want.response, body)

			// Check Content-Type

			assert.Equal(t, v.want.contenttype, resp.Header.Get("Content-Type"))
		})
	}
}

// Test verifies metric type validation: only "gauge" and "counter" are accepted.
func TestTypes(t *testing.T) {
	ts := GetTestRouter()
	defer ts.Close()

	for _, v := range tableTestTypes {
		t.Run(v.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, v.method, v.path)
			// Check Status code
			assert.Equal(t, v.want.code, resp.StatusCode)
			// t.Logf("%v %v\n", v.want.code, resp.StatusCode)

			// Check response
			assert.Equal(t, v.want.response, body)

			// Check Content-Type

			assert.Equal(t, v.want.contenttype, resp.Header.Get("Content-Type"))
		})
	}
}

// Test verifies that missing metric name returns 404.
func TestMetrics(t *testing.T) {
	ts := GetTestRouter()
	defer ts.Close()

	for _, v := range tableTestMetrics {
		t.Run(v.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, v.method, v.path)
			// Check Status code
			assert.Equal(t, v.want.code, resp.StatusCode)
			// t.Logf("%v %v\n", v.want.code, resp.StatusCode)

			// Check response
			assert.Equal(t, v.want.response, body)

			// Check Content-Type

			assert.Equal(t, v.want.contenttype, resp.Header.Get("Content-Type"))
		})
	}
}

// Test verifies metric value parsing: empty, valid, and invalid values for both types.
func TestValues(t *testing.T) {
	ts := GetTestRouter()
	defer ts.Close()

	for _, v := range tableTestValues {
		t.Run(v.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, v.method, v.path)
			// Check Status code
			assert.Equal(t, v.want.code, resp.StatusCode)
			// t.Logf("%v %v\n", v.want.code, resp.StatusCode)

			// Check response
			assert.Equal(t, v.want.response, body)

			// Check Content-Type

			assert.Equal(t, v.want.contenttype, resp.Header.Get("Content-Type"))
		})
	}
}
