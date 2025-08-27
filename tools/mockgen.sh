#!/bin/bash

# Mock generation script for Go interfaces
# This script generates mocks for all interfaces used in testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo -e "${YELLOW}Generating mocks for Go interfaces...${NC}"

# Check if mockgen is installed
if ! command -v mockgen &> /dev/null; then
    echo -e "${YELLOW}Installing mockgen...${NC}"
    go install github.com/golang/mock/mockgen@latest
fi

echo -e "${GREEN}✓ mockgen is installed${NC}"

# Function to generate mocks for a service
generate_service_mocks() {
    local service_name=$1
    local service_dir="$PROJECT_ROOT/services/${service_name}-service"
    
    if [ ! -d "$service_dir" ]; then
        echo -e "${YELLOW}Service directory not found: $service_dir${NC}"
        return
    fi
    
    echo -e "${YELLOW}Generating mocks for $service_name service...${NC}"
    
    # Create mocks directory
    local mocks_dir="$service_dir/internal/mocks"
    mkdir -p "$mocks_dir"
    
    # Generate mocks for repositories
    if [ -d "$service_dir/internal/repository" ]; then
        echo -e "  Generating repository mocks..."
        
        # Find Go files with interfaces
        local go_files=($(find "$service_dir/internal/repository" -name "*.go" -type f))
        
        for go_file in "${go_files[@]}"; do
            local filename=$(basename "$go_file" .go)
            local package_name=$(basename "$(dirname "$go_file")")
            
            # Check if file contains interfaces
            if grep -q "type.*interface" "$go_file"; then
                echo -e "    Processing $filename..."
                
                # Generate mock for each interface
                local interfaces=$(grep -o "type [A-Za-z0-9_]* interface" "$go_file" | sed 's/type \([A-Za-z0-9_]*\) interface/\1/')
                
                for interface in $interfaces; do
                    local mock_file="$mocks_dir/${filename}_${interface}_mock.go"
                    
                    mockgen \
                        -source="$go_file" \
                        -destination="$mock_file" \
                        -package=mocks \
                        "$interface"
                    
                    echo -e "      Generated mock for $interface"
                done
            fi
        done
    fi
    
    # Generate mocks for services
    if [ -d "$service_dir/internal/service" ]; then
        echo -e "  Generating service mocks..."
        
        local go_files=($(find "$service_dir/internal/service" -name "*.go" -type f))
        
        for go_file in "${go_files[@]}"; do
            local filename=$(basename "$go_file" .go)
            
            if grep -q "type.*interface" "$go_file"; then
                echo -e "    Processing $filename..."
                
                local interfaces=$(grep -o "type [A-Za-z0-9_]* interface" "$go_file" | sed 's/type \([A-Za-z0-9_]*\) interface/\1/')
                
                for interface in $interfaces; do
                    local mock_file="$mocks_dir/${filename}_${interface}_mock.go"
                    
                    mockgen \
                        -source="$go_file" \
                        -destination="$mock_file" \
                        -package=mocks \
                        "$interface"
                    
                    echo -e "      Generated mock for $interface"
                done
            fi
        done
    fi
    
    # Generate mocks for external dependencies
    if [ -d "$service_dir/internal" ]; then
        echo -e "  Generating external dependency mocks..."
        
        # Mock OpenAI client for chat service
        if [ "$service_name" = "chat" ] && [ -d "$service_dir/internal/service/openai" ]; then
            local openai_file="$service_dir/internal/service/openai/client.go"
            if [ -f "$openai_file" ]; then
                echo -e "    Processing OpenAI client..."
                
                mockgen \
                    -source="$openai_file" \
                    -destination="$mocks_dir/openai_client_mock.go" \
                    -package=mocks \
                    "OpenAIClient"
                
                echo -e "      Generated mock for OpenAI client"
            fi
        fi
        
        # Mock database connection
        if [ -d "$service_dir/internal/repository" ]; then
            local db_file="$service_dir/internal/repository/storage.go"
            if [ -f "$db_file" ]; then
                echo -e "    Processing database connection..."
                
                # Check if there's a database interface
                if grep -q "type.*Database.*interface" "$db_file"; then
                    mockgen \
                        -source="$db_file" \
                        -destination="$mocks_dir/database_mock.go" \
                        -package=mocks \
                        "Database"
                    
                    echo -e "      Generated mock for database"
                fi
            fi
        fi
    fi
    
    echo -e "${GREEN}✓ Generated mocks for $service_name service${NC}"
}

