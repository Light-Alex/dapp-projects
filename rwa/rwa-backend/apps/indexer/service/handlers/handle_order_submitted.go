package handlers

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"time"

	contractRwa "github.com/AnchoredLabs/rwa-backend/libs/contracts/rwa"
	"github.com/AnchoredLabs/rwa-backend/libs/core/evm_helper"
	"github.com/AnchoredLabs/rwa-backend/libs/core/models/rwa"
	"github.com/AnchoredLabs/rwa-backend/libs/core/trade"
	coreTypes "github.com/AnchoredLabs/rwa-backend/libs/core/types"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/acmestack/gorm-plus/gplus"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type HandleOrderSubmitted struct {
	topic0        string
	orderFilterer *contractRwa.OrderFilterer
	tradeService  trade.TradeService
	usdmAddress   string
	evmClient     *evm_helper.EvmClient
	chainID       uint64
	pocAddress    common.Address
	privateKey    *ecdsa.PrivateKey
}

func NewHandleOrderSubmitted(tradeService trade.TradeService, usdmAddress string, evmClient *evm_helper.EvmClient, chainID uint64, pocAddress string, privateKey *ecdsa.PrivateKey) (*HandleOrderSubmitted, error) {
	orderFilterer, err := contractRwa.NewOrderFilterer(common.Address{}, nil)
	if err != nil {
		return nil, err
	}

	// Get the OrderSubmitted event topic0 from the ABI
	orderAbi, err := contractRwa.OrderMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	orderSubmittedEvent, ok := orderAbi.Events["OrderSubmitted"]
	if !ok {
		return nil, errors.New("orderSubmitted event not found in Order ABI")
	}

	return &HandleOrderSubmitted{
		topic0:        orderSubmittedEvent.ID.Hex(),
		orderFilterer: orderFilterer,
		tradeService:  tradeService,
		usdmAddress:   usdmAddress,
		evmClient:     evmClient,
		chainID:       chainID,
		pocAddress:    common.HexToAddress(pocAddress),
		privateKey:    privateKey,
	}, nil
}

func (h *HandleOrderSubmitted) ContractType() coreTypes.ContractType {
	return coreTypes.ContractTypeOrder
}

func (h *HandleOrderSubmitted) Topic0() string {
	return h.topic0
}

