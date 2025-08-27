# API Documentation

## Overview

The Decentralized Identity & Authentication System provides two main REST APIs:

1. **Auth Service API** - Traditional user authentication with DID integration
2. **DID Manager API** - Decentralized identity management and blockchain integration

## Base URLs

| Service | Development | Production |
|---------|-------------|------------|
| Auth Service | `http://localhost:8080` | `https://auth.yourdomain.com` |
| DID Manager | `http://localhost:8082` | `https://did.yourdomain.com` |

## Authentication

### Auth Service

Uses JWT tokens for authentication:

```bash
Authorization: Bearer <jwt_token>
```

### DID Manager

Most endpoints are public, but some require API keys or signed requests for production use.

---

## Auth Service API

### User Registration

Register a new user and automatically create a DID.

**Endpoint:** `POST /v1/auth/signup`

**Request Body:**
```json
{
  "name": "Alice Smith",
  "email": "alice@example.com",
  "password": "SecurePassword123!"
}
```

**Response:**
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Alice Smith",
    "email": "alice@example.com",
    "did": "did:example:user:hash:signature",
    "user_hash": "sha256_hash_of_identity_data",
    "created_at": "2025-08-27T10:00:00Z",
    "updated_at": "2025-08-27T10:00:00Z"
  },
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "access_expires_at": "2025-08-27T10:15:00Z",
    "refresh_expires_at": "2025-09-03T10:00:00Z"
  }
}
```

**Status Codes:**
- `201` - User created successfully
- `400` - Invalid request data
- `409` - User already exists
- `500` - Internal server error

---

### User Authentication

Authenticate a user and return JWT tokens.

**Endpoint:** `POST /v1/auth/signin`

**Request Body:**
```json
{
  "email": "alice@example.com",
  "password": "SecurePassword123!"
}
```

**Response:**
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Alice Smith",
    "email": "alice@example.com",
    "did": "did:example:user:hash:signature",
    "created_at": "2025-08-27T10:00:00Z"
  },
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Status Codes:**
- `200` - Authentication successful
- `400` - Invalid request data
- `401` - Invalid credentials
- `500` - Internal server error

---

### Token Refresh

Refresh an expired access token.

**Endpoint:** `POST /v1/auth/refresh`

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-08-27T10:15:00Z"
}
```

---

### Sign Out

Invalidate user tokens.

**Endpoint:** `POST /v1/auth/signout`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**
```json
{
  "message": "Successfully signed out"
}
```

---

## DID Manager API

### Create DID

Generate a new Decentralized Identifier for a user.

**Endpoint:** `POST /api/v1/did`

**Request Body:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Alice Smith",
  "email": "alice@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "did": {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "did": "did:example:user:63b748edafe8657c:7f2d7e7d5108b2ac5bc0cab7f69af2f8",
      "user_hash": "63b748edafe8657c96910ffa2487e3e06690a942805b6ea080df31a95e8ba346",
      "public_key": "f992f05c952715bd896ec103a6ebffd0f8e78ff32063d71964a854c680a7e3c1...",
      "status": "pending",
      "created_at": "2025-08-27T10:00:00Z",
      "updated_at": "2025-08-27T10:00:00Z",
      "blockchain_tx": ""
    },
    "user_hash": "63b748edafe8657c96910ffa2487e3e06690a942805b6ea080df31a95e8ba346",
    "status": "pending",
    "message": "DID created successfully and queued for blockchain registration"
  }
}
```

**Status Codes:**
- `201` - DID created successfully
- `400` - Invalid request data
- `409` - DID already exists
- `500` - Internal server error

---

### Verify DID

Verify a DID against user identity data.

**Endpoint:** `POST /api/v1/did/verify`

**Request Body:**
```json
{
  "did": "did:example:user:63b748edafe8657c:7f2d7e7d5108b2ac5bc0cab7f69af2f8",
  "user_hash": "63b748edafe8657c96910ffa2487e3e06690a942805b6ea080df31a95e8ba346"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "is_valid": true,
    "did": "did:example:user:63b748edafe8657c:7f2d7e7d5108b2ac5bc0cab7f69af2f8",
    "user_hash": "63b748edafe8657c96910ffa2487e3e06690a942805b6ea080df31a95e8ba346",
    "status": "registered",
    "message": "DID is valid and registered on blockchain",
    "blockchain_tx": "0x1234567890abcdef..."
  }
}
```

**Status Codes:**
- `200` - Verification completed
- `400` - Invalid request data
- `404` - DID not found
- `500` - Internal server error

---

### Get DID Status

Check the status of a DID without full verification.

**Endpoint:** `GET /api/v1/did/status/{did}`

**Path Parameters:**
- `did` - The DID to check (URL encoded)

**Example:**
```bash
GET /api/v1/did/status/did:example:user:63b748edafe8657c:7f2d7e7d5108b2ac5bc0cab7f69af2f8
```

**Response:**
```json
{
  "success": true,
  "data": {
    "did": "did:example:user:63b748edafe8657c:7f2d7e7d5108b2ac5bc0cab7f69af2f8",
    "status": "registered",
    "is_valid": true,
    "message": "DID is registered on blockchain",
    "blockchain_tx": "0x1234567890abcdef..."
  }
}
```

**Status Values:**
- `pending` - DID created but not yet on blockchain
- `registered` - DID successfully registered on blockchain
- `failed` - Blockchain registration failed
- `revoked` - DID has been revoked

---

### Get DID by User ID

Retrieve DID information for a specific user.

**Endpoint:** `GET /api/v1/did/user/{userID}`

**Path Parameters:**
- `userID` - The user UUID

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "did": "did:example:user:63b748edafe8657c:7f2d7e7d5108b2ac5bc0cab7f69af2f8",
    "user_hash": "63b748edafe8657c96910ffa2487e3e06690a942805b6ea080df31a95e8ba346",
    "status": "registered",
    "created_at": "2025-08-27T10:00:00Z",
    "blockchain_tx": "0x1234567890abcdef..."
  }
}
```

