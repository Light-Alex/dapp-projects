package redis_cache

const ChannelPrefix = "SynFutures-Channel#"

// Key prefixes for different cache types
const (
	OrderBookKeyPrefix           = "orderBook:" // Prefix for order book keys
	InstrumentKeyPrefix          = "instrument:"
	KlineKeyPrefix               = "kline:current" // Prefix for kline keys
	TokenPriceBySymbolKeyPreFix  = "token:price:bySymbol"
	TokenPriceByAddressKeyPreFix = "token:price:byAddress:"
	DepthKeyPrefix      = "perp:pair:depth:"
)

// Redis channel constants
const (
	PubSubTopicSnapshotChanged = "pubsub:event:snapshot:snapshotChanged"
	PubSubTopicKlineUpdated    = "pubsub:event:indexer:klineUpdated"
)
