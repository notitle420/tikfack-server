package event_log

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"

	"github.com/tikfack/server/internal/domain/entity"
	repo "github.com/tikfack/server/internal/domain/repository"
)

// KafkaEventLogRepository implements repo.EventLogRepository by producing to Kafka.
type KafkaEventLogRepository struct {
	writer *kafka.Writer
}

// NewKafkaEventLogRepository returns an EventLogRepository backed by Kafka.
func NewKafkaEventLogRepository(writer *kafka.Writer) repo.EventLogRepository {
	return &KafkaEventLogRepository{writer: writer}
}

// InsertEventLog publishes a single EventLog to the Kafka topic.
func (k *KafkaEventLogRepository) InsertEventLog(ctx context.Context, e *entity.EventLog) error {
	payload, err := json.Marshal(e)
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Key:   []byte(e.UserID + ":" + e.EventTime.String()),
		Value: payload,
	}
	return k.writer.WriteMessages(ctx, msg)
}

// InsertEventLogs publishes multiple EventLogs in one batch to Kafka.
func (k *KafkaEventLogRepository) InsertEventLogs(ctx context.Context, events []*entity.EventLog) error {
	msgs := make([]kafka.Message, len(events))
	for i, e := range events {
		payload, err := json.Marshal(e)
		if err != nil {
			return err
		}
		msgs[i] = kafka.Message{
			Key:   []byte(e.UserID + ":" + e.EventTime.String()),
			Value: payload,
		}
	}
	return k.writer.WriteMessages(ctx, msgs...)
}