# Deployment Guide

## Overview

This guide covers deploying the Decentralized Identity & Authentication System across different environments, from local development to production.

## Prerequisites

- Docker 20.0+ and Docker Compose 2.0+
- Go 1.21+
- Node.js 18+ (for smart contracts)
- PostgreSQL 15+
- Kubernetes 1.25+ (for production)

## Environment Setup

### 1. Local Development

#### Quick Start
```bash
# Clone repository
git clone <repository-url>
cd go-blockchain

# Start all services
cd deployments/local
docker-compose up -d

# Initialize database
make db-init

# Deploy smart contracts
cd ../../contracts
npm install
npx hardhat run scripts/deploy.js --network localhost

# Verify services
curl http://localhost:8080/health
curl http://localhost:8082/api/v1/health
```

#### Environment Variables
```bash
# .env file for local development
# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=starter_db

# DID Manager
DID_MANAGER_URL=http://localhost:8082
ETHEREUM_RPC_URL=http://localhost:8545
ETHEREUM_PRIVATE_KEY=4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f863187
ETHEREUM_CONTRACT_ADDRESS=0xe78A0F7E598Cc8b0Bb87894B0F60dD2a88d6a8Ab

# NATS
NATS_URL=nats://localhost:4222
```

#### Service Status Check
```bash
# Check all services
docker-compose ps

# View logs
docker-compose logs -f auth-service
docker-compose logs -f did-manager

# Test API endpoints
make test-apis
```

### 2. Staging Environment

#### Infrastructure Setup
```bash
# Create staging namespace
kubectl create namespace did-staging

# Apply staging configurations
kubectl apply -f deployments/staging/ -n did-staging

# Deploy database
helm install postgres-staging bitnami/postgresql \
  --namespace did-staging \
  --set auth.postgresPassword=staging_password \
  --set primary.persistence.size=20Gi

# Deploy NATS
helm install nats-staging nats/nats \
  --namespace did-staging \
  --set nats.jetstream.enabled=true
```

#### Configuration
```yaml
# deployments/staging/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: did-system-config
  namespace: did-staging
data:
  POSTGRES_HOST: "postgres-staging"
  POSTGRES_PORT: "5432"
  POSTGRES_DB: "did_staging"
  NATS_URL: "nats://nats-staging:4222"
  ETHEREUM_RPC_URL: "https://rpc-mumbai.maticvigil.com"
  LOG_LEVEL: "info"
  ENVIRONMENT: "staging"
```

#### Secrets Management
```yaml
# deployments/staging/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: did-system-secrets
  namespace: did-staging
type: Opaque
data:
  POSTGRES_PASSWORD: c3RhZ2luZ19wYXNzd29yZA== # base64 encoded
  ETHEREUM_PRIVATE_KEY: NGMwODgzYTY5MTAyOTM3ZA== # base64 encoded
  JWT_SECRET: bXlfc2VjcmV0X2tleQ== # base64 encoded
```

#### Deployment
```bash
# Deploy services
kubectl apply -f deployments/staging/

# Check deployment status
kubectl get pods -n did-staging

# Check service endpoints
kubectl get services -n did-staging

# Deploy smart contracts to testnet
cd contracts
npx hardhat run scripts/deploy.js --network testnet
```

### 3. Production Environment

#### Infrastructure Requirements

**Compute Resources:**
```yaml
# Production resource requirements
Auth Service:
  replicas: 3
  resources:
    requests:
      memory: "512Mi"
      cpu: "500m"
    limits:
      memory: "1Gi"
      cpu: "1000m"

DID Manager:
  replicas: 5
  resources:
    requests:
      memory: "1Gi"
      cpu: "1000m"
    limits:
      memory: "2Gi"
      cpu: "2000m"

Database:
  type: "PostgreSQL 15"
  storage: "100Gi"
  replicas: 3 (HA setup)
  
NATS:
  replicas: 3
  storage: "50Gi"
```

#### Production Security Configuration

```yaml
# deployments/production/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: did-system-network-policy
  namespace: did-production
spec:
  podSelector:
    matchLabels:
      app: did-manager
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: auth-service
    - podSelector:
        matchLabels:
          app: api-gateway
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
  - to:
    - podSelector:
        matchLabels:
          app: nats
```

#### SSL/TLS Configuration
```yaml
# deployments/production/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: did-system-ingress
  namespace: did-production
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - api.yourdomain.com
    - did.yourdomain.com
    secretName: did-system-tls
  rules:
  - host: api.yourdomain.com
    http:
      paths:
      - path: /v1/auth
        pathType: Prefix
        backend:
          service:
            name: auth-service
            port:
              number: 8080
  - host: did.yourdomain.com
    http:
      paths:
      - path: /api/v1
        pathType: Prefix
        backend:
          service:
            name: did-manager
            port:
              number: 8082
```