# Generate mocks for all services
echo -e "${YELLOW}Generating mocks for all services...${NC}"

# Auth service
generate_service_mocks "auth"

# Chat service
generate_service_mocks "chat"

# Generate mocks for shared packages
echo -e "${YELLOW}Generating mocks for shared packages...${NC}"

if [ -d "$PROJECT_ROOT/packages" ]; then
    # Create mocks directory for shared packages
    local pkg_mocks_dir="$PROJECT_ROOT/packages/mocks"
    mkdir -p "$pkg_mocks_dir"
    
    # Mock logger
    if [ -d "$PROJECT_ROOT/pkg/logger" ]; then
        local logger_file="$PROJECT_ROOT/pkg/logger/logger.go"
        if [ -f "$logger_file" ]; then
            echo -e "  Generating logger mock..."
            
            mockgen \
                -source="$logger_file" \
                -destination="$pkg_mocks_dir/logger_mock.go" \
                -package=mocks \
                "Logger"
            
            echo -e "    Generated mock for logger"
        fi
    fi
    
    # Mock auth utilities
    if [ -d "$PROJECT_ROOT/pkg/auth" ]; then
        local auth_file="$PROJECT_ROOT/pkg/auth/auth.go"
        if [ -f "$auth_file" ]; then
            echo -e "  Generating auth utilities mock..."
            
            # Check for interfaces
            if grep -q "type.*interface" "$auth_file"; then
                local interfaces=$(grep -o "type [A-Za-z0-9_]* interface" "$auth_file" | sed 's/type \([A-Za-z0-9_]*\) interface/\1/')
                
                for interface in $interfaces; do
                    local mock_file="$pkg_mocks_dir/auth_${interface}_mock.go"
                    
                    mockgen \
                        -source="$auth_file" \
                        -destination="$mock_file" \
                        -package=mocks \
                        "$interface"
                    
                    echo -e "    Generated mock for $interface"
                done
            fi
        fi
    fi
    
    # Mock database connection
    if [ -d "$PROJECT_ROOT/pkg/database" ]; then
        local db_file="$PROJECT_ROOT/pkg/database/database.go"
        if [ -f "$db_file" ]; then
            echo -e "  Generating database connection mock..."
            
            mockgen \
                -source="$db_file" \
                -destination="$pkg_mocks_dir/database_mock.go" \
                -package=mocks \
                "Connection"
            
            echo -e "    Generated mock for database connection"
        fi
    fi
    
    echo -e "${GREEN}✓ Generated mocks for shared packages${NC}"
fi

# Update go.mod files
echo -e "${YELLOW}Updating Go module dependencies...${NC}"

# Update auth service dependencies
if [ -d "$PROJECT_ROOT/services/auth-service" ]; then
    cd "$PROJECT_ROOT/services/auth-service"
    go mod tidy
    echo -e "${GREEN}✓ Updated auth service dependencies${NC}"
fi



# Update shared packages dependencies
if [ -d "$PROJECT_ROOT/packages" ]; then
    cd "$PROJECT_ROOT/packages"
    if [ -f "go.mod" ]; then
        go mod tidy
        echo -e "${GREEN}✓ Updated shared packages dependencies${NC}"
    fi
fi

echo -e "${GREEN}🎉 Mock generation completed successfully!${NC}"
echo -e "${YELLOW}Generated mocks are available in:${NC}"
echo -e "  - services/auth-service/internal/mocks/"

if [ -d "$PROJECT_ROOT/packages/mocks" ]; then
    echo -e "  - packages/mocks/"
fi

echo -e "${YELLOW}Next steps:${NC}"
echo -e "  1. Review generated mocks"
echo -e "  2. Update your tests to use the new mocks"
echo -e "  3. Run tests to ensure everything works"
echo -e "  4. Commit the generated mock files"

echo -e "${YELLOW}Note:${NC}"
echo -e "  - Mocks are generated in the mocks package"
echo -e "  - Import mocks in your tests: import 'your-service/internal/mocks'"
echo -e "  - Use mocks.NewMockInterfaceName(ctrl) to create mock instances"
