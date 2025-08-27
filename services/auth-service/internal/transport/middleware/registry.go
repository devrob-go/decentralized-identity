package middleware

import (
	"google.golang.org/grpc"
)

// Registry manages gRPC middleware registration
type Registry struct {
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
}

// NewRegistry creates a new middleware registry
func NewRegistry() *Registry {
	return &Registry{
		unaryInterceptors:  make([]grpc.UnaryServerInterceptor, 0),
		streamInterceptors: make([]grpc.StreamServerInterceptor, 0),
	}
}

// AddUnary adds a unary interceptor to the registry
func (r *Registry) AddUnary(interceptor grpc.UnaryServerInterceptor) {
	r.unaryInterceptors = append(r.unaryInterceptors, interceptor)
}

// AddStream adds a stream interceptor to the registry
func (r *Registry) AddStream(interceptor grpc.StreamServerInterceptor) {
	r.streamInterceptors = append(r.streamInterceptors, interceptor)
}

// AddUnaryFirst adds a unary interceptor at the beginning of the chain
func (r *Registry) AddUnaryFirst(interceptor grpc.UnaryServerInterceptor) {
	r.unaryInterceptors = append([]grpc.UnaryServerInterceptor{interceptor}, r.unaryInterceptors...)
}

// AddStreamFirst adds a stream interceptor at the beginning of the chain
func (r *Registry) AddStreamFirst(interceptor grpc.StreamServerInterceptor) {
	r.streamInterceptors = append([]grpc.StreamServerInterceptor{interceptor}, r.streamInterceptors...)
}

// GetUnaryInterceptors returns all registered unary interceptors
func (r *Registry) GetUnaryInterceptors() []grpc.UnaryServerInterceptor {
	return r.unaryInterceptors
}

// GetStreamInterceptors returns all registered stream interceptors
func (r *Registry) GetStreamInterceptors() []grpc.StreamServerInterceptor {
	return r.streamInterceptors
}

// Clear removes all registered middleware
func (r *Registry) Clear() {
	r.unaryInterceptors = make([]grpc.UnaryServerInterceptor, 0)
	r.streamInterceptors = make([]grpc.StreamServerInterceptor, 0)
}

// Count returns the number of registered middleware
func (r *Registry) Count() (unary, stream int) {
	return len(r.unaryInterceptors), len(r.streamInterceptors)
}
