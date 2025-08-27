# Quick Start Guide

## 🚀 Get Started in 5 Minutes

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (optional)
- Node.js 18+ (for smart contracts)

### 1. Clone & Start
```bash
git clone <repository-url>
cd go-blockchain/deployments/local
docker-compose up -d
```

### 2. Initialize Database
```bash
make db-init
```

### 3. Deploy Smart Contract
```bash
cd ../../contracts
npm install
npx hardhat run scripts/deploy.js --network localhost
```

### 4. Test the System
```bash
# Register user (creates DID automatically)
curl -X POST http://localhost:8080/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","password":"Test123!"}'

# Create DID directly
curl -X POST http://localhost:8082/api/v1/did \
  -H "Content-Type: application/json" \
  -d '{"user_id":"550e8400-e29b-41d4-a716-446655440000","name":"Bob","email":"bob@example.com","password":"password"}'

# Process blockchain queue
curl -X POST http://localhost:8082/api/v1/queue/process
```

## 🔗 Quick Links

| Service | URL | Description |
|---------|-----|-------------|
| Auth Service | http://localhost:8080 | User authentication |
| DID Manager | http://localhost:8082 | DID operations |
| PostgreSQL | localhost:5432 | Database |
| Ganache | http://localhost:8545 | Local blockchain |

## 📚 Documentation

- **[API Documentation](docs/API.md)** - Complete API reference
- **[Use Cases](docs/USE_CASES.md)** - Real-world examples
- **[Architecture](docs/ARCHITECTURE.md)** - System design
- **[Deployment](docs/DEPLOYMENT.md)** - Production setup

## 🛠️ Commands

```bash
# Development
make dev-start        # Start development environment
make db-init          # Initialize database
make test            # Run tests
make build           # Build services

# Testing
cd cli
go run did-cli.go demo  # Run CLI demo

# Monitoring
docker-compose logs -f did-manager  # View logs
curl http://localhost:8082/api/v1/health  # Health check
```

## 🎯 What You Get

✅ **Complete DID System** - User registration with blockchain identity  
✅ **Smart Contracts** - Immutable identity registry  
✅ **REST APIs** - Full CRUD operations for DIDs  
✅ **Async Processing** - Non-blocking blockchain operations  
✅ **Security** - Ed25519 cryptography + JWT tokens  
✅ **Production Ready** - Docker, monitoring, documentation  

🚀 **Ready to build the future of digital identity!**
