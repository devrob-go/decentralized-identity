package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

// DIDClient represents a client for interacting with the DID Manager service
type DIDClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewDIDClient creates a new DID client
func NewDIDClient(baseURL string) *DIDClient {
	return &DIDClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DIDCreateRequest represents a request to create a DID
type DIDCreateRequest struct {
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// DIDResponse represents the response after DID creation
type DIDResponse struct {
	Success bool `json:"success"`
	Data    struct {
		DID struct {
			ID        string    `json:"id"`
			UserID    string    `json:"user_id"`
			Did       string    `json:"did"`
			UserHash  string    `json:"user_hash"`
			Status    string    `json:"status"`
			CreatedAt time.Time `json:"created_at"`
		} `json:"did"`
		UserHash string `json:"user_hash"`
		Status   string `json:"status"`
		Message  string `json:"message"`
	} `json:"data"`
}

// DIDVerificationRequest represents a request to verify a DID
type DIDVerificationRequest struct {
	DID      string `json:"did"`
	UserHash string `json:"user_hash"`
}

// DIDVerificationResponse represents the response after DID verification
type DIDVerificationResponse struct {
	Success bool `json:"success"`
	Data    struct {
		IsValid      bool   `json:"is_valid"`
		DID          string `json:"did"`
		UserHash     string `json:"user_hash"`
		Status       string `json:"status"`
		Message      string `json:"message"`
		BlockchainTx string `json:"blockchain_tx"`
	} `json:"data"`
}

// CreateDID creates a new DID
func (c *DIDClient) CreateDID(req *DIDCreateRequest) (*DIDResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/did",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var didResp DIDResponse
	if err := json.Unmarshal(body, &didResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &didResp, nil
}

// VerifyDID verifies a DID
func (c *DIDClient) VerifyDID(req *DIDVerificationRequest) (*DIDVerificationResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/did/verify",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var verifyResp DIDVerificationResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &verifyResp, nil
}

// GetDIDStatus gets the status of a DID
func (c *DIDClient) GetDIDStatus(did string) error {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v1/did/status/" + did)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("DID Status Response: %s\n", string(body))
	return nil
}

// HealthCheck checks the health of the DID Manager service
func (c *DIDClient) HealthCheck() error {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v1/health")
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Health Check Response: %s\n", string(body))
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run did-cli.go <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  health                    - Check service health")
		fmt.Println("  create <name> <email>     - Create a new DID")
		fmt.Println("  verify <did> <userHash>   - Verify a DID")
		fmt.Println("  status <did>              - Get DID status")
		fmt.Println("  demo                      - Run a complete demo workflow")
		return
	}

	// Initialize client
	client := NewDIDClient("http://localhost:8082")

	command := os.Args[1]

	switch command {
	case "health":
		if err := client.HealthCheck(); err != nil {
			fmt.Printf("Health check failed: %v\n", err)
			os.Exit(1)
		}

	case "create":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run did-cli.go create <name> <email>")
			os.Exit(1)
		}
		name := os.Args[2]
		email := os.Args[3]

		req := &DIDCreateRequest{
			UserID:   uuid.New().String(),
			Name:     name,
			Email:    email,
			Password: "password123",
		}

		resp, err := client.CreateDID(req)
		if err != nil {
			fmt.Printf("Failed to create DID: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("DID created successfully!\n")
		fmt.Printf("DID: %s\n", resp.Data.DID.Did)
		fmt.Printf("User Hash: %s\n", resp.Data.UserHash)
		fmt.Printf("Status: %s\n", resp.Data.Status)
		fmt.Printf("Message: %s\n", resp.Data.Message)

	case "verify":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run did-cli.go verify <did> <userHash>")
			os.Exit(1)
		}
		did := os.Args[2]
		userHash := os.Args[3]

		req := &DIDVerificationRequest{
			DID:      did,
			UserHash: userHash,
		}

		resp, err := client.VerifyDID(req)
		if err != nil {
			fmt.Printf("Failed to verify DID: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("DID verification completed!\n")
		fmt.Printf("Is Valid: %t\n", resp.Data.IsValid)
		fmt.Printf("Status: %s\n", resp.Data.Status)
		fmt.Printf("Message: %s\n", resp.Data.Message)
		if resp.Data.BlockchainTx != "" {
			fmt.Printf("Blockchain TX: %s\n", resp.Data.BlockchainTx)
		}

	case "status":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run did-cli.go status <did>")
			os.Exit(1)
		}
		did := os.Args[2]

		if err := client.GetDIDStatus(did); err != nil {
			fmt.Printf("Failed to get DID status: %v\n", err)
			os.Exit(1)
		}

	case "demo":
		fmt.Println("Running complete DID workflow demo...")
		fmt.Println("=====================================")

		// Step 1: Health check
		fmt.Println("\n1. Checking service health...")
		if err := client.HealthCheck(); err != nil {
			fmt.Printf("Health check failed: %v\n", err)
			os.Exit(1)
		}

		// Step 2: Create DID
		fmt.Println("\n2. Creating a new DID...")
		req := &DIDCreateRequest{
			UserID:   uuid.New().String(),
			Name:     "John Doe",
			Email:    "john.doe@example.com",
			Password: "password123",
		}

		resp, err := client.CreateDID(req)
		if err != nil {
			fmt.Printf("Failed to create DID: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ DID created successfully!\n")
		fmt.Printf("  DID: %s\n", resp.Data.DID.Did)
		fmt.Printf("  User Hash: %s\n", resp.Data.UserHash)
		fmt.Printf("  Status: %s\n", resp.Data.Status)

		// Step 3: Verify DID
		fmt.Println("\n3. Verifying the created DID...")
		verifyReq := &DIDVerificationRequest{
			DID:      resp.Data.DID.Did,
			UserHash: resp.Data.UserHash,
		}

		verifyResp, err := client.VerifyDID(verifyReq)
		if err != nil {
			fmt.Printf("Failed to verify DID: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ DID verification completed!\n")
		fmt.Printf("  Is Valid: %t\n", verifyResp.Data.IsValid)
		fmt.Printf("  Status: %s\n", verifyResp.Data.Status)
		fmt.Printf("  Message: %s\n", verifyResp.Data.Message)

		// Step 4: Check status
		fmt.Println("\n4. Checking DID status...")
		if err := client.GetDIDStatus(resp.Data.DID.Did); err != nil {
			fmt.Printf("Failed to get DID status: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n=====================================")
		fmt.Println("Demo completed successfully!")

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