func (h *HandleOrderSubmitted) HandleEvent(ctx context.Context, tx *gorm.DB, event *coreTypes.EventLogWithId) error {
	ethLog, err := event.ConvertToEthLog()
	if err != nil {
		log.ErrorZ(ctx, "failed to convert event to eth log", zap.Error(err))
		return err
	}

	// Parse the OrderSubmitted event using the type-safe method
	parsedEvent, err := h.orderFilterer.ParseOrderSubmitted(*ethLog)
	if err != nil {
		log.ErrorZ(ctx, "failed to parse Order OrderSubmitted event", zap.Error(err))
		return err
	}

	// Get or create account
	accountID, err := getOrCreateAccount(ctx, tx, parsedEvent.User)
	if err != nil {
		log.ErrorZ(ctx, "failed to get or create account", zap.Error(err))
		return err
	}

	// Parse block timestamp
	blockTimestamp, err := parseBlockTimestamp(event.BlockTimestamp)
	if err != nil {
		log.ErrorZ(ctx, "failed to parse block timestamp", zap.Error(err))
		return err
	}

	// Convert big.Int to decimal
	quantity := bigIntToDecimalWithPrecision(parsedEvent.Qty, 18)
	price := bigIntToDecimalWithPrecision(parsedEvent.Price, 18)

	// Convert order type, side, and time in force
	orderType := convertUint8ToOrderType(parsedEvent.OrderType)
	orderSide := convertUint8ToOrderSide(parsedEvent.Side)
	timeInForce := convertUint8ToTimeInForce(parsedEvent.Tif)

	// Check if order already exists (idempotency check)
	orderIdStr := parsedEvent.OrderId.String()
	q, u := gplus.NewQuery[rwa.Order]()
	q.Eq(&u.ClientOrderID, orderIdStr)
	existingOrder, dbRes := gplus.SelectOne(q, gplus.Db(tx.WithContext(ctx)))
	if dbRes.Error == nil {
		// Order already exists, skip processing
		log.InfoZ(ctx, "order already exists, skipping",
			zap.String("orderId", orderIdStr),
			zap.Uint64("dbOrderId", existingOrder.ID))
		return nil
	}
	if !errors.Is(dbRes.Error, gorm.ErrRecordNotFound) {
		log.ErrorZ(ctx, "failed to check existing order", zap.Error(dbRes.Error))
		return dbRes.Error
	}

	// Calculate escrow amount and asset based on order side
	var escrowAmount decimal.Decimal
	var escrowAsset string
	if orderSide == rwa.OrderSideBuy {
		// Buy order: escrowAmount = price * qty (both already in human-readable units)
		escrowAmount = price.Mul(quantity)
		escrowAsset = h.usdmAddress
	} else {
		// Sell order: escrowAmount = qty
		escrowAmount = quantity
		// Query symbolToToken[symbol] from the Order contract to get the escrow asset address
		escrowAsset = h.querySymbolToToken(ctx, parsedEvent.Symbol)
	}

	// Create order in database first
	order := rwa.Order{
		ClientOrderID:     orderIdStr,
		AccountID:         accountID,
		Symbol:            parsedEvent.Symbol,
		AssetType:         rwa.AssetTypeStock,
		Side:              orderSide,
		Type:              orderType,
		TimeInForce:       string(timeInForce),
		Quantity:          quantity,
		Price:             price,
		StopPrice:         decimal.Zero,
		Status:            rwa.OrderStatusPending,
		FilledQuantity:    decimal.Zero,
		FilledPrice:       decimal.Zero,
		RemainingQuantity: quantity,
		EscrowAmount:      escrowAmount,
		EscrowAsset:       escrowAsset,
		ContractTxHash:    event.TxHash,
		SubmittedAt:       &blockTimestamp,
		Provider:          "alpaca",
	}

	// Save order first to get the ID using gorm-plus
	dbRes = gplus.Insert(&order, gplus.Db(tx.WithContext(ctx)))
	if dbRes.Error != nil {
		log.ErrorZ(ctx, "failed to save order", zap.Error(dbRes.Error))
		return dbRes.Error
	}

	// Save event log
	eventLogID, err := saveEventLog(ctx, tx, event, "OrderSubmitted", accountID, order.ID)
	if err != nil {
		log.ErrorZ(ctx, "failed to save event log", zap.Error(err))
		return err
	}

	// Update order with event log ID
	order.EventLogID = eventLogID
	if err := tx.WithContext(ctx).Save(&order).Error; err != nil {
		log.ErrorZ(ctx, "failed to update order with event log ID", zap.Error(err))
		return err
	}

	// Place order with Alpaca if trade service is available
	if h.tradeService != nil {
		// Build place order request based on order type
		placeOrderReq := trade.PlaceOrderRequest{
			ClientOrderID: orderIdStr,
			Symbol:        parsedEvent.Symbol,
			AssetType:     rwa.AssetTypeStock,
			Side:          orderSide,
			Type:          orderType,
			Quantity:      quantity,
			TimeInForce:   timeInForce,
		}

		// Set price and stop price based on order type
		switch orderType {
		case rwa.OrderTypeMarket:
			// Market order: no price needed
			placeOrderReq.Price = decimal.Zero
			placeOrderReq.StopPrice = decimal.Zero
		case rwa.OrderTypeLimit:
			// Limit order: set limit price
			placeOrderReq.Price = price
			placeOrderReq.StopPrice = decimal.Zero
		case rwa.OrderTypeStop:
			// Stop order: set stop price
			placeOrderReq.Price = decimal.Zero
			placeOrderReq.StopPrice = price
		case rwa.OrderTypeStopLimit:
			// Stop limit order: set both limit price and stop price
			placeOrderReq.Price = price
			placeOrderReq.StopPrice = price // TODO: Check if contract has separate stop price field
		default:
			log.WarnZ(ctx, "unknown order type, defaulting to market order",
				zap.String("orderType", string(orderType)),
				zap.String("orderId", orderIdStr))
			placeOrderReq.Type = rwa.OrderTypeMarket
			placeOrderReq.Price = decimal.Zero
			placeOrderReq.StopPrice = decimal.Zero
		}

		// Place order with Alpaca
		alpacaOrder, err := h.tradeService.PlaceOrder(ctx, placeOrderReq)
		if err != nil {
			log.ErrorZ(ctx, "failed to place order with Alpaca",
				zap.Error(err),
				zap.String("orderId", orderIdStr),
				zap.String("symbol", parsedEvent.Symbol))

			// Update order status to rejected and let the transaction commit
			// so the Rejected status is persisted. Do NOT return the error.
			order.Status = rwa.OrderStatusRejected
			order.Notes = "Failed to place order with Alpaca: " + err.Error()
			if saveErr := tx.WithContext(ctx).Save(&order).Error; saveErr != nil {
				log.ErrorZ(ctx, "failed to update order status to rejected", zap.Error(saveErr))
				return saveErr
			}
			// Call on-chain cancelOrder asynchronously to refund escrowed assets.
			// This must happen outside the DB transaction to avoid blocking.
			go h.callCancelOrder(parsedEvent.OrderId)
			// Alpaca failure is not a processing error — order is saved as Rejected.
		}

		// Update order with Alpaca response
		if alpacaOrder != nil {
			order.ExternalOrderID = alpacaOrder.ExternalOrderID
			order.Status = alpacaOrder.Status
			order.FilledQuantity = alpacaOrder.FilledQuantity
			order.FilledPrice = alpacaOrder.FilledPrice
			if alpacaOrder.SubmittedAt != nil {
				order.SubmittedAt = alpacaOrder.SubmittedAt
			}
			if !alpacaOrder.CreatedAt.IsZero() {
				order.CreatedAt = alpacaOrder.CreatedAt
			}
			if !alpacaOrder.UpdatedAt.IsZero() {
				order.UpdatedAt = alpacaOrder.UpdatedAt
			}
			if alpacaOrder.FilledAt != nil {
				order.FilledAt = alpacaOrder.FilledAt
			}
			if alpacaOrder.CancelledAt != nil {
				order.CancelledAt = alpacaOrder.CancelledAt
			}

			// Update remaining quantity
			order.RemainingQuantity = order.Quantity.Sub(order.FilledQuantity)

			if err := tx.WithContext(ctx).Save(&order).Error; err != nil {
				log.ErrorZ(ctx, "failed to update order with Alpaca response", zap.Error(err))
				return err
			}

			log.InfoZ(ctx, "order placed successfully with Alpaca",
				zap.String("txHash", event.TxHash),
				zap.String("user", parsedEvent.User.Hex()),
				zap.String("orderId", orderIdStr),
				zap.String("alpacaOrderId", alpacaOrder.ExternalOrderID),
				zap.Uint64("dbOrderId", order.ID),
				zap.String("symbol", parsedEvent.Symbol),
				zap.String("orderType", string(orderType)),
				zap.String("status", string(order.Status)),
			)
		}
	} else {
		log.WarnZ(ctx, "trade service not available, order saved but not placed",
			zap.String("orderId", orderIdStr))
	}

	log.InfoZ(ctx, "saved OrderSubmitted event and order",
		zap.String("txHash", event.TxHash),
		zap.String("user", parsedEvent.User.Hex()),
		zap.String("orderId", orderIdStr),
		zap.Uint64("dbOrderId", order.ID),
		zap.Uint64("accountId", accountID),
		zap.String("symbol", parsedEvent.Symbol),
		zap.Uint64("eventLogId", eventLogID),
	)

	return nil
}

