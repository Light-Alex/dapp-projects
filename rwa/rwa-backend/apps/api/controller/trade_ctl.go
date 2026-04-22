package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/AnchoredLabs/rwa-backend/apps/api/dto"
	"github.com/AnchoredLabs/rwa-backend/apps/api/utils"
	"github.com/AnchoredLabs/rwa-backend/libs/core/error_msg"
	"github.com/AnchoredLabs/rwa-backend/libs/core/redis_cache"
	"github.com/AnchoredLabs/rwa-backend/libs/core/trade"
	"github.com/AnchoredLabs/rwa-backend/libs/core/web"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type TradeController struct {
	tradeService trade.TradeService
	cache        *redis_cache.MarketDataCacheService
}

func NewTradeController(tradeService trade.TradeService, cache *redis_cache.MarketDataCacheService) *TradeController {
	return &TradeController{
		tradeService: tradeService,
		cache:        cache,
	}
}

// tryCache attempts to get data from cache. Returns true on hit.
// Logs non-Nil Redis errors for observability.
func (ctl *TradeController) tryCache(ctx context.Context, dataType, key string, dest any) bool {
	err := ctl.cache.Get(ctx, dataType, key, dest)
	if err == nil {
		return true
	}
	if err != redis.Nil {
		log.WarnZ(ctx, "Redis cache error, falling back to upstream",
			zap.String("dataType", dataType),
			zap.String("key", key),
			zap.Error(err))
	}
	return false
}

// GetCurrentPrice godoc
// @Summary Get Current Price
// @Description Get the current price for a given symbol
// @Tags Trade
// @Accept json
// @Produce json
// @Param symbol query string true "Stock symbol" default(AAPL) example(AAPL)
// @Success 200 {object} web.Response{data=dto.GetCurrentPriceResponse}
// @Failure 400 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /trade/currentPrice [get]
func (ctl *TradeController) GetCurrentPrice(g *gin.Context) {
	ctx := g.Request.Context()

	var req dto.GetCurrentPriceRequest
	if err := g.ShouldBindQuery(&req); err != nil {
		log.ErrorZ(ctx, "failed to bind query parameters", zap.Error(err))
		web.ResponseError(error_msg.ErrInvalidRequestParams, g)
		return
	}

	log.InfoZ(ctx, "Getting current price", zap.String("symbol", req.Symbol))

	// Try cache first
	var cached dto.GetCurrentPriceResponse
	if ctl.tryCache(ctx, "currentPrice", req.Symbol, &cached) {
		web.ResponseOk(cached, g)
		return
	}

	marketData, err := ctl.tradeService.GetCurrentPrice(ctx, req.Symbol)
	if err != nil {
		log.ErrorZ(ctx, "failed to get current price",
			zap.String("symbol", req.Symbol),
			zap.Error(err))
		web.ResponseError(error_msg.ErrFailedToGetCurrentPrice, g)
		return
	}

	response := dto.GetCurrentPriceResponse{
		Symbol:    marketData.Symbol,
		Price:     marketData.Price,
		Volume:    marketData.Volume,
		Timestamp: utils.TimeToUnix(marketData.Timestamp),
	}

	_ = ctl.cache.Set(ctx, "currentPrice", req.Symbol, response, redis_cache.CurrentPriceTTL)
	web.ResponseOk(response, g)
}

