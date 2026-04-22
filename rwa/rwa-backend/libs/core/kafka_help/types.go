package kafka_help

import (
	"time"

	"github.com/IBM/sarama"
)

// DefaultProducerConfig returns a common sarama config for sync producers.
func DefaultProducerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 3
	config.Net.MaxOpenRequests = 5
	config.Net.KeepAlive = 60 * time.Second
	return config
}

type KafkaConfig struct {
	Brokers               []string `json:"brokers" yaml:"brokers"`
	Enabled               bool     `json:"enabled" yaml:"enabled"`
	OrderUpdatePartitions int32    `json:"orderUpdatePartitions" yaml:"orderUpdatePartitions"`
	BarPartitions         int32    `json:"barPartitions" yaml:"barPartitions"`
}

// GetOrderUpdatePartitions returns the configured partition count, defaulting to 3.
func (c *KafkaConfig) GetOrderUpdatePartitions() int32 {
	if c.OrderUpdatePartitions <= 0 {
		return 3
	}
	return c.OrderUpdatePartitions
}

// GetBarPartitions returns the configured bar partition count, defaulting to 1.
func (c *KafkaConfig) GetBarPartitions() int32 {
	if c.BarPartitions <= 0 {
		return 1
	}
	return c.BarPartitions
}
