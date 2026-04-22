package rwa

import "time"

// FailedEvent stores WebSocket trade update events that failed to process.
// These events are persisted for manual recovery or future retry mechanisms.
type FailedEvent struct {
	ID             uint64    `gorm:"column:id;primaryKey" json:"id"`
	ClientOrderID  string    `gorm:"column:client_order_id;index" json:"clientOrderId"`
	EventType      string    `gorm:"column:event_type" json:"eventType"`
	ExecutionID    string    `gorm:"column:execution_id" json:"executionId"`
	EventData      string    `gorm:"column:event_data;type:text" json:"eventData"`
	ErrorMessage   string    `gorm:"column:error_message;type:text" json:"errorMessage"`
	Source         string    `gorm:"column:source;default:'alpaca'" json:"source"`
	RetryCount     int       `gorm:"column:retry_count;default:0" json:"retryCount"`
	Status         string    `gorm:"column:status;default:'pending'" json:"status"`
	ResolvedAt     *time.Time `gorm:"column:resolved_at" json:"resolvedAt"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (*FailedEvent) TableName() string { return "failed_events" }
