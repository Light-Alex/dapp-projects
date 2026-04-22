package trade

import (
	"context"
	"time"

	"github.com/AnchoredLabs/rwa-backend/libs/errors"
	"github.com/AnchoredLabs/rwa-backend/libs/log"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"go.uber.org/zap"
)

type alpacaMarketDataService struct {
	dataClient *marketdata.Client
}

func newAlpacaMarketDataService(dataClient *marketdata.Client) *alpacaMarketDataService {
	return &alpacaMarketDataService{
		dataClient: dataClient,
	}
}

func (s *alpacaMarketDataService) GetCurrentPrice(ctx context.Context, symbol string) (*MarketDataResponse, error) {
	if err := validateSymbol(symbol); err != nil {
		log.ErrorZ(ctx, "Symbol validation failed", zap.String("symbol", symbol), zap.Error(err))
		return nil, err
	}

	trades, err := s.dataClient.GetLatestTrades([]string{symbol}, marketdata.GetLatestTradeRequest{})
	if err != nil {
		log.ErrorZ(ctx, "Failed to get current price", zap.String("symbol", symbol), zap.Error(err))
		return nil, errors.Errorf("get current price failed: %v", err)
	}

	trade, exists := trades[symbol]
	if !exists {
		log.ErrorZ(ctx, "No data found for symbol", zap.String("symbol", symbol))
		return nil, errors.Errorf("no data for symbol %s", symbol)
	}

	log.InfoZ(ctx, "Current price retrieved from latest trade",
		zap.String("symbol", symbol),
		zap.Float64("price", trade.Price),
		zap.Uint32("size", trade.Size))

	return &MarketDataResponse{
		Symbol:    symbol,
		Price:     trade.Price,
		Volume:    float64(trade.Size),
		Timestamp: trade.Timestamp,
	}, nil
}

func (s *alpacaMarketDataService) GetHistoricalData(ctx context.Context, req GetHistoricalDataRequest) ([]HistoricalDataResponse, error) {
	log.InfoZ(ctx, "Getting historical data",
		zap.String("interval", req.Interval),
		zap.String("start", req.StartTime.Format("2006-01-02")),
		zap.String("end", req.EndTime.Format("2006-01-02")))

	normalizedInterval, err := normalizeInterval(req.Interval)
	if err != nil {
		log.ErrorZ(ctx, "Time interval validation failed", zap.String("interval", req.Interval), zap.Error(err))
		return nil, err
	}

	var alpacaInterval marketdata.TimeFrame
	switch normalizedInterval {
	case "1m":
		alpacaInterval = marketdata.NewTimeFrame(1, marketdata.Min)
	case "5m":
		alpacaInterval = marketdata.NewTimeFrame(5, marketdata.Min)
	case "15m":
		alpacaInterval = marketdata.NewTimeFrame(15, marketdata.Min)
	case "1h":
		alpacaInterval = marketdata.NewTimeFrame(1, marketdata.Hour)
	case "1d":
		alpacaInterval = marketdata.NewTimeFrame(1, marketdata.Day)
	case "1w":
		alpacaInterval = marketdata.NewTimeFrame(1, marketdata.Week)
	case "1M":
		alpacaInterval = marketdata.NewTimeFrame(1, marketdata.Month)
	default:
		return nil, errors.Errorf("unsupported time interval: %s", req.Interval)
	}

	request := marketdata.GetBarsRequest{
		TimeFrame:  alpacaInterval,
		Start:      req.StartTime,
		End:        req.EndTime,
		TotalLimit: req.Limit,
	}

	bars, err := s.dataClient.GetBars(req.Symbol, request)
	if err != nil {
		log.ErrorZ(ctx, "Failed to get historical data", zap.Error(err))
		return nil, errors.Errorf("get historical data failed: %v", err)
	}

	var historicalData []HistoricalDataResponse
	for _, bar := range bars {
		historicalData = append(historicalData, HistoricalDataResponse{
			Symbol:    req.Symbol,
			Open:      bar.Open,
			High:      bar.High,
			Low:       bar.Low,
			Close:     bar.Close,
			Volume:    bar.Volume,
			Timestamp: bar.Timestamp,
		})
	}

	return historicalData, nil
}

