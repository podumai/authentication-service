package interceptor

import (
	"authentication_service/internal/logger"
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
)

func LoggingInterceptor(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		request any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		response, err := handler(ctx, request)
		elapsedTime := fmt.Sprintf("%.2fms", time.Since(start).Seconds()*1000)

		if err == nil {
			log.Info("gRPC request completed",
				logger.Field{Key: "method", Value: info.FullMethod},
				logger.Field{Key: "duration", Value: elapsedTime},
			)
		} else {
			log.Error("gRPC request failed",
				logger.Field{Key: "method", Value: info.FullMethod},
				logger.Field{Key: "duration", Value: elapsedTime},
				logger.Field{Key: "error", Value: err.Error()},
			)
		}

		return response, err
	}
}
