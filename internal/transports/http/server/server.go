package server

import (
	"authentication_service/internal/cache"
	"authentication_service/internal/config"
	"authentication_service/internal/database"
	"authentication_service/internal/logger"
	"authentication_service/internal/observability/metrics"
	"authentication_service/internal/service"
	"authentication_service/internal/transports/http/handler"
	"errors"
	"net"
	"net/http"
)

type Opts struct {
	Config   *config.HTTPServer
	Logger   logger.Logger
	Database database.DatabaseService
	Cache    cache.CacheService
	Metrics  metrics.MetricsService
	Health   service.HealthService
}

type HTTPServer struct {
	Config *config.HTTPServer
	Server *http.Server
	Logger logger.Logger
}

func NewHTTPServer(opts *Opts) *HTTPServer {
	mux := http.NewServeMux()
	healthHandler := handler.NewHealthHandler(&handler.Opts{
		HealthService: opts.Health,
		Logger:        opts.Logger,
	})

	mux.HandleFunc("GET /live", healthHandler.Live)
	mux.HandleFunc("GET /ready", healthHandler.Ready)
	mux.Handle("GET /metrics", opts.Metrics.Handler())

	return &HTTPServer{
		Config: opts.Config,
		Server: &http.Server{
			Addr:    opts.Config.URL,
			Handler: mux,
		},
		Logger: opts.Logger,
	}
}

func (s *HTTPServer) ServeListener(listener net.Listener) error {
	s.Logger.Info("HTTP server started", logger.Field{Key: "address", Value: listener.Addr().String()})
	if err := s.Server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.Logger.Error("HTTP server failed", logger.Field{Key: "error", Value: err.Error()})
		return err
	}
	return nil
}

func (s *HTTPServer) Serve() error {
	listener, err := net.Listen("tcp", s.Config.URL)
	if err != nil {
		s.Logger.Error("Failed to create HTTP listener",
			logger.Field{Key: "address", Value: s.Config.URL},
			logger.Field{Key: "error", Value: err.Error()},
		)
		return err
	}
	return s.ServeListener(listener)
}
