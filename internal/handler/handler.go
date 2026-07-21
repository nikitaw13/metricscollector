package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PrometheRus/metricscollector/internal/model"
	"github.com/PrometheRus/metricscollector/internal/repository"
)

type MetricsHandler struct {
	Storage repository.Repository
}

func (h *MetricsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// Если метод != POST - остальные проверки не имеют смысла
	if req.Method != http.MethodPost {
		http.Error(res, "Эндпоинт принимает метрики только по протоколу HTTP методом POST", http.StatusMethodNotAllowed)
		return
	}

	// 0 - update, 1 - type, 2 - key, 3 - value
	segments := strings.Split(strings.Trim(req.URL.Path, "/"), "/")

	// При попытке передать запрос без типа метрики возвращать http.StatusNotFound
	// !Отсутствует в постановке задачи!
	if len(segments) == 1 {
		http.Error(res, "Значение типа метрики не передано", http.StatusNotFound)
		return
	}

	// При попытке передать запрос с некорректным типом метрики возвращать http.StatusBadRequest
	mType := segments[1]
	if mType != model.Gauge && mType != model.Counter {
		http.Error(res, "Запрос с некорректным типом метрики", http.StatusBadRequest)
		return
	}

	// При попытке передать запрос без имени метрики возвращать http.StatusNotFound
	if len(segments) == 2 || segments[2] == "" {
		http.Error(res, "Значение имени метрики не передано", http.StatusNotFound)
		return
	}

	// При попытке передать запрос без значения метрики возвращать http.StatusNotFound
	if len(segments) == 3 {
		http.Error(res, "Значение метрики не передано", http.StatusBadRequest)
		return
	}

	mName := segments[2]

	switch mType {
	case "gauge":
		mGaugeValue, err := strconv.ParseFloat(segments[3], 64)
		// При попытке передать запрос с некорректным значением метрики возвращать http.StatusBadRequest.
		if err != nil {
			http.Error(res, "Значение метрики некорректно", http.StatusBadRequest)
			return
		}

		// При ошибке обработки запроса обновления Gauge возвращать http.StatusInternalServerError
		if err := h.Storage.UpdateGauge(mName, mGaugeValue); err != nil {
			http.Error(res, "Ошибка при обновлении Gauge", http.StatusInternalServerError)
			return
		}
		//io.WriteString(res, "Запрос обновления Gauge выполнен!")

	case "counter":
		mCounterValue, err := strconv.ParseInt(segments[3], 10, 64)
		// При попытке передать запрос с некорректным значением метрики возвращать http.StatusBadRequest.
		if err != nil {
			http.Error(res, "Значение метрики некорректно", http.StatusBadRequest)
			return
		}
		// При ошибке обработки запроса обновления Counter возвращать http.StatusInternalServerError
		if err := h.Storage.UpdateCounter(mName, mCounterValue); err != nil {
			http.Error(res, "Ошибка при обновлении Counter", http.StatusInternalServerError)
			return
		}
		// io.WriteString(res, "Запрос обновления Counter выполнен!")
	}
	// Никаких ошибок не получили, HTTP запрос успешно обработан
	res.WriteHeader(http.StatusOK)
	log.Printf("The handler recieved request: %s %s %s and the response is %d", req.RemoteAddr, req.Method, req.URL.Path, http.StatusOK)
}
