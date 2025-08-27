# Go Chat AI - Staging Deployment

This directory contains the Helm chart configuration for deploying the Go Chat AI platform to a staging environment.

## Architecture Overview

The Go Chat AI platform consists of two main microservices:

- **Auth Service**: Handles user authentication, JWT token management, and user management
- **Chat Service**: Provides AI chat functionality using OpenAI's API

## Chart Structure

```
staging/
├── Chart.yaml                 # Main chart definition
├── values.yaml               # Default values for all environments
├── values-staging.yaml       # Staging-specific overrides
├── Makefile                  # Build and deployment commands
├── deploy.sh                 # Deployment script
├── charts/                   # Subcharts directory
│   ├── auth-service/         # Authentication service subchart
│   │   ├── Chart.yaml       # Subchart definition
│   │   ├── values.yaml      # Subchart default values
│   │   ├── templates/       # Kubernetes manifests
│   │   └── .helmignore      # Files to ignore

└── templates/                # Main chart templates (shared resources)
```

## Prerequisites

- Kubernetes cluster (1.19+)
- Helm 3.0+
- kubectl configured
- cert-manager (for TLS certificates)
- nginx-ingress controller

## Quick Start

### 1. Install the Chart

```bash
# Using the deployment script
./deploy.sh install

# Or using Make
make install

# Or using Helm directly
helm install go-chat-ai . \
  --namespace staging \
  --create-namespace \
  --values values-staging.yaml \
  --wait \
  --timeout 10m
```

### 2. Check Deployment Status

```bash
./deploy.sh status
# or
make status
```

### 3. Access the Services

- **Auth Service**: https://auth-staging.your-domain.com


## Configuration

### Main Values (values.yaml)

Contains default configuration for all environments:
- Service configurations
- Resource limits
- Security settings
- Database configurations

### Staging Values (values-staging.yaml)

Contains staging-specific overrides:
- Reduced resource limits
- Staging domain names
- Environment-specific variables

### Service-Specific Values

Each subchart has its own `values.yaml`:
- `charts/auth-service/values.yaml` - Auth service defaults


## Available Commands

### Deployment Script

```bash
./deploy.sh [COMMAND]

Commands:
  install      - Install the Helm chart
  upgrade      - Upgrade the Helm chart
  uninstall    - Uninstall the Helm chart
  status       - Check deployment status
  logs         - Show application logs
  port-forward - Set up port forwarding
  lint         - Lint the Helm chart
  template     - Template the Helm chart
  test         - Test the Helm chart
  clean        - Clean up temporary files
  deps         - Update Helm dependencies
  help         - Show help message
```

### Make Commands

```bash
make [TARGET]

Targets:
  install      - Install the Helm chart
  upgrade      - Upgrade the Helm chart
  uninstall    - Uninstall the Helm chart
  status       - Check deployment status
  logs         - Show application logs
  port-forward - Set up port forwarding
  lint         - Lint the Helm chart
  template     - Template the Helm chart
  test         - Test the Helm chart
  clean        - Clean up temporary files
  deps         - Update Helm dependencies
```

## Service Configuration

### Auth Service

- **Ports**: gRPC (8080), REST (8081)
- **Health Check**: `/health` endpoint
- **Database**: PostgreSQL for user data
- **Cache**: Redis for session management
- **Security**: JWT-based authentication

### Chat Service

- **Ports**: gRPC (8080), REST (8081)
- **Health Check**: `/health` endpoint
- **AI Integration**: OpenAI API
- **Security**: JWT token validation via auth service

## Database Configuration

### PostgreSQL

- **Database**: `go_chat_ai_staging`
- **Persistence**: 5Gi storage
- **Resources**: 500m CPU, 512Mi memory

### Redis

- **Purpose**: Session caching and rate limiting
- **Persistence**: 2Gi storage
- **Resources**: 250m CPU, 256Mi memory

## Monitoring

- **ServiceMonitor**: Prometheus monitoring enabled
- **Scrape Interval**: 60 seconds
- **Scrape Timeout**: 15 seconds

## Security Features

- **Pod Security**: Non-root containers
- **Network Policies**: Restricted pod communication
- **RBAC**: Service account with minimal permissions
- **TLS**: Automatic certificate management via cert-manager
- **Security Headers**: HSTS and other security headers

## Scaling

Both services support horizontal pod autoscaling:

- **Min Replicas**: 2
- **Max Replicas**: 5
- **CPU Target**: 80%
- **Memory Target**: 80%

## Troubleshooting

### Common Issues

1. **Pod Startup Failures**
   - Check resource limits
   - Verify database connectivity
   - Review startup probe configuration

2. **Ingress Issues**
   - Verify cert-manager is running
   - Check ingress controller status
   - Validate TLS certificate configuration

3. **Database Connection Issues**
   - Verify PostgreSQL pod status
   - Check database credentials
   - Validate network policies

### Debug Commands

```bash
# Check pod logs
./deploy.sh logs

# Port forward for local debugging
./deploy.sh port-forward

# Check resource status
kubectl get all -n staging

# Check events
kubectl get events -n staging --sort-by='.lastTimestamp'
```

## Development

### Adding New Services

1. Create a new subchart in `charts/`
2. Add service configuration to main `values.yaml`
3. Update deployment scripts and Makefile
4. Test with `helm template` and `helm lint`

### Modifying Existing Services

1. Update the service's `values.yaml`
2. Modify templates as needed
3. Test changes with `helm template`
4. Deploy with `./deploy.sh upgrade`

## Contributing

1. Follow the existing chart structure
2. Use consistent naming conventions
3. Test changes with `helm lint` and `helm template`
4. Update documentation for any new features

## License

This project is licensed under the MIT License.
