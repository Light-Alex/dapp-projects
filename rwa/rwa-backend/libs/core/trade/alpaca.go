package trade

import (
	"context"

	"github.com/AnchoredLabs/rwa-backend/libs/core/models/rwa"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

type AlpacaService struct {
	tradeService      *alpacaTradeService
	marketDataService *alpacaMarketDataService
	config            *AlpacaConfig
}

func NewAlpacaService(cfg *AlpacaConfig) *AlpacaService {
	tradeClient := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    cfg.APIKey,
		APISecret: cfg.APISecret,
		BaseURL:   cfg.BaseURL,
	})

	dataClient := marketdata.NewClient(marketdata.ClientOpts{
		APIKey:    cfg.APIKey,
		APISecret: cfg.APISecret,
		BaseURL:   cfg.DataURL,
	})

	return &AlpacaService{
		tradeService:      newAlpacaTradeService(tradeClient),
		marketDataService: newAlpacaMarketDataService(dataClient),
		config:            cfg,
	}
}

func (s *AlpacaService) PlaceOrder(ctx context.Context, req PlaceOrderRequest) (*rwa.Order, error) {
	return s.tradeService.PlaceOrder(ctx, req)
}

func (s *AlpacaService) CancelOrder(ctx context.Context, orderID string) error {
	return s.tradeService.CancelOrder(ctx, orderID)
}

func (s *AlpacaService) GetOrder(ctx context.Context, orderID string) (*rwa.Order, error) {
	return s.tradeService.GetOrder(ctx, orderID)
}

func (s *AlpacaService) GetOrders(ctx context.Context, req GetOrdersRequest) ([]rwa.Order, error) {
	return s.tradeService.GetOrders(ctx, req)
}

func (s *AlpacaService) GetAccount(ctx context.Context) (*TradingAccountResponse, error) {
	return s.tradeService.GetAccount(ctx)
}

func (s *AlpacaService) GetPositions(ctx context.Context) ([]PositionResponse, error) {
	return s.tradeService.GetPositions(ctx)
}

func (s *AlpacaService) GetAssets(ctx context.Context, req GetAssetsRequest) ([]AssetResponse, error) {
	return s.tradeService.GetAssets(ctx, req)
}

func (s *AlpacaService) GetAsset(ctx context.Context, symbol string) (*AssetResponse, error) {
	return s.tradeService.GetAsset(ctx, symbol)
}

func (s *AlpacaService) GetMarketClock(ctx context.Context) (*MarketClockResponse, error) {
	return s.tradeService.GetMarketClock(ctx)
}

func (s *AlpacaService) GetCurrentPrice(ctx context.Context, symbol string) (*MarketDataResponse, error) {
	return s.marketDataService.GetCurrentPrice(ctx, symbol)
}

func (s *AlpacaService) GetHistoricalData(ctx context.Context, req GetHistoricalDataRequest) ([]HistoricalDataResponse, error) {
	return s.marketDataService.GetHistoricalData(ctx, req)
}

func (s *AlpacaService) GetLatestQuote(ctx context.Context, symbol string) (*QuoteResponse, error) {
	return s.marketDataService.GetLatestQuote(ctx, symbol)
}

func (s *AlpacaService) GetSnapshot(ctx context.Context, symbol string) (*SnapshotResponse, error) {
	return s.marketDataService.GetSnapshot(ctx, symbol)
}
