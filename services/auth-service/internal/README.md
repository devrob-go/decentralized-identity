# Internal Directory Structure

This directory follows a clean architecture pattern with clear separation of concerns:

## Directory Layout

```
internal/
├── config/           # Configuration types and defaults
│   └── transport.go  # Transport configuration
├── domain/           # Domain models and business logic
├── handler/          # HTTP/gRPC handlers
│   ├── http/         # REST handlers
│   │   ├── auth.go   # Authentication HTTP handlers
│   │   ├── users.go  # User management HTTP handlers
│   │   └── rest.go   # REST gateway implementation
│   └── grpc/         # gRPC handlers
│       ├── auth_handler.go  # Authentication gRPC handlers
│       ├── auth_handler_test.go  # Auth handler tests
│       └── users.go  # User management gRPC handlers
├── repository/       # Data access layer
├── services/         # Business logic services
└── transport/        # Transport layer configuration
    ├── http/         # HTTP transport configuration
    │   └── http.go   # HTTP server configuration
    ├── grpc/         # gRPC transport configuration
    │   └── grpc.go   # gRPC server configuration
    ├── server/       # Server implementation and orchestration
    │   ├── server.go      # Main server orchestration
    │   ├── dependencies.go # Dependency injection
    │   ├── register.go    # Service registration
    │   └── health.go      # Health check server
    ├── middleware/   # Middleware components
    │   ├── registry.go    # Middleware registry
    │   ├── middleware.go  # Core middleware implementation
    │   ├── security.go    # Security middleware
    │   └── interceptors.go # gRPC interceptors
    ├── errors/       # Error handling
    ├── lifecycle/    # Server lifecycle management
    └── tls/          # TLS configuration
```

## Handler Structure

### HTTP Handlers (`handler/http/`)
- **auth.go**: Handles authentication endpoints (signup, signin, signout, refresh)
- **users.go**: Handles user management endpoints (CRUD operations)
- **rest.go**: REST gateway implementation for gRPC services

### gRPC Handlers (`handler/grpc/`)
- **auth_handler.go**: Implements the AuthService gRPC interface
- **users.go**: Implements user management gRPC methods

## Transport Configuration

### HTTP Transport (`transport/http/`)
- **http.go**: HTTP server configuration with timeouts and settings

### gRPC Transport (`transport/grpc/`)
- **grpc.go**: gRPC server configuration with TLS support

## Configuration

### Transport Config (`config/`)
- **transport.go**: Centralized transport configuration including TLS, server, gateway, health, and security settings

## Benefits of This Structure

1. **Clear Separation**: HTTP and gRPC handlers are clearly separated
2. **Consistent Naming**: Follows Go conventions and project patterns
3. **Easy Navigation**: Developers can quickly find the right handler type
4. **Scalable**: Easy to add new handlers or transport types
5. **Testable**: Clear structure makes testing straightforward
6. **Maintainable**: Logical grouping reduces cognitive load

## Recent Cleanup and Reorganization

The internal directory has been cleaned up and reorganized to:

- **Consolidated server logic** into `transport/server/` directory
- **Moved middleware components** to `transport/middleware/` directory
- **Eliminated duplicate files** and package conflicts
- **Centralized service registration** in `transport/server/register.go`
- **Moved health server** to `transport/server/health.go`
- **Fixed import paths** after restructuring
- **Removed outdated documentation** and unnecessary files

## Usage

- Add new HTTP handlers in `handler/http/`
- Add new gRPC handlers in `handler/grpc/`
- Configure transport settings in `config/transport.go`
- Implement transport-specific logic in `transport/http/` or `transport/grpc/`
