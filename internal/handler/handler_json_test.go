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

// Function sends a JSON request to the test server
// and returns the response and body in JSON
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

// Template defines a single table-driven test case for json requests
type jsonTestTemplate struct {
	name        string
	method      string
	path        string
	body        string
	contentType string
	want        tableWantJsonTemplate
}

// Template defines expected JSON response for a table-driven test case.
type tableWantJsonTemplate struct {
	code        int
	contentType string
	// metrics     *model.Metrics
}

// Update: successful metric updates for json requests
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
			"application/json; charset=utf-8",
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
			"application/json; charset=utf-8",
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
			"application/json; charset=utf-8",
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
			"application/json; charset=utf-8",
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
			"application/json; charset=utf-8",
		},
	},
}

// Read: successful fetching metric values and lists
var jsonReadTests = []jsonTestTemplate{
	{
		"Read existing gauge",
		http.MethodPost,
		"/value",
		"{\"type\":\"gauge\", \"id\":\"___test___\"}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusOK,
			// Тут надо добавить, что ждем 123 в ответе
			"application/json; charset=utf-8",
		},
	},
	{
		"Read existing gauge",
		http.MethodPost,
		"/value",
		"{\"type\":\"counter\", \"id\":\"___test___\"}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusOK,
			// Тут надо добавить, что ждем 234 в ответе
			"application/json; charset=utf-8",
		},
	},
}

// Validation: invalid type, name, or value
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
			"application/json; charset=utf-8",
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
			"application/json; charset=utf-8",
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
			"application/json; charset=utf-8",
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
			"application/json; charset=utf-8",
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
			"application/json; charset=utf-8",
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
			"application/json; charset=utf-8",
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
			"application/json; charset=utf-8",
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
			"application/json; charset=utf-8",
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
			"application/json; charset=utf-8",
		},
	},
	{
		"Counter metric value is latin",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\", \"id\":\"name\", \"value\":\"LatinText\"}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusBadRequest,
			"application/json; charset=utf-8",
		},
	},
	{
		"Counter metric value is positive float",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\", \"id\":\"name\", \"value\":1001,00}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusBadRequest,
			"application/json; charset=utf-8",
		},
	},
	{
		"Counter metric value is negative float",
		http.MethodPost,
		"/update",
		"{\"type\":\"counter\", \"id\":\"name\", \"value\":-1001,00}",
		"application/json",
		tableWantJsonTemplate{
			http.StatusBadRequest,
			"application/json; charset=utf-8",
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
			"application/json; charset=utf-8",
		},
	},
}

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

// func TestJsonRead(t *testing.T) {
// 	runJsonTests(t, jsonReadTests)
// }

func TestJsonValidate(t *testing.T) {
	runJsonTests(t, validationJsonTests)
}

func TestJsonUpdate(t *testing.T) {
	runJsonTests(t, jsonUpdateTests)
}

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
			assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

			var got model.Metrics
			require.NoError(t, json.Unmarshal([]byte(body), &got))
			assert.Equal(t, tt.want, got)
		})
	}
}

func fPtr(v float64) *float64 { return &v }
func iPtr(v int64) *int64     { return &v }
