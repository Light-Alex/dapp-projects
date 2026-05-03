package service

import (
	"context"
	"fmt"
	"time"

	"github.com/AnchoredLabs/rwa-backend/apps/alpaca-stream/config"
	"github.com/AnchoredLabs/rwa-backend/apps/alpaca-stream/constants"
	"github.com/AnchoredLabs/rwa-backend/apps/alpaca-stream/handlers"
	"github.com/AnchoredLabs/rwa-backend/apps/alpaca-stream/ws"
	"github.com/AnchoredLabs/rwa-backend/libs/core/kafka_help"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type AlpacaWebSocketService struct {
	config             *config.Config
	orderSyncService   *OrderSyncService
	barKafkaService    *kafka_help.BarKafkaService
	tradeUpdatesClient *ws.Client
	marketDataClient   *ws.Client
	subscriptionMgr    *ws.SubscriptionManager
	marketDataSubMgr   *ws.SubscriptionManager
}

func NewAlpacaWebSocketService(
	lc fx.Lifecycle,
	conf *config.Config,
	orderSyncService *OrderSyncService,
	barKafkaService *kafka_help.BarKafkaService,
) *AlpacaWebSocketService {
	service := &AlpacaWebSocketService{
		config:           conf,
		orderSyncService: orderSyncService,
		barKafkaService:  barKafkaService,
	}

	lc.Append(fx.Hook{
		OnStart: service.StartStreaming,
		OnStop:  service.StopStreaming,
	})

	return service
}

func (s *AlpacaWebSocketService) StartStreaming(ctx context.Context) error {
	enableMarketData := s.config.Alpaca != nil && s.config.Alpaca.EnableMarketData
	log.InfoZ(ctx, "Starting Alpaca WebSocket streaming service",
		zap.Bool("enable_trade_updates", constants.EnableTradeUpdates),
		zap.Bool("enable_market_data", enableMarketData),
	)

	if constants.EnableTradeUpdates {
		log.InfoZ(ctx, "Trade updates enabled, setting up client...")
		if err := s.setupTradeUpdatesClient(ctx); err != nil {
			log.ErrorZ(ctx, "Failed to setup trade updates client", zap.Error(err))
			return fmt.Errorf("failed to setup trade updates client: %w", err)
		}
	} else {
		log.WarnZ(ctx, "Trade updates not enabled, cannot listen to order events")
	}

	if enableMarketData {
		symbols := s.config.Alpaca.Symbols
		if len(symbols) == 0 {
			log.WarnZ(ctx, "Market data enabled but no symbols configured, skipping")
		} else if err := s.setupMarketDataClient(ctx, symbols, constants.DefaultMarketDataFeed, true, false, false); err != nil {
			log.WarnZ(ctx, "Failed to setup market data client", zap.Error(err))
		}
	}

	log.InfoZ(ctx, "Alpaca WebSocket streaming service started")
	return nil
}

func (s *AlpacaWebSocketService) StopStreaming(ctx context.Context) error {
	log.InfoZ(ctx, "Stopping Alpaca WebSocket streaming service")

	if s.tradeUpdatesClient != nil {
		if err := s.tradeUpdatesClient.Close(); err != nil {
			log.ErrorZ(ctx, "Failed to close trade updates client", zap.Error(err))
		}
	}

	if s.marketDataClient != nil {
		if err := s.marketDataClient.Close(); err != nil {
			log.ErrorZ(ctx, "Failed to close market data client", zap.Error(err))
		}
	}

	return nil
}

func (s *AlpacaWebSocketService) setupTradeUpdatesClient(ctx context.Context) error {
	wsURL := s.config.Alpaca.WSURL
	if wsURL == "" {
		return fmt.Errorf("alpaca ws url is empty")
	}

	apiKeyPrefix := s.config.Alpaca.APIKey
	if len(apiKeyPrefix) > 8 {
		apiKeyPrefix = apiKeyPrefix[:8]
	}
	log.InfoZ(ctx, "Preparing to connect to trade updates WebSocket",
		zap.String("ws_url", wsURL),
		zap.String("api_key_prefix", apiKeyPrefix+"..."),
	)

	client := ws.NewClient(
		s.config.Alpaca.APIKey,
		s.config.Alpaca.APISecret,
		wsURL,
	)

	client.SetErrorHandler(func(err error) {
		log.ErrorZ(ctx, "WebSocket error", zap.Error(err))
	})

	// Connect
	log.InfoZ(ctx, "Connecting to WebSocket server...")
	if err := client.Connect(ctx); err != nil {
		log.ErrorZ(ctx, "WebSocket connection failed", zap.Error(err))
		return fmt.Errorf("failed to connect: %w", err)
	}
	log.InfoZ(ctx, "WebSocket connected successfully")

	// Setup subscription manager
	subscriptionMgr := ws.NewSubscriptionManager(client)

	// Setup trade updates handler
	tradeUpdatesHandler := handlers.NewTradeUpdatesHandler()
	tradeUpdatesHandler.SetEventHandlers(
		s.onTradeUpdateNew,
		s.onTradeUpdateFill,
		s.onTradeUpdatePartialFill,
		s.onTradeUpdateCanceled,
		s.onTradeUpdateExpired,
		s.onTradeUpdateRejected,
		s.onTradeUpdateReplaced,
		s.onTradeUpdateDoneForDay,
	)

	// Subscribe to trade updates
	log.InfoZ(ctx, "Preparing to subscribe to trade updates stream", zap.String("stream_type", constants.StreamTypeTradeUpdates))
	subConfig := &ws.SubscriptionConfig{
		Type:    constants.StreamTypeTradeUpdates,
		Symbols: []string{},
		Handler: tradeUpdatesHandler.Handle,
	}

	if err := subscriptionMgr.Subscribe(ctx, subConfig); err != nil {
		log.ErrorZ(ctx, "Failed to subscribe to trade updates", zap.Error(err))
		return fmt.Errorf("failed to subscribe to trade updates: %w", err)
	}

	s.tradeUpdatesClient = client
	s.subscriptionMgr = subscriptionMgr

	// 注册重连回调，确保 WebSocket 重连后自动重新订阅
	client.SetReconnectHandler(func(reconnCtx context.Context) {
		if err := subscriptionMgr.Resubscribe(reconnCtx); err != nil {
			log.ErrorZ(reconnCtx, "Failed to resubscribe after reconnect", zap.Error(err))
		}
	})

	log.InfoZ(ctx, "Trade updates client connected and subscribed successfully, now listening to all order events (including new orders)")
	return nil
}

func (s *AlpacaWebSocketService) setupMarketDataClient(ctx context.Context, symbols []string, feed string, enableBars, enableQuotes, enableTrades bool) error {
	if len(symbols) == 0 {
		return fmt.Errorf("market data enabled but no symbols provided")
	}

	if feed == "" {
		feed = constants.DefaultMarketDataFeed
	}

	wsDataURL := s.config.Alpaca.WSDataURL
	if wsDataURL == "" {
		wsDataURL = fmt.Sprintf("wss://stream.data.alpaca.markets/v2/%s", feed)
	}

	client := ws.NewClient(
		s.config.Alpaca.APIKey,
		s.config.Alpaca.APISecret,
		wsDataURL,
	)

	client.SetErrorHandler(func(err error) {
		log.ErrorZ(ctx, "Market data WebSocket error", zap.Error(err))
	})

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	subscriptionMgr := ws.NewSubscriptionManager(client)

	if enableBars {
		barsHandler := handlers.NewBarsHandler()
		barsHandler.SetBarHandler(s.onBar)

		client.RegisterHandler(constants.StreamTypeBars, barsHandler.Handle)

		subscribeMsg := map[string]any{
			"action": "subscribe",
			"bars":   symbols,
		}
		if err := client.WriteJSON(ctx, subscribeMsg); err != nil {
			return fmt.Errorf("failed to subscribe to bars: %w", err)
		}

		log.InfoZ(ctx, "Subscribed to bars", zap.Strings("symbols", symbols))
	}

	// Subscribe to quotes if enabled
	if enableQuotes {
		subscribeMsg := map[string]any{
			"action": "subscribe",
			"quotes": symbols,
		}
		if err := client.WriteJSON(ctx, subscribeMsg); err != nil {
			return fmt.Errorf("failed to subscribe to quotes: %w", err)
		}
		log.InfoZ(ctx, "Subscribed to quotes", zap.Strings("symbols", symbols))
	}

	// Subscribe to trades if enabled
	if enableTrades {
		subscribeMsg := map[string]any{
			"action": "subscribe",
			"trades": symbols,
		}
		if err := client.WriteJSON(ctx, subscribeMsg); err != nil {
			return fmt.Errorf("failed to subscribe to trades: %w", err)
		}
		log.InfoZ(ctx, "Subscribed to trades", zap.Strings("symbols", symbols))
	}

	s.marketDataClient = client
	s.marketDataSubMgr = subscriptionMgr

	log.InfoZ(ctx, "Market data client connected and subscribed")
	return nil
}

// Trade update event handlers
func (s *AlpacaWebSocketService) onTradeUpdateNew(ctx context.Context, data handlers.TradeUpdateMessageData) {
	s.orderSyncService.HandleNew(ctx, data)
}

func (s *AlpacaWebSocketService) onTradeUpdateFill(ctx context.Context, data handlers.TradeUpdateMessageData) {
	s.orderSyncService.HandleFill(ctx, data)
}

func (s *AlpacaWebSocketService) onTradeUpdatePartialFill(ctx context.Context, data handlers.TradeUpdateMessageData) {
	s.orderSyncService.HandlePartialFill(ctx, data)
}

func (s *AlpacaWebSocketService) onTradeUpdateCanceled(ctx context.Context, data handlers.TradeUpdateMessageData) {
	s.orderSyncService.HandleCanceled(ctx, data)
}

func (s *AlpacaWebSocketService) onTradeUpdateExpired(ctx context.Context, data handlers.TradeUpdateMessageData) {
	s.orderSyncService.HandleExpired(ctx, data)
}

func (s *AlpacaWebSocketService) onTradeUpdateRejected(ctx context.Context, data handlers.TradeUpdateMessageData) {
	s.orderSyncService.HandleRejected(ctx, data)
}

func (s *AlpacaWebSocketService) onTradeUpdateDoneForDay(ctx context.Context, data handlers.TradeUpdateMessageData) {
	s.orderSyncService.HandleDoneForDay(ctx, data)
}

func (s *AlpacaWebSocketService) onTradeUpdateReplaced(ctx context.Context, data handlers.TradeUpdateMessageData) {
	log.DebugZ(ctx, "Order replaced event received, no handler implemented yet",
		zap.String("order_id", data.Order.ID))
}

// Market data handlers
func (s *AlpacaWebSocketService) onBar(ctx context.Context, symbol string, bar handlers.BarData) {
	log.DebugZ(ctx, "Received bar",
		zap.String("symbol", symbol),
		zap.Float64("close", bar.Close),
		zap.Int64("volume", bar.Volume),
	)

	// 解析 Alpaca 返回的 RFC3339 时间戳字符串
	timestamp, err := time.Parse(time.RFC3339, bar.Timestamp)
	if err != nil {
		log.ErrorZ(ctx, "Failed to parse bar timestamp", zap.String("timestamp", bar.Timestamp), zap.Error(err))
		return
	}

	s.barKafkaService.Publish(ctx, &kafka_help.BarEvent{
		Symbol:     symbol,
		Open:       bar.Open,
		High:       bar.High,
		Low:        bar.Low,
		Close:      bar.Close,
		Volume:     bar.Volume,
		Timestamp:  timestamp.Unix(),
		TradeCount: bar.TradeCount,
		VWAP:       bar.VWAP,
	})
}
