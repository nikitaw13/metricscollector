package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PrometheRus/metricscollector/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedJSONContentType is the expected Content-Type header for all JSON responses.
const expectedJSONContentType = "application/json; charset=utf-8"

// testJSONRequest sends an HTTP request with a JSON body to the test server
// and returns the full response along with the response body as a string.
func testJSONRequest(t *testing.T, ts *httptest.Server, method, path, body string, contentType string) (*http.Response, string) {
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

	return resp, string(bodyBytes)
}

// JSONTestTemplate holds a single table-driven test case for JSON endpoint tests.
type JSONTestTemplate struct {
	name        string
	method      string
	path        string
	body        string
	contentType string
	want        JSONTestWant
}

// JSONTestWant describes the expected HTTP response for a JSON test case.
type JSONTestWant struct {
	code        int
	contentType string
}

// JSONUpdateTests covers successful POST /update calls (valid counter and gauge payloads).
var JSONUpdateTests = []JSONTestTemplate{
	{
		"Counter positive int",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\", \"id\":\"test\", \"delta\":1000}",
		"application/json",
		JSONTestWant{
			http.StatusOK,
			expectedJSONContentType,
		},
	},
	{
		"Gauge positive int",
		http.MethodPost,
		"/update",
		"{\"type\":\"gauge\", \"id\":\"test\", \"value\":1000}",
		"application/json",
		JSONTestWant{
			http.StatusOK,
			expectedJSONContentType,
		},
	},
	{
		"Gauge negative int",
		http.MethodPost,
		"/update",
		"{\"type\":\"gauge\", \"id\":\"test\", \"value\":-1000}",
		"application/json",
		JSONTestWant{
			http.StatusOK,
			expectedJSONContentType,
		},
	},
	{
		"Gauge positive float",
		http.MethodPost,
		"/update",
		"{\"type\":\"gauge\", \"id\":\"test\", \"value\":1000.00}",
		"application/json",
		JSONTestWant{
			http.StatusOK,
			expectedJSONContentType,
		},
	},
	{
		"Gauge negative float",
		http.MethodPost,
		"/update",
		"{\"type\":\"gauge\", \"id\":\"test\", \"value\":-1000.00}",
		"application/json",
		JSONTestWant{
			http.StatusOK,
			expectedJSONContentType,
		},
	},
}

