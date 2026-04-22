package rwa

import (
	"time"

	"github.com/shopspring/decimal"
)

// Order maps to public.orders
type Order struct {
	ID                uint64          `gorm:"column:id;primaryKey" json:"id"`
	ClientOrderID     string          `gorm:"column:client_order_id;index" json:"clientOrderId"`
	AccountID         uint64          `gorm:"column:account_id;index" json:"accountId"`
	Symbol            string          `gorm:"column:symbol;index" json:"symbol"`
	AssetType         AssetType       `gorm:"column:asset_type" json:"assetType"`
	Side              OrderSide       `gorm:"column:side" json:"side"`
	Type              OrderType       `gorm:"column:type" json:"type"`
	Quantity          decimal.Decimal `gorm:"column:quantity" json:"quantity"`
	Price             decimal.Decimal `gorm:"column:price" json:"price"`
	StopPrice         decimal.Decimal `gorm:"column:stop_price" json:"stopPrice"`
	Status            OrderStatus     `gorm:"column:status" json:"status"`
	FilledQuantity    decimal.Decimal `gorm:"column:filled_quantity" json:"filledQuantity"`
	FilledPrice       decimal.Decimal `gorm:"column:filled_price" json:"filledPrice"`
	RemainingQuantity decimal.Decimal `gorm:"column:remaining_quantity" json:"remainingQuantity"`
	EscrowAmount      decimal.Decimal `gorm:"column:escrow_amount" json:"escrowAmount"`
	EscrowAsset       string          `gorm:"column:escrow_asset" json:"escrowAsset"`
	RefundAmount      decimal.Decimal `gorm:"column:refund_amount" json:"refundAmount"`
	ContractTxHash    string          `gorm:"column:contract_tx_hash;index" json:"contractTxHash"`
	ExecuteTxHash     string          `gorm:"column:execute_tx_hash" json:"executeTxHash"`
	CancelTxHash      string          `gorm:"column:cancel_tx_hash" json:"cancelTxHash"`
	EventLogID        uint64          `gorm:"column:event_log_id;index" json:"eventLogId"`
	ExternalOrderID   string          `gorm:"column:external_order_id" json:"externalOrderId"`
	TimeInForce       string          `gorm:"column:time_in_force" json:"timeInForce"`
	Provider          string          `gorm:"column:provider" json:"provider"`
	Commission        string          `gorm:"column:commission" json:"commission"`
	CommissionAsset   string          `gorm:"column:commission_asset" json:"commissionAsset"`
	Metadata          string          `gorm:"column:metadata;type:text" json:"metadata"`
	Notes             string          `gorm:"column:notes;type:text" json:"notes"`
	CreatedAt         time.Time       `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time       `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
	SubmittedAt       *time.Time      `gorm:"column:submitted_at" json:"submittedAt"`
	AcceptedAt        *time.Time      `gorm:"column:accepted_at" json:"acceptedAt"`
	FilledAt          *time.Time      `gorm:"column:filled_at" json:"filledAt"`
	CancelledAt       *time.Time      `gorm:"column:cancelled_at" json:"cancelledAt"`
	ExpiredAt         *time.Time      `gorm:"column:expired_at" json:"expiredAt"`
}

// OrderExecution maps to public.order_executions
type OrderExecution struct {
	ID              uint64          `gorm:"column:id;primaryKey" json:"id"`
	OrderID         uint64          `gorm:"column:order_id;index" json:"orderId"`
	ExecutionID     string          `gorm:"column:execution_id" json:"executionId"`
	Quantity        decimal.Decimal `gorm:"column:quantity" json:"quantity"`
	Price           decimal.Decimal `gorm:"column:price" json:"price"`
	Commission      string          `gorm:"column:commission" json:"commission"`
	CommissionAsset string          `gorm:"column:commission_asset" json:"commissionAsset"`
	Provider        string          `gorm:"column:provider" json:"provider"`
	ExternalID      string          `gorm:"column:external_id" json:"externalId"`
	ExecutedAt      time.Time       `gorm:"column:executed_at" json:"executedAt"`
	CreatedAt       time.Time       `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

// OrderFilter represents order filter conditions
type OrderFilter struct {
	AccountID string    `json:"account_id"`
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Source    string    `json:"source"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}

func (*Order) TableName() string { return "orders" }

func (*OrderExecution) TableName() string { return "order_executions" }