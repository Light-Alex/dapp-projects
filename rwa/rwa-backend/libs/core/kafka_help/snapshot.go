package kafka_help

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	coreTypes "github.com/AnchoredLabs/rwa-backend/libs/core/types"
	"github.com/AnchoredLabs/rwa-backend/libs/kafka"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type SnapKafkaTxWithEvent struct {
	TxHash         string                      `json:"txHash"`
	TxIndex        uint64                      `json:"txIndex"`
	BlockNum       uint64                      `json:"blockNum"`
	BlockTime      uint64                      `json:"blockTime"`
	RecentBlockNum uint64                      `json:"recentBlockNum"`
	Events         []*coreTypes.EventLogWithId `json:"events"`
	Synced         bool                        `json:"synced"`
}

func (i *SnapKafkaTxWithEvent) ToBytes() []byte {
	marshal, err := json.Marshal(i)
	if err != nil {
		log.ErrorZ(context.Background(), "SnapKafkaTxWithEvent.ToBytes: marshal failed", zap.Error(err))
		return nil
	}
	return marshal
}

type SnapshotKafkaService struct {
	producer    *kafka.Producer
	kafkaConfig *KafkaConfig
}

func NewSnapshotKafkaService(kafkaConfig *KafkaConfig) (*SnapshotKafkaService, error) {
	if !kafkaConfig.Enabled {
		return &SnapshotKafkaService{kafkaConfig: kafkaConfig}, nil
	}

	ctx := context.WithValue(context.Background(), log.TraceID, "SnapshotKafkaServiceInit")
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal  // performance priority: don't wait for all replicas to confirm
	config.Producer.Compression = sarama.CompressionLZ4 // compression improves transmission speed
	config.Producer.Flush.Bytes = 1024 * 512            // 512KB per batch
	config.Producer.Flush.Messages = 1000               // 1000 messages per batch
	config.Producer.Flush.Frequency = 10 * time.Millisecond
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 3
	config.Net.MaxOpenRequests = 10
	config.Net.KeepAlive = 60 * time.Second
	config.Net.WriteTimeout = 30 * time.Second
	config.Net.ReadTimeout = 30 * time.Second
	producer, err := kafka.NewProducer(ctx, kafka.SyncProducerType, kafkaConfig.Brokers, config, func(ctx context.Context, producerError *sarama.ProducerError) {
		log.ErrorZ(ctx, "failed to send snapshot event msg to kafka", zap.Error(producerError.Err), zap.Any("key", producerError.Msg.Key))
	}, func(ctx context.Context, message *sarama.ProducerMessage) {
		log.DebugZ(ctx, "success to send snapshot event msg to kafka", zap.Any("key", message.Key))
	})
	if err != nil {
		log.ErrorZ(ctx, "failed to create kafka producer", zap.Any("brokers", kafkaConfig), zap.Error(err))
		return nil, err
	}
	return &SnapshotKafkaService{
		producer:    producer,
		kafkaConfig: kafkaConfig,
	}, nil
}

// Close closes the underlying Kafka producer.
func (s *SnapshotKafkaService) Close(ctx context.Context) {
	if s.producer != nil {
		s.producer.Close(ctx)
	}
}

// PublishEvent publish event to kafka
func (s *SnapshotKafkaService) PublishEvent(ctx context.Context, eventList []*coreTypes.EventLogWithId, recentBlockNum uint64, synced bool, chainId uint64) error {
	topic := RWA_EVENT_TOPIC + strconv.FormatUint(chainId, 10)
	if !s.kafkaConfig.Enabled {
		return nil
	}
	if len(eventList) == 0 {
		return nil
	}
	list := make([]*SnapKafkaTxWithEvent, 0)
	for txHash, el := range lo.GroupBy(eventList, func(item *coreTypes.EventLogWithId) string {
		return item.TxHash
	}) {
		i := &SnapKafkaTxWithEvent{
			TxHash:         txHash,
			TxIndex:        hexutil.MustDecodeUint64(el[0].TxIndex),
			BlockNum:       hexutil.MustDecodeUint64(el[0].BlockNumber),
			BlockTime:      hexutil.MustDecodeUint64(el[0].BlockTimestamp),
			Events:         el,
			RecentBlockNum: recentBlockNum,
			Synced:         synced,
		}
		//events shot sort by id asc
		sort.Slice(i.Events, func(a, b int) bool {
			return i.Events[a].EventId < i.Events[b].EventId
		})
		list = append(list, i)
	}
	//tx hash sort by tx index asc
	sort.Slice(list, func(a, b int) bool {
		if list[a].BlockNum == list[b].BlockNum {
			return list[a].TxIndex < list[b].TxIndex
		}
		return list[a].BlockNum < list[b].BlockNum
	})
	msgList := lo.Map(list, func(item *SnapKafkaTxWithEvent, index int) *sarama.ProducerMessage {
		return &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(item.TxHash),
			Value: sarama.ByteEncoder(item.ToBytes()),
		}
	})
	var txHashList []string
	txHashMap := make(map[string]bool)
	for _, item := range eventList {
		if _, ok := txHashMap[item.TxHash]; !ok {
			txHashList = append(txHashList, item.TxHash)
			txHashMap[item.TxHash] = true
		}
	}
	startTime := time.Now()
	err := s.producer.SendMessages(ctx, msgList)
	if err != nil {
		log.ErrorZ(ctx, "failed to send event", zap.Error(err))
		return err
	}
	log.InfoZ(ctx, "success to publish event to kafka", zap.Int("event_count", len(eventList)), zap.Int("tx_hash_count", len(txHashList)), zap.Int("cost_time", int(time.Since(startTime).Milliseconds())))
	return nil
}

