package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var broker = []string{"kafka1:39094", "kafka2:39093", "kafka3:39092"}
var groupId = "test_group"
var topics = []string{"test_kafka", "test2_kafka"}

func TestKafkaPubSub(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, log.TraceID, uuid.New().String())
	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	var err error
	cfg.Version, err = sarama.ParseKafkaVersion("3.7.0")
	if err != nil {
		log.ErrorZ(ctx, "initConsumer, parse version", zap.Error(err))
	}
	assert.NoError(t, err)
	_, err = NewKafkaConsumerWithTopics(broker, cfg, groupId, topics, map[string]ReceiveMessageFunc{
		"test_kafka": func(ctx context.Context, session sarama.ConsumerGroupSession, message *sarama.ConsumerMessage) error {
			log.InfoZ(ctx, "receive kafka msg", zap.String("test_kafka", message.Topic), zap.Any("key", message.Key), zap.Any("value", message.Value))
			return nil
		},
		"test2_kafka": func(ctx context.Context, session sarama.ConsumerGroupSession, message *sarama.ConsumerMessage) error {
			log.InfoZ(ctx, "receive kafka msg", zap.String("test2_kafka", message.Topic), zap.Any("key", message.Key), zap.Any("value", message.Value))
			return nil
		},
	}, 10, true)
	if err != nil {
		log.ErrorZ(ctx, "initConsumer", zap.Error(err))
	}
	assert.NoError(t, err)

	// initialize producer
	cfg = sarama.NewConfig()
	producer, err := NewProducer(ctx, AllProducerType, broker, cfg, func(ctx context.Context, producerError *sarama.ProducerError) {
		log.ErrorZ(ctx, "kafka producer error", zap.Any("error", producerError))
	}, func(ctx context.Context, message *sarama.ProducerMessage) {
		log.InfoZ(ctx, "kafka producer success", zap.Any("message", message), zap.Any("key", message.Key), zap.Any("value", message.Value))
	})
	if err != nil {
		log.ErrorZ(ctx, "initProducer", zap.Error(err))
	}
	assert.NoError(t, err)
	for i := 0; i < 100; i++ {
		ctx = context.WithValue(ctx, log.TraceID, uuid.New().String())
		partition, offset, err := producer.SendMessage(ctx, &sarama.ProducerMessage{
			Topic: topics[i%2],
			Key:   sarama.StringEncoder(uuid.NewString()),
			Value: sarama.ByteEncoder(uuid.NewString()),
		})
		if err != nil {
			log.ErrorZ(ctx, "send sync message", zap.Error(err))
		}
		assert.NoError(t, err)
		log.InfoZ(ctx, "send sync message success", zap.Any("partition", partition), zap.Any("offset", offset))
	}
	time.Sleep(time.Duration(5) * time.Second)
}
