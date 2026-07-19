package handler

import (
	"net/http"
	"strconv"
	"strings"
)

type MetricsHandler struct{}

func (h MetricsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// Если метод != POST - остальные проверки не имеют смысла
	if req.Method != http.MethodPost {
		http.Error(res, "Эндпоинт принимает метрики только по протоколу HTTP методом POST", http.StatusMethodNotAllowed)
		return
	}

	// 0 - update, 1 - type, 2 - key, 3 - value
	segments := strings.Split(strings.Trim(req.URL.Path, "/"), "/")

	// При попытке передать запрос без имени метрики возвращать http.StatusNotFound
	if len(segments) == 2 {
		http.Error(res, "Значение имени метрики не передано", http.StatusNotFound)
		return
	}

	// При попытке передать запрос без значения метрики возвращать http.StatusNotFound
	if len(segments) == 3 {
		http.Error(res, "Значение имени метрики не передано", http.StatusBadRequest)
		return
	}

	// При попытке передать запрос с некорректным типом метрики возвращать http.StatusBadRequest.
	mType := segments[1]
	if mType != "gauge" && mType != "counter" {
		http.Error(res, "Запрос с некорректным типом метрики", http.StatusBadRequest)
		return
	}

	// При попытке передать запрос с некорректным значением метрики возвращать http.StatusBadRequest.
	if mType == "gauge" {
		mGaugeValue, err := strconv.ParseFloat(segments[3], 64)
		if err != nil {
			http.Error(res, "Значение метрики некорректно", http.StatusBadRequest)
			return
		}
		// TODO storing logic for mGaugeValue

	} else if mType == "counter" {
		mCounterValue, err := strconv.ParseInt(segments[3], 10, 64)
		if err != nil {
			http.Error(res, "Значение метрики некорректно", http.StatusBadRequest)
			return
		}
		// TODO Storing logic for mCounterValue
	}
}
