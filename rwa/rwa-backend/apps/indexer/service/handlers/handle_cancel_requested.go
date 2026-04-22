package handlers

import (
	"context"
	"errors"

	contractRwa "github.com/AnchoredLabs/rwa-backend/libs/contracts/rwa"
	"github.com/AnchoredLabs/rwa-backend/libs/core/models/rwa"
	"github.com/AnchoredLabs/rwa-backend/libs/core/trade"
	coreTypes "github.com/AnchoredLabs/rwa-backend/libs/core/types"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/acmestack/gorm-plus/gplus"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type HandleCancelRequested struct {
	topic0        string
	orderFilterer *contractRwa.OrderFilterer
	tradeService  trade.TradeService
}

func NewHandleCancelRequested(tradeService trade.TradeService) (*HandleCancelRequested, error) {
	orderFilterer, err := contractRwa.NewOrderFilterer(common.Address{}, nil)
	if err != nil {
		return nil, err
	}

	orderAbi, err := contractRwa.OrderMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	cancelRequestedEvent, ok := orderAbi.Events["CancelRequested"]
	if !ok {
		return nil, errors.New("CancelRequested event not found in Order ABI")
	}

	return &HandleCancelRequested{
		topic0:        cancelRequestedEvent.ID.Hex(),
		orderFilterer: orderFilterer,
		tradeService:  tradeService,
	}, nil
}

func (h *HandleCancelRequested) ContractType() coreTypes.ContractType {
	return coreTypes.ContractTypeOrder
}

func (h *HandleCancelRequested) Topic0() string {
	return h.topic0
}

func (h *HandleCancelRequested) HandleEvent(ctx context.Context, tx *gorm.DB, event *coreTypes.EventLogWithId) error {
	ethLog, err := event.ConvertToEthLog()
	if err != nil {
		log.ErrorZ(ctx, "failed to convert event to eth log", zap.Error(err))
		return err
	}

	parsedEvent, err := h.orderFilterer.ParseCancelRequested(*ethLog)
	if err != nil {
		log.ErrorZ(ctx, "failed to parse Order CancelRequested event", zap.Error(err))
		return err
	}

	orderIdStr := parsedEvent.OrderId.String()
	q, u := gplus.NewQuery[rwa.Order]()
	q.Eq(&u.ClientOrderID, orderIdStr)
	order, dbRes := gplus.SelectOne(q, gplus.Db(tx.WithContext(ctx)))
	if dbRes.Error != nil {
		log.ErrorZ(ctx, "failed to find order", zap.Error(dbRes.Error), zap.String("orderId", orderIdStr))
		return dbRes.Error
	}

	// Idempotency check: if order is already cancel_requested or cancelled, skip
	if order.Status == rwa.OrderStatusCancelRequested || order.Status == rwa.OrderStatusCancelled {
		log.InfoZ(ctx, "order already in cancel_requested or cancelled state, skipping",
			zap.String("orderId", orderIdStr),
			zap.String("status", string(order.Status)),
			zap.Uint64("dbOrderId", order.ID))
		return nil
	}

	// Update order status to cancel_requested
	order.Status = rwa.OrderStatusCancelRequested

	// If trade service is available and order has an external order ID, cancel on Alpaca
	if h.tradeService != nil && order.ExternalOrderID != "" {
		err := h.tradeService.CancelOrder(ctx, order.ExternalOrderID)
		if err != nil {
			log.ErrorZ(ctx, "failed to cancel order with Alpaca",
				zap.Error(err),
				zap.String("orderId", orderIdStr),
				zap.String("externalOrderId", order.ExternalOrderID),
				zap.String("symbol", order.Symbol))
		} else {
			log.InfoZ(ctx, "cancel request sent to Alpaca",
				zap.String("orderId", orderIdStr),
				zap.String("externalOrderId", order.ExternalOrderID),
				zap.String("symbol", order.Symbol))
		}
	}

	if err := tx.WithContext(ctx).Save(order).Error; err != nil {
		log.ErrorZ(ctx, "failed to update order", zap.Error(err))
		return err
	}

	eventLogID, err := saveEventLog(ctx, tx, event, "CancelRequested", order.AccountID, order.ID)
	if err != nil {
		log.ErrorZ(ctx, "failed to save event log", zap.Error(err))
		return err
	}

	log.InfoZ(ctx, "saved CancelRequested event and updated order",
		zap.String("txHash", event.TxHash),
		zap.String("orderId", orderIdStr),
		zap.Uint64("dbOrderId", order.ID),
		zap.String("user", parsedEvent.User.Hex()),
		zap.String("orderStatus", string(order.Status)),
		zap.Uint64("eventLogId", eventLogID),
	)

	return nil
}
