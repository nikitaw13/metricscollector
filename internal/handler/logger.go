package handler

import (
	"net/http"

	"go.uber.org/zap"
)

// Log is a package-level singleton logger.
// Only Initialize should modify this variable.
// Defaults to a no-op logger that produces no output.
var Log *zap.Logger = zap.NewNop()

// Initialize bootstraps the singleton logger at the given level.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}

type (
	// responseData captures the HTTP status code and response body size.
	responseData struct {
		status int
		size   int
	}

	// loggingResponseWriter wraps http.ResponseWriter to capture status and size.
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write delegates to the wrapped ResponseWriter and accumulates bytes written.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size

	if r.responseData.status == 0 {
		r.responseData.status = http.StatusOK
	}

	return size, err
}

// WriteHeader delegates to the wrapped ResponseWriter and records the status code.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}