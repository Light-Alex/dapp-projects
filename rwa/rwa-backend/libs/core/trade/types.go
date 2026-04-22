package trade

import (
	"context"
	"time"

	"github.com/AnchoredLabs/rwa-backend/libs/core/models/rwa"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/shopspring/decimal"
)

// TradeService trading service interface
// TODO: Consider splitting into smaller interfaces by responsibility:
//   - OrderService: PlaceOrder, CancelOrder, GetOrder, GetOrders
//   - MarketDataService: GetCurrentPrice, GetLatestQuote, GetSnapshot, GetHistoricalData
//   - AccountService: GetAccount, GetPositions, GetAssets, GetAsset, GetMarketClock
type TradeService interface {
	PlaceOrder(ctx context.Context, req PlaceOrderRequest) (*rwa.Order, error)
	CancelOrder(ctx context.Context, orderID string) error
	GetOrder(ctx context.Context, orderID string) (*rwa.Order, error)
	GetOrders(ctx context.Context, req GetOrdersRequest) ([]rwa.Order, error)
	GetAccount(ctx context.Context) (*TradingAccountResponse, error)
	GetPositions(ctx context.Context) ([]PositionResponse, error)
	GetAssets(ctx context.Context, req GetAssetsRequest) ([]AssetResponse, error)
	GetAsset(ctx context.Context, symbol string) (*AssetResponse, error)
	GetCurrentPrice(ctx context.Context, symbol string) (*MarketDataResponse, error)
	GetLatestQuote(ctx context.Context, symbol string) (*QuoteResponse, error)
	GetSnapshot(ctx context.Context, symbol string) (*SnapshotResponse, error)
	GetHistoricalData(ctx context.Context, req GetHistoricalDataRequest) ([]HistoricalDataResponse, error)
	GetMarketClock(ctx context.Context) (*MarketClockResponse, error)
}

// PlaceOrderRequest place order request
type PlaceOrderRequest struct {
	ClientOrderID string             `json:"client_order_id"`
	Symbol        string             `json:"symbol"`
	AssetType     rwa.AssetType      `json:"asset_type"`
	Side          rwa.OrderSide      `json:"side"`
	Type          rwa.OrderType      `json:"type"`
	Quantity      decimal.Decimal    `json:"quantity"`
	Price         decimal.Decimal    `json:"price"`
	StopPrice     decimal.Decimal    `json:"stop_price"`
	TimeInForce   alpaca.TimeInForce `json:"time_in_force"`
}

// GetOrdersRequest get orders request
type GetOrdersRequest struct {
	Status    string    `json:"status,omitempty"`
	Symbol    string    `json:"symbol,omitempty"`
	Limit     int       `json:"limit,omitempty"`
	Offset    int       `json:"offset,omitempty"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// GetHistoricalDataRequest get historical data request
type GetHistoricalDataRequest struct {
	Symbol    string    `json:"symbol"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Interval  string    `json:"interval"`
	Limit     int       `json:"limit,omitempty"`
}

// GetAssetsRequest get assets request
type GetAssetsRequest struct {
	Status     string `json:"status,omitempty"`      // active or inactive
	AssetClass string `json:"asset_class,omitempty"` // us_equity or crypto
	Exchange   string `json:"exchange,omitempty"`    // exchange name
}

// MarketHours market hours
type MarketHours struct {
	IsOpen    bool      `json:"is_open"`
	NextOpen  time.Time `json:"next_open"`
	NextClose time.Time `json:"next_close"`
}

// TradeServiceConfig trade service configuration
type TradeServiceConfig struct {
	Provider string                 `json:"provider"` // alpaca
	Config   map[string]any `json:"config"`
}

// AlpacaConfig alpaca configuration
type AlpacaConfig struct {
	APIKey    string `json:"api_key" yaml:"api_key"`
	APISecret string `json:"api_secret" yaml:"api_secret"`
	BaseURL   string `json:"base_url" yaml:"base_url"`
	DataURL   string `json:"data_url" yaml:"data_url"`
}
