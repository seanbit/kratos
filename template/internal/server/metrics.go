package server

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// NewMetricsServer creates a lightweight HTTP server for health probes and Prometheus metrics.
// Used by consumer service (and other non-HTTP services) to satisfy K8s liveness/readiness probes.
func NewMetricsServer() *khttp.Server {
	srv := khttp.NewServer(
		khttp.Address(":9090"),
	)
	srv.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	srv.Handle("/ready", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	srv.Handle("/metrics", promhttp.Handler())
	return srv
}
