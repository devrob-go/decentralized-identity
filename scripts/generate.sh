#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo -e "${YELLOW}Generating code from protobuf definitions...${NC}"

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}✗ protoc is not installed. Please install Protocol Buffers compiler.${NC}"
    exit 1
fi

# Check if Go protobuf plugins are installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo -e "${YELLOW}Installing protoc-gen-go...${NC}"
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo -e "${YELLOW}Installing protoc-gen-go-grpc...${NC}"
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

if ! command -v protoc-gen-grpc-gateway &> /dev/null; then
    echo -e "${YELLOW}Installing protoc-gen-grpc-gateway...${NC}"
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
fi

# Generate Go code for auth service
echo -e "${YELLOW}Generating Go code for auth service...${NC}"
cd "$PROJECT_ROOT/api/auth/v1"
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
       --proto_path=. --proto_path=../common \
       auth.proto health.proto



# Copy generated files to services
echo -e "${YELLOW}Copying generated files to services...${NC}"

# Copy auth service generated files
mkdir -p "$PROJECT_ROOT/services/auth-service/api"
cp "$PROJECT_ROOT/api/auth/v1"/*.go "$PROJECT_ROOT/services/auth-service/api/"



echo -e "${GREEN}🎉 Code generation completed successfully!${NC}"
echo -e "${YELLOW}Generated files are available in:${NC}"
echo -e "  - services/auth-service/api/"

