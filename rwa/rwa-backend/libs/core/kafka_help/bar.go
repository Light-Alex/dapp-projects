package kafka_help

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/AnchoredLabs/rwa-backend/libs/kafka"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"go.uber.org/zap"
)

// BarEvent is the payload published to Kafka for market bar data.
type BarEvent struct {
	Symbol     string  `json:"symbol"`
	Open       float64 `json:"open"`
	High       float64 `json:"high"`
	Low        float64 `json:"low"`
	Close      float64 `json:"close"`
	Volume     int64   `json:"volume"`
	Timestamp  int64   `json:"timestamp"`
	TradeCount int64   `json:"tradeCount,omitempty"`
	VWAP       float64 `json:"vwap,omitempty"`
}

// BarKafkaService publishes market bar events to Kafka.
type BarKafkaService struct {
	producer    *kafka.Producer
	kafkaConfig *KafkaConfig
}

func NewBarKafkaService(kafkaConfig *KafkaConfig) (*BarKafkaService, error) {
	if !kafkaConfig.Enabled {
		return &BarKafkaService{kafkaConfig: kafkaConfig}, nil
	}

	ctx := context.WithValue(context.Background(), log.TraceID, "BarKafkaServiceInit")
	producer, err := kafka.NewProducer(ctx, kafka.SyncProducerType, kafkaConfig.Brokers, DefaultProducerConfig(), nil, nil)
	if err != nil {
		log.ErrorZ(ctx, "BarKafka: failed to create producer", zap.Error(err))
		return nil, err
	}

	return &BarKafkaService{
		producer:    producer,
		kafkaConfig: kafkaConfig,
	}, nil
}

// Close closes the underlying Kafka producer.
func (s *BarKafkaService) Close(ctx context.Context) {
	if s.producer != nil {
		s.producer.Close(ctx)
	}
}

// Publish sends a bar event to Kafka synchronously.
func (s *BarKafkaService) Publish(ctx context.Context, event *BarEvent) {
	if !s.kafkaConfig.Enabled || s.producer == nil {
		return
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		log.ErrorZ(ctx, "BarKafka: failed to marshal event", zap.Error(err))
		return
	}

	msg := &sarama.ProducerMessage{
		Topic: RWA_MARKET_BAR_TOPIC,
		Key:   sarama.StringEncoder(event.Symbol),
		Value: sarama.ByteEncoder(jsonData),
	}

	if _, _, err := s.producer.SendMessage(ctx, msg); err != nil {
		log.ErrorZ(ctx, "BarKafka: failed to publish",
			zap.Error(err),
			zap.String("symbol", event.Symbol))
		return
	}

	log.DebugZ(ctx, "BarKafka: published bar event",
		zap.String("symbol", event.Symbol),
		zap.Float64("close", event.Close))
}
