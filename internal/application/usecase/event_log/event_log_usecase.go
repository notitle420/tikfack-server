package usecase

import (
	"context"

	"github.com/tikfack/server/internal/domain/entity"
	repo "github.com/tikfack/server/internal/domain/repository"
)

// EventLogUsecase defines business operations for processing event logs
// It works with domain entities rather than protobuf types.
type EventLogUsecase interface {
	// Record processes and persists a single EventLog entity
	Record(ctx context.Context, log *entity.EventLog) error

	// RecordBatch processes and persists multiple EventLog entities
	RecordBatch(ctx context.Context, logs []*entity.EventLog) error
}

// eventLogService is a concrete implementation of EventLogUsecase
// delegating to the repository layer.
type eventLogService struct {
	repo repo.EventLogRepository
}

// NewEventLogService constructs a new EventLogUsecase
func NewEventLogService(r repo.EventLogRepository) EventLogUsecase {
	return &eventLogService{repo: r}
}

// Record validates and persists a single EventLog
func (s *eventLogService) Record(ctx context.Context, log *entity.EventLog) error {
	return s.repo.InsertEventLog(ctx, log)
}

// RecordBatch validates and persists multiple EventLog entries
func (s *eventLogService) RecordBatch(ctx context.Context, logs []*entity.EventLog) error {
	return s.repo.InsertEventLogs(ctx, logs)
}