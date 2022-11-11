package server

import (
	"context"
	"github.com/JECSand/go-grpc-server-boilerplate/services"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
		err := i.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
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
		err := i.authorize(stream.Context(), info.FullMethod)
		if err != nil {
			return err
		}
		return handler(srv, stream)
	}
}

func (i *AuthInterceptor) authorize(ctx context.Context, method string) error {
	roleMap, ok := i.accessibleRoles[method]
	if !ok {
		return nil // unprotected endpoint
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}
	values := md["authorization"]
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}
	accessToken := values[0]
	var authorized bool
	var err error
	switch roleMap[0] {
	case "Root":
		authorized, err = i.tokenService.RootAdminTokenVerifyMiddleWare(accessToken)
	case "Admin":
		authorized, err = i.tokenService.AdminTokenVerifyMiddleWare(accessToken)
	case "Member":
		authorized, err = i.tokenService.MemberTokenVerifyMiddleWare(accessToken)
	}
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}
	if authorized {
		return nil
	}
	return status.Error(codes.PermissionDenied, "no permission to access this RPC")
}
