package repository

// Тип для хранения метрик
type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

// TODO: Кто инициализирует map'ы? Пустая структура — это nil-мапы. Нужен конструктор.
