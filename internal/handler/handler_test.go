package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PrometheRus/metricscollector/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestMethods_ServeHTTP(t *testing.T) {
	type response struct {
		code        int
		response    string
		contenttype string
	}
	tests := []struct {
		name   string
		method string
		target string
		want   response
	}{
		// Methods
		{
			name:   "200 if POST",
			method: http.MethodPost,
			target: "http://localhost:8080/update/counter/test/1",
			want: response{
				code:        http.StatusOK,
				response:    "Запрос обновления Counter выполнен!\n",
				contenttype: "text/plain",
			},
		},
		{
			name:   "405 if GET",
			method: http.MethodGet,
			target: "http://localhost:8080/update/counter/test/1",
			want: response{
				code:        http.StatusMethodNotAllowed,
				response:    "Эндпоинт принимает метрики только по протоколу HTTP методом POST\n",
				contenttype: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "405 if PUT",
			method: http.MethodPut,
			target: "http://localhost:8080/update/counter/test/1",
			want: response{
				code:        http.StatusMethodNotAllowed,
				response:    "Эндпоинт принимает метрики только по протоколу HTTP методом POST\n",
				contenttype: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "405 if HEAD",
			method: http.MethodHead,
			target: "http://localhost:8080/update/counter/test/1",
			want: response{
				code:        http.StatusMethodNotAllowed,
				response:    "Эндпоинт принимает метрики только по протоколу HTTP методом POST\n",
				contenttype: "text/plain; charset=utf-8",
			},
		},
	}

	for _, test := range tests {
		var r = repository.NewMemStorage()
		h := &MetricsHandler{Storage: r}

		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()

			h.ServeHTTP(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)

			// проверяем Content-Type
			assert.Equal(t, test.want.contenttype, res.Header.Get("Content-Type"))

			// получаем и проверяем тело запроса
			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			assert.Equal(t, test.want.response, string(resBody))
		})
	}
}

func TestMetricTypes_ServeHTTP(t *testing.T) {
	type response struct {
		code        int
		response    string
		contenttype string
	}
	tests := []struct {
		name   string
		method string
		target string
		want   response
	}{
		{
			name:   "404 if POST without metric type",
			method: http.MethodPost,
			target: "http://localhost:8080/update/",
			want: response{
				code:        http.StatusNotFound,
				response:    "Значение типа метрики не передано\n",
				contenttype: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "400 if POST with 'random' type",
			method: http.MethodPost,
			target: "http://localhost:8080/update/random",
			want: response{
				code:        http.StatusBadRequest,
				response:    "Запрос с некорректным типом метрики\n",
				contenttype: "text/plain; charset=utf-8",
			},
		},
		// Returns 301 (redirect)
		// {
		// 	name:   "404 if POST with empty type",
		// 	method: http.MethodPost,
		// 	target: "http://localhost:8080/update//",
		// 	want: response{
		// 		code:        http.StatusNotFound,
		// 		response:    "Значение типа метрики не передано\n",
		// 		contenttype: "text/plain; charset=utf-8",
		// 	},
		// },
		{
			name:   "200 if POST with 'gauge' type",
			method: http.MethodPost,
			target: "http://localhost:8080/update/gauge/test/1.00",
			want: response{
				code:        http.StatusOK,
				response:    "Запрос обновления Gauge выполнен!\n",
				contenttype: "text/plain",
			},
		},
		{
			name:   "200 if POST 'counter' type",
			method: http.MethodPost,
			target: "http://localhost:8080/update/counter/test/100",
			want: response{
				code:        http.StatusOK,
				response:    "Запрос обновления Counter выполнен!\n",
				contenttype: "text/plain",
			},
		},
	}

	for _, test := range tests {
		var r = repository.NewMemStorage()
		h := &MetricsHandler{Storage: r}

		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()

			h.ServeHTTP(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)

			// проверяем Content-Type
			assert.Equal(t, test.want.contenttype, res.Header.Get("Content-Type"))

			// получаем и проверяем тело запроса
			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			assert.Equal(t, test.want.response, string(resBody))
		})
	}
}

func TestMetricNames_ServeHTTP(t *testing.T) {
	type response struct {
		code        int
		response    string
		contenttype string
	}
	tests := []struct {
		name   string
		method string
		target string
		want   response
	}{
		{
			name:   "404 if POST without gauge metric name",
			method: http.MethodPost,
			target: "http://localhost:8080/update/gauge",
			want: response{
				code:        http.StatusNotFound,
				response:    "Значение имени метрики не передано\n",
				contenttype: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "404 if POST without counter metric name",
			method: http.MethodPost,
			target: "http://localhost:8080/update/counter",
			want: response{
				code:        http.StatusNotFound,
				response:    "Значение имени метрики не передано\n",
				contenttype: "text/plain; charset=utf-8",
			},
		},
	}

	for _, test := range tests {
		var r = repository.NewMemStorage()
		h := &MetricsHandler{Storage: r}

		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()

			h.ServeHTTP(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)

			// проверяем Content-Type
			assert.Equal(t, test.want.contenttype, res.Header.Get("Content-Type"))

			// получаем и проверяем тело запроса
			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			assert.Equal(t, test.want.response, string(resBody))
		})
	}
}

