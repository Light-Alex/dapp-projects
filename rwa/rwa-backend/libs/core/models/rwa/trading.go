package rwa

import (
	"time"

	"github.com/shopspring/decimal"
)

// TradingAccount maps to public.trading_accounts
type TradingAccount struct {
	ID                string    `gorm:"column:id;primaryKey" json:"id"`
	ExternalAccountID string    `gorm:"column:external_account_id" json:"externalAccountId"`
	Provider          string    `gorm:"column:provider" json:"provider"`
	AccountType       string    `gorm:"column:account_type" json:"accountType"`
	Status            string    `gorm:"column:status" json:"status"`
	IsActive          bool      `gorm:"column:is_active" json:"isActive"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

// Position maps to public.positions
type Position struct {
	ID            uint64          `gorm:"column:id;primaryKey" json:"id"`
	AccountID     uint64          `gorm:"column:account_id;index" json:"accountId"`
	Symbol        string          `gorm:"column:symbol;index" json:"symbol"`
	AssetType     AssetType       `gorm:"column:asset_type" json:"assetType"`
	Quantity      decimal.Decimal `gorm:"column:quantity" json:"quantity"`
	AveragePrice  decimal.Decimal `gorm:"column:average_price" json:"averagePrice"`
	MarketValue   decimal.Decimal `gorm:"column:market_value" json:"marketValue"`
	UnrealizedPnL decimal.Decimal `gorm:"column:unrealized_pnl" json:"unrealizedPnl"`
	RealizedPnL   decimal.Decimal `gorm:"column:realized_pnl" json:"realizedPnl"`
	CreatedAt     time.Time       `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time       `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

// AssetType represents asset type
type AssetType string

const (
	AssetTypeStock  AssetType = "stock"
	AssetTypeCrypto AssetType = "crypto"
	AssetTypeForex  AssetType = "forex"
	AssetTypeBond   AssetType = "bond"
	AssetTypeOption AssetType = "option"
)

// OrderSide represents order side
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// OrderType represents order type
type OrderType string

const (
	OrderTypeMarket    OrderType = "market"
	OrderTypeLimit     OrderType = "limit"
	OrderTypeStop      OrderType = "stop"
	OrderTypeStopLimit OrderType = "stop_limit"
)

// OrderStatus represents order status
type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "new"
	OrderStatusPending         OrderStatus = "pending"
	OrderStatusAccepted        OrderStatus = "accepted"
	OrderStatusRejected        OrderStatus = "rejected"
	OrderStatusFilled          OrderStatus = "filled"
	OrderStatusPartiallyFilled OrderStatus = "partially_filled"
	OrderStatusCancelRequested OrderStatus = "cancel_requested"
	OrderStatusCancelled       OrderStatus = "cancelled"
	OrderStatusExpired         OrderStatus = "expired"
)

func (*TradingAccount) TableName() string { return "trading_accounts" }

func (*Position) TableName() string { return "positions" }