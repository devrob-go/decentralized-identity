# Architecture Documentation

## Overview

Go Chat AI is built using a microservices architecture with domain-driven design principles. The system consists of two main services that communicate via gRPC and expose REST APIs through gRPC-Gateway.

## Architecture Principles

### 1. Domain-Driven Design (DDD)
- **Bounded Contexts**: Each service has its own bounded context
- **Domain Models**: Business logic is encapsulated in domain models
- **Ubiquitous Language**: Consistent terminology across the system

### 2. Clean Architecture
- **Dependency Inversion**: High-level modules don't depend on low-level modules
- **Separation of Concerns**: Clear boundaries between layers
- **Testability**: Easy to test each layer independently

### 3. Microservices Best Practices
- **Single Responsibility**: Each service has one clear purpose
- **Independent Deployment**: Services can be deployed independently
- **Technology Diversity**: Each service can use different technologies if needed

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Client Applications                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   REST      │  │   gRPC      │  │   Web       │  │   Mobile/CLI        │ │
│  │   Client    │  │   Client    │  │   Client    │  │   Client            │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                              API Layer                                      │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  Auth Service (Ports: gRPC 50051, REST 8080)                          │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │ │
│  │  │   gRPC      │  │   REST      │  │   Auth      │  │   Health    │    │ │
│  │  │   Server    │  │   Gateway   │  │   Server    │  │   Server    │    │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  Chat Service (Ports: gRPC 50052, REST 8081)                          │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │ │
│  │  │   gRPC      │  │   REST      │  │   Chat      │  │   OpenAI    │    │ │
│  │  │   Server    │  │   Gateway   │  │   Server    │  │   Client    │    │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                              Service Layer                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Auth      │  │   User      │  │   Chat      │  │   OpenAI    │         │
│  │   Service   │  │   Service   │  │   Service   │  │   Service   │         │ │
│  │ • SignUp    │  │ • GetUsers  │  │ • Chat      │  │ • GPT-3.5   │         │ │
│  │ • SignIn    │  │ • Pagination│  │ • History   │  │ • GPT-4     │         │ │
│  │ • SignOut   │  │ • Validation│  │ • Stream    │  │ • Streaming │         │ │
│  │ • Refresh   │  └─────────────┘  └─────────────┘  └─────────────┘         │
│  │ • Revoke    │                                                            │
│  └─────────────┘                                                            │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                              Domain Layer                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │    User     │  │   Auth      │  │   Chat      │  │ Validation  │         │
│  │ • Models    │  │ • Tokens    │  │ • Messages  │  │ • Rules     │         │
│  │ • Credentials│  │ • JWT      │  │ • Conversations│ • Errors    │         │
│  │ • Validation│  │ • Security  │  │ • AI Models │  │ • Messages  │         │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                          Infrastructure Layer                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ PostgreSQL  │  │   Redis     │  │   Logger    │  │   OpenAI    │         │
│  │ • Migrations│  │ • Caching   │  │ • Structured│  │ • API       │         │
│  │ • Connection│  │ • Sessions  │  │ • Correlation│  │ • Rate Limit│        │
│  │ • Pool      │  │ • Rate Limit│  │ • Levels    │  │ • Models    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Service Architecture

### Auth Service

