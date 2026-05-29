package server

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/timlevett/ecowitt-prom/internal/metrics"
	"github.com/timlevett/ecowitt-prom/internal/parser"
)

type Server struct {
	addr        string
	dataPath    string
	metricsPath string
	exporter    *metrics.Exporter
	registry    *prometheus.Registry
	passkey     string
}

type Option func(*Server)

func WithAddr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

func WithDataPath(path string) Option {
	return func(s *Server) {
		s.dataPath = path
	}
}

func WithMetricsPath(path string) Option {
	return func(s *Server) {
		s.metricsPath = path
	}
}

func WithPasskey(passkey string) Option {
	return func(s *Server) {
		s.passkey = passkey
	}
}

func New(opts ...Option) *Server {
	registry := prometheus.NewRegistry()
	exporter := metrics.NewExporter(registry)

	s := &Server{
		addr:        ":8080",
		dataPath:    "/data/report",
		metricsPath: "/metrics",
		exporter:    exporter,
		registry:    registry,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.dataPath, s.handleData)
	mux.Handle(s.metricsPath, promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Printf("ecowitt-prom listening on %s", s.addr)
	log.Printf("  POST %s (data endpoint)", s.dataPath)
	log.Printf("  GET  %s (metrics endpoint)", s.metricsPath)
	log.Printf("  GET  /healthz (health check)")

	srv := &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return srv.ListenAndServe()
}

func maskPasskey(pk string) string {
	if len(pk) <= 4 {
		return "****"
	}
	return pk[:4] + "****"
}

func (s *Server) handleData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := parser.Parse(r)
	if err != nil {
		log.Printf("error parsing request: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if s.passkey != "" && data.Passkey != s.passkey {
		log.Printf("rejecting data from unknown passkey: %s", maskPasskey(data.Passkey))
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	s.exporter.Update(data)

	log.Printf("received data from %s (%s): outdoor=%.1fF humidity=%.0f%% wind=%.1fmph",
		maskPasskey(data.Passkey), data.StationType, data.TempOutdoorF, data.HumidityOutdoor, data.WindSpeedMPH)

	w.WriteHeader(http.StatusNoContent)
}
