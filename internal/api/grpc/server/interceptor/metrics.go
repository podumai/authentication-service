package interceptor

import (
	"authentication_service/internal/observability/metrics"
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		request any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		response, err := handler(ctx, request)

		method := info.FullMethod
		statusCode := status.Code(err).String()
		metrics.GRPCRequestCounter.WithLabelValues(method, statusCode).Inc()
		metrics.GRPCRequestLatency.WithLabelValues(method).Observe(time.Since(start).Seconds())

		return response, err
	}
}