The Auth Service is responsible for user authentication and authorization. It follows a layered architecture:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Transport Layer                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   gRPC      │  │   REST      │  │   Health    │  │   Metrics   │         │
│  │   Server    │  │   Gateway   │  │   Check     │  │   Endpoint  │         │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                              Handler Layer                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Auth      │  │   User      │  │   Health    │  │   Middleware│         │
│  │   Handler   │  │   Handler   │  │   Handler   │  │   (Auth,    │         │
│  │ • SignUp    │  │ • GetUsers  │  │ • Liveness  │  │    Logging, │         │
│  │ • SignIn    │  │ • Update    │  │ • Readiness │  │    Metrics) │         │
│  │ • SignOut   │  │ • Delete    │  │ • Startup   │  │             │         │
│  │ • Refresh   │  └─────────────┘  └─────────────┘  └─────────────┘         │
│  │ • Revoke    │                                                            │
│  └─────────────┘                                                            │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                              Service Layer                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Auth      │  │   User      │  │   Token     │  │   Validation│         │
│  │   Service   │  │   Service   │  │   Service   │  │   Service   │         │
│  │ • Business  │  │ • Business  │  │ • JWT       │  │ • Input     │         │
│  │   Logic     │  │   Logic     │  │   Management│  │   Validation│         │
│  │ • Rules     │  │ • Rules     │  │ • Security  │  │ • Business  │         │
│  │ • Validation│  │ • Validation│  │ • Expiry    │  │   Rules     │         │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                              Domain Layer                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │    User     │  │   Auth      │  │   Token     │  │   Error     │         │
│  │ • Entity    │  │ • Entity    │  │ • Entity    │  │ • Types     │         │
│  │ • Value     │  │ • Value     │  │ • Value     │  │ • Messages  │         │
│  │   Objects  │  │   Objects   │  │   Objects   │  │ • Codes     │         │
│  │ • Business  │  │ • Business  │  │ • Business  │  │ • Handling  │         │
│  │   Rules     │  │   Rules     │  │   Rules     │  │             │         │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                          Repository Layer                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   User      │  │   Token     │  │   Database  │  │   Migration │         │
│  │ Repository  │  │ Repository  │  │   Connection│  │   Manager   │         │
│  │ • CRUD      │  │ • CRUD      │  │ • Pool      │  │ • Schema    │         │
│  │ • Queries   │  │ • Queries   │  │ • Health    │  │   Updates   │         │
│  │ • Pagination│  │ • Cleanup   │  │ • Metrics   │  │ • Rollbacks │         │
│  │ • Search    │  └─────────────┘  └─────────────┘  └─────────────┘         │
│  └─────────────┘                                                            │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Chat Service