// validationJSONTests covers POST /update and POST /value with invalid payloads
// (missing/incorrect type, name, or value) — all expected to return 400 or 404.
var validationJSONTests = []JSONTestTemplate{
	// Missing or invalid metric TYPE
	{
		"Missing metric type",
		http.MethodPost,
		"/update",
		"{}",
		"application/json",
		JSONTestWant{
			http.StatusBadRequest,
			expectedJSONContentType,
		},
	},
	{
		"Invalid metric type",
		http.MethodPost,
		"/update",
		"{\"type\":\"random\"}",
		"application/json",
		JSONTestWant{
			http.StatusBadRequest,
			expectedJSONContentType,
		},
	},

	// Missing or invalid metric NAME
	{
		"Missing gauge metric name",
		http.MethodPost,
		"/update",
		"{\"type\":\"gauge\"}",
		"application/json",
		JSONTestWant{
			http.StatusBadRequest,
			expectedJSONContentType,
		},
	},
	{
		"Missing counter metric name",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\"}",
		"application/json",
		JSONTestWant{
			http.StatusBadRequest,
			expectedJSONContentType,
		},
	},
	{
		"Invalid gauge metric name",
		http.MethodPost,
		"/value",
		"{\"type\":\"gauge\",\"id\":\"unknown\"}",
		"application/json",
		JSONTestWant{
			http.StatusNotFound,
			expectedJSONContentType,
		},
	},
	{
		"Invalid counter metric name",
		http.MethodPost,
		"/value",
		"{\"type\":\"counter\",\"id\":\"unknown\"}",
		"application/json",
		JSONTestWant{
			http.StatusNotFound,
			expectedJSONContentType,
		},
	},

	// Missing or invalid metric VALUE
	{
		"Missing gauge metric value",
		http.MethodPost,
		"/update",
		"{\"type\":\"gauge\", \"id\":\"name\"}",
		"application/json",
		JSONTestWant{
			http.StatusBadRequest,
			expectedJSONContentType,
		},
	},
	{
		"Missing counter metric value",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\", \"id\":\"name\"}",
		"application/json",
		JSONTestWant{
			http.StatusBadRequest,
			expectedJSONContentType,
		},
	},
	{
		"Gauge metric value is latin",
		http.MethodPost,
		"/update",
		"{\"type\":\"gauge\", \"id\":\"name\", \"value\":\"LatinText\"}",
		"application/json",
		JSONTestWant{
			http.StatusBadRequest,
			expectedJSONContentType,
		},
	},
	{
		"Counter metric value is latin",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\", \"id\":\"name\", \"delta\":\"LatinText\"}",
		"application/json",
		JSONTestWant{
			http.StatusBadRequest,
			expectedJSONContentType,
		},
	},
	{
		"Counter metric value is positive float",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\", \"id\":\"name\", \"delta\":1001.00}",
		"application/json",
		JSONTestWant{
			http.StatusBadRequest,
			expectedJSONContentType,
		},
	},
	{
		"Counter metric value is negative float",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\", \"id\":\"name\", \"delta\":-1001.00}",
		"application/json",
		JSONTestWant{
			http.StatusBadRequest,
			expectedJSONContentType,
		},
	},
	{
		"Counter metric value is negative int",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\", \"id\":\"test\", \"delta\":-1000}",
		"application/json",
		JSONTestWant{
			http.StatusBadRequest,
			expectedJSONContentType,
		},
	},
}

// runJSONTests executes a slice of table-driven JSON test cases against a test server.
func runJSONTests(t *testing.T, cases []JSONTestTemplate) {
	ts := GetTestServer()
	defer ts.Close()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, _ := testJSONRequest(t, ts, tc.method, tc.path, tc.body, tc.contentType)
			assert.Equal(t, tc.want.code, resp.StatusCode)
			assert.Equal(t, tc.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

// TestJSONValidate verifies that /update and /value reject invalid JSON payloads.
func TestJSONValidate(t *testing.T) {
	runJSONTests(t, validationJSONTests)
}

// TestJSONUpdate verifies that /update accepts valid metric payloads and returns 200.
func TestJSONUpdate(t *testing.T) {
	runJSONTests(t, JSONUpdateTests)
}

// TestJSONRead verifies that /value returns the correct stored metric values in JSON.
func TestJSONRead(t *testing.T) {
	ts := GetTestServer()
	defer ts.Close()

	tests := []struct {
		name string
		body string
		want model.Metrics
	}{
		{
			"Read existing gauge",
			"{\"type\":\"gauge\", \"id\":\"___test___\"}",
			model.Metrics{ID: "___test___", MType: "gauge", Value: ptrFloat64(defaultGaugeValue)},
		},
		{
			"Read existing counter",
			"{\"type\":\"counter\", \"id\":\"___test___\"}",
			model.Metrics{ID: "___test___", MType: "counter", Delta: ptrInt64(defaultCounterValue)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testJSONRequest(t, ts, http.MethodPost, "/value", tt.body, "application/json")
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, expectedJSONContentType, resp.Header.Get("Content-Type"))

			var got model.Metrics
			require.NoError(t, json.Unmarshal([]byte(body), &got))
			assert.Equal(t, tt.want, got)
		})
	}
}

// ptrFloat64 returns a pointer to v, used to build *float64 for test expectations.
func ptrFloat64(v float64) *float64 { return &v }
// ptrInt64 returns a pointer to v, used to build *int64 for test expectations.
func ptrInt64(v int64) *int64      { return &v }