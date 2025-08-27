# Build stage
FROM golang:1.24.6 AS builder

# Install necessary tools including protoc
RUN apt-get update && apt-get install -y \
    git \
    ca-certificates \
    tzdata \
    wget \
    unzip \
    && rm -rf /var/lib/apt/lists/*

# Install protoc
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-x86_64.zip \
    && unzip protoc-25.1-linux-x86_64.zip -d /usr/local \
    && rm protoc-25.1-linux-x86_64.zip

# Install Go protobuf plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
    && go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# Set working directory
WORKDIR /app

# Copy the entire project structure
COPY . .

# Download dependencies
RUN cd services/auth-service && go mod download

# Generate protobuf code from the root directory where proto files are located
RUN protoc --proto_path=. --proto_path=api/common \
    --go_out=api/auth/v1/proto --go_opt=paths=source_relative \
    --go-grpc_out=api/auth/v1/proto --go-grpc_opt=paths=source_relative \
    api/auth/v1/proto/auth.proto api/auth/v1/proto/health.proto

# Create proto directory in auth-service and copy generated files
RUN mkdir -p services/auth-service/proto && \
    cp api/auth/v1/proto/*.go services/auth-service/proto/

# Build the application
RUN cd services/auth-service/cmd/server && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/services/auth-service/cmd/server/main .

# Copy migrations directory
COPY --from=builder /app/services/auth-service/internal/repository/migrations ./storage/migrations

# Expose both gRPC and REST ports
EXPOSE 8080 8081

# Run the application
CMD ["./main"]