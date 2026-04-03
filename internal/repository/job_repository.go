package repository

import (
	"context"

	"github.com/HARA-DID/did-queueing-engine/internal/domain"
)

// JobRepository defines persistence operations for jobs.
type JobRepository interface {
	Create(ctx context.Context, job *domain.Job) error
	FindByEventID(ctx context.Context, eventID string) (*domain.Job, error)
	UpdateStatus(ctx context.Context, jobID string, status domain.JobStatus, txHashes []string, errMsg string) error
	IncrementRetry(ctx context.Context, jobID string, errMsg string) error
}
