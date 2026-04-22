package handlers

import (
	"context"
	"errors"
	"fmt"

	contractRwa "github.com/AnchoredLabs/rwa-backend/libs/contracts/rwa"
	"github.com/AnchoredLabs/rwa-backend/libs/core/models/rwa"
	"github.com/AnchoredLabs/rwa-backend/libs/core/trade"
	coreTypes "github.com/AnchoredLabs/rwa-backend/libs/core/types"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/acmestack/gorm-plus/gplus"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type HandleOrderExecuted struct {
	topic0        string
	orderFilterer *contractRwa.OrderFilterer
	tradeService  trade.TradeService
}

func NewHandleOrderExecuted(tradeService trade.TradeService) (*HandleOrderExecuted, error) {
	orderFilterer, err := contractRwa.NewOrderFilterer(common.Address{}, nil)
	if err != nil {
		return nil, err
	}

	orderAbi, err := contractRwa.OrderMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	orderExecutedEvent, ok := orderAbi.Events["OrderExecuted"]
	if !ok {
		return nil, errors.New("OrderExecuted event not found in Order ABI")
	}

	return &HandleOrderExecuted{
		topic0:        orderExecutedEvent.ID.Hex(),
		orderFilterer: orderFilterer,
		tradeService:  tradeService,
	}, nil
}

func (h *HandleOrderExecuted) ContractType() coreTypes.ContractType {
	return coreTypes.ContractTypeOrder
}

func (h *HandleOrderExecuted) Topic0() string {
	return h.topic0
}

func (h *HandleOrderExecuted) HandleEvent(ctx context.Context, tx *gorm.DB, event *coreTypes.EventLogWithId) error {
	ethLog, err := event.ConvertToEthLog()
	if err != nil {
		log.ErrorZ(ctx, "failed to convert event to eth log", zap.Error(err))
		return err
	}

	parsedEvent, err := h.orderFilterer.ParseOrderExecuted(*ethLog)
	if err != nil {
		log.ErrorZ(ctx, "failed to parse Order OrderExecuted event", zap.Error(err))
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

	// Idempotency check: if order is already filled, skip
	if order.Status == rwa.OrderStatusFilled {
		log.InfoZ(ctx, "order already filled, skipping",
			zap.String("orderId", orderIdStr),
			zap.Uint64("dbOrderId", order.ID))
		return nil
	}

	blockTimestamp, err := parseBlockTimestamp(event.BlockTimestamp)
	if err != nil {
		log.ErrorZ(ctx, "failed to parse block timestamp", zap.Error(err))
		return err
	}

	refundAmount := bigIntToDecimalWithPrecision(parsedEvent.RefundAmount, 18)

	// Sync with Alpaca if trade service is available
	if h.tradeService != nil && order.ExternalOrderID != "" {
		alpacaOrder, err := h.tradeService.GetOrder(ctx, order.ExternalOrderID)
		if err != nil {
			log.WarnZ(ctx, "failed to get order status from Alpaca, using contract event data",
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

			order.RemainingQuantity = order.Quantity.Sub(order.FilledQuantity)

			log.InfoZ(ctx, "synced order status from Alpaca",
				zap.String("orderId", orderIdStr),
				zap.String("externalOrderId", order.ExternalOrderID),
				zap.String("status", string(order.Status)),
				zap.String("filledQuantity", order.FilledQuantity.String()))
		}
	}

	// Determine status based on filled quantities if not already set from Alpaca
	if order.Status != rwa.OrderStatusFilled && order.Status != rwa.OrderStatusPartiallyFilled {
		if order.FilledQuantity.GreaterThan(decimal.Zero) && order.FilledQuantity.LessThan(order.Quantity) {
			order.Status = rwa.OrderStatusPartiallyFilled
		} else {
			order.Status = rwa.OrderStatusFilled
		}
		order.FilledAt = &blockTimestamp
	}

	// Record refund amount in notes if any
	if refundAmount.GreaterThan(decimal.Zero) {
		if order.Notes != "" {
			order.Notes += "; "
		}
		order.Notes += fmt.Sprintf("refundAmount=%s", refundAmount.String())
	}

	if err := tx.WithContext(ctx).Save(order).Error; err != nil {
		log.ErrorZ(ctx, "failed to update order", zap.Error(err))
		return err
	}

	// Save order execution record
	execution := rwa.OrderExecution{
		OrderID:    order.ID,
		Quantity:   order.FilledQuantity,
		Price:      order.FilledPrice,
		Provider:   order.Provider,
		ExternalID: order.ExternalOrderID,
		ExecutedAt: blockTimestamp,
	}

	dbRes = gplus.Insert(&execution, gplus.Db(tx.WithContext(ctx)))
	if dbRes.Error != nil {
		log.ErrorZ(ctx, "failed to save order execution", zap.Error(dbRes.Error))
		return dbRes.Error
	}

	// Save event log
	eventLogID, err := saveEventLog(ctx, tx, event, "OrderExecuted", order.AccountID, order.ID)
	if err != nil {
		log.ErrorZ(ctx, "failed to save event log", zap.Error(err))
		return err
	}

	log.InfoZ(ctx, "saved OrderExecuted event and updated order",
		zap.String("txHash", event.TxHash),
		zap.String("orderId", orderIdStr),
		zap.Uint64("dbOrderId", order.ID),
		zap.String("orderStatus", string(order.Status)),
		zap.String("filledQuantity", order.FilledQuantity.String()),
		zap.String("filledPrice", order.FilledPrice.String()),
		zap.String("refundAmount", refundAmount.String()),
		zap.Uint64("executionId", execution.ID),
		zap.Uint64("eventLogId", eventLogID),
	)

	return nil
}
