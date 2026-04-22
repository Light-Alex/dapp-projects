package rwa

import (
	"time"
)

// Account maps to public.account
type Account struct {
	ID        uint64    `gorm:"column:id;primaryKey" json:"id"`
	Address   string    `gorm:"column:address;uniqueIndex" json:"address"`
	IsActive  bool      `gorm:"column:is_active;default:true" json:"isActive"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (*Account) TableName() string { return "account" }
