# Go Auth Service - Microservices Monorepo

A modern, scalable microservices architecture built with Go, featuring authentication and user management services. This project follows Go best practices and domain-driven design principles.

## 🏗️ Architecture Overview

This monorepo contains the main microservice:

- **Auth Service**: Handles user authentication, authorization, and user management


The service supports both gRPC and REST APIs, with shared libraries for common functionality.

## 📁 Project Structure

```
go-chat-ai/
├── api/                    # API definitions (protobuf, OpenAPI)
│   ├── auth/              # Auth service API definitions

│   └── common/            # Shared protobuf definitions
├── pkg/                   # Shared, reusable packages
│   ├── auth/              # JWT, authentication utilities
│   ├── logger/            # Structured logging
│   ├── middleware/        # HTTP/gRPC middleware
│   ├── metrics/           # Prometheus metrics
│   ├── database/          # Database connection and migrations
│   ├── config/            # Configuration management
│   ├── errors/            # Error handling utilities
│   └── utils/             # Common utilities
├── services/              # Microservices
│   ├── auth-service/      # Authentication service

├── deployments/           # Deployment configurations
│   ├── local/            # Local development
│   ├── staging/          # Staging environment
│   └── production/       # Production environment
├── scripts/               # Build and deployment scripts
├── docs/                  # Documentation
└── tools/                 # Development tools
```

## 🚀 Quick Start

### Prerequisites

- Go 1.24 or later
- Docker and Docker Compose
- PostgreSQL
- Protocol Buffers compiler (protoc)

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd go-chat-ai
   ```

2. **Set up environment variables**
   ```bash
   cp services/auth-service/env.example services/auth-service/.env
   
   # Edit .env files with your configuration
   ```

3. **Generate protobuf code**
   ```bash
   ./scripts/generate.sh
   ```

4. **Build all services**
   ```bash
   ./scripts/build.sh
   ```

5. **Start local services**
   ```bash
   ./scripts/deploy.sh -e local
   ```

6. **Run tests**
   ```bash
   ./scripts/test.sh
   ```

## 🛠️ Development

### Code Generation

The project uses Protocol Buffers for API definitions. To regenerate code after changes:

```bash
./scripts/generate.sh
```

### Building

Build individual services:
```bash
cd services/auth-service
go build -o bin/auth-service ./cmd/server


```

Or build all services:
```bash
./scripts/build.sh
```

### Testing

Run tests for all services:
```bash
./scripts/test.sh
```

Run tests for a specific service:
```bash
cd services/auth-service
go test ./...
```

## 🚢 Deployment

### Local Development
```bash
./scripts/deploy.sh -e local
```

### Staging
```bash
./scripts/deploy.sh -e staging -s all
```

### Production
```bash
./scripts/deploy.sh -e production -s all
```

### Deploy Specific Service
```bash
./scripts/deploy.sh -e staging -s auth
./scripts/deploy.sh -e production -s chat
```

## 📚 API Documentation

### Auth Service
- **gRPC**: Port 50051
- **REST**: Port 8080
- **Health Check**: `/health`

### Chat Service
- **gRPC**: Port 50052
- **REST**: Port 8081
- **Health Check**: `/health`

## 🔧 Configuration

### Environment Variables

#### Auth Service
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret for JWT token signing
- `JWT_EXPIRY`: JWT token expiry time
- `PORT`: Service port (default: 8080)
- `GRPC_PORT`: gRPC port (default: 50051)

#### Chat Service
- `DATABASE_URL`: PostgreSQL connection string
- `OPENAI_API_KEY`: OpenAI API key
- `PORT`: Service port (default: 8081)
- `GRPC_PORT`: gRPC port (default: 50052)

## 🧪 Testing

### Unit Tests
```bash
go test ./...
```

### Integration Tests
```bash
go test -tags=integration ./...
```

### End-to-End Tests
```bash
go test -tags=e2e ./...
```

## 📊 Monitoring

The services expose Prometheus metrics at `/metrics` endpoints:

- HTTP request metrics
- gRPC request metrics
- Database connection metrics
- Custom business metrics

## 🔒 Security

- JWT-based authentication
- HTTPS/TLS support
- Input validation and sanitization
- SQL injection prevention
- CORS configuration

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:
- Create an issue in the repository
- Check the [documentation](docs/)
- Review the [architecture guide](docs/architecture/)

## 🔮 Roadmap

- [ ] Add notification service
- [ ] Implement payment service
- [ ] Add API Gateway
- [ ] Enhanced monitoring and alerting
- [ ] Multi-region deployment support
- [ ] Performance testing suite

