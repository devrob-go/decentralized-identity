#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Default values
ENVIRONMENT="local"
SERVICE="all"
HELM_NAMESPACE="default"

# Help function
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -e, --environment ENV    Deployment environment (local, staging, production)"
    echo "  -s, --service SERVICE    Service to deploy (auth, all)"
    echo "  -n, --namespace NS      Kubernetes namespace for Helm deployments"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 -e local                    # Deploy all services locally"
    echo "  $0 -e staging -s auth          # Deploy auth service to staging"
    echo ""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -s|--service)
            SERVICE="$2"
            shift 2
            ;;
        -n|--namespace)
            HELM_NAMESPACE="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# Validate environment
case $ENVIRONMENT in
    local|staging|production)
        ;;
    *)
        echo -e "${RED}Invalid environment: $ENVIRONMENT${NC}"
        echo "Valid environments: local, staging, production"
        exit 1
        ;;
esac

# Validate service
case $SERVICE in
    auth|all)
        ;;
    *)
        echo -e "${RED}Invalid service: $SERVICE${NC}"
        echo "Valid services: auth, all"
        exit 1
        ;;
esac

echo -e "${YELLOW}Deploying to $ENVIRONMENT environment...${NC}"

# Function to deploy a service
deploy_service() {
    local service_name=$1
    local env=$2
    
    echo -e "${YELLOW}Deploying $service_name service to $env...${NC}"
    
    case $env in
        local)
            # Local deployment using Docker Compose
            cd "$PROJECT_ROOT/deployments/local"
            if [ -f "docker-compose.yml" ]; then
                docker-compose up -d $service_name
                echo -e "${GREEN}✓ $service_name deployed locally${NC}"
            else
                echo -e "${RED}✗ docker-compose.yml not found in local deployment${NC}"
                return 1
            fi
            ;;
        staging|production)
            # Kubernetes deployment using Helm
            cd "$PROJECT_ROOT/deployments/$env"
            if [ -d "charts/$service_name-service" ]; then
                helm upgrade --install $service_name-service charts/$service_name-service \
                    --namespace $HELM_NAMESPACE \
                    --create-namespace \
                    --values values-$env.yaml
                echo -e "${GREEN}✓ $service_name deployed to $env${NC}"
            else
                echo -e "${RED}✗ Helm chart not found for $service_name service${NC}"
                return 1
            fi
            ;;
    esac
}

# Deploy services based on selection
case $SERVICE in
    auth)
        deploy_service "auth" $ENVIRONMENT
        ;;
    all)
        deploy_service "auth" $ENVIRONMENT
        ;;
esac

echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"

# Show status for local deployments
if [ "$ENVIRONMENT" = "local" ]; then
    echo -e "${YELLOW}Checking service status...${NC}"
    cd "$PROJECT_ROOT/deployments/local"
    if [ -f "docker-compose.yml" ]; then
        docker-compose ps
    fi
fi
