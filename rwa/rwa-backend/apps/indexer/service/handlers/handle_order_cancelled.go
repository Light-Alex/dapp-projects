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

type HandleOrderCancelled struct {
	topic0        string
	orderFilterer *contractRwa.OrderFilterer
	tradeService  trade.TradeService
}

func NewHandleOrderCancelled(tradeService trade.TradeService) (*HandleOrderCancelled, error) {
	orderFilterer, err := contractRwa.NewOrderFilterer(common.Address{}, nil)
	if err != nil {
		return nil, err
	}

	orderAbi, err := contractRwa.OrderMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	orderCancelledEvent, ok := orderAbi.Events["OrderCancelled"]
	if !ok {
		return nil, errors.New("orderCancelled event not found in Order ABI")
	}

	return &HandleOrderCancelled{
		topic0:        orderCancelledEvent.ID.Hex(),
		orderFilterer: orderFilterer,
		tradeService:  tradeService,
	}, nil
}

func (h *HandleOrderCancelled) ContractType() coreTypes.ContractType {
	return coreTypes.ContractTypeOrder
}

func (h *HandleOrderCancelled) Topic0() string {
	return h.topic0
}

func (h *HandleOrderCancelled) HandleEvent(ctx context.Context, tx *gorm.DB, event *coreTypes.EventLogWithId) error {
	ethLog, err := event.ConvertToEthLog()
	if err != nil {
		log.ErrorZ(ctx, "failed to convert event to eth log", zap.Error(err))
		return err
	}

	parsedEvent, err := h.orderFilterer.ParseOrderCancelled(*ethLog)
	if err != nil {
		log.ErrorZ(ctx, "failed to parse Order OrderCancelled event", zap.Error(err))
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

	// Idempotency check: if order is already cancelled, skip
	if order.Status == rwa.OrderStatusCancelled {
		log.InfoZ(ctx, "order already cancelled, skipping",
			zap.String("orderId", orderIdStr),
			zap.Uint64("dbOrderId", order.ID))
		return nil
	}

	blockTimestamp, err := parseBlockTimestamp(event.BlockTimestamp)
	if err != nil {
		log.ErrorZ(ctx, "failed to parse block timestamp", zap.Error(err))
		return err
	}

	if h.tradeService != nil && order.ExternalOrderID != "" {
		err := h.tradeService.CancelOrder(ctx, order.ExternalOrderID)
		if err != nil {
			log.ErrorZ(ctx, "failed to cancel order with Alpaca, will sync status instead",
				zap.Error(err),
				zap.String("orderId", orderIdStr),
				zap.String("externalOrderId", order.ExternalOrderID),
				zap.String("symbol", order.Symbol))
		} else {
			log.InfoZ(ctx, "order cancelled successfully with Alpaca",
				zap.String("orderId", orderIdStr),
				zap.String("externalOrderId", order.ExternalOrderID),
				zap.String("symbol", order.Symbol))
		}

		alpacaOrder, err := h.tradeService.GetOrder(ctx, order.ExternalOrderID)
		if err != nil {
			log.WarnZ(ctx, "failed to get order status from Alpaca, using contract event status",
				zap.Error(err),
				zap.String("orderId", orderIdStr),
				zap.String("externalOrderId", order.ExternalOrderID))
		} else if alpacaOrder != nil {
			order.Status = alpacaOrder.Status
			order.FilledQuantity = alpacaOrder.FilledQuantity
			order.FilledPrice = alpacaOrder.FilledPrice
			if !alpacaOrder.UpdatedAt.IsZero() {
				order.UpdatedAt = alpacaOrder.UpdatedAt
			}
			if alpacaOrder.FilledAt != nil {
				order.FilledAt = alpacaOrder.FilledAt
			}
			if alpacaOrder.CancelledAt != nil {
				order.CancelledAt = alpacaOrder.CancelledAt
			} else {
				order.CancelledAt = &blockTimestamp
			}

			order.RemainingQuantity = order.Quantity.Sub(order.FilledQuantity)

			log.InfoZ(ctx, "synced order status from Alpaca",
				zap.String("orderId", orderIdStr),
				zap.String("externalOrderId", order.ExternalOrderID),
				zap.String("status", string(order.Status)),
				zap.String("filledQuantity", order.FilledQuantity.String()))
		}
	} else {
		order.Status = rwa.OrderStatusCancelled
		order.CancelledAt = &blockTimestamp
	}

	if order.Status != rwa.OrderStatusCancelled {
		log.InfoZ(ctx, "updating order status to cancelled as per contract event",
			zap.String("orderId", orderIdStr),
			zap.String("previousStatus", string(order.Status)))
		order.Status = rwa.OrderStatusCancelled
		if order.CancelledAt == nil {
			order.CancelledAt = &blockTimestamp
		}
	}

	if err := tx.WithContext(ctx).Save(order).Error; err != nil {
		log.ErrorZ(ctx, "failed to update order", zap.Error(err))
		return err
	}

	eventLogID, err := saveEventLog(ctx, tx, event, "OrderCancelled", order.AccountID, order.ID)
	if err != nil {
		log.ErrorZ(ctx, "failed to save event log", zap.Error(err))
		return err
	}

	log.InfoZ(ctx, "saved OrderCancelled event and updated order",
		zap.String("txHash", event.TxHash),
		zap.String("orderId", orderIdStr),
		zap.Uint64("dbOrderId", order.ID),
		zap.String("user", parsedEvent.User.Hex()),
		zap.String("refundAmount", parsedEvent.RefundAmount.String()),
		zap.Uint8("previousStatus", parsedEvent.PreviousStatus),
		zap.String("orderStatus", string(order.Status)),
		zap.Uint64("eventLogId", eventLogID),
	)

	return nil
}