---

### Process Blockchain Queue

Manually trigger processing of pending blockchain operations.

**Endpoint:** `POST /api/v1/queue/process`

**Request Body:** None

**Response:**
```json
{
  "success": true,
  "message": "Queue processing completed",
  "processed_jobs": 5,
  "successful": 4,
  "failed": 1
}
```

---

### Health Check

Check service health and dependencies.

**Endpoint:** `GET /api/v1/health`

**Response:**
```json
{
  "status": "healthy",
  "service": "did-manager",
  "version": "1.0.0",
  "timestamp": "2025-08-27T10:00:00Z",
  "dependencies": {
    "database": "healthy",
    "blockchain": "healthy",
    "queue": "healthy"
  }
}
```

---

## Error Responses

All APIs use consistent error response format:

```json
{
  "success": false,
  "error": "Error type",
  "details": "Detailed error message",
  "timestamp": "2025-08-27T10:00:00Z",
  "path": "/api/v1/did/verify"
}
```

### Common Error Codes

| HTTP Status | Error Type | Description |
|-------------|------------|-------------|
| 400 | `invalid_request` | Malformed request data |
| 401 | `unauthorized` | Missing or invalid authentication |
| 403 | `forbidden` | Insufficient permissions |
| 404 | `not_found` | Resource not found |
| 409 | `conflict` | Resource already exists |
| 422 | `validation_error` | Input validation failed |
| 429 | `rate_limit_exceeded` | Too many requests |
| 500 | `internal_error` | Server error |
| 502 | `bad_gateway` | Upstream service error |
| 503 | `service_unavailable` | Service temporarily unavailable |

---

## Rate Limiting

### Default Limits

| Endpoint | Rate Limit | Window |
|----------|------------|--------|
| POST /v1/auth/signup | 5 requests | 1 minute |
| POST /v1/auth/signin | 10 requests | 1 minute |
| POST /api/v1/did | 10 requests | 1 minute |
| GET endpoints | 100 requests | 1 minute |

### Rate Limit Headers

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1635724800
```

---

## SDK Examples

### JavaScript/TypeScript

```typescript
import axios from 'axios';

class DIDClient {
  private baseURL: string;
  
  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }
  
  async createDID(userData: {
    user_id: string;
    name: string;
    email: string;
    password: string;
  }) {
    const response = await axios.post(`${this.baseURL}/api/v1/did`, userData);
    return response.data;
  }
  
  async verifyDID(did: string, userHash: string) {
    const response = await axios.post(`${this.baseURL}/api/v1/did/verify`, {
      did,
      user_hash: userHash
    });
    return response.data;
  }
  
  async getDIDStatus(did: string) {
    const response = await axios.get(`${this.baseURL}/api/v1/did/status/${encodeURIComponent(did)}`);
    return response.data;
  }
}

// Usage
const client = new DIDClient('http://localhost:8082');
const result = await client.createDID({
  user_id: '550e8400-e29b-41d4-a716-446655440000',
  name: 'Alice Smith',
  email: 'alice@example.com',
  password: 'password123'
});
```

### Python

```python
import requests
from typing import Dict, Any

class DIDClient:
    def __init__(self, base_url: str):
        self.base_url = base_url
        
    def create_did(self, user_data: Dict[str, str]) -> Dict[str, Any]:
        response = requests.post(f"{self.base_url}/api/v1/did", json=user_data)
        response.raise_for_status()
        return response.json()
        
    def verify_did(self, did: str, user_hash: str) -> Dict[str, Any]:
        data = {"did": did, "user_hash": user_hash}
        response = requests.post(f"{self.base_url}/api/v1/did/verify", json=data)
        response.raise_for_status()
        return response.json()
        
    def get_did_status(self, did: str) -> Dict[str, Any]:
        response = requests.get(f"{self.base_url}/api/v1/did/status/{did}")
        response.raise_for_status()
        return response.json()

# Usage
client = DIDClient('http://localhost:8082')
result = client.create_did({
    'user_id': '550e8400-e29b-41d4-a716-446655440000',
    'name': 'Alice Smith',
    'email': 'alice@example.com',
    'password': 'password123'
})
```

---

## WebSocket Support (Future)

Real-time updates for DID status changes:

```javascript
const ws = new WebSocket('ws://localhost:8082/api/v1/ws');

ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  if (update.type === 'did_status_update') {
    console.log('DID status changed:', update.data);
  }
};

// Subscribe to DID updates
ws.send(JSON.stringify({
  type: 'subscribe',
  did: 'did:example:user:...'
}));
```

---

## Testing

### Using curl

```bash
# Register user
curl -X POST http://localhost:8080/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","password":"Test123!"}'

# Create DID
curl -X POST http://localhost:8082/api/v1/did \
  -H "Content-Type: application/json" \
  -d '{"user_id":"550e8400-e29b-41d4-a716-446655440000","name":"Test","email":"test@example.com","password":"password"}'

# Verify DID
curl -X POST http://localhost:8082/api/v1/did/verify \
  -H "Content-Type: application/json" \
  -d '{"did":"did:example:...","user_hash":"hash..."}'
```

### Using the CLI

```bash
cd cli

# Create DID
go run did-cli.go create "Alice Smith" "alice@example.com"

# Verify DID
go run did-cli.go verify "did:example:..." "user_hash"

# Check status
go run did-cli.go status "did:example:..."

# Demo workflow
go run did-cli.go demo
```
