package repository

import (
	"context"

	"github.com/tikfack/server/internal/domain/entity"
)

// EventLogRepository defines the interface for event log persistence operations
// using the domain EventLog entity.
type EventLogRepository interface {
	// InsertEventLog persists a single EventLog to the underlying system (e.g., Kafka).
	InsertEventLog(ctx context.Context, log *entity.EventLog) error

	// InsertEventLogs persists multiple EventLogs in batch to the underlying system.
	InsertEventLogs(ctx context.Context, logs []*entity.EventLog) error
}