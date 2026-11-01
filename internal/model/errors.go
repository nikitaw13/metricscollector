package model

import "errors"

// Errors describing invalid metric payloads and missing metrics.
var (
	ErrMissingType    = errors.New("Metric type is required")
	ErrInvalidType    = errors.New("Invalid metric type")
	ErrMissingID      = errors.New("Metric ID is required")
	ErrMissingValue   = errors.New("Metric value is required")
	ErrInvalidValue   = errors.New("Invalid metric value")
	ErrMetricNotFound = errors.New("not found")
)
