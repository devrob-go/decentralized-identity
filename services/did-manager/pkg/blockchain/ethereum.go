package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthereumClient handles interactions with Ethereum blockchain
type EthereumClient struct {
	client     *ethclient.Client
	privateKey *ecdsa.PrivateKey
	address    common.Address
	contract   common.Address
	chainID    *big.Int
	gasLimit   uint64
	gasPrice   *big.Int
}

// NewEthereumClient creates a new Ethereum client
func NewEthereumClient(rpcURL, privateKeyHex, contractAddress string) (*EthereumClient, error) {
	// Connect to Ethereum node
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Get public key and address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to get public key")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Get chain ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Parse contract address
	contract := common.HexToAddress(contractAddress)

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	return &EthereumClient{
		client:     client,
		privateKey: privateKey,
		address:    address,
		contract:   contract,
		chainID:    chainID,
		gasLimit:   300000, // Adjust based on contract complexity
		gasPrice:   gasPrice,
	}, nil
}

// RegisterDID registers a DID on the blockchain
func (e *EthereumClient) RegisterDID(userHash, did string) (string, error) {
	// DID Registry ABI (simplified)
	didRegistryABI := `[
		{
			"inputs": [
				{"name": "userHash", "type": "bytes32"},
				{"name": "did", "type": "string"}
			],
			"name": "registerDID",
			"outputs": [],
			"stateMutability": "nonpayable",
			"type": "function"
		}
	]`

	parsedABI, err := abi.JSON(strings.NewReader(didRegistryABI))
	if err != nil {
		return "", fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Encode function call
	data, err := parsedABI.Pack("registerDID", common.HexToHash(userHash), did)
	if err != nil {
		return "", fmt.Errorf("failed to pack function call: %w", err)
	}

	// Create transaction
	tx, err := e.sendTransaction(data)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return tx.Hash().Hex(), nil
}

// UpdateDID updates a DID on the blockchain
func (e *EthereumClient) UpdateDID(userHash, did string) (string, error) {
	// DID Registry ABI for update
	didRegistryABI := `[
		{
			"inputs": [
				{"name": "userHash", "type": "bytes32"},
				{"name": "did", "type": "string"}
			],
			"name": "updateDID",
			"outputs": [],
			"stateMutability": "nonpayable",
			"type": "function"
		}
	]`

	parsedABI, err := abi.JSON(strings.NewReader(didRegistryABI))
	if err != nil {
		return "", fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Encode function call
	data, err := parsedABI.Pack("updateDID", common.HexToHash(userHash), did)
	if err != nil {
		return "", fmt.Errorf("failed to pack function call: %w", err)
	}

	// Create transaction
	tx, err := e.sendTransaction(data)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return tx.Hash().Hex(), nil
}

// VerifyDID verifies a DID on the blockchain
func (e *EthereumClient) VerifyDID(did string) (bool, error) {
	// DID Registry ABI for verification
	didRegistryABI := `[
		{
			"inputs": [
				{"name": "did", "type": "string"}
			],
			"name": "verifyDID",
			"outputs": [{"name": "", "type": "bool"}],
			"stateMutability": "view",
			"type": "function"
		}
	]`

	parsedABI, err := abi.JSON(strings.NewReader(didRegistryABI))
	if err != nil {
		return false, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Encode function call
	data, err := parsedABI.Pack("verifyDID", did)
	if err != nil {
		return false, fmt.Errorf("failed to pack function call: %w", err)
	}

	// Call contract (read-only)
	result, err := e.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &e.contract,
		Data: data,
	}, nil)
	if err != nil {
		return false, fmt.Errorf("failed to call contract: %w", err)
	}

	// Decode result
	var isValid bool
	err = parsedABI.UnpackIntoInterface(&isValid, "verifyDID", result)
	if err != nil {
		return false, fmt.Errorf("failed to unpack result: %w", err)
	}

	return isValid, nil
}

// sendTransaction sends a transaction to the blockchain
func (e *EthereumClient) sendTransaction(data []byte) (*types.Transaction, error) {
	// Get nonce
	nonce, err := e.client.PendingNonceAt(context.Background(), e.address)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		e.contract,
		big.NewInt(0), // No ETH transfer
		e.gasLimit,
		e.gasPrice,
		data,
	)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(e.chainID), e.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = e.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	// Wait for transaction to be mined
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Poll for transaction receipt
	var receipt *types.Receipt
	for {
		receipt, err = e.client.TransactionReceipt(ctx, signedTx.Hash())
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("transaction wait timeout")
		case <-time.After(time.Second):
			// Continue polling
		}
	}

	if receipt.Status == 0 {
		return nil, fmt.Errorf("transaction failed")
	}

	log.Printf("Transaction mined: %s", signedTx.Hash().Hex())
	return signedTx, nil
}

// Close closes the Ethereum client connection
func (e *EthereumClient) Close() {
	if e.client != nil {
		e.client.Close()
	}
}
