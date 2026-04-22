package dto

// GetCurrentPriceRequest request for getting current price
type GetCurrentPriceRequest struct {
	Symbol string `form:"symbol" json:"symbol" binding:"required" validate:"required" example:"AAPL"`
}

// GetCurrentPriceResponse response for getting current price
type GetCurrentPriceResponse struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Volume    float64 `json:"volume"`
	Timestamp int64   `json:"timestamp" example:"1704067200"` // Unix timestamp in seconds
}

// GetHistoricalDataRequest request for getting historical data
// Interval supports two formats:
//   - Short format: 1m, 5m, 15m, 1h, 1d, 1w
//   - Full format: 1Min, 5Min, 15Min, 1Hour, 1Day, 1Week
type GetHistoricalDataRequest struct {
	Symbol    string `form:"symbol" json:"symbol" binding:"required" validate:"required" example:"AAPL"`
	StartTime int64  `form:"start_time" json:"start_time" binding:"required" validate:"required" example:"1704067200"` // Unix timestamp in seconds
	EndTime   int64  `form:"end_time" json:"end_time" binding:"required" validate:"required" example:"1706745599"`     // Unix timestamp in seconds
	Interval  string `form:"interval" json:"interval" binding:"required" validate:"required" example:"1d"`             // Examples: 1m/1Min, 5m/5Min, 15m/15Min, 1h/1Hour, 1d/1Day, 1w/1Week
	Limit     int    `form:"limit" json:"limit" example:"100"`
}

// GetHistoricalDataResponse response for getting historical data
type GetHistoricalDataResponse struct {
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    uint64  `json:"volume"`
	Timestamp int64   `json:"timestamp" example:"1704067200"` // Unix timestamp in seconds
}

// GetHistoricalDataListResponse list response for getting historical data
type GetHistoricalDataListResponse struct {
	Symbol string                      `json:"symbol"`
	Data   []GetHistoricalDataResponse `json:"data"`
}

// GetMarketClockResponse response for getting market clock
type GetMarketClockResponse struct {
	Timestamp int64 `json:"timestamp" example:"1704067200"` // Unix timestamp in seconds
	IsOpen    bool  `json:"is_open"`
	NextOpen  int64 `json:"next_open" example:"1704067200"`  // Unix timestamp in seconds
	NextClose int64 `json:"next_close" example:"1704067200"` // Unix timestamp in seconds
}

// GetLatestQuoteRequest request for getting latest quote
type GetLatestQuoteRequest struct {
	Symbol string `form:"symbol" json:"symbol" binding:"required" validate:"required" example:"AAPL"`
}

// GetLatestQuoteResponse response for getting latest quote
type GetLatestQuoteResponse struct {
	Quote QuoteDTO `json:"quote"`
}

// GetSnapshotRequest request for getting snapshot
type GetSnapshotRequest struct {
	Symbol string `form:"symbol" json:"symbol" binding:"required" validate:"required" example:"AAPL"`
}

// GetSnapshotResponse response for getting snapshot
type GetSnapshotResponse struct {
	Snapshot SnapshotData `json:"snapshot"`
}

// SnapshotData represents snapshot data for a symbol
type SnapshotData struct {
	Symbol       string    `json:"symbol"`
	LatestTrade  *TradeDTO `json:"latest_trade,omitempty"`
	LatestQuote  *QuoteDTO `json:"latest_quote,omitempty"`
	MinuteBar    *BarDTO   `json:"minute_bar,omitempty"`
	DailyBar     *BarDTO   `json:"daily_bar,omitempty"`
	PrevDailyBar *BarDTO   `json:"prev_daily_bar,omitempty"`
}

// TradeDTO represents trade data in DTO format
type TradeDTO struct {
	Timestamp  int64    `json:"timestamp" example:"1704067200"` // Unix timestamp in seconds
	Price      float64  `json:"price"`
	Size       uint32   `json:"size"`
	Exchange   string   `json:"exchange"`
	ID         int64    `json:"id"`
	Conditions []string `json:"conditions,omitempty"`
	Tape       string   `json:"tape,omitempty"`
	Update     string   `json:"update,omitempty"`
}

// QuoteDTO represents quote data in DTO format
type QuoteDTO struct {
	Timestamp   int64    `json:"timestamp" example:"1704067200"` // Unix timestamp in seconds
	BidPrice    float64  `json:"bid_price"`
	BidSize     uint32   `json:"bid_size"`
	AskPrice    float64  `json:"ask_price"`
	AskSize     uint32   `json:"ask_size"`
	BidExchange string   `json:"bid_exchange,omitempty"`
	AskExchange string   `json:"ask_exchange,omitempty"`
	Conditions  []string `json:"conditions,omitempty"`
	Tape        string   `json:"tape,omitempty"`
}

// BarDTO represents bar/OHLCV data in DTO format
type BarDTO struct {
	Timestamp  int64   `json:"timestamp" example:"1704067200"` // Unix timestamp in seconds
	Open       float64 `json:"open"`
	High       float64 `json:"high"`
	Low        float64 `json:"low"`
	Close      float64 `json:"close"`
	Volume     uint64  `json:"volume"`
	TradeCount uint64  `json:"trade_count,omitempty"`
	VWAP       float64 `json:"vwap,omitempty"`
}

// GetAssetsRequest request for getting assets
type GetAssetsRequest struct {
	Status     string `form:"status" json:"status" example:"active"`              // active or inactive
	AssetClass string `form:"asset_class" json:"asset_class" example:"us_equity"` // us_equity or crypto
	Exchange   string `form:"exchange" json:"exchange" example:"NASDAQ"`          // exchange name
}

// GetAssetsResponse response for getting assets
type GetAssetsResponse struct {
	Assets []AssetDTO `json:"assets"`
}

// GetAssetRequest request for getting a single asset
type GetAssetRequest struct {
	Symbol string `form:"symbol" json:"symbol" binding:"required" validate:"required" example:"AAPL"`
}

// GetAssetResponse response for getting a single asset
type GetAssetResponse struct {
	Asset AssetDTO `json:"asset"`
}

// AssetDTO represents asset data in DTO format
type AssetDTO struct {
	ID                           string   `json:"id"`
	Class                        string   `json:"class"`                           // us_equity or crypto
	Exchange                     string   `json:"exchange"`                        // exchange name
	Symbol                       string   `json:"symbol"`                          // asset symbol
	Name                         string   `json:"name"`                            // asset name
	Status                       string   `json:"status"`                          // active or inactive
	Tradable                     bool     `json:"tradable"`                        // whether the asset is tradable
	Marginable                   bool     `json:"marginable"`                      // whether the asset is marginable
	MaintenanceMarginRequirement uint     `json:"maintenance_margin_requirement"`  // maintenance margin requirement
	Shortable                    bool     `json:"shortable"`                       // whether the asset is shortable
	EasyToBorrow                 bool     `json:"easy_to_borrow"`                  // whether the asset is easy to borrow
	Fractionable                 bool     `json:"fractionable"`                    // whether the asset is fractionable
	Attributes                   []string `json:"attributes,omitempty"`            // additional attributes
}
