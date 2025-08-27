#!/bin/bash

# Protocol Buffers to Go code generation script
# This script generates Go code from .proto files for all services

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo -e "${YELLOW}Generating Go code from Protocol Buffers...${NC}"

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}✗ protoc is not installed. Please install Protocol Buffers compiler.${NC}"
    echo "Installation guide: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

# Check if Go protobuf plugins are installed
check_plugin() {
    local plugin_name=$1
    local install_cmd=$2
    
    if ! command -v $plugin_name &> /dev/null; then
        echo -e "${YELLOW}Installing $plugin_name...${NC}"
        eval $install_cmd
    else
        echo -e "${GREEN}✓ $plugin_name is already installed${NC}"
    fi
}

# Check and install required plugins
check_plugin "protoc-gen-go" "go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
check_plugin "protoc-gen-go-grpc" "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
check_plugin "protoc-gen-grpc-gateway" "go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest"

echo -e "${GREEN}All required plugins are installed${NC}"

# Function to generate code for a service
generate_service_code() {
    local service_name=$1
    local proto_dir="$PROJECT_ROOT/api/$service_name/v1"
    local output_dir="$PROJECT_ROOT/services/${service_name}-service/api"
    
    if [ ! -d "$proto_dir" ]; then
        echo -e "${YELLOW}No protobuf files found for $service_name, skipping...${NC}"
        return
    fi
    
    echo -e "${YELLOW}Generating Go code for $service_name service...${NC}"
    
    # Create output directory
    mkdir -p "$output_dir"
    
    # Find all .proto files
    local proto_files=($(find "$proto_dir" -name "*.proto" -type f))
    
    if [ ${#proto_files[@]} -eq 0 ]; then
        echo -e "${YELLOW}No .proto files found in $proto_dir${NC}"
        return
    fi
    
    # Generate Go code
    cd "$proto_dir"
    
    for proto_file in "${proto_files[@]}"; do
        local filename=$(basename "$proto_file")
        echo -e "  Processing $filename..."
        
        protoc \
            --go_out="$output_dir" \
            --go_opt=paths=source_relative \
            --go-grpc_out="$output_dir" \
            --go-grpc_opt=paths=source_relative \
            --grpc-gateway_out="$output_dir" \
            --grpc-gateway_opt=paths=source_relative \
            --proto_path="$proto_dir" \
            --proto_path="$PROJECT_ROOT/api/common" \
            "$filename"
    done
    
    echo -e "${GREEN}✓ Generated Go code for $service_name service${NC}"
}

# Generate code for all services
echo -e "${YELLOW}Generating code for all services...${NC}"

# Auth service
generate_service_code "auth"



# Generate common protobuf code if it exists
if [ -d "$PROJECT_ROOT/api/common" ]; then
    echo -e "${YELLOW}Generating common protobuf code...${NC}"
    
    # Find common .proto files
    local common_proto_files=($(find "$PROJECT_ROOT/api/common" -name "*.proto" -type f))
    
    if [ ${#common_proto_files[@]} -gt 0 ]; then
        local common_output_dir="$PROJECT_ROOT/packages/proto"
        mkdir -p "$common_output_dir"
        
        cd "$PROJECT_ROOT/api/common"
        
        for proto_file in "${common_proto_files[@]}"; do
            local filename=$(basename "$proto_file")
            echo -e "  Processing common $filename..."
            
            protoc \
                --go_out="$common_output_dir" \
                --go_opt=paths=source_relative \
                "$filename"
        done
        
        echo -e "${GREEN}✓ Generated common protobuf code${NC}"
    fi
fi

# Update go.mod files if needed
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

echo -e "${GREEN}🎉 Protocol Buffers code generation completed successfully!${NC}"
echo -e "${YELLOW}Generated files are available in:${NC}"
echo -e "  - services/auth-service/api/"

if [ -d "$PROJECT_ROOT/packages/proto" ]; then
    echo -e "  - packages/proto/"
fi

echo -e "${YELLOW}Next steps:${NC}"
echo -e "  1. Review generated code"
echo -e "  2. Update your handlers to use new interfaces"
echo -e "  3. Run tests to ensure everything works"
echo -e "  4. Build and deploy your services"
