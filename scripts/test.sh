#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo -e "${YELLOW}Running tests for Go Auth Service...${NC}"

# Test shared packages
echo -e "${YELLOW}Testing shared packages...${NC}"
cd "$PROJECT_ROOT/packages"
if [ -f "go.mod" ]; then
    go test -v ./...
    echo -e "${GREEN}✓ Shared packages tests passed${NC}"
else
    echo -e "${YELLOW}No go.mod found in packages directory, skipping${NC}"
fi

# Test auth service
echo -e "${YELLOW}Testing auth service...${NC}"
cd "$PROJECT_ROOT/services/auth-service"
if [ -f "go.mod" ]; then
    go test -v ./...
    echo -e "${GREEN}✓ Auth service tests passed${NC}"
else
    echo -e "${RED}✗ No go.mod found in auth service${NC}"
    exit 1
fi



# Run integration tests if they exist
if [ -d "$PROJECT_ROOT/tests/integration" ]; then
    echo -e "${YELLOW}Running integration tests...${NC}"
    cd "$PROJECT_ROOT/tests/integration"
    go test -v ./...
    echo -e "${GREEN}✓ Integration tests passed${NC}"
fi

echo -e "${GREEN}🎉 All tests passed successfully!${NC}"
