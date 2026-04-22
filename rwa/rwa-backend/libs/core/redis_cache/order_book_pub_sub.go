package redis_cache

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/redis/go-redis/v9"
)

// SnapshotEventType represents the type of snapshot event
type SnapshotEventType string

const (
	SnapshotEventTypeInstrument SnapshotEventType = "instrument"
	SnapshotEventTypeOrder      SnapshotEventType = "order"
	SnapshotEventTypeOrderBook  SnapshotEventType = "orderBook"
	SnapshotEventTypeRange      SnapshotEventType = "range"
	SnapshotEventTypePosition   SnapshotEventType = "position"
	SnapshotEventTypeGate       SnapshotEventType = "gate"
)

// SnapshotEvent represents a snapshot change event
type SnapshotEvent struct {
	ChainId     uint64            `json:"chainId"` // Non-pointer, use 0 for empty
	Type        SnapshotEventType `json:"type"`
	Instrument  string            `json:"instrument"`
	Expiry      uint32            `json:"expiry"`
	UserAddress string            `json:"userAddress"` // Non-pointer, use empty string for empty
	// Additional fields for Go version (not sent to Redis to maintain compatibility)
	BlockNum uint64         `json:"-"`
	EventId  uint64         `json:"-"`
	Address  common.Address `json:"-"`
}

type OrderBookPubSub struct {
	client redis.UniversalClient
}

func NewOrderBookPubSub(rdb redis.UniversalClient) *OrderBookPubSub {
	return &OrderBookPubSub{
		client: rdb,
	}
}
