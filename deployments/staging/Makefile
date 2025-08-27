# Helm Chart Makefile for Go Auth Service
# Provides common operations for managing the Helm chart

.PHONY: help install upgrade uninstall status logs port-forward lint template test clean

# Configuration
CHART_NAME := go-auth-service
NAMESPACE := staging
RELEASE_NAME := go-auth-service
VALUES_FILE := values-staging.yaml

# Default target
help:
	@echo "Available commands:"
	@echo "  install      - Install the Helm chart"
	@echo "  upgrade      - Upgrade the Helm chart"
	@echo "  uninstall    - Uninstall the Helm chart"
	@echo "  status       - Check deployment status"
	@echo "  logs         - Show application logs"
	@echo "  port-forward - Set up port forwarding"
	@echo "  lint         - Lint the Helm chart"
	@echo "  template     - Template the Helm chart"
	@echo "  test         - Test the Helm chart"
	@echo "  clean        - Clean up temporary files"
	@echo "  deps         - Update Helm dependencies"

# Install the Helm chart
install:
	@echo "Installing Helm chart..."
	helm install $(RELEASE_NAME) . \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--values $(VALUES_FILE) \
		--wait \
		--timeout 10m

# Upgrade the Helm chart
upgrade:
	@echo "Upgrading Helm chart..."
	helm upgrade $(RELEASE_NAME) . \
		--namespace $(NAMESPACE) \
		--values $(VALUES_FILE) \
		--wait \
		--timeout 10m

# Uninstall the Helm chart
uninstall:
	@echo "Uninstalling Helm chart..."
	helm uninstall $(RELEASE_NAME) --namespace $(NAMESPACE)

# Check deployment status
status:
	@echo "Checking deployment status..."
	@echo "=== Pods ==="
	kubectl get pods -n $(NAMESPACE) -l app.kubernetes.io/name=$(CHART_NAME)
	@echo ""
	@echo "=== Services ==="
	kubectl get services -n $(NAMESPACE) -l app.kubernetes.io/name=$(CHART_NAME)
	@echo ""
	@echo "=== Ingress ==="
	kubectl get ingress -n $(NAMESPACE) -l app.kubernetes.io/name=$(CHART_NAME)

# Show application logs
logs:
	@echo "Showing application logs..."
	@POD_NAME=$$(kubectl get pods -n $(NAMESPACE) -l app.kubernetes.io/name=$(CHART_NAME) -o jsonpath="{.items[0].metadata.name}") && \
	if [ -n "$$POD_NAME" ]; then \
		kubectl logs -f $$POD_NAME -n $(NAMESPACE); \
	else \
		echo "No pods found"; \
	fi

# Set up port forwarding
port-forward:
	@echo "Setting up port forwarding..."
	@POD_NAME=$$(kubectl get pods -n $(NAMESPACE) -l app.kubernetes.io/name=$(CHART_NAME) -o jsonpath="{.items[0].metadata.name}") && \
	if [ -n "$$POD_NAME" ]; then \
		echo "Port forwarding to $$POD_NAME..."; \
		echo "REST API: http://localhost:8081"; \
		echo "gRPC: localhost:8080"; \
		echo "Press Ctrl+C to stop"; \
		kubectl port-forward $$POD_NAME 8081:8081 8080:8080 -n $(NAMESPACE); \
	else \
		echo "No pods found"; \
	fi

# Lint the Helm chart
lint:
	@echo "Linting Helm chart..."
	helm lint .

# Template the Helm chart
template:
	@echo "Templating Helm chart..."
	helm template $(RELEASE_NAME) . \
		--namespace $(NAMESPACE) \
		--values $(VALUES_FILE)

# Test the Helm chart
test:
	@echo "Testing Helm chart..."
	helm test $(RELEASE_NAME) --namespace $(NAMESPACE)

# Clean up temporary files
clean:
	@echo "Cleaning up temporary files..."
	rm -rf charts/*/charts
	rm -rf charts/*/requirements.lock

# Update Helm dependencies
deps:
	@echo "Updating Helm dependencies..."
	helm dependency update .
	helm dependency update charts/auth-service
