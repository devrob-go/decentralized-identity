package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DIDClient handles communication with the DID Manager service
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

// DIDRecord represents the DID record structure
type DIDRecord struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	DID          string `json:"did"`
	UserHash     string `json:"user_hash"`
	PublicKey    string `json:"public_key"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	BlockchainTx string `json:"blockchain_tx"`
}

// DIDCreateResponseData represents the data section of the DID creation response
type DIDCreateResponseData struct {
	DIDRecord DIDRecord `json:"did"`
	UserHash  string    `json:"user_hash"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
}

// DIDCreateResponse represents the full response from DID creation
type DIDCreateResponse struct {
	Success bool                  `json:"success"`
	Data    DIDCreateResponseData `json:"data"`
}

// CreateDID creates a new DID for a user
func (c *DIDClient) CreateDID(req *DIDCreateRequest) (*DIDCreateResponse, error) {
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
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response DIDCreateResponse

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("DID creation failed: %s", string(body))
	}

	return &response, nil
}
