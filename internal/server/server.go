package server

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// StartServer starts the web server.
func StartServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.Handle("/metrics", promhttp.Handler())

	return http.ListenAndServe(":2114", mux)
}
