# Decentralized Identity & Authentication System

A comprehensive system that integrates traditional user authentication with blockchain-based Decentralized Identifiers (DIDs) using Go microservices and Ethereum smart contracts.

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Auth Service  │◄──►│  DID Manager    │◄──►│   Blockchain    │
│   (Port 8080)   │    │  (Port 8081)    │    │  (Ethereum)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │      NATS       │    │   Smart        │
│   Database      │    │   Message Queue │    │   Contract     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 Features

### Core Functionality
- **User Registration & Authentication**: Traditional user management with JWT tokens
- **DID Generation**: Cryptographically secure Decentralized Identifiers
- **Blockchain Integration**: Ethereum smart contract for immutable identity proofs
- **Asynchronous Processing**: NATS-based job queue for blockchain operations
- **Verifiable Credentials**: Cryptographic proof of identity ownership

### Security Features
- **Ed25519 Key Pairs**: Modern cryptography for DID authentication
- **Hash-based Identity**: SHA256 hashing of user data for privacy
- **Private Key Protection**: Keys never stored in plain text
- **JWT Token Management**: Secure session handling

### Optional Features
- **Wallet Integration**: MetaMask/Phantom support (extensible)
- **Audit Logging**: Complete trail of DID operations
- **Rate Limiting**: API protection against abuse
- **Health Monitoring**: Service health checks and metrics

## 📁 Project Structure

```
├── services/
│   ├── auth-service/          # Existing authentication service
│   └── did-manager/           # New DID management service
│       ├── cmd/server/        # Main application entry point
│       ├── internal/          # Private application code
│       │   ├── config/        # Configuration management
│       │   ├── domain/        # Business logic interfaces
│       │   ├── handler/       # HTTP request handlers
│       │   ├── repository/    # Data access layer
│       │   └── services/      # Business logic implementation
│       ├── pkg/               # Public packages
│       │   ├── blockchain/    # Ethereum integration
│       │   ├── crypto/        # Cryptographic utilities
│       │   ├── did/           # DID generation and validation
│       │   └── queue/         # NATS message queuing
│       └── scripts/           # Database initialization
├── contracts/                  # Ethereum smart contracts
│   ├── DIDRegistry.sol        # Main DID registry contract
│   ├── package.json           # Node.js dependencies
│   ├── hardhat.config.js      # Hardhat configuration
│   └── scripts/               # Deployment scripts
├── cli/                       # Command-line interface
├── docker-compose.yml         # Local development setup
└── README.md                  # This file
```

## 🛠️ Technology Stack

### Backend Services
- **Go 1.21+**: High-performance microservices
- **Gin**: Fast HTTP web framework
- **PostgreSQL**: Primary data storage
- **NATS**: Message queuing and streaming
- **Zerolog**: Structured logging

### Blockchain
- **Ethereum/Polygon**: Smart contract platform
- **Solidity 0.8.19**: Smart contract language
- **Hardhat**: Development and deployment framework
- **OpenZeppelin**: Secure contract libraries

### Infrastructure
- **Docker**: Containerization
- **Docker Compose**: Local development orchestration
- **Nginx**: Reverse proxy (optional)

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Node.js 18+ (for smart contract deployment)
- PostgreSQL 15+
- NATS Server

### 1. Clone and Setup
```bash
git clone <repository-url>
cd go-blockchain
```

### 2. Start Local Development Environment
```bash
# Start all services
docker-compose up -d

# Check service status
docker-compose ps
```

### 3. Deploy Smart Contract
```bash
cd contracts
npm install
npx hardhat compile
npx hardhat node  # In separate terminal
npx hardhat run scripts/deploy.js --network localhost
```

### 4. Test the System
```bash
# Test DID Manager health
curl http://localhost:8081/api/v1/health

# Run CLI demo
cd cli
go run did-cli.go demo
```

## 📖 API Documentation

### DID Manager Service (Port 8081)

#### Create DID
```http
POST /api/v1/did
Content-Type: application/json

{
  "user_id": "uuid",
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123"
}
```

#### Verify DID
```http
POST /api/v1/did/verify
Content-Type: application/json

{
  "did": "did:example:user:hash:key",
  "user_hash": "user_hash_string"
}
```

#### Get DID Status
```http
GET /api/v1/did/status/{did}
```

#### Health Check
```http
GET /api/v1/health
```

### Auth Service Integration

The existing auth service can be extended to integrate with DID Manager:

