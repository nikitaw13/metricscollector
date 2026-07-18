package repository

// Интерфейс для взаимодействия с хранилищем
type repository interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}
