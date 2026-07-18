package repository

// Тип для хранения метрик
type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}
