package main

import (
	"authentication_service/internal/cache"
	"authentication_service/internal/config"
	"authentication_service/internal/database"
	"authentication_service/internal/observability/metrics"
	"authentication_service/internal/observability/tracing"
	"authentication_service/internal/service"
	"authentication_service/internal/transports/grpc/server"
	httpserver "authentication_service/internal/transports/http/server"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"

	slog "authentication_service/internal/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logger := slog.NewLogger("debug", os.Stderr)

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal(err.Error())
	}

	tracerService, err := tracing.NewTracerService(ctx, &tracing.Opts{
		Config: cfg.Tracing,
		Logger: logger,
	})
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer tracerService.Shutdown(ctx)

	db, err := database.NewDatabase(&database.Opts{
		Config: cfg.Database,
		Logger: logger,
	})
	if err != nil {
		logger.Fatal(err.Error())
	}

	cacheService, err := cache.NewRedisCache(ctx, &cache.Opts{
		Config: cfg.Redis,
		Logger: logger,
	})
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer cacheService.Close()

	userService := service.NewUserService(&service.UserServiceOpts{
		Database: db,
		Cache:    cacheService,
	})
	if err != nil {
		logger.Fatal("user-service", slog.Field{Key: "error", Value: err})
	}
	defer userService.Shutdown()

	tokenService := service.NewTokenService(&service.TokenOpts{
		Database: db,
	})

	healthService := service.NewHealthService(&service.HealthServiceOpts{
		Database: db,
		Cache:    cacheService,
	})

	metricsServer := metrics.NewMetricsService(cfg.Metrics, metrics.GRPCMetrics{})

	httpServer := httpserver.NewHTTPServer(&httpserver.Opts{
		Config:   cfg.HTTPServer,
		Logger:   logger,
		Database: db,
		Cache:    cacheService,
		Metrics:  metricsServer,
		Health:   healthService,
	})
	go func() {
		if err := httpServer.Serve(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			stop()
		}
	}()
	defer httpServer.Server.Shutdown(ctx)

	grpcServer := server.NewGRPCServer(&server.Opts{
		Config:       cfg.GRPCServer,
		UserService:  userService,
		TokenService: tokenService,
		AuthService: service.AuthService{Config: &config.AuthService{
			SecretKey: cfg.AuthService.SecretKey,
			TokenTTL:  cfg.AuthService.TokenTTL,
		}},
		Logger: logger,
	})
	defer grpcServer.Server.GracefulStop()
	go func() {
		if err := grpcServer.Serve(); err != nil {
			logger.Fatal(err.Error())
		}
	}()
	<-ctx.Done()
	healthService.SetReady(false)
	logger.Info("srv", slog.Field{Key: "message", Value: "server stopped"})
}
