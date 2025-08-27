.PHONY: proto build run test clean install-tools generate dev install-air

# Protobuf generation with both gRPC and REST gateway
proto:
	@echo "Generating protobuf code..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
		proto/auth.proto proto/health.proto
	@echo "Protobuf generation completed"

# Build the application
build: proto
	@echo "Building application..."
	go build -o bin/auth-service main.go
	@echo "Build completed"

# Run the application
run: build
	@echo "Running application..."
	./bin/auth-service

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f proto/*.pb.go
	rm -f proto/*.pb.gw.go
	@echo "Clean completed"

# Install protobuf tools
install-tools:
	@echo "Installing protobuf tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@echo "Tools installation completed"

# Generate and format code
generate: proto
	@echo "Formatting generated code..."
	go fmt ./...
	go mod tidy
	@echo "Code generation and formatting completed"

# Run with hot reload (requires air)
dev:
	@echo "Running with hot reload..."
	air

# Install air for hot reload
install-air:
	@echo "Installing air for hot reload..."
	go install github.com/cosmtrek/air@latest
