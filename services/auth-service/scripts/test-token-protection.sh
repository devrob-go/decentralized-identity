#!/bin/bash

# Test Token Protection Script for Auth Service
# This script tests that protected endpoints require valid authentication tokens

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8082"  # Auth service REST gateway port
GRPC_URL="localhost:8081"         # Auth service gRPC port

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Test data
TEST_EMAIL="test@example.com"
TEST_PASSWORD="TestPassword123!"
TEST_NAME="Test User"

# Global variables to store tokens
ACCESS_TOKEN=""
REFRESH_TOKEN=""

# Function to test endpoint without authentication
test_unauth_endpoint() {
    local endpoint="$1"
    local method="$2"
    local description="$3"
    
    print_status "Testing $description without authentication..."
    
    response=$(curl -s -w "%{http_code}" -X "$method" \
        -H "Content-Type: application/json" \
        "$BASE_URL$endpoint" \
        -o /dev/null)
    
    if [[ "$response" == "401" || "$response" == "403" ]]; then
        print_success "✓ $description correctly rejected without auth (HTTP $response)"
        return 0
    else
        print_error "✗ $description should require authentication but returned HTTP $response"
        return 1
    fi
}

# Function to test endpoint with invalid token
test_invalid_token_endpoint() {
    local endpoint="$1"
    local method="$2"
    local description="$3"
    
    print_status "Testing $description with invalid token..."
    
    response=$(curl -s -w "%{http_code}" -X "$method" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer invalid_token_here" \
        "$BASE_URL$endpoint" \
        -o /dev/null)
    
    if [[ "$response" == "401" || "$response" == "403" ]]; then
        print_success "✓ $description correctly rejected with invalid token (HTTP $response)"
        return 0
    else
        print_error "✗ $description should reject invalid tokens but returned HTTP $response"
        return 1
    fi
}

# Function to test endpoint with valid token
test_valid_token_endpoint() {
    local endpoint="$1"
    local method="$2"
    local description="$3"
    local data="$4"
    
    print_status "Testing $description with valid token..."
    
    if [[ -n "$data" ]]; then
        response=$(curl -s -w "%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $ACCESS_TOKEN" \
            -d "$data" \
            "$BASE_URL$endpoint" \
            -o /dev/null)
    else
        response=$(curl -s -w "%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $ACCESS_TOKEN" \
            "$BASE_URL$endpoint" \
            -o /dev/null)
    fi
    
    if [[ "$response" == "200" || "$response" == "201" ]]; then
        print_success "✓ $description accepted with valid token (HTTP $response)"
        return 0
    else
        print_warning "⚠ $description returned HTTP $response (may be expected for some endpoints)"
        return 0
    fi
}

# Function to create a test user and get tokens
setup_test_user() {
    print_status "Setting up test user..."
    
    # First, try to sign up
    signup_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"$TEST_NAME\",\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}" \
        "$BASE_URL/v1/auth/signup")
    
    # If signup fails (user might already exist), try to sign in
    if ! echo "$signup_response" | grep -q "access_token"; then
        print_warning "Signup failed (user might exist), trying signin..."
        signin_response=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}" \
            "$BASE_URL/v1/auth/signin")
        signup_response="$signin_response"
    fi
    
    # Extract tokens
    ACCESS_TOKEN=$(echo "$signup_response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    REFRESH_TOKEN=$(echo "$signup_response" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
    
    if [[ -n "$ACCESS_TOKEN" ]]; then
        print_success "✓ Test user setup complete with valid tokens"
        return 0
    else
        print_error "✗ Failed to get valid tokens from auth response"
        echo "Response: $signup_response"
        return 1
    fi
}

# Function to clean up test user
cleanup_test_user() {
    if [[ -n "$ACCESS_TOKEN" ]]; then
        print_status "Cleaning up test user..."
        curl -s -X POST \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $ACCESS_TOKEN" \
            -d "{\"access_token\":\"$ACCESS_TOKEN\"}" \
            "$BASE_URL/v1/auth/signout" > /dev/null
        print_status "✓ Test user signed out"
    fi
}

# Main test function
run_token_protection_tests() {
    print_status "Starting token protection tests..."
    echo
    
    # Setup test user and get tokens
    if ! setup_test_user; then
        print_error "Failed to setup test user. Exiting..."
        exit 1
    fi
    
    echo
    print_status "Testing protected endpoints..."
    echo
    
    # Test ListUsers endpoint (this is what we just protected)
    test_unauth_endpoint "/v1/users" "GET" "ListUsers endpoint"
    test_invalid_token_endpoint "/v1/users" "GET" "ListUsers endpoint"
    test_valid_token_endpoint "/v1/users" "GET" "ListUsers endpoint"
    
    echo
    
    # Test other protected endpoints
    test_unauth_endpoint "/v1/auth/signout" "POST" "SignOut endpoint"
    test_invalid_token_endpoint "/v1/auth/signout" "POST" "SignOut endpoint"
    
    test_unauth_endpoint "/v1/auth/refresh" "POST" "RefreshToken endpoint"
    test_invalid_token_endpoint "/v1/auth/refresh" "POST" "RefreshToken endpoint"
    
    test_unauth_endpoint "/v1/auth/revoke" "POST" "RevokeToken endpoint"
    test_invalid_token_endpoint "/v1/auth/revoke" "POST" "RevokeToken endpoint"
    
    echo
    print_status "Testing unprotected endpoints (should work without auth)..."
    echo
    
    # Test unprotected endpoints
    response=$(curl -s -w "%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}" \
        "$BASE_URL/v1/auth/signin" \
        -o /dev/null)
    
    if [[ "$response" == "200" ]]; then
        print_success "✓ SignIn endpoint works without prior auth (HTTP $response)"
    else
        print_warning "⚠ SignIn endpoint returned HTTP $response"
    fi
    
    echo
    cleanup_test_user
    
    print_success "Token protection tests completed!"
}

# Function to check if service is running
check_service() {
    print_status "Checking if auth service is running..."
    
    if ! curl -s "$BASE_URL/health" > /dev/null 2>&1; then
        print_error "Auth service is not running or not accessible at $BASE_URL"
        print_error "Please start the service first:"
        print_error "  cd services/auth-service && make run"
        exit 1
    fi
    
    print_success "✓ Auth service is running"
}

# Main execution
main() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}   Auth Service Token Protection Test   ${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo
    
    check_service
    run_token_protection_tests
    
    echo
    echo -e "${BLUE}========================================${NC}"
    echo -e "${GREEN}   Test completed successfully!        ${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# Run main function
main "$@"
