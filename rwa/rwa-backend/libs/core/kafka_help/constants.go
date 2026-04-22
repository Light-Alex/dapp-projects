package kafka_help

// topics
const (
	RWA_EVENT_TOPIC        = "rwa.poc.event."      // poc event topic
	RWA_ORDER_UPDATE_TOPIC = "rwa.order.update"     // order update topic
	RWA_MARKET_BAR_TOPIC   = "rwa.market.bar"       // market bar data topic
)

// consumer groups
const (
	RWA_INDEXER_CONSUMER_GROUP      = "rwa_indexer_consumer_group_"      // poc indexer consumer group
	RWA_ORDER_UPDATE_CONSUMER_GROUP = "rwa_order_update_consumer_group"  // order update consumer group
	RWA_MARKET_BAR_CONSUMER_GROUP   = "rwa_market_bar_consumer_group"    // market bar consumer group
)
