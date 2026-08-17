package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testPlainRequest sends an HTTP request to the test server and returns the response and body.
func testPlainRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(body)
}

// plainTestTemplate holds a single table-driven test case for plain-text endpoints.
type plainTestTemplate struct {
	name   string
	method string
	path   string
	want   plainTestWant
}

// plainTestWant describes the expected HTTP response for a plain-text test case.
type plainTestWant struct {
	code        int
	response    string
	contentType string
}

// plainUpdateTests covers successful metric updates via URL-encoded endpoints.
var plainUpdateTests = []plainTestTemplate{
	{
		"Counter positive int", http.MethodPost, "/update/counter/test/1000", plainTestWant{
			http.StatusOK,
			"Metric 'test' updated✅\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Gauge positive int", http.MethodPost, "/update/gauge/test/1000", plainTestWant{
			http.StatusOK,
			"Metric 'test' updated✅\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Gauge negative int", http.MethodPost, "/update/gauge/test/-1000", plainTestWant{
			http.StatusOK,
			"Metric 'test' updated✅\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Gauge positive float", http.MethodPost, "/update/gauge/test/1000.00", plainTestWant{
			http.StatusOK,
			"Metric 'test' updated✅\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Gauge negative float", http.MethodPost, "/update/gauge/test/-1000.00", plainTestWant{
			http.StatusOK,
			"Metric 'test' updated✅\n",
			"text/plain; charset=utf-8",
		},
	},
}

// plainReadTests covers successful metric reads via URL-encoded endpoints.
var plainReadTests = []plainTestTemplate{
	{
		"Read all metrics", http.MethodGet, "/", plainTestWant{
			http.StatusOK,
			htmlResponse,
			"text/html; charset=UTF-8",
		},
	},
	{
		"Read existing gauge", http.MethodGet, "/value/gauge/___test___", plainTestWant{
			code:        http.StatusOK,
			response:    "123\n",
			contentType: "text/plain; charset=utf-8",
		},
	},
	{
		"Read existing counter", http.MethodGet, "/value/counter/___test___", plainTestWant{
			code:        http.StatusOK,
			response:    "234\n",
			contentType: "text/plain; charset=utf-8",
		},
	},
}

// plainValidationTests covers invalid requests to URL-encoded endpoints.
var plainValidationTests = []plainTestTemplate{
	// Invalid metric type
	{
		"Invalid metric type", http.MethodPost, "/update/random/", plainTestWant{
			http.StatusBadRequest,
			"Invalid metric type\n",
			"text/plain; charset=utf-8",
		},
	},

	// Missing or invalid metric name
	{
		"Missing gauge metric name", http.MethodPost, "/update/gauge", plainTestWant{
			http.StatusNotFound,
			"Metric ID is required\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Missing counter metric name", http.MethodPost, "/update/counter", plainTestWant{
			http.StatusNotFound,
			"Metric ID is required\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Invalid gauge metric name", http.MethodGet, "/value/gauge/unknown", plainTestWant{
			http.StatusNotFound,
			"Gauge unknown not found\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Invalid counter metric name", http.MethodGet, "/value/counter/unknown", plainTestWant{
			http.StatusNotFound,
			"Counter unknown not found\n",
			"text/plain; charset=utf-8",
		},
	},

	// Missing or invalid metric value
	{
		"Missing gauge metric value", http.MethodPost, "/update/gauge/name", plainTestWant{
			http.StatusBadRequest,
			"Metric value is required\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Missing counter metric value", http.MethodPost, "/update/counter/name", plainTestWant{
			http.StatusBadRequest,
			"Metric value is required\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Gauge metric value is latin", http.MethodPost, "/update/gauge/name/value", plainTestWant{
			http.StatusBadRequest,
			"Invalid metric value\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Counter metric value is latin", http.MethodPost, "/update/counter/name/value", plainTestWant{
			http.StatusBadRequest,
			"Invalid metric value\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Counter metric value is positive float", http.MethodPost, "/update/counter/name/1000.00", plainTestWant{
			http.StatusBadRequest,
			"Invalid metric value\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Counter metric value is negative float", http.MethodPost, "/update/counter/name/-1001,00", plainTestWant{
			http.StatusBadRequest,
			"Invalid metric value\n",
			"text/plain; charset=utf-8",
		},
	},
	{
		"Counter metric value is negative int", http.MethodPost, "/update/counter/test/-1001", plainTestWant{
			http.StatusBadRequest,
			"Invalid metric value\n",
			"text/plain; charset=utf-8",
		},
	},
}

// runPlainTests executes a slice of table-driven plain-text test cases against a test server.
func runPlainTests(t *testing.T, cases []plainTestTemplate) {
	ts := GetTestServer()
	defer ts.Close()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, body := testPlainRequest(t, ts, tc.method, tc.path)
			assert.Equal(t, tc.want.code, resp.StatusCode)
			assert.Equal(t, tc.want.response, body)
			assert.Equal(t, tc.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

func TestPlainUpdate(t *testing.T) {
	runPlainTests(t, plainUpdateTests)
}

func TestPlainRead(t *testing.T) {
	runPlainTests(t, plainReadTests)
}

func TestPlainValidate(t *testing.T) {
	runPlainTests(t, plainValidationTests)
}