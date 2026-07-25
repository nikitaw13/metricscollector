package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PrometheRus/metricscollector/internal/repository"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		name:   "405 if GET",
		method: http.MethodGet,
		path:   "/update/counter/test/1",
		want: table_want_template{
			code:        http.StatusMethodNotAllowed,
			response:    "",
			contenttype: "",
		},
	},
	{
		name:   "405 if PUT /update",
		method: http.MethodPut,
		path:   "/update/counter/test/1",
		want: table_want_template{
			code:        http.StatusMethodNotAllowed,
			response:    "",
			contenttype: "",
		},
	},
	{
		name:   "405 if HEAD /update",
		method: http.MethodHead,
		path:   "/update/counter/test/1",
		want: table_want_template{
			code:        http.StatusMethodNotAllowed,
			response:    "",
			contenttype: "",
		},
	},
}

var table_test_types = []table_test_template{
	{
		name:   "400 if POST without metric type",
		method: http.MethodPost,
		path:   "/update",
		want: table_want_template{
			code:        http.StatusBadRequest,
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
		path:   "/update/counter/test/1",
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
			response:    "Metric is required\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "404 if POST without counter metric name",
		method: http.MethodPost,
		path:   "/update/counter",
		want: table_want_template{
			code:        http.StatusNotFound,
			response:    "Metric is required\n",
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
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "200 if POST gauge value is negative float",
		method: http.MethodPost,
		path:   "/update/gauge/name/-1000.00",
		want: table_want_template{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "200 if POST gauge value is positive int",
		method: http.MethodPost,
		path:   "/update/gauge/name/1000",
		want: table_want_template{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "200 if POST gauge value is negative int",
		method: http.MethodPost,
		path:   "/update/gauge/name/-1000",
		want: table_want_template{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
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
			contenttype: "text/plain; charset=utf-8",
		},
	},
	{
		name:   "200 if POST counter value is negative int",
		method: http.MethodPost,
		path:   "/update/counter/name/-1001",
		want: table_want_template{
			code:        http.StatusOK,
			response:    "Metric 'name' updated✅\n",
			contenttype: "text/plain; charset=utf-8",
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

func TestMethods(t *testing.T) {
	var s = repository.NewServerStorage()
	var h = MetricsHandler{Storage: s}

	router := chi.NewRouter()

	router.Get("/", h.GetRootHandler)
	router.Get("/value/{TYPE}/{METRIC}", h.GetMetricHandler)

	router.Post("/update", h.PostNoTypeHandler)
	router.Route("/update/{TYPE}", func(r chi.Router) {
		r.Use(TypeMiddleware)                     // проверит TYPE для всех трёх ниже
		r.Post("/", h.PostNoMetricHandler)        // 404 — нет имени
		r.Post("/{METRIC}", h.PostNoValueHandler) // 400 — нет значения
		r.Post("/{METRIC}/{VALUE}", h.PostFullHandler)
	})

	ts := httptest.NewServer(router)

	defer ts.Close()

	for _, v := range table_test_methods {
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

func TestTypes(t *testing.T) {
	var s = repository.NewServerStorage()
	var h = MetricsHandler{Storage: s}

	router := chi.NewRouter()

	router.Get("/", h.GetRootHandler)
	router.Get("/value/{TYPE}/{METRIC}", h.GetMetricHandler)

	router.Post("/update", h.PostNoTypeHandler)
	router.Route("/update/{TYPE}", func(r chi.Router) {
		r.Use(TypeMiddleware)                     // проверит TYPE для всех трёх ниже
		r.Post("/", h.PostNoMetricHandler)        // 404 — нет имени
		r.Post("/{METRIC}", h.PostNoValueHandler) // 400 — нет значения
		r.Post("/{METRIC}/{VALUE}", h.PostFullHandler)
	})

	ts := httptest.NewServer(router)

	defer ts.Close()

	for _, v := range table_test_types {
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

func TestMetrics(t *testing.T) {
	var s = repository.NewServerStorage()
	var h = MetricsHandler{Storage: s}

	router := chi.NewRouter()

	router.Get("/", h.GetRootHandler)
	router.Get("/value/{TYPE}/{METRIC}", h.GetMetricHandler)

	router.Post("/update", h.PostNoTypeHandler)
	router.Route("/update/{TYPE}", func(r chi.Router) {
		r.Use(TypeMiddleware)                     // проверит TYPE для всех трёх ниже
		r.Post("/", h.PostNoMetricHandler)        // 404 — нет имени
		r.Post("/{METRIC}", h.PostNoValueHandler) // 400 — нет значения
		r.Post("/{METRIC}/{VALUE}", h.PostFullHandler)
	})

	ts := httptest.NewServer(router)

	defer ts.Close()

	for _, v := range table_test_metrics {
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

func TestValues(t *testing.T) {
	var s = repository.NewServerStorage()
	var h = MetricsHandler{Storage: s}

	router := chi.NewRouter()

	router.Get("/", h.GetRootHandler)
	router.Get("/value/{TYPE}/{METRIC}", h.GetMetricHandler)

	router.Post("/update", h.PostNoTypeHandler)
	router.Route("/update/{TYPE}", func(r chi.Router) {
		r.Use(TypeMiddleware)                     // проверит TYPE для всех трёх ниже
		r.Post("/", h.PostNoMetricHandler)        // 404 — нет имени
		r.Post("/{METRIC}", h.PostNoValueHandler) // 400 — нет значения
		r.Post("/{METRIC}/{VALUE}", h.PostFullHandler)
	})

	ts := httptest.NewServer(router)

	defer ts.Close()

	for _, v := range table_test_values {
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