// querySymbolToToken queries the Order contract's symbolToToken mapping
// to resolve the token address for a given symbol. Returns the address as hex string,
// or empty string if the query fails.
func (h *HandleOrderSubmitted) querySymbolToToken(ctx context.Context, symbol string) string {
	if h.evmClient == nil {
		log.WarnZ(ctx, "evmClient not configured, cannot query symbolToToken",
			zap.String("symbol", symbol))
		return ""
	}

	ethClient := h.evmClient.MustGetHttpClient(h.chainID)
	orderCaller, err := contractRwa.NewOrderCaller(h.pocAddress, ethClient)
	if err != nil {
		log.ErrorZ(ctx, "failed to create OrderCaller for symbolToToken query",
			zap.Error(err), zap.String("symbol", symbol))
		return ""
	}

	tokenAddr, err := orderCaller.SymbolToToken(&bind.CallOpts{Context: ctx}, symbol)
	if err != nil {
		log.ErrorZ(ctx, "failed to query symbolToToken",
			zap.Error(err), zap.String("symbol", symbol))
		return ""
	}

	if tokenAddr == (common.Address{}) {
		log.WarnZ(ctx, "symbolToToken returned zero address",
			zap.String("symbol", symbol))
		return ""
	}

	log.InfoZ(ctx, "resolved symbolToToken",
		zap.String("symbol", symbol),
		zap.String("tokenAddress", tokenAddr.Hex()))
	return tokenAddr.Hex()
}

// callCancelOrder sends a cancelOrder transaction to the on-chain OrderContract
// to refund the user's escrowed assets when Alpaca PlaceOrder fails.
// Runs asynchronously to avoid blocking the indexer's DB transaction.
func (h *HandleOrderSubmitted) callCancelOrder(orderId *big.Int) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if h.privateKey == nil {
		log.WarnZ(ctx, "callCancelOrder: backend private key not configured, skipping",
			zap.String("order_id", orderId.String()))
		return
	}

	ethClient := h.evmClient.MustGetHttpClient(h.chainID)
	orderTransactor, err := contractRwa.NewOrderTransactor(h.pocAddress, ethClient)
	if err != nil {
		log.ErrorZ(ctx, "callCancelOrder: failed to create OrderTransactor",
			zap.Error(err), zap.String("order_id", orderId.String()))
		return
	}

	auth, err := bind.NewKeyedTransactorWithChainID(h.privateKey, new(big.Int).SetUint64(h.chainID))
	if err != nil {
		log.ErrorZ(ctx, "callCancelOrder: failed to create transact opts",
			zap.Error(err), zap.String("order_id", orderId.String()))
		return
	}

	tx, err := orderTransactor.CancelOrder(auth, orderId)
	if err != nil {
		log.ErrorZ(ctx, "callCancelOrder: contract call failed",
			zap.Error(err), zap.String("order_id", orderId.String()))
		return
	}

	log.InfoZ(ctx, "callCancelOrder: cancel tx sent for rejected order",
		zap.String("order_id", orderId.String()),
		zap.String("tx_hash", tx.Hash().Hex()))
}

