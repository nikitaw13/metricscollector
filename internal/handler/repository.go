package handler

type StorageGetter interface {
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

type StorageSetter interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
}

// Интерфейс для взаимодействия с хранилищем
type Repository interface {
	StorageGetter
	StorageSetter
}
