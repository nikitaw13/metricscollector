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

// Update: successful metric updates
var updateTests = []tableTestTemplate{
	{
		"Counter positive int", http.MethodPost, "/update/counter/test/1000", tableWantTemplate{
			http.StatusOK,
			"Metric 'test' updated✅\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Counter negative int", http.MethodPost, "/update/counter/test/-1001", tableWantTemplate{
			http.StatusOK,
			"Metric 'test' updated✅\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Gauge positive int", http.MethodPost, "/update/gauge/test/1000", tableWantTemplate{
			http.StatusOK,
			"Metric 'test' updated✅\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Gauge negative int", http.MethodPost, "/update/gauge/test/-1000", tableWantTemplate{
			http.StatusOK,
			"Metric 'test' updated✅\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Gauge positive float", http.MethodPost, "/update/gauge/test/1000.00", tableWantTemplate{
			http.StatusOK,
			"Metric 'test' updated✅\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Gauge negative float", http.MethodPost, "/update/gauge/test/-1000.00", tableWantTemplate{
			http.StatusOK,
			"Metric 'test' updated✅\n",
			"text/plain; charset=utf-8",
		},
	},
}

var htmlResponse = `<html><body>
<h1>List of names and results of all currently known metrics</h1>
<b>___test___</b>:   <code>234</code><br><b>___test___</b>:   <code>123</code><br><hr></body></html>
`

// Read: successful fetching metric values and lists
var readTests = []tableTestTemplate{
	{
		"Read all metrics", http.MethodGet, "/", tableWantTemplate{
			http.StatusOK,
			htmlResponse,
			"text/html; charset=UTF-8",
		},
	},
	{
		"Read existing gauge", http.MethodGet, "/value/gauge/___test___", tableWantTemplate{
			code:        http.StatusOK,
			response:    "123\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		"Read existing counter", http.MethodGet, "/value/counter/___test___", tableWantTemplate{
			code:        http.StatusOK,
			response:    "234\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
}

// Validation: invalid type, name, or value
var validationTests = []tableTestTemplate{
	// Missing or invalid metric type
	{
		"Missing metric type", http.MethodPost, "/update", tableWantTemplate{
			http.StatusBadRequest,
			"Type is required\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Invalid metric type", http.MethodPost, "/update/random/", tableWantTemplate{
			http.StatusBadRequest,
			"Invalid metric type\n",
			"text/plain; charset=utf-8",
		},
	},

	// Missing or invalid metric name
	{
		"Missing gauge metric name", http.MethodPost, "/update/gauge", tableWantTemplate{
			http.StatusNotFound,
			"Metric ID is required\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Missing counter metric name", http.MethodPost, "/update/counter", tableWantTemplate{
			http.StatusNotFound,
			"Metric ID is required\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Invalid gauge metric name", http.MethodGet, "/value/gauge/unknown", tableWantTemplate{
			http.StatusNotFound,
			"Gauge unknown not found\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Invalid counter metric name", http.MethodGet, "/value/counter/unknown", tableWantTemplate{
			http.StatusNotFound,
			"Counter unknown not found\n",
			"text/plain; charset=utf-8",
		},
	},

	// Missing or invalid metric value
	{
		"Missing gauge metric value", http.MethodPost, "/update/gauge/name", tableWantTemplate{
			http.StatusBadRequest,
			"Metric value is required\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Missing counter metric value", http.MethodPost, "/update/counter/name", tableWantTemplate{
			http.StatusBadRequest,
			"Metric value is required\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Gauge metric value is latin", http.MethodPost, "/update/gauge/name/value", tableWantTemplate{
			http.StatusBadRequest,
			"Invalid metric value\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Counter metric value is latin", http.MethodPost, "/update/counter/name/value", tableWantTemplate{
			http.StatusBadRequest,
			"Invalid metric value\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Counter metric value is positive float", http.MethodPost, "/update/counter/name/1000.00", tableWantTemplate{
			http.StatusBadRequest,
			"Invalid metric value\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Counter metric value is negative float", http.MethodPost, "/update/counter/name/-1001,00", tableWantTemplate{
			http.StatusBadRequest,
			"Invalid metric value\n",
			"text/plain; charset=utf-8",
		},
	},
}

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

func runTableTests(t *testing.T, cases []tableTestTemplate) {
	ts := GetTestRouter()
	defer ts.Close()

	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, v.method, v.path)
			assert.Equal(t, v.want.code, resp.StatusCode)
			assert.Equal(t, v.want.response, body)
			assert.Equal(t, v.want.contenttype, resp.Header.Get("Content-Type"))
		})
	}
}

func TestUpdate(t *testing.T) {
	runTableTests(t, updateTests)
}

func TestRead(t *testing.T) {
	runTableTests(t, readTests)
}
func TestValidate(t *testing.T) {
	runTableTests(t, validationTests)
}
