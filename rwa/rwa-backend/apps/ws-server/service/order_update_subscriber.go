package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/AnchoredLabs/rwa-backend/apps/ws-server/types"
	"github.com/AnchoredLabs/rwa-backend/apps/ws-server/ws"
	"github.com/AnchoredLabs/rwa-backend/libs/core/kafka_help"
	"github.com/AnchoredLabs/rwa-backend/libs/kafka"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/olahol/melody"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type OrderUpdateSubscriber struct {
	kafkaConfig *kafka_help.KafkaConfig
	wsServer    *ws.Server
	consumer    *kafka.Consumer
}

func NewOrderUpdateSubscriber(
	lc fx.Lifecycle,
	kafkaConfig *kafka_help.KafkaConfig,
	wsServer *ws.Server,
) *OrderUpdateSubscriber {
	s := &OrderUpdateSubscriber{
		kafkaConfig: kafkaConfig,
		wsServer:    wsServer,
	}
	lc.Append(fx.Hook{
		OnStart: s.Start,
		OnStop:  s.Stop,
	})
	return s
}

func (s *OrderUpdateSubscriber) Start(ctx context.Context) error {
	if !s.kafkaConfig.Enabled {
		log.WarnZ(ctx, "OrderUpdateSubscriber: Kafka not enabled, skipping")
		return nil
	}

	consumer, err := kafka.NewKafkaConsumer(
		s.kafkaConfig.Brokers,
		nil, // use default config
		kafka_help.RWA_ORDER_UPDATE_CONSUMER_GROUP,
		kafka_help.RWA_ORDER_UPDATE_TOPIC,
		s.handleMessage,
		s.kafkaConfig.GetOrderUpdatePartitions(),
		true, // auto-ack
	)
	if err != nil {
		log.ErrorZ(ctx, "OrderUpdateSubscriber: failed to create Kafka consumer", zap.Error(err))
		return fmt.Errorf("failed to create order update Kafka consumer: %w", err)
	}
	s.consumer = consumer

	log.InfoZ(ctx, "OrderUpdateSubscriber started with Kafka")
	return nil
}

func (s *OrderUpdateSubscriber) Stop(ctx context.Context) error {
	if s.consumer != nil {
		s.consumer.Close(ctx)
	}
	return nil
}

func (s *OrderUpdateSubscriber) handleMessage(ctx context.Context, _ sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) error {
	var event kafka_help.OrderUpdateEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.ErrorZ(ctx, "OrderUpdateSubscriber: failed to unmarshal order update event",
			zap.Error(err),
			zap.Int32("partition", msg.Partition),
			zap.Int64("offset", msg.Offset),
			zap.ByteString("value", msg.Value))
		return nil // return nil to ack and skip malformed messages
	}

	// Broadcast to WS clients subscribed to this account
	key := fmt.Sprintf("order_%d", event.AccountID)
	res := &types.WsStream{
		Stream: types.WsStreamTypeOrder,
		Data:   event,
	}
	if err := s.wsServer.GetMelody().BroadcastFilter(res.ToByte(), func(session *melody.Session) bool {
		_, ok := session.Get(key)
		return ok
	}); err != nil {
		log.ErrorZ(ctx, "OrderUpdateSubscriber: failed to broadcast order update", zap.Error(err))
	}

	log.DebugZ(ctx, "OrderUpdateSubscriber: broadcasted order update to WS clients",
		zap.Uint64("accountId", event.AccountID),
		zap.Uint64("orderId", event.OrderID),
		zap.String("event", event.Event))

	return nil
}