// GetHistoricalData godoc
// @Summary Get Historical Data
// @Description Get historical price data for a given symbol
// @Tags Trade
// @Accept json
// @Produce json
// @Param symbol query string true "Stock symbol" default(AAPL) example(AAPL)
// @Param start_time query int true "Start time as Unix timestamp in seconds" default(1704067200) example(1704067200)
// @Param end_time query int true "End time as Unix timestamp in seconds" default(1706745599) example(1706745599)
// @Param interval query string true "Time interval. Short format: 1m, 5m, 15m, 1h, 1d, 1w. Full format: 1Min, 5Min, 15Min, 1Hour, 1Day, 1Week" default(1d) example(1d)
// @Param limit query int false "Maximum number of records to return" default(100) example(100)
// @Success 200 {object} web.Response{data=dto.GetHistoricalDataListResponse}
// @Failure 400 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /trade/historicalData [get]
func (ctl *TradeController) GetHistoricalData(g *gin.Context) {
	ctx := g.Request.Context()

	var req dto.GetHistoricalDataRequest
	if err := g.ShouldBindQuery(&req); err != nil {
		log.ErrorZ(ctx, "failed to bind query parameters", zap.Error(err))
		web.ResponseError(error_msg.ErrInvalidRequestParams, g)
		return
	}

	// Try cache first
	cacheKey := fmt.Sprintf("%s_%d_%d_%s_%d", req.Symbol, req.StartTime, req.EndTime, req.Interval, req.Limit)
	var cached dto.GetHistoricalDataListResponse
	if ctl.tryCache(ctx, "historicalData", cacheKey, &cached) {
		web.ResponseOk(cached, g)
		return
	}

	// Convert Unix timestamps to time.Time
	startTime := utils.UnixToTime(req.StartTime)
	endTime := utils.UnixToTime(req.EndTime)

	serviceReq := trade.GetHistoricalDataRequest{
		Symbol:    req.Symbol,
		StartTime: startTime,
		EndTime:   endTime,
		Interval:  req.Interval,
		Limit:     req.Limit,
	}

	log.InfoZ(ctx, "Getting historical data",
		zap.String("symbol", req.Symbol),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime),
		zap.String("interval", req.Interval))

	historicalData, err := ctl.tradeService.GetHistoricalData(ctx, serviceReq)
	if err != nil {
		log.ErrorZ(ctx, "failed to get historical data",
			zap.String("symbol", req.Symbol),
			zap.Error(err))
		web.ResponseError(error_msg.ErrFailedToGetHistoricalData, g)
		return
	}

	responseData := make([]dto.GetHistoricalDataResponse, 0, len(historicalData))
	for _, item := range historicalData {
		responseData = append(responseData, dto.GetHistoricalDataResponse{
			Open:      item.Open,
			High:      item.High,
			Low:       item.Low,
			Close:     item.Close,
			Volume:    item.Volume,
			Timestamp: utils.TimeToUnix(item.Timestamp),
		})
	}

	response := dto.GetHistoricalDataListResponse{
		Symbol: req.Symbol,
		Data:   responseData,
	}

	_ = ctl.cache.Set(ctx, "historicalData", cacheKey, response, redis_cache.HistoricalDataTTL)

	log.InfoZ(ctx, "Successfully retrieved historical data",
		zap.String("symbol", req.Symbol),
		zap.Int("count", len(responseData)))

	web.ResponseOk(response, g)
}

// GetMarketClock godoc
// @Summary Get Market Clock
// @Description Get the current market clock status including whether the market is open and next open/close times
// @Tags Trade
// @Accept json
// @Produce json
// @Success 200 {object} web.Response{data=dto.GetMarketClockResponse}
// @Failure 500 {object} web.Response
// @Router /trade/marketClock [get]
func (ctl *TradeController) GetMarketClock(g *gin.Context) {
	ctx := g.Request.Context()

	log.InfoZ(ctx, "Getting market clock")

	// Try cache first
	var cached dto.GetMarketClockResponse
	if ctl.tryCache(ctx, "marketClock", "global", &cached) {
		web.ResponseOk(cached, g)
		return
	}

	marketClock, err := ctl.tradeService.GetMarketClock(ctx)
	if err != nil {
		log.ErrorZ(ctx, "failed to get market clock", zap.Error(err))
		web.ResponseError(error_msg.ErrFailedToGetMarketClock, g)
		return
	}

	response := dto.GetMarketClockResponse{
		Timestamp: utils.TimeToUnix(marketClock.Timestamp),
		IsOpen:    marketClock.IsOpen,
		NextOpen:  utils.TimeToUnix(marketClock.NextOpen),
		NextClose: utils.TimeToUnix(marketClock.NextClose),
	}

	_ = ctl.cache.Set(ctx, "marketClock", "global", response, redis_cache.MarketClockTTL)

	log.InfoZ(ctx, "Successfully retrieved market clock",
		zap.Bool("is_open", response.IsOpen),
		zap.Int64("next_open", response.NextOpen),
		zap.Int64("next_close", response.NextClose))

	web.ResponseOk(response, g)
}

