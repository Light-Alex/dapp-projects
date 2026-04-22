package rwa

import "time"

// EventClientRecord maps to public.event_client_record
type EventClientRecord struct {
	ChainID     uint64    `gorm:"column:chain_id;primaryKey" json:"chainId"`
	LastBlock   uint64    `gorm:"column:last_block" json:"lastBlock"`
	LastEventID uint64    `gorm:"column:last_event_id" json:"lastEventId"`
	UpdateAt    time.Time `gorm:"column:update_at" json:"updateAt"`
}

func (*EventClientRecord) TableName() string { return "event_client_record" }

