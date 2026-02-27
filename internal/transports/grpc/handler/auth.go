package handler

import (
	"authentication_service/internal/api/grpc/proto/auth"
	"authentication_service/internal/service"
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "google.golang.org/protobuf/types/known/emptypb"
)

type AuthServer struct {
	auth.AuthServer
	userService  service.UserService
	tokenService service.TokenService
	authService  service.AuthService
}

type AuthServerOpts struct {
	UserService  service.UserService
	TokenService service.TokenService
	AuthService  service.AuthService
}

func NewAuthServer(opts *AuthServerOpts) *AuthServer {
	return &AuthServer{
		userService:  opts.UserService,
		tokenService: opts.TokenService,
		authService:  opts.AuthService,
	}
}

func (as *AuthServer) Login(ctx context.Context, request *auth.UserCredentials) (*auth.TokenPair, error) {
	tr := otel.Tracer("auth_service.Auth")
	ctx, span := tr.Start(ctx, "authservice.Login")
	defer span.End()

	user, err := as.userService.FindByEmail(ctx, request.GetEmail())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err = as.authService.ValidatePassword(request.GetPassword(), user.PasswordHash); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	tokens, err := as.authService.CreateTokenPair(user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err = as.tokenService.Insert(ctx, fmt.Sprint(user.ID), user.Email); err != nil {
		return nil, err
	}

	return &auth.TokenPair{
		AccessToken:  &tokens.AccessToken,
		RefreshToken: &tokens.RefreshToken,
	}, nil
}

func (as *AuthServer) Register(ctx context.Context, request *auth.UserCredentials) (*pb.Empty, error) {
	tr := otel.Tracer("auth_service.Auth")
	ctx, span := tr.Start(ctx, "authservice.Register")
	defer span.End()

	hashedPassword, err := as.authService.CreatePasswordHash(request.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err = as.userService.Create(ctx, request.GetEmail(), string(hashedPassword)); err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	return &pb.Empty{}, nil
}

func (as *AuthServer) Refresh(ctx context.Context, request *auth.TokenPair) (*auth.TokenPair, error) {
	tr := otel.Tracer("auth_service.Auth")
	ctx, span := tr.Start(ctx, "authservice.Refresh")
	defer span.End()

	claims, err := as.authService.ParseToken(request.GetAccessToken())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	tokenPair, err := as.authService.CreateTokenPairWithClaims(claims)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err = as.tokenService.Remove(ctx, claims.ID, request.GetRefreshToken()); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &auth.TokenPair{
		AccessToken:  &tokenPair.AccessToken,
		RefreshToken: &tokenPair.RefreshToken,
	}, nil
}

func (as *AuthServer) ChangePassword(ctx context.Context, request *auth.UserCredentials) (*pb.Empty, error) {
	tr := otel.Tracer("auth_service.Auth")
	ctx, span := tr.Start(ctx, "authservice.ChangePassword")
	defer span.End()

	hashedPassword, err := as.authService.CreatePasswordHash(request.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	user, err := as.userService.FindByEmail(ctx, request.GetEmail())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	err = as.userService.UpdatePassword(ctx, user.ID, hashedPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Empty{}, nil
}
