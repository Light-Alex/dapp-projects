package kafka

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"go.uber.org/zap"
)

var ErrBrokerAddrEmpty = errors.New("broker addrs empty")

type delayTaskResult int8

const (
	addDelayTaskOk delayTaskResult = iota
	addDelayTaskFailed
	addDelayTaskAbort
)

// Topic type define
type Topic = string

// ReceiveMessageFunc message callback function
// error is the error returned. Kafka will re-push this message to the consumer until it returns nil.
// This means if message processing fails, the partition will be blocked.
// If processing fails and you don't want to block other successfully processable messages, use KafkaRetry together.
// With retry enabled, returning !=nil will automatically retry.
// With auto-ack enabled, returning nil will automatically acknowledge. Note: this auto-ack is not Kafka's auto-ack mechanism.
// Note: the callback function is prohibited from starting another goroutine for separate processing unless you know what you're doing.
type ReceiveMessageFunc func(context.Context, sarama.ConsumerGroupSession, *sarama.ConsumerMessage) error

// WatchSetupFunc watch setup func
type WatchSetupFunc func(session sarama.ConsumerGroupSession)

// Consumer kafka consumer
type Consumer struct {
	client                 sarama.ConsumerGroup
	topics                 []string
	messageCallbackMap     map[string]ReceiveMessageFunc
	closedChan             chan struct{}
	autoAck                bool
	brokerAddr             string
	partitions             map[string][]int32
	gracefullyCloseTimeout time.Duration
	watchSetup             WatchSetupFunc
}

// NewKafkaConsumer usage notes please refer to NewKafkaConsumerWithTopics
func NewKafkaConsumer(broker []string, cfg *sarama.Config, groupID string,
	topic string, callback ReceiveMessageFunc, numPartitions int32,
	autoAck bool,
) (*Consumer, error) {
	return NewKafkaConsumerWithTopics(
		broker, cfg, groupID,
		[]string{topic},
		map[string]ReceiveMessageFunc{topic: callback},
		numPartitions, autoAck)
}

// NewKafkaConsumerWithGracefullyClose new kafka consumer with gracefully close
func NewKafkaConsumerWithGracefullyClose(broker []string, cfg *sarama.Config, groupID string, topics []string,
	callbackMap map[string]ReceiveMessageFunc, numPartitions int32,
	autoAck bool, gracefullyCloseTimeout time.Duration) (*Consumer, error) {
	return NewConsumerWithWatchSetup(broker, cfg, groupID, topics, callbackMap, numPartitions, autoAck, gracefullyCloseTimeout, nil)
}

// NewKafkaConsumerWithTopics new kafka consumer with topics
// broker kafka broker
// cfg kafka config. If nil, defaults to sarama.NewConfig(). Users must understand kafka parameters
// groupID consumer group ID. Same ID means consumers are in the same group, only one consumer in a group can consume a partition
// topics supports consuming messages from multiple topics
// callbackMap callback functions corresponding to topics
// numPartitions number of partitions for the topic. Note: if specified less than existing, it won't delete original partitions.
// Deleting is strongly recommended to be handled manually by ops
// autoAck auto-acknowledge mechanism. Auto-ack means automatically calling MarkMessage when message processing returns normally
// retry auto-retry mechanism. Automatically retries when message processing fails
// Relationship between autoAck and retry:
// autoAck false disables auto mechanism, users must call MarkMessage to acknowledge, even if processing is correct. Has no relation to retry. retry only works when autoAck is true
// TODO recommend setting autoAck to true, whether retry is needed depends on use case
func NewKafkaConsumerWithTopics(broker []string, cfg *sarama.Config, groupID string, topics []string,
	callbackMap map[string]ReceiveMessageFunc, numPartitions int32,
	autoAck bool) (*Consumer, error) {
	return NewKafkaConsumerWithGracefullyClose(broker, cfg, groupID, topics, callbackMap, numPartitions, autoAck, time.Second*10)
}

