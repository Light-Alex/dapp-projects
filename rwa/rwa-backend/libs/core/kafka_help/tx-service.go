package kafka_help

// Transaction status enum
type TxStatus string

const (
	TxStatusPending   TxStatus = "pending"
	TxStatusConfirmed TxStatus = "confirmed"
	TxStatusFailed    TxStatus = "failed"
)

// Source enum for tracking
type TrackingSource string

const (
	SourceKafkaEvent TrackingSource = "kafka_event"
	SourceChainQuery TrackingSource = "chain_query"
)

// TxStatusResult represents the result of a transaction status update
type TxStatusResult struct {
	TxHash      string                `json:"txHash"`
	RequestID   string                `json:"requestId"` // Added RequestID for tracking
	ChainID     uint64                `json:"chain_id"`
	Status      TxStatus              `json:"status"` // TxStatusConfirmed or TxStatusFailed
	BlockNumber uint64                `json:"blockNumber"`
	Event       *SnapKafkaTxWithEvent `json:"events,omitempty"`  // Event returned on success
	Receipt     string                `json:"receipt,omitempty"` // Receipt returned on failure
	Timestamp   int64                 `json:"timestamp"`         // Timestamp in milliseconds
	BlockTime   int64                 `json:"blockTime"`         // Block timestamp in milliseconds
	Source      TrackingSource        `json:"source"`            // SourceKafkaEvent or SourceChainQuery
}
