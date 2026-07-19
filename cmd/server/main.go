package main

import (
	"log"
	"net/http"
	"time"

	"github.com/PrometheRus/metricscollector/internal/handler"
)

func main() {
	var h handler.MetricsHandler
	mux := http.NewServeMux()
	mux.Handle("/update/", h)

	srv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