// NewConsumerWithWatchSetup new kafka consumer with topics
func NewConsumerWithWatchSetup(broker []string, cfg *sarama.Config, groupID string, topics []string,
	callbackMap map[string]ReceiveMessageFunc, numPartitions int32,
	autoAck bool, gracefullyCloseTimeout time.Duration, watchSetup WatchSetupFunc,
) (*Consumer, error) {
	if len(broker) == 0 {
		return nil, ErrBrokerAddrEmpty
	}
	if cfg == nil {
		cfg = sarama.NewConfig()
		cfg.Net.MaxOpenRequests = 255

		// for financial business scenarios, don't miss any messages
		cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

		cfg.Consumer.MaxWaitTime = 500 * time.Millisecond
		cfg.Consumer.MaxProcessingTime = 2000 * time.Millisecond
		cfg.Consumer.Group.Session.Timeout = 15 * time.Second
		cfg.Consumer.Group.Heartbeat.Interval = 5 * time.Second
		cfg.Version, _ = sarama.ParseKafkaVersion("2.2.0")
	}
	cfg.Consumer.Return.Errors = true
	if numPartitions > 0 {
		kafkaUtil, err := NewKafkaUtil(broker, cfg)
		if err != nil {
			return nil, err
		}
		defer kafkaUtil.Close()
		for _, topic := range topics {
			if interErr := kafkaUtil.CreateTopic(context.Background(), topic, numPartitions); interErr != nil {
				log.ErrorZ(context.Background(), "create topic partition failed", zap.String("kafka_topic", topic), zap.Error(err))
				err = interErr
			}
		}
		if err != nil {
			return nil, err
		}
	}
	k := &Consumer{
		topics:                 topics,
		messageCallbackMap:     callbackMap,
		autoAck:                autoAck,
		brokerAddr:             strings.Join(broker, ","),
		gracefullyCloseTimeout: gracefullyCloseTimeout,
		watchSetup:             watchSetup,
	}
	var err error
	k.client, err = sarama.NewConsumerGroup(broker, groupID, cfg)
	if err != nil {
		log.ErrorZ(context.Background(), "error creating consumer group client", zap.Error(err), zap.Strings("broker", broker), zap.Strings("topics", topics))
		return nil, err
	}
	k.run()
	return k, nil
}

// run loop consumer topics message
func (k *Consumer) run() {
	go func() {
		for {
			ctx := context.Background()
			err := k.client.Consume(ctx, k.topics, k)
			if errors.Is(err, sarama.ErrClosedConsumerGroup) {
				log.InfoZ(ctx, "consume normal exited", zap.Strings("topics", k.topics))
				return
			} else if err != nil {
				log.ErrorZ(ctx, "error from consumer ", zap.Error(err), zap.Strings("topics", k.topics))
			}
			log.InfoZ(ctx, "consume retry", zap.Strings("topics", k.topics))
		}
	}()

	go func() {
		for {
			select {
			case err, ok := <-k.client.Errors():
				if !ok {
					log.InfoZ(context.Background(), "consume kafka err chan closed")
					return
				}
				log.ErrorZ(context.Background(), "consume kafka message error", zap.Error(err))
			}
		}
	}()
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (k *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	defer func() {
		if err := recover(); err != nil {
			log.ErrorZ(context.Background(), "setup unknown panic", zap.Stack("stack"), zap.Reflect("error", err))
		}
	}()
	k.closedChan = make(chan struct{})
	k.partitions = session.Claims()
	log.InfoZ(context.Background(), "consumer setup ok", zap.String("id", session.MemberID()), zap.Reflect("partitions", k.partitions))
	if k.watchSetup != nil {
		k.watchSetup(session)
	}
	return nil
}

// GetConsumedPartitionType type define
type GetConsumedPartitionType = int64

const (
	// MaxConsumedPartition get maximum partition number of the consumed topic
	MaxConsumedPartition GetConsumedPartitionType = iota
	// MinConsumedPartition get the minimum partition number of the consumed topic
	MinConsumedPartition
)

func _max(slice []int32) int32 {
	if len(slice) == 0 {
		return 0
	}

	m := slice[0]
	for _, i := range slice[1:] {
		if i <= m {
			continue
		}
		m = i
	}
	return m
}

func _min(slice []int32) int32 {
	if len(slice) == 0 {
		return 0
	}

	m := slice[0]
	for _, i := range slice[1:] {
		if i >= m {
			continue
		}
		m = i
	}
	return m
}

// GetConsumedPartition  get one partition number of the consumed topic, default 0
func (k *Consumer) GetConsumedPartition(topic string, which int64) int32 {
	partitions := k.partitions[topic]
	if len(partitions) == 0 {
		return 0
	}
	switch which {
	case MaxConsumedPartition:
		return _max(partitions)
	case MinConsumedPartition:
		return _min(partitions)
	default:
		return partitions[0]
	}
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (k *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	close(k.closedChan)
	log.InfoZ(context.Background(), "consumer will exited or rebanlance", zap.Reflect("id", session.MemberID()))
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (k *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil // channel closed, exit
			}
			traceLog := GetTraceID(message.Headers)
			// TODO task timeout is not using kafka's max task timeout config. Reason: if developer sets inappropriate timeout, it easily causes continuous retries, adding complexity.
			// For kafka tasks, optimization can be done from monitoring system for tasks with long processing times
			ctx := context.WithValue(context.Background(), log.TraceID, traceLog)
			// Note: the callback function is prohibited from starting another goroutine for separate processing unless you know what you're doing.
			if callback, ok := k.messageCallbackMap[message.Topic]; ok {
				if err := k.handleMessageByTakeTime(ctx, session, message, callback); err != nil {
					return err
				}
			} else {
				log.ErrorZ(ctx, "kafka message not find topic to callback function", zap.String("kafka_topic", message.Topic),
					zap.Int32("partition", message.Partition), zap.ByteString("key", message.Key))
				session.MarkMessage(message, "")
			}
		case <-session.Context().Done():
			log.InfoZ(context.Background(), "session context done, will exit consume claim", zap.Reflect("id", session.MemberID()))
			return nil
		}
	}
}