#### Production Deployment Steps

1. **Infrastructure Setup**
```bash
# Create production namespace
kubectl create namespace did-production

# Set up monitoring
helm install prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace did-production

# Deploy database with HA
helm install postgres-prod bitnami/postgresql-ha \
  --namespace did-production \
  --set postgresql.replicaCount=3 \
  --set postgresql.auth.postgresPassword=$POSTGRES_PASSWORD
```

2. **Secret Management**
```bash
# Create secrets using external secret manager
kubectl create secret generic did-system-secrets \
  --namespace did-production \
  --from-literal=POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
  --from-literal=ETHEREUM_PRIVATE_KEY="$ETHEREUM_PRIVATE_KEY" \
  --from-literal=JWT_SECRET="$JWT_SECRET"
```

3. **Service Deployment**
```bash
# Deploy services
envsubst < deployments/production/auth-service.yaml | kubectl apply -f -
envsubst < deployments/production/did-manager.yaml | kubectl apply -f -

# Deploy ingress
kubectl apply -f deployments/production/ingress.yaml

# Verify deployment
kubectl get pods -n did-production
kubectl get services -n did-production
kubectl get ingress -n did-production
```

4. **Smart Contract Deployment**
```bash
# Deploy to mainnet (use with caution)
cd contracts
npx hardhat run scripts/deploy.js --network mainnet

# Verify on Etherscan
npx hardhat verify --network mainnet $CONTRACT_ADDRESS
```

## Database Migration

### Migration Scripts
```sql
-- migrations/001_initial_schema.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    did VARCHAR(255),
    user_hash VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE dids (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    did VARCHAR(255) UNIQUE NOT NULL,
    user_hash VARCHAR(255) NOT NULL,
    public_key TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    blockchain_tx VARCHAR(255)
);

CREATE INDEX idx_dids_user_id ON dids(user_id);
CREATE INDEX idx_dids_did ON dids(did);
CREATE INDEX idx_dids_status ON dids(status);
```

### Migration Execution
```bash
# Run migrations
make db-migrate

# Rollback if needed
make db-rollback

# Check migration status
make db-status
```

## Smart Contract Deployment

### Contract Configuration
```javascript
// hardhat.config.js
require("@nomicfoundation/hardhat-toolbox");
require("dotenv").config();

module.exports = {
  solidity: {
    version: "0.8.19",
    settings: {
      optimizer: {
        enabled: true,
        runs: 200,
      },
    },
  },
  networks: {
    localhost: {
      url: "http://127.0.0.1:8545",
      chainId: 1337,
    },
    mumbai: {
      url: process.env.MUMBAI_RPC_URL,
      accounts: [process.env.PRIVATE_KEY],
      chainId: 80001,
    },
    polygon: {
      url: process.env.POLYGON_RPC_URL,
      accounts: [process.env.PRIVATE_KEY],
      chainId: 137,
    },
  },
  etherscan: {
    apiKey: {
      polygon: process.env.POLYGONSCAN_API_KEY,
      polygonMumbai: process.env.POLYGONSCAN_API_KEY,
    },
  },
};
```

### Deployment Script
```javascript
// scripts/deploy.js
const { ethers } = require("hardhat");

async function main() {
  console.log("Deploying DID Registry contract...");
  
  const [deployer] = await ethers.getSigners();
  console.log("Deploying with account:", deployer.address);
  
  const balance = await deployer.getBalance();
  console.log("Account balance:", ethers.utils.formatEther(balance));
  
  // Deploy contract
  const DIDRegistry = await ethers.getContractFactory("DIDRegistry");
  const didRegistry = await DIDRegistry.deploy();
  await didRegistry.deployed();
  
  console.log("DID Registry deployed to:", didRegistry.address);
  
  // Add deployer as authorized operator
  await didRegistry.addAuthorizedOperator(deployer.address);
  console.log("Deployer added as authorized operator");
  
  // Verify deployment
  const stats = await didRegistry.getContractStats();
  console.log("Contract stats:", {
    total: stats.total.toString(),
    active: stats.active.toString(),
    revoked: stats.revoked.toString()
  });
  
  // Save deployment info
  const deploymentInfo = {
    contractAddress: didRegistry.address,
    deployer: deployer.address,
    network: network.name,
    blockNumber: didRegistry.deployTransaction.blockNumber,
    transactionHash: didRegistry.deployTransaction.hash,
    timestamp: new Date().toISOString()
  };
  
  const fs = require("fs");
  fs.writeFileSync("deployment-info.json", JSON.stringify(deploymentInfo, null, 2));
  
  console.log("Deployment successful!");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
```