// GetLatestQuote godoc
// @Summary Get Latest Quote
// @Description Get the latest bid/ask quote for a given symbol
// @Tags Trade
// @Accept json
// @Produce json
// @Param symbol query string true "Stock symbol" default(AAPL) example(AAPL)
// @Success 200 {object} web.Response{data=dto.GetLatestQuoteResponse}
// @Failure 400 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /trade/latestQuote [get]
func (ctl *TradeController) GetLatestQuote(g *gin.Context) {
	ctx := g.Request.Context()

	var req dto.GetLatestQuoteRequest
	if err := g.ShouldBindQuery(&req); err != nil {
		log.ErrorZ(ctx, "failed to bind query parameters", zap.Error(err))
		web.ResponseError(error_msg.ErrInvalidRequestParams, g)
		return
	}

	log.InfoZ(ctx, "Getting latest quote", zap.String("symbol", req.Symbol))

	// Try cache first
	var cached dto.GetLatestQuoteResponse
	if ctl.tryCache(ctx, "latestQuote", req.Symbol, &cached) {
		web.ResponseOk(cached, g)
		return
	}

	quote, err := ctl.tradeService.GetLatestQuote(ctx, req.Symbol)
	if err != nil {
		log.ErrorZ(ctx, "failed to get latest quote",
			zap.String("symbol", req.Symbol),
			zap.Error(err))
		web.ResponseError(error_msg.ErrFailedToGetLatestQuote, g)
		return
	}

	// Parse timestamp string to time.Time, then convert to Unix timestamp
	quoteTime, err := time.Parse(time.RFC3339, quote.Timestamp)
	if err != nil {
		log.ErrorZ(ctx, "failed to parse quote timestamp", zap.String("timestamp", quote.Timestamp), zap.Error(err))
		web.ResponseError(error_msg.ErrInvalidTimestampFormat, g)
		return
	}

	quoteDTO := dto.QuoteDTO{
		Timestamp:   utils.TimeToUnix(quoteTime),
		BidPrice:    quote.BidPrice,
		BidSize:     quote.BidSize,
		AskPrice:    quote.AskPrice,
		AskSize:     quote.AskSize,
		BidExchange: quote.BidExchange,
		AskExchange: quote.AskExchange,
		Conditions:  quote.Conditions,
		Tape:        quote.Tape,
	}

	response := dto.GetLatestQuoteResponse{
		Quote: quoteDTO,
	}

	_ = ctl.cache.Set(ctx, "latestQuote", req.Symbol, response, redis_cache.LatestQuoteTTL)

	log.InfoZ(ctx, "Successfully retrieved latest quote",
		zap.String("symbol", req.Symbol))

	web.ResponseOk(response, g)
}

