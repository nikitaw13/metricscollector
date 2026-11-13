package handler

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// handleDatabasePing responds to GET /ping by checking the database connectivity.
func (h *MetricsHandler) handleDatabasePing(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DatabasePing(r.Context()); err != nil {
		h.logger.Error("failed ping database", zap.Error(err))
		http.Error(w, internalErrorMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Pong\n")
}
