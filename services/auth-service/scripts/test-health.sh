#!/bin/bash

# Test script for health endpoint
echo "Testing health endpoint..."

# Test REST health endpoint
echo "Testing REST health endpoint at http://localhost:8080/v1/health"
curl -v -w "\nResponse time: %{time_total}s\nHTTP Status: %{http_code}\n" \
     -H "Content-Type: application/json" \
     http://localhost:8080/v1/health

echo -e "\n\nTesting gRPC health endpoint (if grpcurl is available)"
if command -v grpcurl &> /dev/null; then
    grpcurl -plaintext localhost:8081 grpc.health.v1.Health/Check
else
    echo "grpcurl not available, skipping gRPC test"
fi

echo -e "\nHealth check test completed."
