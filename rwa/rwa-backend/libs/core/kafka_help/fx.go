package kafka_help

import (
	"context"

	"go.uber.org/fx"
)

// LoadModule loads all Kafka producers (for services that produce messages).
func LoadModule(config *KafkaConfig) fx.Option {
	return fx.Module("kafka", fx.Supply(config), fx.Provide(
		NewSnapshotKafkaService,
		NewOrderUpdateKafkaService,
		NewBarKafkaService,
	), fx.Invoke(registerProducerShutdown))
}

// LoadConsumerModule only supplies KafkaConfig (for services that only consume).
func LoadConsumerModule(config *KafkaConfig) fx.Option {
	return fx.Module("kafka", fx.Supply(config))
}

func registerProducerShutdown(lc fx.Lifecycle, snapshot *SnapshotKafkaService, orderUpdate *OrderUpdateKafkaService, bar *BarKafkaService) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			snapshot.Close(ctx)
			orderUpdate.Close(ctx)
			bar.Close(ctx)
			return nil
		},
	})
}
