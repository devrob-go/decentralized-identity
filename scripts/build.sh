#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo -e "${YELLOW}Building Go Auth Service...${NC}"

# Build shared packages first
echo -e "${YELLOW}Building shared packages...${NC}"
cd "$PROJECT_ROOT/packages"
if [ -f "go.mod" ]; then
    go mod tidy
    go build ./...
    echo -e "${GREEN}✓ Shared packages built successfully${NC}"
else
    echo -e "${YELLOW}No go.mod found in packages directory, skipping${NC}"
fi

# Build auth service
echo -e "${YELLOW}Building auth service...${NC}"
cd "$PROJECT_ROOT/services/auth-service"
if [ -f "go.mod" ]; then
    go mod tidy
    go build -o bin/auth-service ./cmd/server
    echo -e "${GREEN}✓ Auth service built successfully${NC}"
else
    echo -e "${RED}✗ No go.mod found in auth service${NC}"
    exit 1
fi





echo -e "${GREEN}🎉 All services built successfully!${NC}"
echo -e "${YELLOW}Binaries are available in:${NC}"
echo -e "  - services/auth-service/bin/"

