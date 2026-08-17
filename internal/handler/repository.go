package handler

// StorageGetter defines read-only operations for retrieving metrics.
type StorageGetter interface {
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

// StorageSetter defines write operations for updating metrics.
type StorageSetter interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
}

// Repository combines read and write operations for the server-side metric store.
type Repository interface {
	StorageGetter
	StorageSetter
}