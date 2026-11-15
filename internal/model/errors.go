package model

import "errors"

// Errors describing invalid metric payloads and missing metrics.
var (
	ErrMissingType    = errors.New("metric type is required")
	ErrInvalidType    = errors.New("invalid metric type")
	ErrMissingID      = errors.New("metric ID is required")
	ErrMissingValue   = errors.New("metric value is required")
	ErrInvalidValue   = errors.New("invalid metric value")
	ErrMetricNotFound = errors.New("not found")
)
