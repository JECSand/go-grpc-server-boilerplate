package server

import (
	"context"
	"github.com/JECSand/go-grpc-server-boilerplate/services"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
	"google.golang.org/grpc"
	"log"
)

// AuthInterceptor enforces JWT based authentication on gRPC services
type AuthInterceptor struct {
	log             utilities.Logger
	tokenService    *services.TokenService
	accessibleRoles map[string][]string
}

// NewAuthInterceptor constructs an AuthInterceptor
func NewAuthInterceptor(log utilities.Logger, tokenService *services.TokenService, accessibleRoles map[string][]string) *AuthInterceptor {
	return &AuthInterceptor{log, tokenService, accessibleRoles}
}

// Unary method creates and returns a gRPC unary server interceptor
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Println("--> unary interceptor: ", info.FullMethod)

		// TODO: implement authorization

		return handler(ctx, req)
	}
}

// Stream creates and returns a gRPC stream server interceptor
func (i *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		log.Println("--> stream interceptor: ", info.FullMethod)

		// TODO: implement authorization

		return handler(srv, stream)
	}
}
