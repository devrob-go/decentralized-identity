package repository

import (
	"database/sql"
	"fmt"
	"time"

	"did-manager/internal/domain"

	"github.com/google/uuid"
)

// BlockchainJobRepository implements the blockchain job repository interface
type BlockchainJobRepository struct {
	db *sql.DB
}

// NewBlockchainJobRepository creates a new blockchain job repository
func NewBlockchainJobRepository(db *sql.DB) *BlockchainJobRepository {
	return &BlockchainJobRepository{db: db}
}

// Create creates a new blockchain job record
func (r *BlockchainJobRepository) Create(job *domain.BlockchainJob) error {
	query := `
		INSERT INTO blockchain_jobs (id, job_type, did_id, user_hash, did, status, retry_count, max_retries, error, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(query,
		job.ID,
		job.JobType,
		job.DIDID,
		job.UserHash,
		job.DID,
		job.Status,
		job.RetryCount,
		job.MaxRetries,
		job.Error,
		job.CreatedAt,
		job.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create blockchain job: %w", err)
	}

	return nil
}

// GetByID retrieves a blockchain job by ID
func (r *BlockchainJobRepository) GetByID(id uuid.UUID) (*domain.BlockchainJob, error) {
	query := `
		SELECT id, job_type, did_id, user_hash, did, status, retry_count, max_retries, error, created_at, updated_at, processed_at
		FROM blockchain_jobs WHERE id = $1
	`

	var job domain.BlockchainJob
	err := r.db.QueryRow(query, id).Scan(
		&job.ID,
		&job.JobType,
		&job.DIDID,
		&job.UserHash,
		&job.DID,
		&job.Status,
		&job.RetryCount,
		&job.MaxRetries,
		&job.Error,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.ProcessedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blockchain job not found")
		}
		return nil, fmt.Errorf("failed to get blockchain job: %w", err)
	}

	return &job, nil
}

// GetPendingJobs retrieves pending blockchain jobs
func (r *BlockchainJobRepository) GetPendingJobs(limit int) ([]*domain.BlockchainJob, error) {
	query := `
		SELECT id, job_type, did_id, user_hash, did, status, retry_count, max_retries, error, created_at, updated_at, processed_at
		FROM blockchain_jobs 
		WHERE status IN ($1, $2) AND retry_count < max_retries
		ORDER BY created_at ASC
		LIMIT $3
	`

	rows, err := r.db.Query(query, domain.JobStatusPending, domain.JobStatusRetrying, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*domain.BlockchainJob
	for rows.Next() {
		var job domain.BlockchainJob
		err := rows.Scan(
			&job.ID,
			&job.JobType,
			&job.DIDID,
			&job.UserHash,
			&job.DID,
			&job.Status,
			&job.RetryCount,
			&job.MaxRetries,
			&job.Error,
			&job.CreatedAt,
			&job.UpdatedAt,
			&job.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan blockchain job: %w", err)
		}
		jobs = append(jobs, &job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return jobs, nil
}

// UpdateStatus updates the status of a blockchain job
func (r *BlockchainJobRepository) UpdateStatus(id uuid.UUID, status string, errorMsg string) error {
	query := `
		UPDATE blockchain_jobs 
		SET status = $2, error = $3, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(query, id, status, errorMsg)
	if err != nil {
		return fmt.Errorf("failed to update blockchain job status: %w", err)
	}

	return nil
}

// MarkCompleted marks a blockchain job as completed
func (r *BlockchainJobRepository) MarkCompleted(id uuid.UUID) error {
	query := `
		UPDATE blockchain_jobs 
		SET status = $2, processed_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(query, id, domain.JobStatusCompleted)
	if err != nil {
		return fmt.Errorf("failed to mark blockchain job completed: %w", err)
	}

	return nil
}

// IncrementRetryCount increments the retry count for a blockchain job
func (r *BlockchainJobRepository) IncrementRetryCount(id uuid.UUID) error {
	query := `
		UPDATE blockchain_jobs 
		SET retry_count = retry_count + 1, status = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(query, id, domain.JobStatusRetrying)
	if err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}

	return nil
}

// CleanupCompletedJobs removes old completed jobs
func (r *BlockchainJobRepository) CleanupCompletedJobs(daysOld int) error {
	cutoffDate := time.Now().AddDate(0, 0, -daysOld)

	query := `
		DELETE FROM blockchain_jobs 
		WHERE status = $1 AND processed_at < $2
	`

	_, err := r.db.Exec(query, domain.JobStatusCompleted, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to cleanup completed jobs: %w", err)
	}

	return nil
}
