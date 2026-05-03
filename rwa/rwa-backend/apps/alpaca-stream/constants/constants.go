package constants

const (
	StreamTypeTradeUpdates = "trade_updates"
	StreamTypeBars         = "bars"
	StreamTypeQuotes       = "quotes"
	StreamTypeTrades       = "trades"

	// StreamName 常量用于 WebSocket 消息中的 stream 字段
	StreamNameAuthorization = "authorization"
	StreamNameListening     = "listening"
)

const (
	EventTypeNew            = "new"
	EventTypeFill           = "fill"
	EventTypePartialFill    = "partial_fill"
	EventTypeCanceled       = "canceled"
	EventTypeExpired        = "expired"
	EventTypeRejected       = "rejected"
	EventTypeReplaced       = "replaced"
	EventTypePendingNew     = "pending_new"
	EventTypePendingCancel  = "pending_cancel"
	EventTypePendingReplace = "pending_replace"
	EventTypeCancelRejected = "cancel_rejected"
	EventTypeDoneForDay      = "done_for_day"
	EventTypeReplaceRejected = "replace_rejected"
)

const (
	FeedIEX = "iex"
	FeedSIP = "sip"
)

const EnableTradeUpdates = true

const (
	DefaultMarketDataFeed = FeedIEX
)

const (
	DefaultReconnectDelay      = 1
	DefaultMaxReconnectDelay   = 30
	DefaultMaxReconnectAttempts = -1
)

const (
	AuthTimeout   = 5
	ReadDeadline  = 300 // 5 分钟，避免非交易时段频繁断连
	WriteDeadline = 10
	PingInterval  = 120 // 2 分钟发送一次 ping 保活
)

// WebSocket 消息类型常量（数组格式中的 T 字段）
const (
	MessageTypeSuccess     = "success"
	MessageTypeError       = "error"
	MessageTypeBars        = "b"
	MessageTypeQuotes      = "q"
	MessageTypeTrades      = "t"
	MessageTypeSubscription = "subscription"
)

// JSON 字段名常量
const (
	FieldType         = "T"
	FieldMessage      = "msg"
	FieldStream       = "stream"
	FieldData         = "data"
	FieldAction       = "action"
	FieldStatus       = "status"
	FieldErrorMessage = "error_message"
	FieldStreams      = "streams"
)

// 状态值常量
const (
	StatusAuthorized  = "authorized"
	MessageAuthenticated = "authenticated"
	ActionError       = "error"
)

