package rwa

import (
	"time"
)

// EventLog maps to public.event_logs
type EventLog struct {
	ID              uint64     `gorm:"column:id;primaryKey" json:"id"`
	TxHash          string     `gorm:"column:tx_hash;index" json:"txHash"`
	BlockNumber     int64      `gorm:"column:block_number;index" json:"blockNumber"`
	BlockHash       string     `gorm:"column:block_hash" json:"blockHash"`
	LogIndex        int        `gorm:"column:log_index" json:"logIndex"`
	ContractAddress string     `gorm:"column:contract_address;index" json:"contractAddress"`
	EventName       string     `gorm:"column:event_name" json:"eventName"`
	EventData       string     `gorm:"column:event_data;type:text" json:"eventData"`
	AccountID       uint64     `gorm:"column:account_id;index" json:"accountId"`
	OrderID         uint64     `gorm:"column:order_id;index" json:"orderId"`
	ProcessedAt     *time.Time `gorm:"column:processed_at" json:"processedAt"`
	Status          string     `gorm:"column:status;default:'pending'" json:"status"`
	ErrorMessage    string     `gorm:"column:error_message;type:text" json:"errorMessage"`
	RetryCount      int        `gorm:"column:retry_count;default:0" json:"retryCount"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (*EventLog) TableName() string { return "event_logs" }