## Monitoring & Alerting

### Prometheus Configuration
```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'auth-service'
    static_configs:
      - targets: ['auth-service:8080']
    metrics_path: '/metrics'
    
  - job_name: 'did-manager'
    static_configs:
      - targets: ['did-manager:8082']
    metrics_path: '/metrics'
    
  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']
```

### Grafana Dashboards
```json
{
  "dashboard": {
    "title": "DID System Overview",
    "panels": [
      {
        "title": "DID Creation Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(did_creation_total[5m])",
            "legendFormat": "{{status}}"
          }
        ]
      },
      {
        "title": "Blockchain Operations",
        "type": "graph", 
        "targets": [
          {
            "expr": "blockchain_ops_duration_seconds",
            "legendFormat": "{{operation}}"
          }
        ]
      }
    ]
  }
}
```

### Alerting Rules
```yaml
# monitoring/alerts.yml
groups:
  - name: did-system
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{code=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          
      - alert: DatabaseDown
        expr: up{job="postgres"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Database is down"
          
      - alert: BlockchainSyncLag
        expr: blockchain_block_lag > 100
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Blockchain sync lagging"
```

## Backup & Recovery

### Database Backup
```bash
#!/bin/bash
# scripts/backup-db.sh

BACKUP_DIR="/backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="$BACKUP_DIR/did_system_backup_$TIMESTAMP.sql"

# Create backup
pg_dump -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB > $BACKUP_FILE

# Compress backup
gzip $BACKUP_FILE

# Upload to cloud storage
aws s3 cp $BACKUP_FILE.gz s3://your-backup-bucket/database/

# Cleanup old backups (keep 7 days)
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete
```

### Disaster Recovery
```bash
#!/bin/bash
# scripts/restore-db.sh

BACKUP_FILE=$1

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

# Download from cloud storage if needed
if [[ $BACKUP_FILE == s3://* ]]; then
    aws s3 cp $BACKUP_FILE /tmp/restore.sql.gz
    BACKUP_FILE="/tmp/restore.sql.gz"
fi

# Restore database
gunzip -c $BACKUP_FILE | psql -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB

echo "Database restored from $BACKUP_FILE"
```

## Scaling Strategies

### Horizontal Pod Autoscaling
```yaml
# deployments/production/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: did-manager-hpa
  namespace: did-production
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: did-manager
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Database Scaling
```bash
# Read replicas for scaling reads
kubectl apply -f - <<EOF
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: postgres-cluster
  namespace: did-production
spec:
  instances: 3
  postgresql:
    parameters:
      max_connections: "500"
      shared_buffers: "256MB"
      effective_cache_size: "1GB"
  storage:
    size: 100Gi
    storageClass: fast-ssd
EOF
```

## Performance Tuning

### Application Optimization
```go
// Connection pooling
db, err := sql.Open("postgres", dsn)
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)

// Caching layer
cache := redis.NewClient(&redis.Options{
    Addr:     "redis:6379",
    PoolSize: 10,
})

// Rate limiting
limiter := rate.NewLimiter(rate.Limit(100), 200) // 100 req/sec, burst 200
```

### Database Optimization
```sql
-- Optimize queries
EXPLAIN ANALYZE SELECT * FROM dids WHERE user_id = $1;

-- Add indexes
CREATE INDEX CONCURRENTLY idx_dids_user_id_status ON dids(user_id, status);

-- Vacuum and analyze
VACUUM ANALYZE dids;
```

## Troubleshooting

### Common Issues

1. **Service Won't Start**
```bash
# Check logs
kubectl logs -f deployment/did-manager -n did-production

# Check configuration
kubectl describe configmap did-system-config -n did-production

# Check secrets
kubectl get secrets -n did-production
```

2. **Database Connection Issues**
```bash
# Test connection
kubectl exec -it postgres-0 -n did-production -- psql -U postgres -d did_db -c "SELECT 1;"

# Check network policies
kubectl get networkpolicies -n did-production
```

3. **Blockchain Connection Issues**
```bash
# Test RPC endpoint
curl -X POST $ETHEREUM_RPC_URL \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# Check smart contract
npx hardhat verify --network mainnet $CONTRACT_ADDRESS
```

### Performance Issues
```bash
# Check resource usage
kubectl top pods -n did-production

# Check database performance
kubectl exec -it postgres-0 -n did-production -- psql -U postgres -c "
SELECT query, calls, total_time, mean_time 
FROM pg_stat_statements 
ORDER BY total_time DESC 
LIMIT 10;"

# Check cache hit rates
redis-cli info stats
```

This deployment guide provides comprehensive instructions for deploying the DID system across all environments with proper security, monitoring, and scaling considerations.
