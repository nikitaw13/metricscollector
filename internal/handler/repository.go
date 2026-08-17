package handler

// StorageGetter defines read-only operations for metric storage used by the handler.
type StorageGetter interface {
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

// StorageSetter defines write operations for metric storage used by the handler.
type StorageSetter interface {
	SetGauge(name string, value float64) error
	AddCounter(name string, value int64) error
}

// Repository defines the interface for interacting with metric storage.
type Repository interface {
	StorageGetter
	StorageSetter
}
