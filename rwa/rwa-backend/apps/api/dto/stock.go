package dto

// GetStockListRequest request for getting stock list with pagination
type GetStockListRequest struct {
	Page     int `form:"page" json:"page" example:"1"`
	PageSize int `form:"page_size" json:"page_size" example:"20"`
}

// GetStockListResponse response for getting stock list
type GetStockListResponse struct {
	List  []StockInfo `json:"list"`
	Total int64       `json:"total"`
}

// GetStockDetailRequest request for getting stock detail
type GetStockDetailRequest struct {
	Symbol string `form:"symbol" json:"symbol" binding:"required" validate:"required" example:"AAPL"`
}

// StockInfo stock information
type StockInfo struct {
	ID        uint64 `json:"id"`
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
	Exchange  string `json:"exchange"`
	About     string `json:"about,omitempty"`
	Status    string `json:"status,omitempty"`
	Contract  string `json:"contract"`
	CreatedAt int64  `json:"createdAt,omitempty" example:"1704067200"` // Unix timestamp in seconds
	UpdatedAt int64  `json:"updatedAt,omitempty" example:"1704067200"` // Unix timestamp in seconds
}
