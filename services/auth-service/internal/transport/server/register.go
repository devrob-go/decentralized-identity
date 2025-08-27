package server

import (
	"api/auth/v1/proto"
	"auth-service/config"
	grpchandler "auth-service/internal/handler/grpc"
	"auth-service/internal/repository"
	"auth-service/internal/services"
	zlog "packages/logger"

	"google.golang.org/grpc"
)

// RegisterServices registers all gRPC services and their REST gateways
func RegisterServices(grpcServer *grpc.Server, svc *services.Service, logger *zlog.Logger, db *repository.DB, cfg *config.Config) {
	// Register gRPC services
	proto.RegisterAuthServiceServer(grpcServer, grpchandler.NewAuthHandler(svc, logger))
	proto.RegisterHealthServer(grpcServer, NewHealthServer(db, logger, cfg))

	// Note: REST gateway handlers are registered in the server.go file
	// when creating the REST gateway to avoid conflicts
}
