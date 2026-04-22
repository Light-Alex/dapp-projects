package rwa

import (
	"time"
)

// StockStatus represents stock active/inactive status
type StockStatus string

const (
	StockStatusActive   StockStatus = "active"
	StockStatusInactive StockStatus = "inactive"
)

// Stock maps to public.stocks
type Stock struct {
	ID        uint64      `gorm:"column:id;primaryKey" json:"id"`
	Symbol    string      `gorm:"column:symbol;uniqueIndex" json:"symbol"`
	Name      string      `gorm:"column:name" json:"name"`
	Exchange  string      `gorm:"column:exchange" json:"exchange"`
	About     string      `gorm:"column:about;type:text" json:"about"`
	Status    StockStatus `gorm:"column:status;default:'active'" json:"status"`
	Contract  string      `gorm:"column:contract" json:"contract"`
	CreatedAt time.Time   `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time   `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (*Stock) TableName() string { return "stocks" }
