package main

import (
	"flag"
	"os"

	"github.com/timlevett/ecowitt-prom/internal/server"
)

func main() {
	addr := flag.String("addr", envOrDefault("LISTEN_ADDR", ":8080"), "HTTP listen address")
	dataPath := flag.String("data-path", envOrDefault("DATA_PATH", "/data/report"), "Path for Ecowitt data POST endpoint")
	metricsPath := flag.String("metrics-path", envOrDefault("METRICS_PATH", "/metrics"), "Path for Prometheus metrics endpoint")
	passkey := flag.String("passkey", envOrDefault("STATION_PASSKEY", ""), "Only accept data from this PASSKEY (empty = accept all)")
	flag.Parse()

	srv := server.New(
		server.WithAddr(*addr),
		server.WithDataPath(*dataPath),
		server.WithMetricsPath(*metricsPath),
		server.WithPasskey(*passkey),
	)

	if err := srv.ListenAndServe(); err != nil {
		os.Exit(1)
	}
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