func TestMetricValues_ServeHTTP(t *testing.T) {
	type response struct {
		code        int
		response    string
		contenttype string
	}
	tests := []struct {
		name   string
		method string
		target string
		want   response
	}{
		// Empty value
		{
			name:   "400 if POST gauge value empty",
			method: http.MethodPost,
			target: "http://localhost:8080/update/gauge/name",
			want: response{
				code:        http.StatusBadRequest,
				response:    "Значение метрики не передано\n",
				contenttype: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "400 if POST counter value empty",
			method: http.MethodPost,
			target: "http://localhost:8080/update/counter/name",
			want: response{
				code:        http.StatusBadRequest,
				response:    "Значение метрики не передано\n",
				contenttype: "text/plain; charset=utf-8",
			},
		},
		// Parse Gauge values
		{
			name:   "200 if POST gauge value is positive float",
			method: http.MethodPost,
			target: "http://localhost:8080/update/gauge/name/1.00",
			want: response{
				code:        http.StatusOK,
				response:    "Запрос обновления Gauge выполнен!\n",
				contenttype: "text/plain",
			},
		},
		{
			name:   "200 if POST gauge value is negative float",
			method: http.MethodPost,
			target: "http://localhost:8080/update/gauge/name/-1000.00",
			want: response{
				code:        http.StatusOK,
				response:    "Запрос обновления Gauge выполнен!\n",
				contenttype: "text/plain",
			},
		},
		{
			name:   "200 if POST gauge value is positive int",
			method: http.MethodPost,
			target: "http://localhost:8080/update/gauge/name/1000",
			want: response{
				code:        http.StatusOK,
				response:    "Запрос обновления Gauge выполнен!\n",
				contenttype: "text/plain",
			},
		},
		{
			name:   "200 if POST gauge value is negative int",
			method: http.MethodPost,
			target: "http://localhost:8080/update/gauge/name/-1000",
			want: response{
				code:        http.StatusOK,
				response:    "Запрос обновления Gauge выполнен!\n",
				contenttype: "text/plain",
			},
		},
		{
			name:   "400 if POST gauge value is latin",
			method: http.MethodPost,
			target: "http://localhost:8080/update/gauge/name/value",
			want: response{
				code:        http.StatusBadRequest,
				response:    "Значение метрики некорректно\n",
				contenttype: "text/plain; charset=utf-8",
			},
		},

		// Parse Counter values
		{
			name:   "200 if POST counter value is positive int",
			method: http.MethodPost,
			target: "http://localhost:8080/update/counter/name/1",
			want: response{
				code:        http.StatusOK,
				response:    "Запрос обновления Counter выполнен!\n",
				contenttype: "text/plain",
			},
		},
		{
			name:   "200 if POST counter value is negative int",
			method: http.MethodPost,
			target: "http://localhost:8080/update/counter/name/-1001",
			want: response{
				code:        http.StatusOK,
				response:    "Запрос обновления Counter выполнен!\n",
				contenttype: "text/plain",
			},
		},
		{
			name:   "400 if POST counter value is positive float",
			method: http.MethodPost,
			target: "http://localhost:8080/update/counter/name/1000.00",
			want: response{
				code:        http.StatusBadRequest,
				response:    "Значение метрики некорректно\n",
				contenttype: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "400 if POST counter value is negative float",
			method: http.MethodPost,
			target: "http://localhost:8080/update/counter/name/-1001,00",
			want: response{
				code:        http.StatusBadRequest,
				response:    "Значение метрики некорректно\n",
				contenttype: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "400 if POST counter value is latin",
			method: http.MethodPost,
			target: "http://localhost:8080/update/counter/name/value",
			want: response{
				code:        http.StatusBadRequest,
				response:    "Значение метрики некорректно\n",
				contenttype: "text/plain; charset=utf-8",
			},
		},
	}

	for _, test := range tests {
		var r = repository.NewMemStorage()
		h := &MetricsHandler{Storage: r}

		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()

			h.ServeHTTP(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)

			// проверяем Content-Type
			assert.Equal(t, test.want.contenttype, res.Header.Get("Content-Type"))

			// получаем и проверяем тело запроса
			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			assert.Equal(t, test.want.response, string(resBody))
		})
	}
}
