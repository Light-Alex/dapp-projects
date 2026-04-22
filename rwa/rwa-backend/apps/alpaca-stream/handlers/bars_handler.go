package handlers

import (
	"context"
	"encoding/json"

	"github.com/AnchoredLabs/rwa-backend/apps/alpaca-stream/constants"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"go.uber.org/zap"
)

// BarMessage represents a bar (candle) message
type BarMessage struct {
	Stream string       `json:"stream"`
	Data   []BarData    `json:"data"`
}

// BarData represents a single bar
type BarData struct {
	Symbol    string  `json:"S"` // Symbol
	Open      float64 `json:"o"` // Open price
	High      float64 `json:"h"` // High price
	Low       float64 `json:"l"` // Low price
	Close     float64 `json:"c"` // Close price
	Volume    int64   `json:"v"` // Volume
	Timestamp int64   `json:"t"` // Timestamp (Unix nanoseconds)
	TradeCount int64  `json:"n,omitempty"` // Trade count
	VWAP      float64 `json:"vw,omitempty"` // Volume weighted average price
}

// BarsHandler handles bar messages
type BarsHandler struct {
	onBar func(ctx context.Context, symbol string, bar BarData)
}

// NewBarsHandler creates a new bars handler
func NewBarsHandler() *BarsHandler {
	return &BarsHandler{}
}

// SetBarHandler sets the bar handler
func (h *BarsHandler) SetBarHandler(handler func(ctx context.Context, symbol string, bar BarData)) {
	h.onBar = handler
}

// Handle handles a bar message
func (h *BarsHandler) Handle(ctx context.Context, message json.RawMessage) error {
	var msg BarMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		return err
	}

	if msg.Stream != constants.StreamTypeBars {
		return nil
	}

	for _, bar := range msg.Data {
		if h.onBar != nil {
			h.onBar(ctx, bar.Symbol, bar)
		}

		log.DebugZ(ctx, "Received bar data",
			zap.String("symbol", bar.Symbol),
			zap.Float64("open", bar.Open),
			zap.Float64("high", bar.High),
			zap.Float64("low", bar.Low),
			zap.Float64("close", bar.Close),
			zap.Int64("volume", bar.Volume),
			zap.Int64("timestamp", bar.Timestamp),
		)
	}

	return nil
}

