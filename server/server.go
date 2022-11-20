package server

import (
	"context"
	"crypto/tls"
	"github.com/JECSand/go-grpc-server-boilerplate/config"
	authsService "github.com/JECSand/go-grpc-server-boilerplate/protos/auth"
	groupsService "github.com/JECSand/go-grpc-server-boilerplate/protos/group"
	tasksService "github.com/JECSand/go-grpc-server-boilerplate/protos/task"
	usersService "github.com/JECSand/go-grpc-server-boilerplate/protos/user"
	"github.com/JECSand/go-grpc-server-boilerplate/services"
	"github.com/JECSand/go-grpc-server-boilerplate/utilities"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func accessibleRoles() map[string][]string {
	const authServicePath = "/authService.AuthService/"
	const userServicePath = "/usersService.UserService/"
	const groupServicePath = "/groupsService.GroupService/"
	const taskServicePath = "/tasksService.TaskService/"
	return map[string][]string{
		authServicePath + "Logout":         {"Member"},
		authServicePath + "Refresh":        {"Member"},
		authServicePath + "GenerateKey":    {"Member"},
		authServicePath + "UpdatePassword": {"Member"},
		userServicePath + "Create":         {"Admin"},
		userServicePath + "Update":         {"Admin"},
		userServicePath + "Get":            {"Member"},
		userServicePath + "GetGroupUsers":  {"Member"},
		userServicePath + "Find":           {"Member"},
		userServicePath + "Delete":         {"Admin"},
		groupServicePath + "Create":        {"Root"},
		groupServicePath + "Update":        {"Admin"},
		groupServicePath + "Get":           {"Member"},
		groupServicePath + "Find":          {"Member"},
		groupServicePath + "Delete":        {"Root"},
		taskServicePath + "Create":         {"Member"},
		taskServicePath + "Update":         {"Member"},
		taskServicePath + "Get":            {"Member"},
		taskServicePath + "GetGroupTasks":  {"Member"},
		taskServicePath + "GetUserTasks":   {"Member"},
		taskServicePath + "Find":           {"Member"},
		taskServicePath + "Delete":         {"Member"},
	}
}

// Server is a struct that stores the API Apps high level attributes such as the router, config, and services
type Server struct {
	log              utilities.Logger
	cfg              *config.Configuration
	TokenService     *services.TokenService
	UserDataService  services.UserDataService
	GroupDataService services.GroupDataService
	TaskDataService  services.TaskDataService
	FileDataService  services.FileDataService
}

// NewServer is a function used to initialize a new Server struct
func NewServer(log utilities.Logger, cfg *config.Configuration, u services.UserDataService, g services.GroupDataService,
	t services.TaskDataService, f services.FileDataService, ts *services.TokenService) *Server {
	return &Server{
		log:              log,
		cfg:              cfg,
		TokenService:     ts,
		UserDataService:  u,
		GroupDataService: g,
		TaskDataService:  t,
		FileDataService:  f,
	}
}

// Start starts the initialized Server
func (s *Server) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	li := NewLoggerInterceptor(s.log, s.cfg)
	ai := NewAuthInterceptor(s.log, s.TokenService, accessibleRoles())
	l, err := net.Listen("tcp", os.Getenv("PORT"))
	if err != nil {
		return errors.Wrap(err, "net.Listen")
	}
	defer l.Close()
	cert, err := tls.LoadX509KeyPair(s.cfg.Cert, s.cfg.Key)
	if err != nil {
		s.log.Fatalf("failed to load key pair: %s", err)
	}
	grpcServer := grpc.NewServer(
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: s.cfg.Server.MaxConnectionIdle * time.Minute,
			Timeout:           s.cfg.Server.Timeout * time.Second,
			MaxConnectionAge:  s.cfg.Server.MaxConnectionAge * time.Minute,
			Time:              s.cfg.Server.Timeout * time.Minute,
		}),
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpcrecovery.UnaryServerInterceptor(),
			ai.Unary(),
			li.Logger,
		),
		grpc.StreamInterceptor(ai.Stream()),
	)
	userService := services.NewUserService(s.log, s.TokenService, s.UserDataService, s.GroupDataService, s.TaskDataService, s.FileDataService)
	usersService.RegisterUserServiceServer(grpcServer, userService)
	groupService := services.NewGroupService(s.log, s.TokenService, s.UserDataService, s.GroupDataService, s.TaskDataService, s.FileDataService)
	groupsService.RegisterGroupServiceServer(grpcServer, groupService)
	taskService := services.NewTaskService(s.log, s.TokenService, s.UserDataService, s.GroupDataService, s.TaskDataService, s.FileDataService)
	tasksService.RegisterTaskServiceServer(grpcServer, taskService)
	authService := services.NewAuthService(s.log, s.TokenService, s.UserDataService, s.GroupDataService)
	authsService.RegisterAuthServiceServer(grpcServer, authService)
	go func() {
		s.log.Infof("GRPC Server is listening on port: %s", s.cfg.Server.Port)
		s.log.Fatal(grpcServer.Serve(l))
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	select {
	case v := <-quit:
		s.log.Errorf("signal.Notify: %v", v)
	case done := <-ctx.Done():
		s.log.Errorf("ctx.Done: %v", done)
	}
	grpcServer.GracefulStop()
	s.log.Info("Server Exited Properly")
	return nil
}

// StartTest starts the initialized Server in a test state
func (s *Server) StartTest(ctx context.Context) (*grpc.ClientConn, func()) {
	buffer := 101024 * 1024
	l := bufconn.Listen(buffer)
	li := NewLoggerInterceptor(s.log, s.cfg)
	ai := NewAuthInterceptor(s.log, s.TokenService, accessibleRoles())
	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: s.cfg.Server.MaxConnectionIdle * time.Minute,
			Timeout:           s.cfg.Server.Timeout * time.Second,
			MaxConnectionAge:  s.cfg.Server.MaxConnectionAge * time.Minute,
			Time:              s.cfg.Server.Timeout * time.Minute,
		}),
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpcrecovery.UnaryServerInterceptor(),
			ai.Unary(),
			li.Logger,
		),
		grpc.StreamInterceptor(ai.Stream()),
	)
	userService := services.NewUserService(s.log, s.TokenService, s.UserDataService, s.GroupDataService, s.TaskDataService, s.FileDataService)
	usersService.RegisterUserServiceServer(grpcServer, userService)
	groupService := services.NewGroupService(s.log, s.TokenService, s.UserDataService, s.GroupDataService, s.TaskDataService, s.FileDataService)
	groupsService.RegisterGroupServiceServer(grpcServer, groupService)
	taskService := services.NewTaskService(s.log, s.TokenService, s.UserDataService, s.GroupDataService, s.TaskDataService, s.FileDataService)
	tasksService.RegisterTaskServiceServer(grpcServer, taskService)
	authService := services.NewAuthService(s.log, s.TokenService, s.UserDataService, s.GroupDataService)
	authsService.RegisterAuthServiceServer(grpcServer, authService)
	go func() {
		s.log.Infof("GRPC Test Server is starting...")
		if err := grpcServer.Serve(l); err != nil {
			s.log.Fatal("error serving GRPC Test Server: %v", err)
		}
	}()
	conn, err := grpc.DialContext(ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return l.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("error connecting to server: %v", err)
	}
	closer := func() {
		err = l.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}
		grpcServer.Stop()
	}
	return conn, closer
}
