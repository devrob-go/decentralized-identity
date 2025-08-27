package services

import (
	"fmt"
	"log"
	"time"

	"did-manager/internal/domain"
	"did-manager/pkg/blockchain"
	"did-manager/pkg/did"
	"did-manager/pkg/queue"

	"github.com/google/uuid"
)

// DIDService implements the DID business logic
type DIDService struct {
	didRepo    domain.DIDRepository
	queueRepo  domain.BlockchainJobRepository
	didGen     *did.Generator
	blockchain *blockchain.EthereumClient
	queue      *queue.NATSQueue
}

// NewDIDService creates a new DID service
func NewDIDService(
	didRepo domain.DIDRepository,
	queueRepo domain.BlockchainJobRepository,
	didGen *did.Generator,
	blockchain *blockchain.EthereumClient,
	queue *queue.NATSQueue,
) *DIDService {
	return &DIDService{
		didRepo:    didRepo,
		queueRepo:  queueRepo,
		didGen:     didGen,
		blockchain: blockchain,
		queue:      queue,
	}
}

// CreateDID creates a new DID for a user
func (s *DIDService) CreateDID(req *domain.DIDCreateRequest) (*domain.DIDResponse, error) {
	// Generate DID, user hash, and keys
	didString, userHash, privateKey, err := s.didGen.GenerateDID(req.UserID, req.Name, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate DID: %w", err)
	}

	// Create DID record in database
	didRecord := &domain.DID{
		ID:        uuid.New(),
		UserID:    req.UserID,
		Did:       didString,
		UserHash:  userHash,
		PublicKey: privateKey, // In production, this should be encrypted
		Status:    string(domain.DIDStatusPending),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.didRepo.Create(didRecord); err != nil {
		return nil, fmt.Errorf("failed to create DID record: %w", err)
	}

	// Create blockchain job for async processing
	blockchainJob := &domain.BlockchainJob{
		ID:         uuid.New(),
		JobType:    string(domain.JobTypeRegisterDID),
		DIDID:      didRecord.ID,
		UserHash:   userHash,
		DID:        didString,
		Status:     string(domain.JobStatusPending),
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.queueRepo.Create(blockchainJob); err != nil {
		log.Printf("Warning: failed to create blockchain job: %v", err)
		// Continue with DID creation even if job creation fails
	}

	// Publish job to NATS queue for async processing
	queueJob := &queue.BlockchainJob{
		ID:        blockchainJob.ID.String(),
		JobType:   blockchainJob.JobType,
		DIDID:     blockchainJob.DIDID.String(),
		UserHash:  blockchainJob.UserHash,
		DID:       blockchainJob.DID,
		CreatedAt: blockchainJob.CreatedAt,
	}

	if err := s.queue.PublishJob(queueJob); err != nil {
		log.Printf("Warning: failed to publish job to queue: %v", err)
		// Continue with DID creation even if queue publishing fails
	}

	return &domain.DIDResponse{
		DID:      didRecord,
		UserHash: userHash,
		Status:   string(domain.DIDStatusPending),
		Message:  "DID created successfully and queued for blockchain registration",
	}, nil
}

// VerifyDID verifies a DID on the blockchain
func (s *DIDService) VerifyDID(req *domain.DIDVerificationRequest) (*domain.DIDVerificationResponse, error) {
	log.Printf("DEBUG SERVICE: Starting verification for DID: %s", req.DID)

	// Check if repository is nil
	if s.didRepo == nil {
		log.Printf("DEBUG SERVICE: didRepo is nil!")
		return &domain.DIDVerificationResponse{
			IsValid:  false,
			DID:      req.DID,
			UserHash: req.UserHash,
			Status:   "not_found",
			Message:  "Service not properly initialized",
		}, nil
	}

	// First check if DID exists in our database
	didRecord, err := s.didRepo.GetByDID(req.DID)
	if err != nil {
		log.Printf("DEBUG SERVICE: GetByDID failed: %v", err)
		return &domain.DIDVerificationResponse{
			IsValid:  false,
			DID:      req.DID,
			UserHash: req.UserHash,
			Status:   "not_found",
			Message:  "DID not found in local database: " + err.Error(),
		}, nil
	}

	log.Printf("DEBUG SERVICE: Found DID record: %+v", didRecord)

	// Verify user hash matches (skip if empty for status checks)
	if req.UserHash != "" && didRecord.UserHash != req.UserHash {
		return &domain.DIDVerificationResponse{
			IsValid:  false,
			DID:      req.DID,
			UserHash: req.UserHash,
			Status:   "hash_mismatch",
			Message:  "User hash does not match",
		}, nil
	}

	// Verify on blockchain
	isValid, err := s.blockchain.VerifyDID(req.DID)
	if err != nil {
		log.Printf("Blockchain verification failed: %v", err)
		// Return local verification result if blockchain is unavailable
		return &domain.DIDVerificationResponse{
			IsValid:      didRecord.Status == string(domain.DIDStatusActive),
			DID:          req.DID,
			UserHash:     req.UserHash,
			Status:       didRecord.Status,
			Message:      "Blockchain verification failed, using local status",
			BlockchainTx: didRecord.BlockchainTx,
		}, nil
	}

	// Update local status if blockchain verification succeeds
	if isValid && didRecord.Status != string(domain.DIDStatusActive) {
		didRecord.Status = string(domain.DIDStatusActive)
		didRecord.UpdatedAt = time.Now()
		if err := s.didRepo.Update(didRecord); err != nil {
			log.Printf("Warning: failed to update DID status: %v", err)
		}
	}

	return &domain.DIDVerificationResponse{
		IsValid:      isValid,
		DID:          req.DID,
		UserHash:     req.UserHash,
		Status:       didRecord.Status,
		Message:      "DID verification completed",
		BlockchainTx: didRecord.BlockchainTx,
	}, nil
}

// GetDIDByUserID retrieves a DID by user ID
func (s *DIDService) GetDIDByUserID(userID uuid.UUID) (*domain.DID, error) {
	return s.didRepo.GetByUserID(userID)
}

// UpdateDIDStatus updates the status of a DID
func (s *DIDService) UpdateDIDStatus(didID uuid.UUID, status string, txHash string) error {
	return s.didRepo.UpdateStatus(didID, status, txHash)
}

// ProcessBlockchainQueue processes pending blockchain jobs
func (s *DIDService) ProcessBlockchainQueue() error {
	// Get pending jobs
	jobs, err := s.queueRepo.GetPendingJobs(10) // Process 10 jobs at a time
	if err != nil {
		return fmt.Errorf("failed to get pending jobs: %w", err)
	}

	for _, job := range jobs {
		if err := s.processJob(job); err != nil {
			log.Printf("Failed to process job %s: %v", job.ID, err)

			// Update job status to failed
			if err := s.queueRepo.UpdateStatus(job.ID, string(domain.JobStatusFailed), err.Error()); err != nil {
				log.Printf("Failed to update job status: %v", err)
			}
		}
	}

	return nil
}

// processJob processes a single blockchain job
func (s *DIDService) processJob(job *domain.BlockchainJob) error {
	// Update job status to processing
	if err := s.queueRepo.UpdateStatus(job.ID, string(domain.JobStatusProcessing), ""); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	var txHash string
	var err error

	// Process based on job type
	switch job.JobType {
	case string(domain.JobTypeRegisterDID):
		txHash, err = s.blockchain.RegisterDID(job.UserHash, job.DID)
	case string(domain.JobTypeUpdateDID):
		txHash, err = s.blockchain.UpdateDID(job.UserHash, job.DID)
	default:
		return fmt.Errorf("unknown job type: %s", job.JobType)
	}

	if err != nil {
		return fmt.Errorf("blockchain operation failed: %w", err)
	}

	// Update DID status to active
	if err := s.didRepo.UpdateStatus(job.DIDID, string(domain.DIDStatusActive), txHash); err != nil {
		return fmt.Errorf("failed to update DID status: %w", err)
	}

	// Mark job as completed
	if err := s.queueRepo.MarkCompleted(job.ID); err != nil {
		return fmt.Errorf("failed to mark job completed: %w", err)
	}

	log.Printf("Successfully processed job %s, transaction: %s", job.ID, txHash)
	return nil
}

// GetDIDRepo returns the DID repository for direct access (debug purposes)
func (s *DIDService) GetDIDRepo() domain.DIDRepository {
	return s.didRepo
}
