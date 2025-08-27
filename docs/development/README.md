# Development Guide

## Getting Started

This guide will help you get started with development on the Go Chat AI project. It covers the development environment setup, coding standards, and common development tasks.

## Prerequisites

### Required Software

- **Go 1.24+**: [Download from golang.org](https://golang.org/dl/)
- **Docker**: [Download from docker.com](https://www.docker.com/products/docker-desktop)
- **Docker Compose**: Usually included with Docker Desktop
- **Protocol Buffers Compiler**: [Installation guide](https://grpc.io/docs/protoc-installation/)
- **PostgreSQL 15+**: [Download from postgresql.org](https://www.postgresql.org/download/)

### Go Tools

Install required Go tools:

```bash
# Install protobuf plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install github.com/cosmtrek/air@latest
```

## Development Environment Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd go-chat-ai
```

### 2. Install Dependencies

```bash
# Install dependencies for all modules
go work sync
```

### 3. Environment Configuration

```bash
# Copy environment files
cp services/auth-service/env.example services/auth-service/.env


# Edit .env files with your configuration
```

### 4. Database Setup

```bash
# Start PostgreSQL with Docker
cd deployments/local
docker-compose up -d postgres

# Wait for database to be ready, then run migrations
cd ../../services/auth-service
go run cmd/migrate/main.go


```

### 5. Generate Code

```bash
# Generate protobuf code
./scripts/generate.sh
```

## Project Structure

### Understanding the Layout

```
go-chat-ai/
├── api/                    # API definitions
│   ├── auth/              # Auth service protobuf files

│   └── common/            # Shared protobuf definitions
├── pkg/                   # Shared packages
│   ├── auth/              # Authentication utilities
│   ├── logger/            # Logging package
│   ├── middleware/        # HTTP/gRPC middleware
│   ├── metrics/           # Prometheus metrics
│   ├── database/          # Database utilities
│   ├── config/            # Configuration management
│   ├── errors/            # Error handling
│   └── utils/             # Common utilities
├── services/              # Microservices
│   ├── auth-service/      # Authentication service

└── deployments/           # Deployment configurations
```

### Service Structure

Each service follows this structure:

```
service-name/
├── cmd/                   # Application entry points
│   ├── server/           # Main server binary
│   └── migrate/          # Database migration tool
├── internal/              # Private service code
│   ├── domain/           # Domain models and business logic
│   ├── repository/        # Data access layer
│   ├── service/           # Business logic services
│   ├── handler/           # HTTP/gRPC handlers
│   └── transport/         # Transport layer configuration
├── configs/               # Service-specific configuration
├── Dockerfile             # Service containerization
└── go.mod                 # Go module definition
```

## Development Workflow

### 1. Making Changes

#### Adding New Features

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the coding standards below

3. **Run tests**:
   ```bash
   ./scripts/test.sh
   ```

4. **Build the project**:
   ```bash
   ./scripts/build.sh
   ```

5. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

#### Modifying APIs

1. **Update protobuf definitions** in the `api/` directory
2. **Regenerate code**:
   ```bash
   ./scripts/generate.sh
   ```
3. **Update handlers** to implement new API methods
4. **Add tests** for new functionality
5. **Update documentation**

### 2. Testing

#### Running Tests

```bash
# Run all tests
./scripts/test.sh

# Run tests for a specific service
cd services/auth-service
go test ./...

# Run tests with coverage
go test -coverpkg=./... -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Test Structure

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test service layer interactions
- **End-to-End Tests**: Test complete user workflows

#### Writing Tests

```go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    service := NewUserService(mockRepo)
    
    // Act
    user, err := service.CreateUser(context.Background(), "test@example.com", "password")
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "test@example.com", user.Email)
}
```

### 3. Code Quality

#### Linting

```bash
# Run linter
golangci-lint run

# Run linter for specific service
cd services/auth-service
golangci-lint run
```

#### Code Formatting

```bash
# Format all Go code
go fmt ./...

# Or use goimports for better import organization
go install golang.org/x/tools/cmd/goimports@latest
goimports -w .
```

## Coding Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for code formatting
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Naming Conventions

- **Packages**: Use lowercase, single-word names
- **Functions**: Use camelCase, descriptive names
- **Variables**: Use camelCase, descriptive names
- **Constants**: Use UPPER_SNAKE_CASE
- **Interfaces**: Use descriptive names ending in -er

### Error Handling

```go
// Good: Check errors immediately
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// Good: Use custom error types
var ErrUserNotFound = errors.New("user not found")

// Good: Wrap errors with context
return fmt.Errorf("failed to process request: %w", err)
```

### Logging

```go
// Use structured logging
logger.Info("user created successfully",
    "user_id", user.ID,
    "email", user.Email,
    "created_at", user.CreatedAt,
)

// Use appropriate log levels
logger.Debug("processing user request", "user_id", userID)
logger.Info("user authenticated", "user_id", userID)
logger.Warn("rate limit exceeded", "user_id", userID)
logger.Error("failed to process request", "error", err)
```

### Testing

- Write tests for all new functionality
- Aim for >80% test coverage
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Test both success and error cases

## API Development

### Adding New Endpoints

1. **Define in protobuf**:
   ```protobuf
   service UserService {
       rpc GetUser(GetUserRequest) returns (GetUserResponse) {
           option (google.api.http) = {
               get: "/v1/users/{id}"
           };
       }
   }
   ```

2. **Implement handler**:
   ```go
   func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
       // Implementation
   }
   ```

3. **Add to service**:
   ```go
   func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
       // Business logic
   }
   ```

4. **Add tests** for all layers

### Validation

- Validate all input at the handler level
- Use business rules in the service layer
- Return appropriate HTTP status codes
- Provide clear error messages

## Database Development

### Adding New Tables

1. **Create migration file**:
   ```sql
   -- 0003_new_table.sql
   CREATE TABLE new_table (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       name VARCHAR(255) NOT NULL,
       created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
       updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
   );
   ```

2. **Add to repository**:
   ```go
   type NewTableRepository interface {
       Create(ctx context.Context, table *domain.NewTable) error
       GetByID(ctx context.Context, id string) (*domain.NewTable, error)
       // ... other methods
   }
   ```

3. **Update domain models**:
   ```go
   type NewTable struct {
       ID        string    `json:"id"`
       Name      string    `json:"name"`
       CreatedAt time.Time `json:"created_at"`
       UpdatedAt time.Time `json:"updated_at"`
   }
   ```

### Database Best Practices

- Use transactions for multi-table operations
- Implement proper indexing
- Use prepared statements
- Handle connection pooling
- Implement retry logic for transient failures

## Debugging

### Local Development

```bash
# Start services with hot reload
cd services/auth-service
air


```

### Debugging with Delve

```bash
# Debug a service
cd services/auth-service
dlv debug cmd/server/main.go

# Debug tests
dlv test ./...
```

### Logging

- Use structured logging with correlation IDs
- Log at appropriate levels
- Include relevant context in log messages
- Use consistent log format across services

## Performance Considerations

### Database Optimization

- Use connection pooling
- Implement proper indexing
- Use pagination for large result sets
- Consider caching for frequently accessed data

### API Performance

- Implement rate limiting
- Use streaming for large responses
- Implement proper error handling
- Monitor response times

### Memory Management

- Avoid memory leaks in goroutines
- Use object pools for frequently allocated objects
- Profile memory usage regularly

## Security Best Practices

### Input Validation

- Validate all user input
- Use parameterized queries
- Implement proper authentication
- Use HTTPS in production

### Authentication

- Implement proper JWT handling
- Use secure token storage
- Implement token refresh
- Add rate limiting

## Troubleshooting

### Common Issues

#### Build Failures

```bash
# Clean and rebuild
go clean -cache
go mod tidy
./scripts/build.sh
```

#### Test Failures

```bash
# Run tests with verbose output
go test -v ./...

# Check for race conditions
go test -race ./...
```

#### Database Connection Issues

```bash
# Check database status
docker-compose ps postgres

# Check logs
docker-compose logs postgres
```

### Getting Help

- Check the [architecture documentation](../architecture/README.md)
- Review existing code examples
- Create an issue in the repository
- Ask questions in team discussions

## Deployment

### Local Deployment

```bash
# Deploy all services locally
./scripts/deploy.sh -e local

# Deploy specific service
./scripts/deploy.sh -e local -s auth
```

### Staging Deployment

```bash
# Deploy to staging
./scripts/deploy.sh -e staging -s all
```

### Production Deployment

```bash
# Deploy to production
./scripts/deploy.sh -e production -s all
```

## Continuous Integration

### Pre-commit Hooks

- Run tests
- Check code formatting
- Run linter
- Check for security vulnerabilities

### CI Pipeline

- Automated testing
- Code quality checks
- Security scanning
- Build verification

## Contributing

### Pull Request Process

1. Create a feature branch
2. Make your changes
3. Add tests
4. Update documentation
5. Submit pull request
6. Address review comments
7. Merge after approval

### Code Review Guidelines

- Review for functionality
- Check code quality
- Verify test coverage
- Ensure security best practices
- Validate documentation updates
