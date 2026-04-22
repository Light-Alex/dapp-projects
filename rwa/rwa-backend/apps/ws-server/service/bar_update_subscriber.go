package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

type BarUpdateSubscriber struct {
	kafkaConfig *kafka_help.KafkaConfig
	wsServer    *ws.Server
	consumer    *kafka.Consumer
}

func NewBarUpdateSubscriber(
	lc fx.Lifecycle,
	kafkaConfig *kafka_help.KafkaConfig,
	wsServer *ws.Server,
) *BarUpdateSubscriber {
	s := &BarUpdateSubscriber{
		kafkaConfig: kafkaConfig,
		wsServer:    wsServer,
	}
	lc.Append(fx.Hook{
		OnStart: s.Start,
		OnStop:  s.Stop,
	})
	return s
}

func (s *BarUpdateSubscriber) Start(ctx context.Context) error {
	if !s.kafkaConfig.Enabled {
		log.WarnZ(ctx, "BarUpdateSubscriber: Kafka not enabled, skipping")
		return nil
	}

	consumer, err := kafka.NewKafkaConsumer(
		s.kafkaConfig.Brokers,
		nil,
		kafka_help.RWA_MARKET_BAR_CONSUMER_GROUP,
		kafka_help.RWA_MARKET_BAR_TOPIC,
		s.handleMessage,
		s.kafkaConfig.GetBarPartitions(),
		true, // auto-ack
	)
	if err != nil {
		log.ErrorZ(ctx, "BarUpdateSubscriber: failed to create Kafka consumer", zap.Error(err))
		return fmt.Errorf("failed to create bar update Kafka consumer: %w", err)
	}
	s.consumer = consumer

	log.InfoZ(ctx, "BarUpdateSubscriber started with Kafka")
	return nil
}

func (s *BarUpdateSubscriber) Stop(ctx context.Context) error {
	if s.consumer != nil {
		s.consumer.Close(ctx)
	}
	return nil
}

func (s *BarUpdateSubscriber) handleMessage(ctx context.Context, _ sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) error {
	var event kafka_help.BarEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.ErrorZ(ctx, "BarUpdateSubscriber: failed to unmarshal bar event",
			zap.Error(err),
			zap.Int32("partition", msg.Partition),
			zap.Int64("offset", msg.Offset),
			zap.ByteString("value", msg.Value))
		return nil
	}

	key := fmt.Sprintf("bar_%s", strings.ToUpper(event.Symbol))
	res := &types.WsStream{
		Stream: types.WsStreamTypeBar,
		Data:   event,
	}
	if err := s.wsServer.GetMelody().BroadcastFilter(res.ToByte(), func(session *melody.Session) bool {
		_, ok := session.Get(key)
		return ok
	}); err != nil {
		log.ErrorZ(ctx, "BarUpdateSubscriber: failed to broadcast bar update", zap.Error(err))
	}

	log.DebugZ(ctx, "BarUpdateSubscriber: broadcasted bar update to WS clients",
		zap.String("symbol", event.Symbol),
		zap.Float64("close", event.Close))

	return nil
}
