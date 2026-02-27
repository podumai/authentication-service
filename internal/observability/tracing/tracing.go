package tracing

import (
	"authentication_service/internal/config"
	"authentication_service/internal/logger"
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type Opts struct {
	Config *config.Tracing
	Logger logger.Logger
}

type TracerService struct {
	Config *config.Tracing
	Logger logger.Logger
	tp     *sdktrace.TracerProvider
}

func (ts *TracerService) Shutdown(ctx context.Context) error {
	return ts.tp.Shutdown(ctx)
}

func NewTracerService(ctx context.Context, opts *Opts) (*TracerService, error) {
	cfg := opts.Config

	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(cfg.CollectorURL),
			otlptracegrpc.WithInsecure()),
	)
	if err != nil {
		return nil, err
	}

	resource, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(cfg.ServiceName)))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	opts.Logger.Info(
		"Tracing initialized (exported: otlptracehttp)",
		logger.Field{Key: "serviceName", Value: cfg.ServiceName},
	)

	return &TracerService{
		Config: cfg,
		Logger: opts.Logger,
		tp:     tp,
	}, nil
}