// GetSnapshot godoc
// @Summary Get Snapshot
// @Description Get comprehensive market snapshot for a given symbol including latest trade, quote, and bar data
// @Tags Trade
// @Accept json
// @Produce json
// @Param symbol query string true "Stock symbol" default(AAPL) example(AAPL)
// @Success 200 {object} web.Response{data=dto.GetSnapshotResponse}
// @Failure 400 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /trade/snapshot [get]
func (ctl *TradeController) GetSnapshot(g *gin.Context) {
	ctx := g.Request.Context()

	var req dto.GetSnapshotRequest
	if err := g.ShouldBindQuery(&req); err != nil {
		log.ErrorZ(ctx, "failed to bind query parameters", zap.Error(err))
		web.ResponseError(error_msg.ErrInvalidRequestParams, g)
		return
	}

	log.InfoZ(ctx, "Getting snapshot", zap.String("symbol", req.Symbol))

	// Try cache first
	var cached dto.GetSnapshotResponse
	if ctl.tryCache(ctx, "snapshot", req.Symbol, &cached) {
		web.ResponseOk(cached, g)
		return
	}

	snapshot, err := ctl.tradeService.GetSnapshot(ctx, req.Symbol)
	if err != nil {
		log.ErrorZ(ctx, "failed to get snapshot",
			zap.String("symbol", req.Symbol),
			zap.Error(err))
		web.ResponseError(error_msg.ErrFailedToGetSnapshot, g)
		return
	}

	snapshotDTO := dto.SnapshotData{
		Symbol: snapshot.Symbol,
	}

	if snapshot.LatestTrade != nil {
		snapshotDTO.LatestTrade = &dto.TradeDTO{
			Timestamp:  utils.TimeToUnix(snapshot.LatestTrade.Timestamp),
			Price:      snapshot.LatestTrade.Price,
			Size:       snapshot.LatestTrade.Size,
			Exchange:   snapshot.LatestTrade.Exchange,
			ID:         snapshot.LatestTrade.ID,
			Conditions: snapshot.LatestTrade.Conditions,
			Tape:       snapshot.LatestTrade.Tape,
			Update:     snapshot.LatestTrade.Update,
		}
	}

	if snapshot.LatestQuote != nil {
		snapshotDTO.LatestQuote = &dto.QuoteDTO{
			Timestamp:   utils.TimeToUnix(snapshot.LatestQuote.Timestamp),
			BidPrice:    snapshot.LatestQuote.BidPrice,
			BidSize:     snapshot.LatestQuote.BidSize,
			AskPrice:    snapshot.LatestQuote.AskPrice,
			AskSize:     snapshot.LatestQuote.AskSize,
			BidExchange: snapshot.LatestQuote.BidExchange,
			AskExchange: snapshot.LatestQuote.AskExchange,
			Conditions:  snapshot.LatestQuote.Conditions,
			Tape:        snapshot.LatestQuote.Tape,
		}
	}

	if snapshot.MinuteBar != nil {
		snapshotDTO.MinuteBar = &dto.BarDTO{
			Timestamp:  utils.TimeToUnix(snapshot.MinuteBar.Timestamp),
			Open:       snapshot.MinuteBar.Open,
			High:       snapshot.MinuteBar.High,
			Low:        snapshot.MinuteBar.Low,
			Close:      snapshot.MinuteBar.Close,
			Volume:     snapshot.MinuteBar.Volume,
			TradeCount: snapshot.MinuteBar.TradeCount,
			VWAP:       snapshot.MinuteBar.VWAP,
		}
	}

	if snapshot.DailyBar != nil {
		snapshotDTO.DailyBar = &dto.BarDTO{
			Timestamp:  utils.TimeToUnix(snapshot.DailyBar.Timestamp),
			Open:       snapshot.DailyBar.Open,
			High:       snapshot.DailyBar.High,
			Low:        snapshot.DailyBar.Low,
			Close:      snapshot.DailyBar.Close,
			Volume:     snapshot.DailyBar.Volume,
			TradeCount: snapshot.DailyBar.TradeCount,
			VWAP:       snapshot.DailyBar.VWAP,
		}
	}

	if snapshot.PrevDailyBar != nil {
		snapshotDTO.PrevDailyBar = &dto.BarDTO{
			Timestamp:  utils.TimeToUnix(snapshot.PrevDailyBar.Timestamp),
			Open:       snapshot.PrevDailyBar.Open,
			High:       snapshot.PrevDailyBar.High,
			Low:        snapshot.PrevDailyBar.Low,
			Close:      snapshot.PrevDailyBar.Close,
			Volume:     snapshot.PrevDailyBar.Volume,
			TradeCount: snapshot.PrevDailyBar.TradeCount,
			VWAP:       snapshot.PrevDailyBar.VWAP,
		}
	}

	response := dto.GetSnapshotResponse{
		Snapshot: snapshotDTO,
	}

	_ = ctl.cache.Set(ctx, "snapshot", req.Symbol, response, redis_cache.SnapshotTTL)
	web.ResponseOk(response, g)
}

