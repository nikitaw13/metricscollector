package repository

import "errors"

// ErrMetricNotFound is returned when a requested metric does not exist in storage.
var ErrMetricNotFound = errors.New("not found")
