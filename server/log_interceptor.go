package server

import (
	"context"
	"github.com/JECSand/go-grpc-server-boilerplate/config"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
)

// LoggerInterceptor struct
type LoggerInterceptor struct {
	log utilities.Logger
	cfg *config.Configuration
}

// NewLoggerInterceptor constructs a LoggerInterceptor
func NewLoggerInterceptor(logger utilities.Logger, cfg *config.Configuration) *LoggerInterceptor {
	return &LoggerInterceptor{log: logger, cfg: cfg}
}

// Logger Interceptor
func (i *LoggerInterceptor) Logger(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	start := time.Now()
	md, _ := metadata.FromIncomingContext(ctx)
	reply, err := handler(ctx, req)
	i.log.Infof("Method: %s, Time: %v, Metadata: %v, Err: %v", info.FullMethod, time.Since(start), md, err)
	return reply, err
}
