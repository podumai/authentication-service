package server

import (
	"authentication_service/internal/api/grpc/auth"
	"authentication_service/internal/api/grpc/server/interceptor"
	"authentication_service/internal/config"
	"authentication_service/internal/database"
	"authentication_service/internal/logger"
	"authentication_service/internal/service"
	"authentication_service/internal/transports/grpc/handler"
	"net"

	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
)

type Opts struct {
	Config       *config.GRPCServer
	Logger       logger.Logger
	UserService  service.UserService
	TokenService service.TokenService
	AuthService  service.AuthService
}

type GRPCServer struct {
	Server   *grpc.Server
	Config   *config.GRPCServer
	Logger   logger.Logger
	Database database.DatabaseService
}

func NewGRPCServer(opts *Opts) *GRPCServer {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.LoggingInterceptor(opts.Logger),
			interceptor.MetricsInterceptor(),
		),
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
			otelgrpc.WithPropagators(otel.GetTextMapPropagator()),
		)),
	)
	auth.RegisterAuthServer(server, handler.NewAuthServer(&handler.AuthServerOpts{
		UserService:  opts.UserService,
		TokenService: opts.TokenService,
		AuthService:  opts.AuthService,
	},
	))
	return &GRPCServer{
		Server: server,
		Config: opts.Config,
		Logger: opts.Logger,
	}
}

func (s *GRPCServer) ServeListener(listener net.Listener) error {
	return s.Server.Serve(listener)
}

func (s *GRPCServer) Serve() error {
	listener, err := net.Listen("tcp", s.Config.URL)
	if err != nil {
		return err
	}
	s.Logger.Info("gRPC server started", logger.Field{Key: "Addr", Value: s.Config.URL})
	return s.ServeListener(listener)
}
