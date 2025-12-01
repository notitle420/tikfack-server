package di

import (
	"os"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/segmentio/kafka-go"
	eventloguc "github.com/tikfack/server/internal/application/usecase/event_log"
	video "github.com/tikfack/server/internal/application/usecase/video"
	eventlogrepo "github.com/tikfack/server/internal/infrastructure/repository/event_log"
	connecthandler "github.com/tikfack/server/internal/presentation/connect"
)

func provideVideoHandler(vu video.VideoUsecase, opts []connect.HandlerOption) *connecthandler.VideoServiceServer {
	return connecthandler.NewVideoServiceHandler(vu, opts...)
}

func provideEventLogHandler(uc eventloguc.EventLogUsecase, opts []connect.HandlerOption) *connecthandler.EventLogServiceServer {
	return connecthandler.NewEventLogServiceHandler(uc, opts...)
}

func provideKafkaWriter() *kafka.Writer {
	brokers := parseKafkaBrokers(os.Getenv("KAFKA_BROKER_ADDRESSES"))
	if len(brokers) == 0 {
		brokers = []string{"localhost:9094"}
	}
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   "event-logs",
	})
}

func provideEventLogUsecase(writer *kafka.Writer) eventloguc.EventLogUsecase {
	return eventloguc.NewEventLogService(eventlogrepo.NewKafkaEventLogRepository(writer))
}

func parseKafkaBrokers(raw string) []string {
	var brokers []string
	for _, addr := range strings.Split(strings.TrimSpace(raw), ",") {
		addr = strings.TrimSpace(addr)
		if addr != "" {
			brokers = append(brokers, addr)
		}
	}
	return brokers
}
