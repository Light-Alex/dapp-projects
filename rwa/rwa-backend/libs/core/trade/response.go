package trade

import (
	"time"

	"github.com/shopspring/decimal"
)

// TradingAccountResponse represents a trading account response
type TradingAccountResponse struct {
	ID                string    `json:"id"`
	ExternalAccountID string    `json:"external_account_id"`
	Provider          string    `json:"provider"`
	AccountType       string    `json:"account_type"`
	Status            string    `json:"status"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// PositionResponse represents a position response
type PositionResponse struct {
	Symbol        string          `json:"symbol"`
	AssetType     string          `json:"asset_type"`
	Quantity      decimal.Decimal `json:"quantity"`
	AveragePrice  decimal.Decimal `json:"average_price"`
	MarketValue   decimal.Decimal `json:"market_value"`
	UnrealizedPnL decimal.Decimal `json:"unrealized_pnl"`
	RealizedPnL   decimal.Decimal `json:"realized_pnl"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// MarketDataResponse represents market data response
type MarketDataResponse struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

// HistoricalDataResponse represents historical data response
type HistoricalDataResponse struct {
	Symbol    string    `json:"symbol"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    uint64    `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

// MarketClockResponse represents market clock response
type MarketClockResponse struct {
	Timestamp time.Time `json:"timestamp"`
	IsOpen    bool      `json:"is_open"`
	NextOpen  time.Time `json:"next_open"`
	NextClose time.Time `json:"next_close"`
}

// QuoteResponse represents a quote response for a symbol
type QuoteResponse struct {
	Symbol      string   `json:"symbol"`
	Timestamp   string   `json:"timestamp"` // ISO 8601 format
	BidPrice    float64  `json:"bid_price"`
	BidSize     uint32   `json:"bid_size"`
	AskPrice    float64  `json:"ask_price"`
	AskSize     uint32   `json:"ask_size"`
	BidExchange string   `json:"bid_exchange,omitempty"`
	AskExchange string   `json:"ask_exchange,omitempty"`
	Conditions  []string `json:"conditions,omitempty"`
	Tape        string   `json:"tape,omitempty"`
	MidPrice    float64  `json:"mid_price,omitempty"`
}

// SnapshotResponse represents a snapshot of market data for a symbol
type SnapshotResponse struct {
	Symbol       string     `json:"symbol"`
	LatestTrade  *TradeData `json:"latest_trade,omitempty"`
	LatestQuote  *QuoteData `json:"latest_quote,omitempty"`
	MinuteBar    *BarData   `json:"minute_bar,omitempty"`
	DailyBar     *BarData   `json:"daily_bar,omitempty"`
	PrevDailyBar *BarData   `json:"prev_daily_bar,omitempty"`
}

// TradeData represents trade data
type TradeData struct {
	Timestamp  time.Time `json:"timestamp"`
	Price      float64   `json:"price"`
	Size       uint32    `json:"size"`
	Exchange   string    `json:"exchange"`
	ID         int64     `json:"id"`
	Conditions []string  `json:"conditions,omitempty"`
	Tape       string    `json:"tape,omitempty"`
	Update     string    `json:"update,omitempty"`
}

// QuoteData represents quote data
type QuoteData struct {
	Timestamp   time.Time `json:"timestamp"`
	BidPrice    float64   `json:"bid_price"`
	BidSize     uint32    `json:"bid_size"`
	AskPrice    float64   `json:"ask_price"`
	AskSize     uint32    `json:"ask_size"`
	BidExchange string    `json:"bid_exchange,omitempty"`
	AskExchange string    `json:"ask_exchange,omitempty"`
	Conditions  []string  `json:"conditions,omitempty"`
	Tape        string    `json:"tape,omitempty"`
}

// BarData represents bar/OHLCV data
type BarData struct {
	Timestamp  time.Time `json:"timestamp"`
	Open       float64   `json:"open"`
	High       float64   `json:"high"`
	Low        float64   `json:"low"`
	Close      float64   `json:"close"`
	Volume     uint64    `json:"volume"`
	TradeCount uint64    `json:"trade_count,omitempty"`
	VWAP       float64   `json:"vwap,omitempty"`
}

// AssetResponse represents an asset response
type AssetResponse struct {
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
