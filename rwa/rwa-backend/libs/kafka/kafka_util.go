package kafka

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/AnchoredLabs/rwa-backend/libs/log"

	"go.uber.org/zap"
)

var ErrFetchPartitions = errors.New("fetch partition total numbers")

type Util struct {
	ca    sarama.ClusterAdmin
	addrs []string
	cfg   *sarama.Config
}

func NewKafkaUtil(addrs []string, cfg *sarama.Config) (*Util, error) {
	if !cfg.Version.IsAtLeast(sarama.V2_0_0_0) {
		log.WarnZ(context.Background(), "at least kafka version 2.0")
		return nil, sarama.ErrUnsupportedVersion
	}
	ca, err := sarama.NewClusterAdmin(addrs, cfg)
	if err != nil {
		return nil, err
	}
	kafkaUtil := &Util{ca: ca, addrs: addrs, cfg: cfg}

	return kafkaUtil, nil
}

func (k *Util) getController() (*sarama.Broker, error) {
	client, err := sarama.NewClient(k.addrs, k.cfg)
	if err != nil {
		return nil, err
	}
	// make sure we can retrieve the controller
	controller, err := client.Controller()
	if err != nil {
		err := client.Close()
		if err != nil {
			return nil, err
		}
		return nil, err
	}
	return controller, nil
}

// GetDefaultReplicationFactor get kafka cluster default replicate factor
func (k *Util) GetDefaultReplicationFactor(ctx context.Context) (int16, error) {
	b, err := k.getController()
	if err != nil {
		return 0, err
	}
	defer func(b *sarama.Broker) {
		err := b.Close()
		if err != nil {
			log.ErrorZ(ctx, "close broker", zap.Error(err))
		}
	}(b)
	controllerID := b.ID()
	var defaultReplicateFactor int
	configs, err := k.ca.DescribeConfig(sarama.ConfigResource{
		Type:        sarama.BrokerResource,
		Name:        fmt.Sprintf("%d", controllerID),
		ConfigNames: []string{"default.replication.factor"}})
	if err != nil {
		return 0, err
	}
	if defaultReplicateFactor, err = strconv.Atoi(configs[0].Value); err != nil {
		return 0, err
	}
	log.InfoZ(ctx, "describe cluster info", zap.Int32("controller_id", controllerID), zap.Int("default_replicate_factor", defaultReplicateFactor))
	return int16(defaultReplicateFactor), nil
}

// GetNumPartitions get topic total partitions
func (k *Util) GetNumPartitions(topic string) (int32, error) {
	metadata, err := k.ca.DescribeTopics([]string{topic})
	if err != nil {
		return 0, err
	}
	if len(metadata) == 0 {
		return 0, ErrFetchPartitions
	}
	meta := metadata[0]
	if errors.Is(meta.Err, sarama.ErrUnknownTopicOrPartition) {
		return 0, ErrFetchPartitions
	}
	if !errors.Is(meta.Err, sarama.ErrNoError) {
		return 0, meta.Err
	}
	return int32(len(meta.Partitions)), nil
}

// CreateTopic manual create topics with number partitions
func (k *Util) CreateTopic(ctx context.Context, topic string, numPartitions int32) error {
	var oldPartitions int32
	var err error
	if oldPartitions, err = k.GetNumPartitions(topic); err != nil {
		if errors.Is(err, ErrFetchPartitions) {
			return k.createTopic(ctx, topic, numPartitions)
		}
		return err
	}
	log.InfoZ(ctx, "query topic old partitions", zap.String("kafka_topic", topic), zap.Int32("old_partitions", oldPartitions))
	if oldPartitions < numPartitions {
		log.InfoZ(ctx, "topic total partitions updated", zap.Int32("partitions", numPartitions), zap.String("kafka_topic", topic))
		return k.ca.CreatePartitions(topic, numPartitions, nil, false)
	} else {
		log.WarnZ(ctx, "query topic total partitions greater than ", zap.Int32("old", oldPartitions), zap.Int32("new", numPartitions))
		return nil
	}
}

func (k *Util) createTopic(ctx context.Context, topic string, numPartitions int32) error {
	replicationFactor, err := k.GetDefaultReplicationFactor(ctx)
	if err != nil {
		return err
	}
	if err := k.ca.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
		ReplicaAssignment: nil,
		ConfigEntries:     nil,
	}, false); err != nil {
		return err
	}
	log.InfoZ(ctx, "create topic ok", zap.String("kafka_topic", topic), zap.Int32("partitions", numPartitions), zap.Int16("replicationFactor", replicationFactor))
	return nil
}

func (k *Util) DeleteTopics(topics []string) error {
	var lastErr error
	for _, topic := range topics {
		lastErr = k.ca.DeleteTopic(topic)
	}
	return lastErr
}

func (k *Util) Close() {
	if k.ca != nil {
		err := k.ca.Close()
		if err != nil {
			log.ErrorZ(context.Background(), "close cluster admin failed", zap.Error(err))
		}
	}
}
