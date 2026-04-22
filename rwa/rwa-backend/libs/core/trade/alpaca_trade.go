package trade

import (
	"context"
	"time"

	"github.com/AnchoredLabs/rwa-backend/libs/core/models/rwa"
	"github.com/AnchoredLabs/rwa-backend/libs/errors"
	"github.com/AnchoredLabs/rwa-backend/libs/log"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// alpacaTradeService handles trading operations using tradeClient
type alpacaTradeService struct {
	tradeClient *alpaca.Client
}

func newAlpacaTradeService(tradeClient *alpaca.Client) *alpacaTradeService {
	return &alpacaTradeService{
		tradeClient: tradeClient,
	}
}

func (s *alpacaTradeService) PlaceOrder(ctx context.Context, req PlaceOrderRequest) (*rwa.Order, error) {
	log.InfoZ(ctx, "Placing order",
		zap.String("symbol", req.Symbol),
		zap.String("side", string(req.Side)),
		zap.String("type", string(req.Type)),
		zap.String("quantity", req.Quantity.String()))

	if err := validateOrderRequest(req.Symbol, string(req.Side), string(req.Type), req.Quantity); err != nil {
		log.ErrorZ(ctx, "Order validation failed", zap.Error(err))
		return nil, err
	}

	alpacaReq := alpaca.PlaceOrderRequest{
		Symbol:        req.Symbol,
		Qty:           &req.Quantity,
		Side:          alpaca.Side(req.Side),
		Type:          alpaca.OrderType(req.Type),
		TimeInForce:   req.TimeInForce,
		ClientOrderID: req.ClientOrderID,
	}

	if !req.Price.IsZero() {
		alpacaReq.LimitPrice = &req.Price
	}
	if !req.StopPrice.IsZero() {
		alpacaReq.StopPrice = &req.StopPrice
	}

	alpacaOrder, err := s.tradeClient.PlaceOrder(alpacaReq)
	if err != nil {
		log.ErrorZ(ctx, "Failed to place order with Alpaca", zap.Error(err))
		return nil, errors.Errorf("place order failed: %v", err)
	}

	log.InfoZ(ctx, "Order placed successfully",
		zap.String("alpaca_order_id", alpacaOrder.ID),
		zap.String("client_order_id", req.ClientOrderID))

	return convertOrder(alpacaOrder, req.ClientOrderID), nil
}

func (s *alpacaTradeService) CancelOrder(ctx context.Context, orderID string) error {
	log.InfoZ(ctx, "Cancelling order", zap.String("order_id", orderID))

	err := s.tradeClient.CancelOrder(orderID)
	if err != nil {
		log.ErrorZ(ctx, "Failed to cancel order", zap.String("order_id", orderID), zap.Error(err))
		return errors.Errorf("cancel order failed: %v", err)
	}

	return nil
}

func (s *alpacaTradeService) GetOrder(ctx context.Context, orderID string) (*rwa.Order, error) {
	log.InfoZ(ctx, "Getting order", zap.String("order_id", orderID))

	alpacaOrder, err := s.tradeClient.GetOrder(orderID)
	if err != nil {
		log.ErrorZ(ctx, "Failed to get order", zap.String("order_id", orderID), zap.Error(err))
		return nil, errors.Errorf("get order failed: %v", err)
	}

	return convertOrder(alpacaOrder, ""), nil
}

func (s *alpacaTradeService) GetOrders(ctx context.Context, req GetOrdersRequest) ([]rwa.Order, error) {
	log.InfoZ(ctx, "Getting orders",
		zap.String("status", req.Status),
		zap.String("symbol", req.Symbol),
		zap.Int("limit", req.Limit))

	alpacaReq := alpaca.GetOrdersRequest{
		Status: req.Status,
		Limit:  req.Limit,
	}

	if !req.StartTime.IsZero() {
		alpacaReq.After = req.StartTime
	}
	if !req.EndTime.IsZero() {
		alpacaReq.Until = req.EndTime
	}

	alpacaOrders, err := s.tradeClient.GetOrders(alpacaReq)
	if err != nil {
		log.ErrorZ(ctx, "Failed to get orders", zap.Error(err))
		return nil, errors.Errorf("get orders failed: %v", err)
	}

	orders := make([]rwa.Order, len(alpacaOrders))
	for i, alpacaOrder := range alpacaOrders {
		orders[i] = *convertOrder(&alpacaOrder, "")
	}

	return orders, nil
}

func (s *alpacaTradeService) GetAccount(ctx context.Context) (*TradingAccountResponse, error) {
	log.InfoZ(ctx, "Getting account information")

	alpacaAccount, err := s.tradeClient.GetAccount()
	if err != nil {
		log.ErrorZ(ctx, "Failed to get account", zap.Error(err))
		return nil, errors.Errorf("get account failed: %v", err)
	}

	log.InfoZ(ctx, "Account retrieved successfully",
		zap.String("account_id", alpacaAccount.ID),
		zap.String("status", alpacaAccount.Status))
	return convertTradingAccount(alpacaAccount), nil
}

func (s *alpacaTradeService) GetPositions(ctx context.Context) ([]PositionResponse, error) {
	log.InfoZ(ctx, "Getting positions")

	alpacaPositions, err := s.tradeClient.GetPositions()
	if err != nil {
		log.ErrorZ(ctx, "Failed to get positions", zap.Error(err))
		return nil, errors.Errorf("get positions failed: %v", err)
	}

	positions := make([]PositionResponse, len(alpacaPositions))
	for i, alpacaPos := range alpacaPositions {
		positions[i] = *convertPosition(&alpacaPos)
	}

	return positions, nil
}

func (s *alpacaTradeService) GetMarketClock(ctx context.Context) (*MarketClockResponse, error) {
	log.InfoZ(ctx, "Getting market clock")

	clock, err := s.tradeClient.GetClock()
	if err != nil {
		log.ErrorZ(ctx, "Failed to get market clock", zap.Error(err))
		return nil, errors.Errorf("get market clock failed: %v", err)
	}

	response := &MarketClockResponse{
		Timestamp: clock.Timestamp,
		IsOpen:    clock.IsOpen,
		NextOpen:  clock.NextOpen,
		NextClose: clock.NextClose,
	}

	log.InfoZ(ctx, "Market clock retrieved",
		zap.Time("timestamp", response.Timestamp),
		zap.Bool("is_open", response.IsOpen),
		zap.String("next_open", response.NextOpen.Format("2006-01-02 15:04:05")),
		zap.String("next_close", response.NextClose.Format("2006-01-02 15:04:05")))

	return response, nil
}

func (s *alpacaTradeService) GetAssets(ctx context.Context, req GetAssetsRequest) ([]AssetResponse, error) {
	log.InfoZ(ctx, "Getting assets",
		zap.String("status", req.Status),
		zap.String("asset_class", req.AssetClass),
		zap.String("exchange", req.Exchange))

	alpacaReq := alpaca.GetAssetsRequest{
		Status:     req.Status,
		AssetClass: req.AssetClass,
		Exchange:   req.Exchange,
	}

	alpacaAssets, err := s.tradeClient.GetAssets(alpacaReq)
	if err != nil {
		log.ErrorZ(ctx, "Failed to get assets", zap.Error(err))
		return nil, errors.Errorf("get assets failed: %v", err)
	}

	assets := make([]AssetResponse, len(alpacaAssets))
	for i, alpacaAsset := range alpacaAssets {
		assets[i] = *convertAsset(&alpacaAsset)
	}

	log.InfoZ(ctx, "Assets retrieved successfully",
		zap.Int("count", len(assets)))

	return assets, nil
}

func (s *alpacaTradeService) GetAsset(ctx context.Context, symbol string) (*AssetResponse, error) {
	log.InfoZ(ctx, "Getting asset", zap.String("symbol", symbol))

	alpacaAsset, err := s.tradeClient.GetAsset(symbol)
	if err != nil {
		log.ErrorZ(ctx, "Failed to get asset", zap.String("symbol", symbol), zap.Error(err))
		return nil, errors.Errorf("get asset failed: %v", err)
	}

	log.InfoZ(ctx, "Asset retrieved successfully",
		zap.String("symbol", symbol),
		zap.String("name", alpacaAsset.Name))

	return convertAsset(alpacaAsset), nil
}

// Helper functions for converting Alpaca types to internal types

// derefDecimal safely dereferences a *decimal.Decimal pointer.
// Returns decimal.Zero if the pointer is nil.
func derefDecimal(d *decimal.Decimal) decimal.Decimal {
	if d == nil {
		return decimal.Zero
	}
	return *d
}

func convertOrder(alpacaOrder *alpaca.Order, clientOrderID string) *rwa.Order {
	qty := derefDecimal(alpacaOrder.Qty)
	order := &rwa.Order{
		ClientOrderID:     clientOrderID,
		Symbol:            alpacaOrder.Symbol,
		AssetType:         rwa.AssetType(alpacaOrder.AssetClass),
		Side:              rwa.OrderSide(alpacaOrder.Side),
		Type:              rwa.OrderType(alpacaOrder.Type),
		Quantity:          qty,
		Status:            rwa.OrderStatus(alpacaOrder.Status),
		FilledQuantity:    alpacaOrder.FilledQty,
		RemainingQuantity: qty.Sub(alpacaOrder.FilledQty),
		TimeInForce:       string(alpacaOrder.TimeInForce),
		Provider:          "alpaca",
		ExternalOrderID:   alpacaOrder.ID,
		CreatedAt:         alpacaOrder.CreatedAt,
		UpdatedAt:         alpacaOrder.UpdatedAt,
	}

	if alpacaOrder.LimitPrice != nil {
		order.Price = *alpacaOrder.LimitPrice
	}
	if alpacaOrder.StopPrice != nil {
		order.StopPrice = *alpacaOrder.StopPrice
	}
	if alpacaOrder.FilledAvgPrice != nil {
		order.FilledPrice = *alpacaOrder.FilledAvgPrice
	}
	if !alpacaOrder.SubmittedAt.IsZero() {
		order.SubmittedAt = &alpacaOrder.SubmittedAt
	}
	if alpacaOrder.FilledAt != nil {
		order.FilledAt = alpacaOrder.FilledAt
	}
	if alpacaOrder.CanceledAt != nil {
		order.CancelledAt = alpacaOrder.CanceledAt
	}

	return order
}

func convertTradingAccount(alpacaAccount *alpaca.Account) *TradingAccountResponse {
	return &TradingAccountResponse{
		ID:                "alpaca_" + alpacaAccount.ID,
		ExternalAccountID: alpacaAccount.ID,
		Provider:          "alpaca",
		AccountType:       "paper", // default to paper account
		Status:            alpacaAccount.Status,
		IsActive:          true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

func convertPosition(alpacaPosition *alpaca.Position) *PositionResponse {
	return &PositionResponse{
		Symbol:        alpacaPosition.Symbol,
		AssetType:     string(alpacaPosition.AssetClass),
		Quantity:      alpacaPosition.Qty,
		AveragePrice:  alpacaPosition.AvgEntryPrice,
		MarketValue:   derefDecimal(alpacaPosition.MarketValue),
		UnrealizedPnL: derefDecimal(alpacaPosition.UnrealizedPL),
		RealizedPnL:   decimal.Zero, // Alpaca Position API does not expose realized P&L
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func convertAsset(alpacaAsset *alpaca.Asset) *AssetResponse {
	return &AssetResponse{
		ID:                           alpacaAsset.ID,
		Class:                        string(alpacaAsset.Class),
		Exchange:                     alpacaAsset.Exchange,
		Symbol:                       alpacaAsset.Symbol,
		Name:                         alpacaAsset.Name,
		Status:                       string(alpacaAsset.Status),
		Tradable:                     alpacaAsset.Tradable,
		Marginable:                   alpacaAsset.Marginable,
		MaintenanceMarginRequirement: alpacaAsset.MaintenanceMarginRequirement,
		Shortable:                    alpacaAsset.Shortable,
		EasyToBorrow:                 alpacaAsset.EasyToBorrow,
		Fractionable:                 alpacaAsset.Fractionable,
		Attributes:                   alpacaAsset.Attributes,
	}
}