```go
// Example: After user registration, create DID
func (s *AuthService) SignUpWithDID(req *UserCreateRequest) (*AuthResponse, error) {
    // 1. Create user in local database
    user, err := s.createUser(req)
    if err != nil {
        return nil, err
    }
    
    // 2. Create DID via DID Manager
    didReq := &DIDCreateRequest{
        UserID:   user.ID,
        Name:     req.Name,
        Email:    req.Email,
        Password: req.Password,
    }
    
    didResp, err := s.didManager.CreateDID(didReq)
    if err != nil {
        // Log error but don't fail user creation
        s.logger.Error("Failed to create DID", "error", err)
    }
    
    // 3. Return auth response with DID info
    return &AuthResponse{
        User:   user,
        DID:    didResp.DID,
        Tokens: s.generateTokens(user),
    }, nil
}
```

## 🔧 Configuration

### Environment Variables

#### DID Manager Service
```bash
# Server
PORT=8081
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=did_manager
DB_PASSWORD=did_manager_password
DB_NAME=did_manager_db
DB_SSLMODE=disable

# Blockchain
ETHEREUM_RPC_URL=http://localhost:8545
ETHEREUM_PRIVATE_KEY=your_private_key
ETHEREUM_CONTRACT_ADDRESS=0x...

# NATS
NATS_URL=nats://localhost:4222
```

#### Smart Contract Deployment
```bash
# .env file in contracts/ directory
PRIVATE_KEY=your_deployer_private_key
TESTNET_RPC_URL=https://rpc-mumbai.maticvigil.com
MAINNET_RPC_URL=https://polygon-rpc.com
POLYGONSCAN_API_KEY=your_api_key
```

## 🧪 Testing

### Unit Tests
```bash
# Test DID Manager
cd services/did-manager
go test ./...

# Test smart contracts
cd contracts
npx hardhat test
```

### Integration Tests
```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test ./tests/integration/...
```

### Load Testing
```bash
# Using k6
k6 run load-tests/did-creation.js
```

## 🔒 Security Considerations

### Private Key Management
- **Never store private keys in databases**
- **Use environment variables or secure key management systems**
- **Implement key rotation policies**
- **Consider hardware security modules (HSM) for production**

### DID Security
- **Validate DID format before processing**
- **Implement rate limiting on DID creation**
- **Use secure random number generation**
- **Implement DID revocation mechanisms**

### API Security
- **JWT token validation**
- **Rate limiting and DDoS protection**
- **Input validation and sanitization**
- **HTTPS enforcement in production**

## 📊 Monitoring & Observability

### Health Checks
- Service health endpoints
- Database connectivity checks
- Blockchain node status
- Queue health monitoring

### Metrics
- DID creation/verification rates
- Blockchain transaction success rates
- Queue processing latency
- Error rates and types

### Logging
- Structured JSON logging
- Request/response correlation
- Error tracking and alerting
- Audit trail for compliance

## 🚀 Deployment

### Production Deployment
```bash
# Build production images
docker build -f services/did-manager/Dockerfile.prod -t did-manager:prod .

# Deploy to Kubernetes
kubectl apply -f deployments/production/

# Deploy smart contract
npx hardhat run scripts/deploy.js --network mainnet
```

### Staging Deployment
```bash
# Deploy to staging environment
./scripts/deploy.sh staging

# Run smoke tests
./scripts/test-staging.sh
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

### Development Guidelines
- Follow Go best practices
- Write comprehensive tests
- Update documentation
- Use conventional commit messages

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

### Common Issues
- **Database connection errors**: Check PostgreSQL service status
- **Blockchain errors**: Verify Ganache is running and accessible
- **NATS connection issues**: Ensure NATS server is started
- **Contract deployment failures**: Check private key and network configuration

### Getting Help
- Check the [Issues](../../issues) page
- Review the [Documentation](docs/)
- Contact the development team

## 🔮 Roadmap

### Phase 1 (Current)
- ✅ Basic DID creation and verification
- ✅ Blockchain integration
- ✅ Asynchronous job processing
- ✅ REST API endpoints

### Phase 2 (Next)
- 🔄 Wallet-based authentication
- 🔄 Verifiable credentials issuance
- 🔄 Advanced DID methods
- 🔄 Cross-chain compatibility

### Phase 3 (Future)
- 📋 Zero-knowledge proofs
- 📋 Decentralized storage integration
- 📋 Multi-signature DID support
- 📋 Mobile SDK development

---

**Built with ❤️ using Go, Ethereum, and modern microservices architecture**

