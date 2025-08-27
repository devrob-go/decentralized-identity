# Deployment Configuration

This directory contains deployment configurations for all microservices in the Go monorepo.

## Structure

```
deployment/
├── local/                          # Local development configurations
│   ├── docker-compose.yml          # Main local development stack
│   └── docker-compose-chat.yml     # Chat service local development
├── staging/                        # Staging/production configurations
│   ├── Chart.yaml                  # Main Helm chart
│   ├── values.yaml                 # Main values file
│   ├── values-staging.yaml         # Staging-specific values
│   ├── deploy.sh                   # Main deployment script
│   ├── deploy-chat.sh              # Chat service deployment script
│   ├── templates/                  # Main chart templates
│   │   ├── deployment.yaml         # Main service deployment
│   │   ├── service.yaml            # Main service
│   │   ├── ingress.yaml            # Main ingress
│   │   └── ...                     # Other templates
│   └── charts/                     # Subcharts for microservices

└── README.md                       # This file
```

## Services

### Main Service (go-starter-grpc)
- **Ports**: gRPC 8080, REST 8081
- **Chart**: Main Helm chart in `staging/`
- **Deploy**: `./deploy.sh`



## Local Development

### Start all services
```bash
cd deployment/local
docker-compose up
```



## Kubernetes Deployment

### Deploy main service
```bash
cd deployment/staging
./deploy.sh
```



### Deploy service
```bash
cd deployment/staging
./deploy.sh
```

## Configuration

### Environment Variables

- `NAMESPACE`: Kubernetes namespace (default: default)
- `DRY_RUN`: Set to 'true' for dry-run mode

### Values Files
- `values.yaml`: Main service configuration


## Helm Charts

### Main Chart
- **Name**: go-starter-grpc
- **Type**: Application
- **Dependencies**: PostgreSQL, Redis



## Deployment Scripts

### deploy.sh
- Deploys the main service
- Creates namespace if needed
- Waits for deployment completion
- Shows deployment status



## Monitoring

### Prometheus
- Service monitors for both services
- Metrics collection enabled
- Configurable scrape intervals

### Grafana
- Dashboard templates ready
- Service-specific dashboards
- Performance monitoring

## Security

### TLS
- Configurable TLS versions
- Certificate management
- mTLS between services

### Network Policies
- Ingress/egress rules
- Service-to-service communication
- External access control

## Troubleshooting

### Common Issues
1. **Namespace not found**: Scripts create namespaces automatically
2. **Image pull errors**: Check image registry and credentials
3. **Service not accessible**: Verify ingress and service configuration

### Debug Commands
```bash
# Check pod status
kubectl get pods -n <namespace>

# View logs
kubectl logs -n <namespace> -l app.kubernetes.io/name=<service-name>

# Check service status
kubectl get services -n <namespace>

# Test endpoints
kubectl port-forward -n <namespace> <pod-name> <local-port>:<pod-port>
```

## Next Steps

1. **Customize values**: Modify `values.yaml` files for your environment
2. **Add monitoring**: Configure Prometheus and Grafana
3. **Set up CI/CD**: Integrate with your deployment pipeline
4. **Security review**: Review and update security policies
5. **Performance tuning**: Adjust resource limits and scaling policies
