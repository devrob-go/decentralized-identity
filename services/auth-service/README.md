# Auth Service - gRPC Architecture

This is a complete gRPC-based authentication and user management service built with Go, featuring a clean architecture with comprehensive logging, database integration, and JWT token management.

## Architecture Overview

The service has been converted from REST to gRPC architecture while maintaining the existing logging system and code patterns. The architecture follows these principles:

- **Protocol Buffers**: Define service contracts and data structures
- **gRPC Server**: Handles all RPC calls with interceptors for logging
- **Service Layer**: Business logic implementation
- **Storage Layer**: Database operations and migrations
- **Logging**: Structured logging with correlation IDs for request tracing

## Features

- **User Management**: Registration, authentication, and user listing
- **JWT Tokens**: Access and refresh token management with automatic expiration
- **Database Integration**: PostgreSQL with automatic migrations
- **Structured Logging**: JSON and console logging with correlation IDs
- **gRPC Interceptors**: Request/response logging and correlation ID handling
- **Graceful Shutdown**: Proper cleanup and resource management
- **Configuration Management**: Environment-based configuration
- **Docker Support**: Containerized deployment

## Project Structure

```
auth-service/
├── proto/           # Protocol Buffer definitions
├── server/          # gRPC server implementation
├── services/        # Business logic layer
├── storage/         # Database operations
├── models/          # Data models
├── utils/           # Utility functions
├── config/          # Configuration management
├── client/          # Example gRPC client
├── Dockerfile       # Container configuration
├── Makefile         # Build and development tasks
└── README.md        # This file
```

## Prerequisites

- Go 1.24.6 or later
- Protocol Buffers compiler (`protoc`)
- PostgreSQL database
- Docker (optional)

## Installation

### 1. Install Protocol Buffer Tools

```bash
# Install protoc compiler
# On Ubuntu/Debian:
sudo apt install protobuf-compiler

# On macOS:
brew install protobuf

# Install Go protobuf plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 2. Generate Protocol Buffer Code

```bash
cd auth-service
make proto
```

### 3. Install Dependencies

```bash
go mod download
go mod tidy
```

## Configuration

The service uses environment variables for configuration. Create a `.env` file:

```env
# Application
APP_ENV=development
APP_PORT=8080
LOG_LEVEL=debug
LOG_JSON_FORMAT=false

# Database
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=starter_db
POSTGRES_HOST=localhost
POSTGRES_PORT=5432

# JWT
JWT_ACCESS_TOKEN_SECRET=your-access-secret
JWT_REFRESH_TOKEN_SECRET=your-refresh-secret
```

## Running the Service

### Development Mode

```bash
# Generate protobuf code and run
make run

# Or manually:
make proto
go run main.go
```

### Production Mode

```bash
# Build and run
make build
./bin/auth-service
```

### Docker

```bash
# Build and run with Docker
docker build -t auth-service .
docker run -p 8080:8080 auth-service
```

## API Reference

### Authentication Service

The service implements the following gRPC methods:

#### User Management
- `SignUp(UserCreateRequest) → AuthResponse`
- `SignIn(Credentials) → AuthResponse`
- `SignOut(SignOutRequest) → Empty`

#### Token Management
- `RefreshToken(RefreshTokenRequest) → TokenResponse`
- `RevokeToken(RevokeTokenRequest) → Empty`

#### User Operations
- `ListUsers(ListUsersRequest) → ListUsersResponse`

### Protocol Buffer Definitions

All service definitions are in `proto/auth.proto`. The service uses:
- `google.protobuf.Timestamp` for time fields
- Standard gRPC status codes for error handling
- Correlation IDs in metadata for request tracing

## Client Usage

### Go Client Example

```go
package main

import (
    "context"
    "log"
    
    "api/auth/v1/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
)

