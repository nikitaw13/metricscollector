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

// desiredContentType is the expected Content-Type header for all JSON responses.
const desiredContentType = "application/json; charset=utf-8"

// testJsonRequest sends an HTTP request with a JSON body to the test server
// and returns the full response along with the response body as a string.
func testJsonRequest(t *testing.T, ts *httptest.Server, method, path, body string, contentType string) (*http.Response, string) {
	var buf io.Reader
	if body != "" {
		buf = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, ts.URL+path, buf)
	require.NoError(t, err)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

// jsonTestTemplate holds a single table-driven test case for JSON endpoint tests.
type jsonTestTemplate struct {
	name        string
	method      string
	path        string
	body        string
	contentType string
	want        tableWantJsonTemplate
}

// tableWantJsonTemplate describes the expected HTTP response for a JSON test case.
type tableWantJsonTemplate struct {
	code        int
	contentType string
	// metrics     *model.Metrics
}

// jsonUpdateTests covers successful POST /update calls (valid counter and gauge payloads).
var jsonUpdateTests = []jsonTestTemplate{
	{
		"Counter positive int",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\", \"id\":\"test\", \"delta\":1000}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusOK,
			// Тут надо добавить, что ждем 1000 в ответе
			desiredContentType,
		},
	},
	{
		"Gauge positive int",
		http.MethodPost,
		"/update",
		"{\"type\":\"gauge\", \"id\":\"test\", \"value\":1000}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusOK,
			// Тут надо добавить, что ждем 1000 в ответе
			desiredContentType,
		},
	},
	{
		"Gauge negative int",
		http.MethodPost,
		"/update",
		"{\"type\":\"gauge\", \"id\":\"test\", \"value\":-1000}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusOK,
			// Тут надо добавить, что ждем -1000 в ответе
			desiredContentType,
		},
	},
	{
		"Gauge positive float",
		http.MethodPost,
		"/update",
		"{\"type\":\"gauge\", \"id\":\"test\", \"value\":1000.00}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusOK,
			// Тут надо добавить, что ждем 1000 в ответе
			desiredContentType,
		},
	},
	{
		"Gauge positive float",
		http.MethodPost,
		"/update",
		"{\"type\":\"gauge\", \"id\":\"test\", \"value\":-1000.00}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusOK,
			// Тут надо добавить, что ждем -1000 в ответе
			desiredContentType,
		},
	},
}

// validationJsonTests covers POST /update and POST /value with invalid payloads
// (missing/incorrect type, name, or value) — all expected to return 400 or 404.
var validationJsonTests = []jsonTestTemplate{
	// Missing or invalid metric TYPE
	{
		"Missing metric type",
		http.MethodPost,
		"/update",
		"{}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusBadRequest,
			desiredContentType,
		},
	},
	{
		"Invalid metric type",
		http.MethodPost,
		"/update",
		"{\"type\":\"random\"}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusBadRequest,
			desiredContentType,
		},
	},

	// Missing or invalid metric NAME
	{
		"Missing gauge metric name",
		http.MethodPost,
		"/update",
		"{\"type\":\"gauge\"}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusBadRequest,
			desiredContentType,
		},
	},
	{
		"Missing counter metric name",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\"}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusBadRequest,
			desiredContentType,
		},
	},
	{
		"Invalid gauge metric name",
		http.MethodPost,
		"/value",
		"{\"type\":\"gauge\",\"id\":\"unknown\"}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusNotFound,
			desiredContentType,
		},
	},
	{
		"Invalid counter metric name",
		http.MethodPost,
		"/value",
		"{\"type\":\"counter\",\"id\":\"unknown\"}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusNotFound,
			desiredContentType,
		},
	},

	// Missing or invalid metric VALUE
	{
		"Missing gauge metric value",
		http.MethodPost,
		"/update",
		"{\"type\":\"gauge\", \"id\":\"name\"}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusBadRequest,
			desiredContentType,
		},
	},
	{
		"Missing counter metric value",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\", \"id\":\"name\"}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusBadRequest,
			desiredContentType,
		},
	},
	{
		"Gauge metric value is latin",
		http.MethodPost,
		"/update",
		"{\"type\":\"gauge\", \"id\":\"name\", \"value\":\"LatinText\"}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusBadRequest,
			desiredContentType,
		},
	},
	{
		"Counter metric value is latin",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\", \"id\":\"name\", \"delta\":\"LatinText\"}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusBadRequest,
			desiredContentType,
		},
	},
	{
		"Counter metric value is positive float",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\", \"id\":\"name\", \"delta\":1001.00}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusBadRequest,
			desiredContentType,
		},
	},
	{
		"Counter metric value is negative float",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\", \"id\":\"name\", \"delta\":-1001.00}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusBadRequest,
			desiredContentType,
		},
	},
	{
		"Counter metric value is negative int",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\", \"id\":\"test\", \"delta\":-1000}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusBadRequest,
			desiredContentType,
		},
	},
}

// runJsonTests executes a slice of table-driven JSON test cases against a test server.
func runJsonTests(t *testing.T, cases []jsonTestTemplate) {
	ts := GetTestRouter()
	defer ts.Close()

	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			resp, _ := testJsonRequest(t, ts, v.method, v.path, v.body, v.contentType)
			assert.Equal(t, v.want.code, resp.StatusCode)
			assert.Equal(t, v.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

// TestJsonValidate verifies that /update and /value reject invalid JSON payloads.
func TestJsonValidate(t *testing.T) {
	runJsonTests(t, validationJsonTests)
}

// TestJsonUpdate verifies that /update accepts valid metric payloads and returns 200.
func TestJsonUpdate(t *testing.T) {
	runJsonTests(t, jsonUpdateTests)
}

// TestJsonRead verifies that /value returns the correct stored metric values in JSON.
func TestJsonRead(t *testing.T) {
	ts := GetTestRouter()
	defer ts.Close()

	tests := []struct {
		name string
		body string
		want model.Metrics
	}{
		{
			"Read existing gauge",
			"{\"type\":\"gauge\", \"id\":\"___test___\"}",
			model.Metrics{ID: "___test___", MType: "gauge", Value: fPtr(defaultGaugeValue)},
		},
		{
			"Read existing counter",
			"{\"type\":\"counter\", \"id\":\"___test___\"}",
			model.Metrics{ID: "___test___", MType: "counter", Delta: iPtr(defaultCounterValue)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testJsonRequest(t, ts, http.MethodPost, "/value", tt.body, "application/json")
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, desiredContentType, resp.Header.Get("Content-Type"))

			var got model.Metrics
			require.NoError(t, json.Unmarshal([]byte(body), &got))
			assert.Equal(t, tt.want, got)
		})
	}
}

// fPtr returns a pointer to v, used to build *float64 for test expectations.
func fPtr(v float64) *float64 { return &v }
// iPtr returns a pointer to v, used to build *int64 for test expectations.
func iPtr(v int64) *int64     { return &v }
