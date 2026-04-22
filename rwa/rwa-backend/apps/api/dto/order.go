package dto

// GetOrdersRequest request for listing orders with filters and pagination
type GetOrdersRequest struct {
	AccountID uint64 `form:"account_id" json:"account_id" example:"1"`
	Symbol    string `form:"symbol" json:"symbol" example:"AAPL"`
	Side      string `form:"side" json:"side" example:"buy"`
	Status    string `form:"status" json:"status" example:"filled"`
	Page      int    `form:"page" json:"page" example:"1"`
	PageSize  int    `form:"page_size" json:"page_size" example:"20"`
}

// GetOrdersResponse response for listing orders
type GetOrdersResponse struct {
	List  []OrderDTO `json:"list"`
	Total int64      `json:"total"`
}

// GetOrderDetailRequest request for getting order detail
type GetOrderDetailRequest struct {
	ID            uint64 `form:"id" json:"id" example:"1"`
	ClientOrderID string `form:"client_order_id" json:"client_order_id" example:""`
}

// GetOrderDetailResponse response for getting order detail
type GetOrderDetailResponse struct {
	Order OrderDTO `json:"order"`
}

// GetOrderExecutionsRequest request for getting order executions
type GetOrderExecutionsRequest struct {
	OrderID uint64 `form:"order_id" json:"order_id" binding:"required" validate:"required" example:"1"`
}

// GetOrderExecutionsResponse response for getting order executions
type GetOrderExecutionsResponse struct {
	List []OrderExecutionDTO `json:"list"`
}

// OrderDTO JSON representation of an order
type OrderDTO struct {
	ID                uint64 `json:"id"`
	ClientOrderID     string `json:"clientOrderId"`
	AccountID         uint64 `json:"accountId"`
	Symbol            string `json:"symbol"`
	AssetType         string `json:"assetType"`
	Side              string `json:"side"`
	Type              string `json:"type"`
	Quantity          string `json:"quantity"`
	Price             string `json:"price"`
	StopPrice         string `json:"stopPrice"`
	Status            string `json:"status"`
	FilledQuantity    string `json:"filledQuantity"`
	FilledPrice       string `json:"filledPrice"`
	RemainingQuantity string `json:"remainingQuantity"`
	ContractTxHash    string `json:"contractTxHash"`
	ExternalOrderID   string `json:"externalOrderId"`
	Provider          string `json:"provider"`
	Commission        string `json:"commission"`
	CommissionAsset   string `json:"commissionAsset"`
	Metadata          string `json:"metadata,omitempty"`
	Notes             string `json:"notes,omitempty"`
	CreatedAt         int64  `json:"createdAt,omitempty" example:"1704067200"`
	UpdatedAt         int64  `json:"updatedAt,omitempty" example:"1704067200"`
	SubmittedAt       int64  `json:"submittedAt,omitempty" example:"1704067200"`
	FilledAt          int64  `json:"filledAt,omitempty" example:"1704067200"`
	CancelledAt       int64  `json:"cancelledAt,omitempty" example:"1704067200"`
	ExpiredAt         int64  `json:"expiredAt,omitempty" example:"1704067200"`
}

// OrderExecutionDTO JSON representation of an order execution
type OrderExecutionDTO struct {
	ID              uint64 `json:"id"`
	OrderID         uint64 `json:"orderId"`
	ExecutionID     string `json:"executionId"`
	Quantity        string `json:"quantity"`
	Price           string `json:"price"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	Provider        string `json:"provider"`
	ExternalID      string `json:"externalId"`
	ExecutedAt      int64  `json:"executedAt,omitempty" example:"1704067200"`
	CreatedAt       int64  `json:"createdAt,omitempty" example:"1704067200"`
}