func main() {
    // Connect to server
    conn, err := grpc.Dial("localhost:8080", 
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()
    
    // Create client
    client := proto.NewAuthServiceClient(conn)
    
    // Add correlation ID
    md := metadata.Pairs("x-correlation-id", "client-123")
    ctx := metadata.NewOutgoingContext(context.Background(), md)
    
    // Make request
    resp, err := client.SignUp(ctx, &proto.UserCreateRequest{
        Name:     "John Doe",
        Email:    "john@example.com",
        Password: "password123",
    })
    if err != nil {
        log.Printf("SignUp failed: %v", err)
        return
    }
    
    log.Printf("User created: %s", resp.User.Name)
}
```

### Testing with grpcurl

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:8080 list

# Call SignUp method
grpcurl -plaintext -d '{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123"
}' localhost:8080 auth.AuthService/SignUp
```

## Logging

The service uses structured logging with the following features:

- **Correlation IDs**: Each request gets a unique correlation ID for tracing
- **Structured Fields**: JSON and console logging with contextual information
- **Request/Response Logging**: Automatic logging of all gRPC calls
- **Error Context**: Detailed error logging with stack traces

### Log Format

```json
{
  "level": "info",
  "time": "2024-01-15T10:30:00Z",
  "correlation_id": "req-123",
  "message": "gRPC request completed",
  "method": "/auth.AuthService/SignUp",
  "duration": "15ms",
  "status_code": "OK"
}
```

## Development

### Available Make Commands

```bash
make proto      # Generate protobuf code
make build      # Build the application
make run        # Run the application
make test       # Run tests
make clean      # Clean build artifacts
make generate   # Generate code and format
```

### Adding New Services

1. **Define Protocol Buffer**: Add new messages and service methods to `proto/auth.proto`
2. **Generate Code**: Run `make proto` to generate Go code
3. **Implement Server**: Add implementation in `server/` directory
4. **Register Service**: Add to `RegisterServices()` function
5. **Add Tests**: Create corresponding test files

### Code Generation Workflow

```bash
# 1. Modify proto files
vim proto/auth.proto

# 2. Generate Go code
make proto

# 3. Build and test
make build
make test

# 4. Run
make run
```

## Database

### Migrations

Database migrations are automatically applied on startup:

```bash

# Manual migration
goose -dir storage/migrations postgres "host=localhost user=postgres dbname=starter_db sslmode=disable" up
```

### Schema

The service uses two main tables:
- `users`: User account information
- `user_tokens`: JWT token storage and management

## Monitoring and Observability

- **Request Tracing**: Correlation IDs for request tracking
- **Performance Metrics**: Request duration logging
- **Error Tracking**: Structured error logging with context
- **Health Checks**: gRPC health check service (can be added)

## Security Considerations

- **JWT Secrets**: Use strong, unique secrets for production
- **TLS**: Enable TLS for production deployments
- **Input Validation**: All inputs are validated at the service layer
- **Token Expiration**: Automatic token expiration and refresh

## Performance

- **Connection Pooling**: Database connection pooling
- **gRPC Streaming**: Support for streaming RPCs (can be extended)
- **Efficient Serialization**: Protocol Buffers for fast serialization
- **Graceful Shutdown**: Proper resource cleanup

## Troubleshooting

### Common Issues

1. **Protobuf Generation Errors**
   ```bash
   # Ensure protoc is installed
   protoc --version
   
   # Reinstall Go plugins
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

2. **Database Connection Issues**
   - Check PostgreSQL is running
   - Verify connection parameters in `.env`
   - Check database exists and migrations are applied

3. **Port Conflicts**
   - Change `APP_PORT` in configuration
   - Check if port is already in use

### Debug Mode

Enable debug logging by setting:
```env
LOG_LEVEL=debug
LOG_JSON_FORMAT=false
```

## Contributing

1. Follow the existing code patterns and logging conventions
2. Add tests for new functionality
3. Update documentation for API changes
4. Use correlation IDs for all logging
5. Follow gRPC best practices

## License

This project is part of the go-starter-grpc template.
