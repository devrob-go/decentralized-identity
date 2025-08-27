package domain

import (
	"time"

	"github.com/google/uuid"
)

// BlockchainJob represents a job to be processed on the blockchain
type BlockchainJob struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	JobType     string     `json:"job_type" db:"job_type"` // register_did, update_did, revoke_did
	DIDID       uuid.UUID  `json:"did_id" db:"did_id"`
	UserHash    string     `json:"user_hash" db:"user_hash"`
	DID         string     `json:"did" db:"did"`
	Status      string     `json:"status" db:"status"` // pending, processing, completed, failed
	RetryCount  int        `json:"retry_count" db:"retry_count"`
	MaxRetries  int        `json:"max_retries" db:"max_retries"`
	Error       string     `json:"error" db:"error"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	ProcessedAt *time.Time `json:"processed_at" db:"processed_at"`
}

// JobStatus represents the current status of a blockchain job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusRetrying   JobStatus = "retrying"
)

// JobType represents the type of blockchain operation
type JobType string

const (
	JobTypeRegisterDID JobType = "register_did"
	JobTypeUpdateDID   JobType = "update_did"
	JobTypeRevokeDID   JobType = "revoke_did"
)

// BlockchainJobRepository defines the interface for blockchain job data operations
type BlockchainJobRepository interface {
	Create(job *BlockchainJob) error
	GetByID(id uuid.UUID) (*BlockchainJob, error)
	GetPendingJobs(limit int) ([]*BlockchainJob, error)
	UpdateStatus(id uuid.UUID, status string, error string) error
	MarkCompleted(id uuid.UUID) error
	IncrementRetryCount(id uuid.UUID) error
	CleanupCompletedJobs(daysOld int) error
}

// QueueService defines the interface for blockchain queue management
type QueueService interface {
	EnqueueJob(jobType JobType, didID uuid.UUID, userHash, did string) error
	ProcessNextJob() error
	RetryFailedJob(jobID uuid.UUID) error
	GetJobStatus(jobID uuid.UUID) (*BlockchainJob, error)
	CleanupOldJobs() error
}
