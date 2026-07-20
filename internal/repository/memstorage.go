package repository

// Тип для хранения метрик
type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

var _ Repository = (*MemStorage)(nil)

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   map[string]float64{},
		counter: map[string]int64{},
	}
}

func (ms *MemStorage) UpdateGauge(name string, value float64) error {
	ms.gauge[name] = value
	return nil
}

func (ms *MemStorage) UpdateCounter(name string, value int64) error {
	ms.counter[name] += value
	return nil
}
