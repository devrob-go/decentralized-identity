#!/bin/bash

# Go Auth Service Helm Chart Deployment Script
# This script provides easy deployment, upgrade, and management of the Helm chart

set -e

# Configuration
CHART_NAME="go-auth-service"
NAMESPACE="staging"
RELEASE_NAME="go-auth-service"
VALUES_FILE="values-staging.yaml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check if helm is installed
    if ! command -v helm &> /dev/null; then
        print_error "helm is not installed. Please install helm first."
        exit 1
    fi
    
    # Check if namespace exists, create if not
    if ! kubectl get namespace $NAMESPACE &> /dev/null; then
        print_status "Creating namespace $NAMESPACE..."
        kubectl create namespace $NAMESPACE
        print_success "Namespace $NAMESPACE created"
    fi
    
    print_success "Prerequisites check passed"
}

# Function to add Helm repositories
add_repositories() {
    print_status "Adding Helm repositories..."
    
    # Add Bitnami repository for PostgreSQL and Redis
    helm repo add bitnami https://charts.bitnami.com/bitnami
    
    # Update repositories
    helm repo update
    
    print_success "Helm repositories added and updated"
}

# Function to update dependencies
update_dependencies() {
    print_status "Updating Helm chart dependencies..."
    
    helm dependency update .
    helm dependency update charts/auth-service

    
    print_success "Dependencies updated"
}

# Function to install the chart
install_chart() {
    print_status "Installing Helm chart..."
    
    helm install $RELEASE_NAME . \
        --namespace $NAMESPACE \
        --values $VALUES_FILE \
        --wait \
        --timeout 10m
    
    print_success "Chart installed successfully"
}

# Function to upgrade the chart
upgrade_chart() {
    print_status "Upgrading Helm chart..."
    
    helm upgrade $RELEASE_NAME . \
        --namespace $NAMESPACE \
        --values $VALUES_FILE \
        --wait \
        --timeout 10m
    
    print_success "Chart upgraded successfully"
}

# Function to uninstall the chart
uninstall_chart() {
    print_status "Uninstalling Helm chart..."
    
    helm uninstall $RELEASE_NAME --namespace $NAMESPACE
    
    print_success "Chart uninstalled successfully"
}

# Function to check deployment status
check_status() {
    print_status "Checking deployment status..."
    
    echo "=== Pods ==="
    kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=$CHART_NAME
    
    echo ""
    echo "=== Services ==="
    kubectl get services -n $NAMESPACE -l app.kubernetes.io/name=$CHART_NAME
    
    echo ""
    echo "=== Ingress ==="
    kubectl get ingress -n $NAMESPACE -l app.kubernetes.io/name=$CHART_NAME
    
    echo ""
    echo "=== HPA ==="
    kubectl get hpa -n $NAMESPACE -l app.kubernetes.io/name=$CHART_NAME
}

# Function to show logs
show_logs() {
    print_status "Showing application logs..."
    
    POD_NAME=$(kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=$CHART_NAME -o jsonpath="{.items[0].metadata.name}")
    
    if [ -n "$POD_NAME" ]; then
        kubectl logs -f $POD_NAME -n $NAMESPACE
    else
        print_error "No pods found"
        exit 1
    fi
}

# Function to set up port forwarding
port_forward() {
    print_status "Setting up port forwarding..."
    
    POD_NAME=$(kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=$CHART_NAME -o jsonpath="{.items[0].metadata.name}")
    
    if [ -n "$POD_NAME" ]; then
        echo "Port forwarding to $POD_NAME..."
        echo "REST API: http://localhost:8081"
        echo "gRPC: localhost:8080"
        echo "Press Ctrl+C to stop"
        kubectl port-forward $POD_NAME 8081:8081 8080:8080 -n $NAMESPACE
    else
        print_error "No pods found"
        exit 1
    fi
}

# Function to lint the chart
lint_chart() {
    print_status "Linting Helm chart..."
    
    helm lint .
    
    print_success "Chart linting completed"
}

# Function to template the chart
template_chart() {
    print_status "Templating Helm chart..."
    
    helm template $RELEASE_NAME . \
        --namespace $NAMESPACE \
        --values $VALUES_FILE
    
    print_success "Chart templating completed"
}

# Function to test the chart
test_chart() {
    print_status "Testing Helm chart..."
    
    helm test $RELEASE_NAME --namespace $NAMESPACE
    
    print_success "Chart testing completed"
}

# Function to clean up
cleanup() {
    print_status "Cleaning up temporary files..."
    
    rm -rf charts/*/charts
    rm -rf charts/*/requirements.lock
    
    print_success "Cleanup completed"
}

# Function to show help
show_help() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  install      - Install the Helm chart"
    echo "  upgrade      - Upgrade the Helm chart"
    echo "  uninstall    - Uninstall the Helm chart"
    echo "  status       - Check deployment status"
    echo "  logs         - Show application logs"
    echo "  port-forward - Set up port forwarding"
    echo "  lint         - Lint the Helm chart"
    echo "  template     - Template the Helm chart"
    echo "  test         - Test the Helm chart"
    echo "  clean        - Clean up temporary files"
    echo "  deps         - Update Helm dependencies"
    echo "  help         - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 install"
    echo "  $0 upgrade"
    echo "  $0 status"
}

# Main script logic
main() {
    case "${1:-help}" in
        install)
            check_prerequisites
            add_repositories
            update_dependencies
            install_chart
            ;;
        upgrade)
            check_prerequisites
            add_repositories
            update_dependencies
            upgrade_chart
            ;;
        uninstall)
            check_prerequisites
            uninstall_chart
            ;;
        status)
            check_prerequisites
            check_status
            ;;
        logs)
            check_prerequisites
            show_logs
            ;;
        port-forward)
            check_prerequisites
            port_forward
            ;;
        lint)
            check_prerequisites
            lint_chart
            ;;
        template)
            check_prerequisites
            template_chart
            ;;
        test)
            check_prerequisites
            test_chart
            ;;
        clean)
            cleanup
            ;;
        deps)
            check_prerequisites
            update_dependencies
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
