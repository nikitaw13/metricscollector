package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

var testTable = []struct {
	name   string
	method string
	path   string
	want   string
	status int
}{
	{"405 GET", "GET", "/update/counter/test/1", "...", 405},
	{"400 invalid type", "POST", "/update/random/test/1", "...", 400},
	{"404 no metric name", "POST", "/update/gauge", "...", 404},
	{"400 no value", "POST", "/update/gauge/name", "...", 400},
	{"200 ok gauge", "POST", "/update/gauge/name/1.00", "...", 200},
	// ...
}

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

type table_test_template struct {
	name   string
	method string
	path   string
	want   table_want_template
}

type table_want_template struct {
	code        int
	response    string
	contenttype string
}

var table_test_methods = []table_test_template{
	{
		name:   "200 if proper POST",
		method: http.MethodPost,
		path:   "/update/counter/test/1",
		want: table_want_template{
			code:        http.StatusOK,
			response:    "Metric 'test' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "405 if GET /update",
		method: http.MethodGet,
		path:   "/update/counter/test/1",
		want: table_want_template{
			code:        http.StatusMethodNotAllowed,
			response:    "",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "405 if PUT /update",
		method: http.MethodPut,
		path:   "/update/counter/test/1",
		want: table_want_template{
			code:        http.StatusMethodNotAllowed,
			response:    "",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "405 if HEAD /update",
		method: http.MethodHead,
		path:   "/update/counter/test/1",
		want: table_want_template{
			code:        http.StatusMethodNotAllowed,
			response:    "",
			contenttype: "text/plain; charset=utf-8",
		},
	},
}

var table_test_types = []table_test_template{
	{
		name:   "404 if POST without metric type",
		method: http.MethodPost,
		path:   "/update/",
		want: table_want_template{
			code:        http.StatusNotFound,
			response:    "Metric type is required\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "400 if POST with 'random' type",
		method: http.MethodPost,
		path:   "/update/random/",
		want: table_want_template{
			code:        http.StatusBadRequest,
			response:    "Invalid metric type\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "200 if POST with 'gauge' type",
		method: http.MethodPost,
		path:   "/update/gauge/test/1.00",
		want: table_want_template{
			code:        http.StatusOK,
			response:    "Metric 'test' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "200 if POST with 'counter' type",
		method: http.MethodPost,
		path:   "/update/counter/test/1.00",
		want: table_want_template{
			code:        http.StatusOK,
			response:    "Metric 'test' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
}

var table_test_metrics = []table_test_template{
	{
		name:   "404 if POST without gauge metric name",
		method: http.MethodPost,
		path:   "/update/gauge",
		want: table_want_template{
			code:        http.StatusNotFound,
			response:    "Invalid metric value",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "404 if POST without counter metric name",
		method: http.MethodPost,
		path:   "/update/counter",
		want: table_want_template{
			code:        http.StatusNotFound,
			response:    "Invalid metric value",
			contenttype: "text/plain; charset=utf-8",
		},
	},
}

var table_test_values = []table_test_template{
	// Empty value
	{
		name:   "400 if POST gauge value empty",
		method: http.MethodPost,
		path:   "/update/gauge/name",
		want: table_want_template{
			code:        http.StatusBadRequest,
			response:    "Metric value is required\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "400 if counter gauge value empty",
		method: http.MethodPost,
		path:   "/update/counter/name",
		want: table_want_template{
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
		want: table_want_template{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain",
		},
	},
	{
		name:   "200 if POST gauge value is negative float",
		method: http.MethodPost,
		path:   "/update/gauge/name/-1000.00",
		want: table_want_template{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain",
		},
	},
	{
		name:   "200 if POST gauge value is positive int",
		method: http.MethodPost,
		path:   "/update/gauge/name/1000",
		want: table_want_template{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain",
		},
	},
	{
		name:   "200 if POST gauge value is negative int",
		method: http.MethodPost,
		path:   "/update/gauge/name/-1000",
		want: table_want_template{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain",
		},
	},
	{
		name:   "400 if POST gauge value is latin",
		method: http.MethodPost,
		path:   "/update/gauge/name/value",
		want: table_want_template{
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
		want: table_want_template{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain",
		},
	},
	{
		name:   "200 if POST counter value is negative int",
		method: http.MethodPost,
		path:   "/update/counter/name/-1001",
		want: table_want_template{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain",
		},
	},
	{
		name:   "400 if POST counter value is positive float",
		method: http.MethodPost,
		path:   "/update/counter/name/1000.00",
		want: table_want_template{
			code:        http.StatusBadRequest,
			response:    "Invalid metric value\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "400 if POST counter value is negative float",
		method: http.MethodPost,
		path:   "/update/counter/name/-1001,00",
		want: table_want_template{
			code:        http.StatusBadRequest,
			response:    "Invalid metric value\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "400 if POST counter value is latin",
		method: http.MethodPost,
		path:   "/update/counter/name/value",
		want: table_want_template{
			code:        http.StatusBadRequest,
			response:    "Invalid metric value\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
}

func Test_ServeHTTP(t *testing.T) {
	// TODO()
}
