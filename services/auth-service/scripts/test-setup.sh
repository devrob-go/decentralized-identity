#!/bin/bash

# Test script to verify the gRPC + REST setup

echo "🚀 Testing Go Starter gRPC + REST Setup"
echo "========================================"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.24+"
    exit 1
fi

echo "✅ Go version: $(go version)"

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "❌ Protocol Buffers compiler (protoc) is not installed"
    echo "   Please install protoc: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

echo "✅ Protocol Buffers compiler: $(protoc --version)"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker"
    exit 1
fi

echo "✅ Docker version: $(docker --version)"

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose"
    exit 1
fi

echo "✅ Docker Compose version: $(docker-compose --version)"

# Install Go tools
echo ""
echo "📦 Installing Go tools..."
cd auth-service
make install-tools

# Generate protobuf code
echo ""
echo "🔧 Generating protobuf code..."
make proto

# Check if generation was successful
if [ ! -f "proto/auth.pb.go" ] || [ ! -f "proto/auth_grpc.pb.go" ] || [ ! -f "proto/auth.pb.gw.go" ]; then
    echo "❌ Protobuf code generation failed"
    exit 1
fi

echo "✅ Protobuf code generated successfully"

# Build the application
echo ""
echo "🔨 Building application..."
make build

if [ ! -f "bin/auth-service" ]; then
    echo "❌ Application build failed"
    exit 1
fi

echo "✅ Application built successfully"

# Test the application
echo ""
echo "🧪 Running tests..."
make test

echo ""
echo "🎉 Setup verification completed successfully!"
echo ""
echo "Next steps:"
echo "1. Copy env.example to .env and configure your environment"
echo "2. Run 'docker-compose up -d' to start services"
echo "3. Run 'make dev' for development with hot reload"
echo ""
echo "Your service will be available at:"
echo "  - gRPC: localhost:8080"
echo "  - REST: localhost:8081"
