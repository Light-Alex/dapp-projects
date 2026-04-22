package kafka

import (
	"context"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/AnchoredLabs/rwa-backend/libs/log"

	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var ErrProducerType = errors.New("producer async or sync type invalid")

type PartitionFunc func(message *sarama.ProducerMessage, numPartitions int32) (int32, error)
type ErrorsFunc func(context.Context, *sarama.ProducerError)
type SuccessFunc func(context.Context, *sarama.ProducerMessage)

type ProducerType uint8

const (
	SyncProducerType  = 1
	ASyncProducerType = 2
	AllProducerType   = 3
)

// Producer send message to kafka broker
type Producer struct {
	syncProducer    sarama.SyncProducer
	asyncProducer   sarama.AsyncProducer
	errCallback     ErrorsFunc
	successCallback SuccessFunc
}

// NewProducer supports initializing sync and async message producers simultaneously
// producerType producer type: sync, async, or both
// addrs kafka broker addresses
// cfg kafka config. If nil, defaults to sarama.NewConfig(). Users must understand kafka parameters
// errCallback error callback for async sending
// successCallback success callback for async sending
func NewProducer(ctx context.Context, producerType ProducerType, addrs []string, cfg *sarama.Config,
	errCallback ErrorsFunc, successCallback SuccessFunc) (*Producer, error) {
	k := &Producer{errCallback: errCallback, successCallback: successCallback}
	var err error
	if cfg == nil {
		cfg = sarama.NewConfig()
	}
	// waitForAll ensures no message loss when sending to kafka
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	if producerType&SyncProducerType == SyncProducerType {
		cfg.Producer.Return.Successes = true
		if k.syncProducer, err = sarama.NewSyncProducer(addrs, cfg); err != nil {
			k.Close(ctx)
			return nil, err
		}
	}
	if producerType&ASyncProducerType == ASyncProducerType {
		if successCallback != nil {
			cfg.Producer.Return.Successes = true
		} else {
			cfg.Producer.Return.Successes = false
		}
		// error information, force enable
		cfg.Producer.Return.Errors = true
		if errCallback == nil {
			k.Close(ctx)
			return nil, fmt.Errorf("async producer, error callback must not nil")
		}
		if k.asyncProducer, err = sarama.NewAsyncProducer(addrs, cfg); err != nil {
			k.Close(ctx)
			return nil, err
		}
		k.receiveAsyncSendMsgResult(ctx, 0)
	}
	return k, nil
}

// SendMessage synchronously sends data to kafka
func (k *Producer) SendMessage(ctx context.Context, msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	if k.syncProducer == nil {
		return 0, 0, ErrProducerType
	}
	k.addTraceIDToMsg(ctx, msg)
	msg.Timestamp = time.Now()
	return k.syncProducer.SendMessage(msg)
}

// SendMessages synchronously sends multiple messages to kafka
func (k *Producer) SendMessages(ctx context.Context, msgs []*sarama.ProducerMessage) (err error) {
	if k.syncProducer == nil {
		return ErrProducerType
	}
	nowTime := time.Now()
	for _, msg := range msgs {
		msg.Timestamp = nowTime
		k.addTraceIDToMsg(ctx, msg)
	}
	return k.syncProducer.SendMessages(msgs)
}

// AsyncSendMessage asynchronously sends data to kafka
func (k *Producer) AsyncSendMessage(ctx context.Context, msg *sarama.ProducerMessage) error {
	if k.asyncProducer == nil {
		return ErrProducerType
	}
	msg.Timestamp = time.Now()
	k.addTraceIDToMsg(ctx, msg)
	k.asyncProducer.Input() <- msg
	return nil
}

func (k *Producer) addTraceIDToMsg(ctx context.Context, msg *sarama.ProducerMessage) {
	for ix, header := range msg.Headers {
		if string(header.Key) != string(log.TraceID) {
			continue
		}
		if len(header.Value) == 0 {
			msg.Headers[ix].Value = []byte(uuid.NewString())
		}
		return
	}
	traceID, _ := ctx.Value(log.TraceID).(string)
	if traceID == "" {
		traceID = uuid.NewString()
	}
	msg.Headers = append(msg.Headers, sarama.RecordHeader{Key: []byte(log.TraceID), Value: []byte(traceID)})
}

// receiveAsyncSendMsgResult receives async send message results from kafka
func (k *Producer) receiveAsyncSendMsgResult(ctx context.Context, retry int) {
	// retry up to 3 times, then consider it failed
	if retry >= 3 {
		panic("receive async send msg result unknown error")
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.ErrorZ(ctx, "handle error message unknown error", zap.Reflect("error", err))
				k.receiveAsyncSendMsgResult(ctx, retry+1)
			} else {
				log.InfoZ(ctx, "receive async send msg result exited")
			}
		}()
		log.InfoZ(ctx, "begin receive async send msg result")
		for {
			select {
			case err, ok := <-k.asyncProducer.Errors():
				if !ok {
					return
				}
				if k.errCallback != nil {
					traceLog := GetTraceIDByProducer(err.Msg.Headers)
					k.errCallback(context.WithValue(context.Background(), log.TraceID, traceLog), err)
				} else {
					log.ErrorZ(ctx, "async send message to kafka failed", zap.Error(err.Err))
				}
				retry = 0
			case success, ok := <-k.asyncProducer.Successes():
				if !ok {
					return
				}
				if k.successCallback != nil {
					traceLog := GetTraceIDByProducer(success.Headers)
					k.successCallback(context.WithValue(context.Background(), log.TraceID, traceLog), success)
				}
				retry = 0
			}
		}
	}()
}

// Close  sync producer  and async producer
func (k *Producer) Close(ctx context.Context) {
	log.InfoZ(ctx, "kafka producer closed")
	if k.asyncProducer != nil {
		if err := k.asyncProducer.Close(); err != nil {
			log.ErrorZ(ctx, "close async producer failed")
		}
	}
	if k.syncProducer != nil {
		if err := k.syncProducer.Close(); err != nil {
			log.ErrorZ(ctx, "close sync producer failed")
		}
	}
}