The Chat Service handles AI-powered chat functionality:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Transport Layer                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   gRPC      │  │   REST      │  │   Health    │  │   Metrics   │         │
│  │   Server    │  │   Gateway   │  │   Check     │  │   Endpoint  │         │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                              Handler Layer                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Chat      │  │   Health    │  │   Middleware│  │   Interceptor│        │
│  │   Handler   │  │   Handler   │  │   (Auth,    │  │   (Auth     │         │
│  │ • Send      │  │ • Liveness  │  │    Logging, │  │    Token    │         │
│  │   Message   │  │ • Readiness │  │    Metrics) │  │    Validation│        │
│  │ • Get       │  │ • Startup   │  │             │  │    Rate     │         │
│  │   History   │  └─────────────┘  └─────────────┘  │    Limiting │         │
│  │ • Stream    │                                     └─────────────┘         │
│  │   Chat      │                                                            │
│  └─────────────┘                                                            │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                              Service Layer                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Chat      │  │   OpenAI    │  │   Conversation│  │   Message   │         │
│  │   Service   │  │   Service   │  │   Service   │  │   Service   │         │
│  │ • Business  │  │ • API       │  │ • Business  │  │ • Business  │         │
│  │   Logic     │  │   Client    │  │   Logic     │  │   Logic     │         │
│  │ • Rules     │  │ • Rate      │  │ • Rules     │  │ • Rules     │         │
│  │ • Validation│  │   Limiting  │  │ • Validation│  │ • Validation│         │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                              Domain Layer                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Chat      │  │   Message   │  │   Conversation│  │   AI       │         │
│  │ • Entity    │  │ • Entity    │  │ • Entity    │  │ • Models    │         │
│  │ • Value     │  │ • Value     │  │ • Value     │  │ • Responses │         │
│  │   Objects  │  │   Objects   │  │   Objects   │  │ • Tokens    │         │
│  │ • Business  │  │ • Business  │  │ • Business  │  │ • Context   │         │
│  │   Rules     │  │   Rules     │  │   Rules     │  │ • Limits    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                          Repository Layer                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Chat      │  │   Message   │  │   Database  │  │   Migration │         │
│  │ Repository  │  │ Repository  │  │   Connection│  │   Manager   │         │
│  │ • CRUD      │  │ • CRUD      │  │ • Pool      │  │ • Schema    │         │
│  │ • Queries   │  │ • Queries   │  │ • Health    │  │   Updates   │         │
│  │ • Pagination│  │ • Pagination│  │ • Metrics   │  │ • Rollbacks │         │
│  │ • Search    │  │ • Search    │  └─────────────┘  └─────────────┘         │
│  └─────────────┘  └─────────────┘                                            │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Data Flow

### Authentication Flow

1. **Client Request**: Client sends credentials to auth service
2. **Validation**: Service validates input and business rules
3. **Authentication**: Service verifies credentials against database
4. **Token Generation**: JWT tokens are generated and stored
5. **Response**: Tokens are returned to client

### Chat Flow

1. **Client Request**: Authenticated client sends message
2. **Auth Validation**: Service validates JWT token
3. **Message Processing**: Message is stored and processed
4. **AI Integration**: OpenAI API is called for response
5. **Response**: AI response is stored and returned

## Security Architecture

### Authentication & Authorization

- **JWT Tokens**: Secure, stateless authentication
- **Token Refresh**: Automatic token renewal
- **Role-Based Access**: User permissions and roles
- **Token Revocation**: Secure token invalidation

### Data Protection

- **Input Validation**: Comprehensive input sanitization
- **SQL Injection Prevention**: Parameterized queries
- **HTTPS/TLS**: Encrypted communication
- **Rate Limiting**: API abuse prevention

### Network Security

- **Service Isolation**: Network policies between services
- **CORS Configuration**: Cross-origin request handling
- **Health Checks**: Service availability monitoring

## Scalability Considerations

### Horizontal Scaling

- **Stateless Services**: Services can be scaled horizontally
- **Load Balancing**: Traffic distribution across instances
- **Database Sharding**: Data partitioning for large datasets
- **Caching**: Redis for session and data caching

### Performance Optimization

- **Connection Pooling**: Database connection management
- **Async Processing**: Non-blocking operations
- **Streaming**: Real-time chat capabilities
- **Metrics**: Performance monitoring and alerting

## Monitoring & Observability

### Metrics

- **Application Metrics**: Request rates, response times, error rates
- **Infrastructure Metrics**: CPU, memory, disk usage
- **Business Metrics**: User activity, chat volume

### Logging

- **Structured Logging**: JSON-formatted logs
- **Correlation IDs**: Request tracing across services
- **Log Levels**: Configurable logging verbosity

### Tracing

- **Distributed Tracing**: Request flow across services
- **Performance Analysis**: Bottleneck identification
- **Error Tracking**: Error propagation analysis

## Deployment Architecture

### Local Development

- **Docker Compose**: Local service orchestration
- **Hot Reload**: Development server with auto-restart
- **Local Database**: PostgreSQL with sample data

### Staging Environment

- **Kubernetes**: Container orchestration
- **Helm Charts**: Deployment templating
- **Monitoring**: Prometheus and Grafana

### Production Environment

- **High Availability**: Multi-zone deployment
- **Auto-scaling**: Horizontal pod autoscaling
- **Backup & Recovery**: Automated backup strategies
- **Disaster Recovery**: Multi-region failover

## Technology Stack

### Backend

- **Language**: Go 1.24+
- **Framework**: Standard library + gRPC
- **Database**: PostgreSQL 15+
- **Cache**: Redis (planned)

### API

- **Protocol**: gRPC + REST (gRPC-Gateway)
- **Serialization**: Protocol Buffers
- **Authentication**: JWT

### Infrastructure

- **Containerization**: Docker
- **Orchestration**: Kubernetes
- **Package Manager**: Helm
- **Monitoring**: Prometheus, Grafana

### Development Tools

- **Code Generation**: protoc, mockgen
- **Testing**: Go testing framework
- **Linting**: golangci-lint
- **CI/CD**: GitHub Actions

## Future Enhancements

### Planned Features

- **API Gateway**: Centralized routing and authentication
- **Service Mesh**: Advanced service-to-service communication
- **Event Streaming**: Asynchronous event processing
- **Multi-tenancy**: Support for multiple organizations

### Scalability Improvements

- **Micro-frontends**: Frontend service decomposition
- **GraphQL**: Flexible data querying
- **Real-time Updates**: WebSocket support
- **Mobile Apps**: Native mobile applications
