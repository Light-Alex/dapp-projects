package types

import (
	"context"
	"encoding/json"

	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"go.uber.org/zap"
)

type WsIncomeMessageType string

const (
	WsIncomeMessageTypeSub   WsIncomeMessageType = "SUBSCRIBE"
	WsIncomeMessageTypeUnsub WsIncomeMessageType = "UNSUBSCRIBE"
)

type WsIncomeMessage struct {
	Id     uint64              `json:"id"`
	Method WsIncomeMessageType `json:"method"`
	Params map[string]any      `json:"params"`
}

type WsResult struct {
	Id     uint64 `json:"id"`
	Result any    `json:"result"`
}

type WsStream struct {
	Stream WsStreamType `json:"stream"`
	Data   any          `json:"data"`
}

func (s WsStream) ToByte() []byte {
	bytes, err := json.Marshal(s)
	if err != nil {
		log.ErrorZ(context.Background(), "WsStream.ToByte: marshal failed", zap.Error(err))
		return nil
	}
	return bytes
}

func (r *WsResult) ToByte() []byte {
	bytes, err := json.Marshal(r)
	if err != nil {
		log.ErrorZ(context.Background(), "WsResult.ToByte: marshal failed", zap.Error(err))
		return nil
	}
	return bytes
}

type WsSubType string

const (
	WsSubTypeBar   WsSubType = "bar"
	WsSubTypeOrder WsSubType = "order"
)

type WsStreamType string

const (
	WsStreamTypeBar   WsStreamType = "bar"
	WsStreamTypeOrder WsStreamType = "order"
)


