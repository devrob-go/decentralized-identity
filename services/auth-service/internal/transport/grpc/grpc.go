package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GRPCTransportConfig holds configuration for gRPC transport
type GRPCTransportConfig struct {
	Port     string
	TLS      bool
	CertFile string
	KeyFile  string
}

// NewGRPCServer creates a new gRPC server with the given configuration
func NewGRPCServer(config *GRPCTransportConfig) (*grpc.Server, error) {
	var opts []grpc.ServerOption

	if config.TLS {
		creds, err := credentials.NewServerTLSFromFile(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(creds))
	}

	return grpc.NewServer(opts...), nil
}