func (k *Consumer) handleMessageByTakeTime(ctx context.Context, session sarama.ConsumerGroupSession, message *sarama.ConsumerMessage, callback ReceiveMessageFunc) error {
	defer func() {
		if e := recover(); e != nil {
			log.ErrorZ(ctx, "handleMessageByTakeTime unknown panic error", zap.String("kafka_topic", message.Topic),
				zap.Int32("partition", message.Partition),
				zap.ByteString("key", message.Key),
				zap.ByteString("kafka_message", message.Value),
				zap.Duration("message_in_kafka_time", time.Since(message.Timestamp)),
				zap.String("kafka_message_time", message.Timestamp.Format(time.RFC3339Nano)), zap.Stack("stack"), zap.Reflect("error", e))
		}
	}()
	startTime := time.Now()
	if err := callback(ctx, session, message); err != nil {
		if innerErr := k.handleError(ctx, session, message, err); innerErr != nil {
			log.ErrorZ(ctx, "receive kafka msg", zap.Reflect("kafka_topic", message.Topic), zap.Int32("partition", message.Partition),
				zap.ByteString("key", message.Key),
				zap.ByteString("kafka_message", message.Value),
				zap.Error(innerErr),
				zap.Duration("take_time", time.Since(startTime)),
				zap.Duration("message_in_kafka_time", time.Since(message.Timestamp)),
				zap.String("kafka_message_time", message.Timestamp.Format(time.RFC3339Nano)))
			return innerErr
		}
	} else {
		if k.autoAck {
			session.MarkMessage(message, "")
		}
		log.InfoZ(ctx, "receive kafka msg", zap.String("kafka_topic", message.Topic),
			zap.Int32("partition", message.Partition),
			zap.ByteString("key", message.Key),
			zap.ByteString("kafka_message", message.Value),
			zap.Duration("handle_message_time", time.Since(startTime)),
			zap.Duration("message_in_kafka_time", time.Since(message.Timestamp)),
			zap.String("kafka_message_time", message.Timestamp.Format(time.RFC3339Nano)))
	}
	return nil
}

func (k *Consumer) handleError(ctx context.Context, _ sarama.ConsumerGroupSession, message *sarama.ConsumerMessage, err error) error {
	log.ErrorZ(ctx, "receive kafka msg failed", zap.Any("kafka_topic", message.Topic), zap.Int32("partition", message.Partition), zap.Any("key", message.Key))
	return err
}

func (k *Consumer) messageHeaderToMap(message *sarama.ConsumerMessage) map[string]string {
	header := make(map[string]string)
	for _, recordHeader := range message.Headers {
		key := string(recordHeader.Key)
		value := string(recordHeader.Value)
		header[key] = value
	}
	return header
}

// Close  kafka client
func (k *Consumer) Close(ctx context.Context) {
	log.InfoZ(ctx, "kafka consumer start close")
	if err := k.client.Close(); err != nil {
		log.ErrorZ(ctx, "close kafka consumer", zap.Error(err))
	}
	// wait task exec finished
	select {
	case <-k.closedChan:
		log.InfoZ(ctx, "kafka consumer end closed ")
		return
	case <-time.After(k.gracefullyCloseTimeout):
		log.WarnZ(ctx, "kafka consumer close timeout")
		return
	}
}
