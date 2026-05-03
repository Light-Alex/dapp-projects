package handlers

import (
	"context"
	"encoding/json"

	"github.com/AnchoredLabs/rwa-backend/apps/alpaca-stream/constants"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"go.uber.org/zap"
)

// TradeUpdateMessage represents a trade update message
type TradeUpdateMessage struct {
	Stream string                 `json:"stream"`
	Data   TradeUpdateMessageData `json:"data"`
}

// AlpacaOrderData represents the order object from Alpaca's trade update WebSocket.
// Field names use snake_case to match Alpaca's JSON format.
type AlpacaOrderData struct {
	ID             string `json:"id"`
	ClientOrderID  string `json:"client_order_id"`
	Status         string `json:"status"`
	Symbol         string `json:"symbol"`
	Qty            string `json:"qty"`
	FilledQty      string `json:"filled_qty"`
	FilledAvgPrice string `json:"filled_avg_price"`
	Side           string `json:"side"`
	Type           string `json:"type"`
	TimeInForce    string `json:"time_in_force"`
	LimitPrice     string `json:"limit_price"`
	StopPrice      string `json:"stop_price"`
	RejectReason   string `json:"reject_reason,omitempty"`
}

// TradeUpdateMessageData contains the trade update data
type TradeUpdateMessageData struct {
	Event       string          `json:"event"`
	ExecutionID string          `json:"execution_id"`
	Order       AlpacaOrderData `json:"order"`
	Timestamp   string          `json:"timestamp"`
	Price       string          `json:"price,omitempty"`
	Qty         string          `json:"qty,omitempty"`
	PositionQty string          `json:"position_qty,omitempty"`
}

// TradeUpdatesHandler handles trade update messages
type TradeUpdatesHandler struct {
	onNew             func(ctx context.Context, data TradeUpdateMessageData)
	onFill            func(ctx context.Context, data TradeUpdateMessageData)
	onPartialFill     func(ctx context.Context, data TradeUpdateMessageData)
	onCanceled        func(ctx context.Context, data TradeUpdateMessageData)
	onExpired         func(ctx context.Context, data TradeUpdateMessageData)
	onRejected        func(ctx context.Context, data TradeUpdateMessageData)
	onReplaced        func(ctx context.Context, data TradeUpdateMessageData)
	onDoneForDay      func(ctx context.Context, data TradeUpdateMessageData)
	onPendingNew      func(ctx context.Context, data TradeUpdateMessageData)
	onPendingCancel   func(ctx context.Context, data TradeUpdateMessageData)
	onPendingReplace  func(ctx context.Context, data TradeUpdateMessageData)
	onCancelRejected  func(ctx context.Context, data TradeUpdateMessageData)
	onReplaceRejected func(ctx context.Context, data TradeUpdateMessageData)
}

// NewTradeUpdatesHandler creates a new trade updates handler
func NewTradeUpdatesHandler() *TradeUpdatesHandler {
	return &TradeUpdatesHandler{}
}

// SetEventHandlers sets event handlers
func (h *TradeUpdatesHandler) SetEventHandlers(
	onNew, onFill, onPartialFill, onCanceled, onExpired, onRejected, onReplaced, onDoneForDay func(ctx context.Context, data TradeUpdateMessageData),
) {
	h.onNew = onNew
	h.onFill = onFill
	h.onPartialFill = onPartialFill
	h.onCanceled = onCanceled
	h.onExpired = onExpired
	h.onRejected = onRejected
	h.onReplaced = onReplaced
	h.onDoneForDay = onDoneForDay
}

// SetAdvancedEventHandlers sets advanced event handlers
func (h *TradeUpdatesHandler) SetAdvancedEventHandlers(
	onPendingNew, onPendingCancel, onPendingReplace, onCancelRejected, onReplaceRejected func(ctx context.Context, data TradeUpdateMessageData),
) {
	h.onPendingNew = onPendingNew
	h.onPendingCancel = onPendingCancel
	h.onPendingReplace = onPendingReplace
	h.onCancelRejected = onCancelRejected
	h.onReplaceRejected = onReplaceRejected
}

// Handle 处理 trade_updates 流的消息，根据事件类型分发到对应的处理函数
//
// 订单生命周期事件：
//   - pending_new:      新订单待交易所确认，触发 onPendingNew
//   - new:              订单已被交易所接收，触发 onNew
//   - rejected:         订单被拒绝（如资金不足），触发 onRejected
//   - partial_fill:     订单部分成交，触发 onPartialFill
//   - fill:             订单完全成交，触发 onFill
//   - pending_cancel:   取消请求待确认，触发 onPendingCancel
//   - cancel_rejected:  取消请求被拒绝，触发 onCancelRejected
//   - canceled:         订单已取消，触发 onCanceled
//   - pending_replace:  修改请求待确认，触发 onPendingReplace
//   - replace_rejected: 修改请求被拒绝，触发 onReplaceRejected
//   - replaced:         订单被修改替换，触发 onReplaced
//   - expired:          订单已过期，触发 onExpired
//   - done_for_day:     当日交易结束，触发 onDoneForDay
//
// 未知事件类型会记录警告日志但不会返回错误
func (h *TradeUpdatesHandler) Handle(ctx context.Context, message json.RawMessage) error {
	var msg TradeUpdateMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		return err
	}

	if msg.Stream != constants.StreamTypeTradeUpdates {
		return nil
	}

	data := msg.Data

	// Route to appropriate handler based on event type
	switch data.Event {
	case constants.EventTypeNew:
		if h.onNew != nil {
			h.onNew(ctx, data)
		}
	case constants.EventTypeFill:
		if h.onFill != nil {
			h.onFill(ctx, data)
		}
	case constants.EventTypePartialFill:
		if h.onPartialFill != nil {
			h.onPartialFill(ctx, data)
		}
	case constants.EventTypeCanceled:
		if h.onCanceled != nil {
			h.onCanceled(ctx, data)
		}
	case constants.EventTypeExpired:
		if h.onExpired != nil {
			h.onExpired(ctx, data)
		}
	case constants.EventTypeRejected:
		if h.onRejected != nil {
			h.onRejected(ctx, data)
		}
	case constants.EventTypeReplaced:
		if h.onReplaced != nil {
			h.onReplaced(ctx, data)
		}
	case constants.EventTypeDoneForDay:
		if h.onDoneForDay != nil {
			h.onDoneForDay(ctx, data)
		}
	case constants.EventTypePendingNew:
		if h.onPendingNew != nil {
			h.onPendingNew(ctx, data)
		}
	case constants.EventTypePendingCancel:
		if h.onPendingCancel != nil {
			h.onPendingCancel(ctx, data)
		}
	case constants.EventTypePendingReplace:
		if h.onPendingReplace != nil {
			h.onPendingReplace(ctx, data)
		}
	case constants.EventTypeCancelRejected:
		if h.onCancelRejected != nil {
			h.onCancelRejected(ctx, data)
		}
	case constants.EventTypeReplaceRejected:
		if h.onReplaceRejected != nil {
			h.onReplaceRejected(ctx, data)
		}
	default:
		log.WarnZ(ctx, "Unknown trade update event type",
			zap.String("event", data.Event),
			zap.String("execution_id", data.ExecutionID),
			zap.String("order_id", data.Order.ID),
			zap.String("raw_message", string(message)),
		)
	}
	return nil
}
