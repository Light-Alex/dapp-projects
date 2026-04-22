package kafka_help

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/AnchoredLabs/rwa-backend/libs/kafka"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"go.uber.org/zap"
)

// OrderUpdateEvent is the payload published to Kafka and pushed to WS clients.
type OrderUpdateEvent struct {
	AccountID         uint64 `json:"accountId"`
	OrderID           uint64 `json:"orderId"`
	ClientOrderID     string `json:"clientOrderId"`
	Symbol            string `json:"symbol"`
	Side              string `json:"side"`
	Status            string `json:"status"`
	FilledQuantity    string `json:"filledQuantity"`
	FilledPrice       string `json:"filledPrice"`
	RemainingQuantity string `json:"remainingQuantity"`
	Quantity          string `json:"quantity"`
	Event             string `json:"event"` // new, fill, partial_fill, cancelled, rejected, expired
	Timestamp         int64  `json:"timestamp"`
}

// OrderUpdateKafkaService publishes order update events to Kafka.
type OrderUpdateKafkaService struct {
	producer    *kafka.Producer
	kafkaConfig *KafkaConfig
}

func NewOrderUpdateKafkaService(kafkaConfig *KafkaConfig) (*OrderUpdateKafkaService, error) {
	if !kafkaConfig.Enabled {
		return &OrderUpdateKafkaService{kafkaConfig: kafkaConfig}, nil
	}

	ctx := context.WithValue(context.Background(), log.TraceID, "OrderUpdateKafkaServiceInit")
	producer, err := kafka.NewProducer(ctx, kafka.SyncProducerType, kafkaConfig.Brokers, DefaultProducerConfig(), nil, nil)
	if err != nil {
		log.ErrorZ(ctx, "OrderUpdateKafka: failed to create producer", zap.Error(err))
		return nil, err
	}

	return &OrderUpdateKafkaService{
		producer:    producer,
		kafkaConfig: kafkaConfig,
	}, nil
}

// Close closes the underlying Kafka producer.
func (s *OrderUpdateKafkaService) Close(ctx context.Context) {
	if s.producer != nil {
		s.producer.Close(ctx)
	}
}

// Publish sends an order update event to Kafka synchronously.
func (s *OrderUpdateKafkaService) Publish(ctx context.Context, event *OrderUpdateEvent) {
	if !s.kafkaConfig.Enabled || s.producer == nil {
		return
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		log.ErrorZ(ctx, "OrderUpdateKafka: failed to marshal event", zap.Error(err))
		return
	}

	msg := &sarama.ProducerMessage{
		Topic: RWA_ORDER_UPDATE_TOPIC,
		Key:   sarama.StringEncoder(fmt.Sprintf("%d", event.AccountID)),
		Value: sarama.ByteEncoder(jsonData),
	}

	if _, _, err := s.producer.SendMessage(ctx, msg); err != nil {
		log.ErrorZ(ctx, "OrderUpdateKafka: failed to publish",
			zap.Error(err),
			zap.Uint64("orderId", event.OrderID),
			zap.String("event", event.Event))
		return
	}

	log.DebugZ(ctx, "OrderUpdateKafka: published order update",
		zap.Uint64("orderId", event.OrderID),
		zap.String("event", event.Event),
		zap.String("status", event.Status))
}
