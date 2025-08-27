# Deployment Guide

## Overview

This guide covers deploying the Go Chat AI services to different environments: local development, staging, and production. Each environment has its own configuration and deployment strategy.

## Prerequisites

### Required Tools

- **Docker**: [Download from docker.com](https://www.docker.com/products/docker-desktop)
- **Docker Compose**: Usually included with Docker Desktop
- **Kubernetes**: For staging and production deployments
- **Helm**: [Installation guide](https://helm.sh/docs/intro/install/)
- **kubectl**: [Installation guide](https://kubernetes.io/docs/tasks/tools/)

### Environment Variables

Ensure you have the following environment variables configured:

```bash
# Database
DATABASE_URL=postgresql://username:password@host:port/database?sslmode=disable

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# OpenAI (for chat service)
OPENAI_API_KEY=your-openai-api-key

# Service ports
AUTH_SERVICE_PORT=8080
AUTH_SERVICE_GRPC_PORT=50051
CHAT_SERVICE_PORT=8081
CHAT_SERVICE_GRPC_PORT=50052
```

## Local Development Deployment

### Quick Start

```bash
# Deploy all services locally
./scripts/deploy.sh -e local

# Or deploy specific services
./scripts/deploy.sh -e local -s auth
./scripts/deploy.sh -e local -s chat
```

### Manual Deployment

1. **Start the database**:
   ```bash
   cd deployments/local
   docker-compose up -d postgres
   ```

2. **Wait for database to be ready**:
   ```bash
   docker-compose logs postgres
   ```

3. **Run migrations**:
   ```bash
   # Auth service migrations
   cd ../../services/auth-service
   go run cmd/migrate/main.go


   ```

4. **Start services**:
   ```bash
   cd ../../deployments/local
   docker-compose up -d
   ```

5. **Verify deployment**:
   ```bash
   docker-compose ps
   docker-compose logs
   ```

### Local Configuration

The local deployment uses `deployments/local/docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: gochat
      POSTGRES_USER: gochat
      POSTGRES_PASSWORD: gochat
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  auth-service:
    build:
      context: ../../services/auth-service
      dockerfile: Dockerfile.dev
    environment:
      - DATABASE_URL=postgresql://gochat:gochat@postgres:5432/gochat?sslmode=disable
      - JWT_SECRET=dev-secret
    ports:
      - "8080:8080"
      - "50051:50051"
    depends_on:
      - postgres



volumes:
  postgres_data:
```

### Health Checks

Verify services are running:

```bash
# Auth service health
curl http://localhost:8080/health

# gRPC health checks
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
```

## Staging Environment Deployment

### Prerequisites

- Kubernetes cluster access
- Helm installed and configured
- Container registry access
- Environment-specific configuration

### Deployment

1. **Build and push images**:
   ```bash
   # Build images
   ./scripts/build.sh

   # Tag and push to registry
   docker tag go-chat-ai/auth-service:latest your-registry/auth-service:staging

   docker push your-registry/auth-service:staging

   ```

2. **Deploy to staging**:
   ```bash
   ./scripts/deploy.sh -e staging -s all
   ```

3. **Verify deployment**:
   ```bash
   kubectl get pods -n go-chat-ai-staging
   kubectl get services -n go-chat-ai-staging
   kubectl get ingress -n go-chat-ai-staging
   ```

### Staging Configuration

The staging environment uses `deployments/staging/values-staging.yaml`:

```yaml
global:
  environment: staging
  imageRegistry: your-registry
  imageTag: staging

auth-service:
  replicaCount: 2
  resources:
    requests:
      memory: "128Mi"
      cpu: "100m"
    limits:
      memory: "256Mi"
      cpu: "200m"
  ingress:
    enabled: true
    host: auth-staging.yourdomain.com
    tls: true



postgresql:
  enabled: true
  auth:
    database: gochat
    username: gochat
    password: staging-password
  primary:
    persistence:
      size: 10Gi

redis:
  enabled: true
  auth:
    password: staging-redis-password
  master:
    persistence:
      size: 5Gi
```

### Staging Helm Charts

The staging environment uses Helm charts in `deployments/staging/charts/`:

- **auth-service**: Authentication service deployment

- **PostgreSQL**: Database deployment
- **Redis**: Caching layer

## Production Environment Deployment

### Prerequisites

- Production Kubernetes cluster
- Production container registry
- SSL certificates
- Monitoring and logging infrastructure
- Backup and disaster recovery setup

### Deployment

1. **Build production images**:
   ```bash
   # Build with production optimizations
   ./scripts/build.sh

   # Tag and push to production registry
   docker tag go-chat-ai/auth-service:latest your-registry/auth-service:production

   docker push your-registry/auth-service:production

   ```

2. **Deploy to production**:
   ```bash
   ./scripts/deploy.sh -e production -s all
   ```

3. **Verify deployment**:
   ```bash
   kubectl get pods -n go-chat-ai-production
   kubectl get services -n go-chat-ai-production
   kubectl get ingress -n go-chat-ai-production
   ```

### Production Configuration

The production environment uses `deployments/production/values-production.yaml`:

```yaml
global:
  environment: production
  imageRegistry: your-registry
  imageTag: production

auth-service:
  replicaCount: 3
  resources:
    requests:
      memory: "256Mi"
      cpu: "200m"
    limits:
      memory: "512Mi"
      cpu: "500m"
  autoscaling:
    enabled: true
    minReplicas: 3
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70
  ingress:
    enabled: true
    host: auth.yourdomain.com
    tls: true
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    readOnlyRootFilesystem: true



postgresql:
  enabled: true
  auth:
    database: gochat
    username: gochat
    password: production-password
  primary:
    persistence:
      size: 100Gi
    resources:
      requests:
        memory: "1Gi"
        cpu: "500m"
      limits:
        memory: "4Gi"
        cpu: "2000m"
  backup:
    enabled: true
    schedule: "0 2 * * *"
    retention: 30

redis:
  enabled: true
  auth:
    password: production-redis-password
  master:
    persistence:
      size: 20Gi
    resources:
      requests:
        memory: "512Mi"
        cpu: "250m"
      limits:
        memory: "2Gi"
        cpu: "1000m"

monitoring:
  prometheus:
    enabled: true
    retention: 30d
  grafana:
    enabled: true
    adminPassword: production-grafana-password

logging:
  elasticsearch:
    enabled: true
    replicas: 3
  kibana:
    enabled: true
  fluentd:
    enabled: true
```

## Monitoring and Observability

### Metrics

Services expose Prometheus metrics at `/metrics` endpoints:

```bash
# Auth service metrics
curl http://localhost:8080/metrics

# Chat service metrics
curl http://localhost:8081/metrics
```

### Health Checks

Implement health checks for all services:

```go
func (h *HealthHandler) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
    // Check database connection
    if err := h.db.Ping(); err != nil {
        return &healthpb.HealthCheckResponse{
            Status: healthpb.HealthCheckResponse_NOT_SERVING,
        }, nil
    }

    // Check external dependencies
    if err := h.checkOpenAI(); err != nil {
        return &healthpb.HealthCheckResponse{
            Status: healthpb.HealthCheckResponse_NOT_SERVING,
        }, nil
    }

    return &healthpb.HealthCheckResponse{
        Status: healthpb.HealthCheckResponse_SERVING,
    }, nil
}
```

### Logging

Configure structured logging across all services:

```go
logger := log.New(
    log.NewJSONHandler(os.Stdout, &log.HandlerOptions{
        Level: log.LevelInfo,
        AddSource: true,
    }),
)

logger.Info("service started",
    "service", "auth-service",
    "version", "1.0.0",
    "port", config.Port,
)
```

## Security Considerations

### Network Policies

Implement network policies to restrict service communication:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: auth-service-network-policy
  namespace: go-chat-ai-production
spec:
  podSelector:
    matchLabels:
      app: auth-service
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
    - protocol: TCP
      port: 50051
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: postgresql
    ports:
    - protocol: TCP
      port: 5432
```

### RBAC

Configure proper role-based access control:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ServiceAccount
metadata:
  name: auth-service
  namespace: go-chat-ai-production
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: auth-service-role
  namespace: go-chat-ai-production
rules:
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: auth-service-role-binding
  namespace: go-chat-ai-production
subjects:
- kind: ServiceAccount
  name: auth-service
  namespace: go-chat-ai-production
roleRef:
  kind: Role
  name: auth-service-role
  apiGroup: rbac.authorization.k8s.io
```

### Secrets Management

Use Kubernetes secrets for sensitive data:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: auth-service-secrets
  namespace: go-chat-ai-production
type: Opaque
data:
  jwt-secret: <base64-encoded-jwt-secret>
  database-url: <base64-encoded-database-url>

```

## Backup and Recovery

### Database Backups

Configure automated database backups:

```yaml
# PostgreSQL backup configuration
backup:
  enabled: true
  schedule: "0 2 * * *"  # Daily at 2 AM
  retention: 30           # Keep 30 days
  storage:
    type: s3
    bucket: your-backup-bucket
    region: us-west-2
```

### Disaster Recovery

Implement disaster recovery procedures:

1. **Regular backups** to multiple locations
2. **Cross-region replication** for critical data
3. **Automated failover** procedures
4. **Recovery testing** on a regular basis

## Performance Optimization

### Resource Limits

Set appropriate resource limits:

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "200m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### Horizontal Pod Autoscaling

Enable automatic scaling:

```yaml
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
```

### Database Optimization

- Connection pooling
- Proper indexing
- Query optimization
- Read replicas for read-heavy workloads

## Troubleshooting

### Common Issues

#### Service Not Starting

```bash
# Check pod status
kubectl get pods -n go-chat-ai-production

# Check pod logs
kubectl logs -f pod/auth-service-xxx -n go-chat-ai-production

# Check pod events
kubectl describe pod auth-service-xxx -n go-chat-ai-production
```

#### Database Connection Issues

```bash
# Check database pod status
kubectl get pods -n postgresql

# Check database logs
kubectl logs -f postgresql-0 -n postgresql

# Test database connectivity
kubectl exec -it postgresql-0 -n postgresql -- psql -U gochat -d gochat
```

#### Ingress Issues

```bash
# Check ingress status
kubectl get ingress -n go-chat-ai-production

# Check ingress controller logs
kubectl logs -f deployment/ingress-nginx-controller -n ingress-nginx
```

### Debugging Commands

```bash
# Port forward to service
kubectl port-forward svc/auth-service 8080:8080 -n go-chat-ai-production

# Execute command in pod
kubectl exec -it auth-service-xxx -n go-chat-ai-production -- /bin/sh

# Copy files from/to pod
kubectl cp auth-service-xxx:/app/logs ./logs -n go-chat-ai-production
```

## Rollback Procedures

### Rolling Back Deployments

```bash
# Check deployment history
kubectl rollout history deployment/auth-service -n go-chat-ai-production

# Rollback to previous version
kubectl rollout undo deployment/auth-service -n go-chat-ai-production

# Rollback to specific version
kubectl rollout undo deployment/auth-service --to-revision=2 -n go-chat-ai-production
```

### Database Rollback

```bash
# Restore from backup
kubectl exec -it postgresql-0 -n postgresql -- pg_restore -U gochat -d gochat /backups/backup-file.sql
```

## Maintenance

### Regular Tasks

- **Security updates**: Keep base images updated
- **Dependency updates**: Update Go modules regularly
- **Backup verification**: Test backup restoration
- **Performance monitoring**: Review metrics and optimize
- **Log rotation**: Manage log storage and retention

### Scheduled Maintenance

- **Database maintenance**: Regular VACUUM and ANALYZE
- **Certificate renewal**: Monitor SSL certificate expiration
- **Resource cleanup**: Remove unused resources
- **Monitoring review**: Update alerting thresholds