func (s *alpacaMarketDataService) GetLatestQuote(ctx context.Context, symbol string) (*QuoteResponse, error) {
	log.InfoZ(ctx, "Getting latest quote", zap.String("symbol", symbol))

	if err := validateSymbol(symbol); err != nil {
		log.ErrorZ(ctx, "Symbol validation failed", zap.String("symbol", symbol), zap.Error(err))
		return nil, err
	}

	quote, err := s.dataClient.GetLatestQuote(symbol, marketdata.GetLatestQuoteRequest{})
	if err != nil {
		log.ErrorZ(ctx, "Failed to get latest quote", zap.String("symbol", symbol), zap.Error(err))
		return nil, errors.Errorf("get latest quote failed: %v", err)
	}

	if quote == nil {
		log.ErrorZ(ctx, "No data found for symbol", zap.String("symbol", symbol))
		return nil, errors.Errorf("no data for symbol %s", symbol)
	}

	midPrice := (quote.BidPrice + quote.AskPrice) / 2.0

	result := &QuoteResponse{
		Symbol:      symbol,
		Timestamp:   quote.Timestamp.Format(time.RFC3339),
		BidPrice:    quote.BidPrice,
		BidSize:     quote.BidSize,
		AskPrice:    quote.AskPrice,
		AskSize:     quote.AskSize,
		BidExchange: quote.BidExchange,
		AskExchange: quote.AskExchange,
		Conditions:  quote.Conditions,
		Tape:        quote.Tape,
		MidPrice:    midPrice,
	}

	log.InfoZ(ctx, "Successfully retrieved latest quote", zap.String("symbol", symbol))
	return result, nil
}

func (s *alpacaMarketDataService) GetSnapshot(ctx context.Context, symbol string) (*SnapshotResponse, error) {
	log.InfoZ(ctx, "Getting snapshot", zap.String("symbol", symbol))

	if err := validateSymbol(symbol); err != nil {
		log.ErrorZ(ctx, "Symbol validation failed", zap.String("symbol", symbol), zap.Error(err))
		return nil, err
	}

	snapshot, err := s.dataClient.GetSnapshot(symbol, marketdata.GetSnapshotRequest{})
	if err != nil {
		log.ErrorZ(ctx, "Failed to get snapshot", zap.String("symbol", symbol), zap.Error(err))
		return nil, errors.Errorf("get snapshot failed: %v", err)
	}

	if snapshot == nil {
		log.ErrorZ(ctx, "No data found for symbol", zap.String("symbol", symbol))
		return nil, errors.Errorf("no data for symbol %s", symbol)
	}

	snapshotResp := &SnapshotResponse{
		Symbol: symbol,
	}

	if snapshot.LatestTrade != nil {
		snapshotResp.LatestTrade = &TradeData{
			Timestamp:  snapshot.LatestTrade.Timestamp,
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
		snapshotResp.LatestQuote = &QuoteData{
			Timestamp:   snapshot.LatestQuote.Timestamp,
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
		snapshotResp.MinuteBar = &BarData{
			Timestamp:  snapshot.MinuteBar.Timestamp,
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
		snapshotResp.DailyBar = &BarData{
			Timestamp:  snapshot.DailyBar.Timestamp,
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
		snapshotResp.PrevDailyBar = &BarData{
			Timestamp:  snapshot.PrevDailyBar.Timestamp,
			Open:       snapshot.PrevDailyBar.Open,
			High:       snapshot.PrevDailyBar.High,
			Low:        snapshot.PrevDailyBar.Low,
			Close:      snapshot.PrevDailyBar.Close,
			Volume:     snapshot.PrevDailyBar.Volume,
			TradeCount: snapshot.PrevDailyBar.TradeCount,
			VWAP:       snapshot.PrevDailyBar.VWAP,
		}
	}

	log.InfoZ(ctx, "Successfully retrieved snapshot", zap.String("symbol", symbol))
	return snapshotResp, nil
}