// GetAssets godoc
// @Summary Get Assets
// @Description Get list of assets (stocks, crypto, etc.) with filtering options
// @Tags Trade
// @Accept json
// @Produce json
// @Param status query string false "Asset status filter (active or inactive)" example(active)
// @Param asset_class query string false "Asset class filter (us_equity or crypto)" example(us_equity)
// @Param exchange query string false "Exchange name filter" example(NASDAQ)
// @Success 200 {object} web.Response{data=dto.GetAssetsResponse}
// @Failure 400 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /trade/assets [get]
func (ctl *TradeController) GetAssets(g *gin.Context) {
	ctx := g.Request.Context()

	var req dto.GetAssetsRequest
	if err := g.ShouldBindQuery(&req); err != nil {
		log.ErrorZ(ctx, "failed to bind query parameters", zap.Error(err))
		web.ResponseError(error_msg.ErrInvalidRequestParams, g)
		return
	}

	serviceReq := trade.GetAssetsRequest{
		Status:     req.Status,
		AssetClass: req.AssetClass,
		Exchange:   req.Exchange,
	}

	assets, err := ctl.tradeService.GetAssets(ctx, serviceReq)
	if err != nil {
		log.ErrorZ(ctx, "failed to get assets", zap.Error(err))
		web.ResponseError(error_msg.ErrFailedToGetAssets, g)
		return
	}
	assetDTOs := make([]dto.AssetDTO, len(assets))
	for i, asset := range assets {
		assetDTOs[i] = dto.AssetDTO{
			ID:                           asset.ID,
			Class:                        asset.Class,
			Exchange:                     asset.Exchange,
			Symbol:                       asset.Symbol,
			Name:                         asset.Name,
			Status:                       asset.Status,
			Tradable:                     asset.Tradable,
			Marginable:                   asset.Marginable,
			MaintenanceMarginRequirement: asset.MaintenanceMarginRequirement,
			Shortable:                    asset.Shortable,
			EasyToBorrow:                 asset.EasyToBorrow,
			Fractionable:                 asset.Fractionable,
			Attributes:                   asset.Attributes,
		}
	}

	response := dto.GetAssetsResponse{
		Assets: assetDTOs,
	}

	log.InfoZ(ctx, "Successfully retrieved assets",
		zap.Int("count", len(assetDTOs)))

	web.ResponseOk(response, g)
}

// GetAsset godoc
// @Summary Get Asset
// @Description Get a single asset information by symbol
// @Tags Trade
// @Accept json
// @Produce json
// @Param symbol query string true "Asset symbol" example(AAPL)
// @Success 200 {object} web.Response{data=dto.GetAssetResponse}
// @Failure 400 {object} web.Response
// @Failure 404 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /trade/asset [get]
func (ctl *TradeController) GetAsset(g *gin.Context) {
	ctx := g.Request.Context()

	var req dto.GetAssetRequest
	if err := g.ShouldBindQuery(&req); err != nil {
		log.ErrorZ(ctx, "failed to bind query parameters", zap.Error(err))
		web.ResponseError(error_msg.ErrInvalidRequestParams, g)
		return
	}

	if req.Symbol == "" {
		log.ErrorZ(ctx, "symbol is empty")
		web.ResponseError(error_msg.ErrSymbolRequired, g)
		return
	}

	log.InfoZ(ctx, "Getting asset", zap.String("symbol", req.Symbol))

	asset, err := ctl.tradeService.GetAsset(ctx, req.Symbol)
	if err != nil {
		log.ErrorZ(ctx, "failed to get asset",
			zap.String("symbol", req.Symbol),
			zap.Error(err))
		web.ResponseError(error_msg.ErrFailedToGetAsset, g)
		return
	}

	assetDTO := dto.AssetDTO{
		ID:                           asset.ID,
		Class:                        asset.Class,
		Exchange:                     asset.Exchange,
		Symbol:                       asset.Symbol,
		Name:                         asset.Name,
		Status:                       asset.Status,
		Tradable:                     asset.Tradable,
		Marginable:                   asset.Marginable,
		MaintenanceMarginRequirement: asset.MaintenanceMarginRequirement,
		Shortable:                    asset.Shortable,
		EasyToBorrow:                 asset.EasyToBorrow,
		Fractionable:                 asset.Fractionable,
		Attributes:                   asset.Attributes,
	}

	response := dto.GetAssetResponse{
		Asset: assetDTO,
	}

	log.InfoZ(ctx, "Successfully retrieved asset",
		zap.String("symbol", req.Symbol),
		zap.String("name", assetDTO.Name))

	web.ResponseOk(response, g)
}
