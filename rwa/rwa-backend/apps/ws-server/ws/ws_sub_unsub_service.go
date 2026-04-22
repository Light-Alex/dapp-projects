package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AnchoredLabs/rwa-backend/apps/ws-server/types"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/olahol/melody"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type SubUnsubService struct {
	wsServer *Server
}

func NewSubUnsubService(wsServer *Server, lc fx.Lifecycle) *SubUnsubService {
	s := &SubUnsubService{
		wsServer: wsServer,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			s.bindEvent(context.WithValue(context.Background(), log.TraceID, "ws_sub_unsub_service"))
			return nil
		},
	})
	return s
}

func (s *SubUnsubService) bindEvent(ctx context.Context) {
	s.wsServer.GetMelody().HandleConnect(func(session *melody.Session) {
		log.InfoZ(ctx, "ws client connected", zap.String("client", session.Request.RemoteAddr))
	})
	s.wsServer.GetMelody().HandleDisconnect(func(session *melody.Session) {
		log.InfoZ(ctx, "ws client disconnected", zap.String("client", session.Request.RemoteAddr))
	})
	s.wsServer.GetMelody().HandleMessage(func(session *melody.Session, msg []byte) {
		if string(msg) == "ping" {
			_ = session.Write([]byte("pong"))
			return
		}
		var message *types.WsIncomeMessage
		if err := json.Unmarshal(msg, &message); err != nil {
			return
		}
		switch message.Method {
		case types.WsIncomeMessageTypeSub:
			s.handleMessage(ctx, session, message, true)
		case types.WsIncomeMessageTypeUnsub:
			s.handleMessage(ctx, session, message, false)
		}
	})
}

func (s *SubUnsubService) handleMessage(ctx context.Context, session *melody.Session, message *types.WsIncomeMessage, isSub bool) {
	t, ok := message.Params["type"].(string)
	if !ok {
		return
	}
	switch types.WsSubType(t) {
	case types.WsSubTypeBar:
		symbols, _ := toStringSlice(message.Params["symbols"])
		if len(symbols) == 0 {
			return
		}
		for _, symbol := range symbols {
			key := fmt.Sprintf("bar_%s", strings.ToUpper(symbol))
			s.setSession(ctx, session, key, message.Id, isSub)
		}
	case types.WsSubTypeOrder:
		accountID, ok := message.Params["account_id"].(float64)
		if !ok || accountID <= 0 {
			return
		}
		key := fmt.Sprintf("order_%d", uint64(accountID))
		s.setSession(ctx, session, key, message.Id, isSub)
	}
}

func (s *SubUnsubService) setSession(ctx context.Context, session *melody.Session, key string, msgId uint64, isSub bool) {
	if isSub {
		session.Set(key, true)
	} else {
		session.UnSet(key)
	}
	log.InfoZ(ctx, "ws sub/unsub", zap.Bool("sub", isSub), zap.String("client", session.Request.RemoteAddr), zap.String("key", key))
	res := &types.WsResult{
		Id:     msgId,
		Result: "success",
	}
	_ = session.Write(res.ToByte())
}

// toStringSlice extracts []string from an interface{} (typically []any from JSON).
func toStringSlice(v any) ([]string, bool) {
	arr, ok := v.([]any)
	if !ok {
		return nil, false
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result, len(result) > 0
}
