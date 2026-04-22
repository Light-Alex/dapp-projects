package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/AnchoredLabs/rwa-backend/libs/core/models/rwa"
	"github.com/AnchoredLabs/rwa-backend/libs/core/types"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/acmestack/gorm-plus/gplus"
	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// getOrCreateAccount gets an account by address or creates a new one if it doesn't exist
func getOrCreateAccount(ctx context.Context, db *gorm.DB, address common.Address) (uint64, error) {
	addressStr := address.Hex()

	// Try to find existing account using gorm-plus
	q, u := gplus.NewQuery[rwa.Account]()
	q.Eq(&u.Address, addressStr)
	account, dbRes := gplus.SelectOne(q, gplus.Db(db.WithContext(ctx)))
	if dbRes.Error == nil {
		return account.ID, nil
	}

	if !errors.Is(dbRes.Error, gorm.ErrRecordNotFound) {
		log.ErrorZ(ctx, "failed to query account", zap.Error(dbRes.Error), zap.String("address", addressStr))
		return 0, dbRes.Error
	}

	// Create new account
	newAccount := rwa.Account{
		Address:  addressStr,
		IsActive: true,
	}
	dbRes = gplus.Insert(&newAccount, gplus.Db(db.WithContext(ctx)))
	if dbRes.Error != nil {
		log.ErrorZ(ctx, "failed to create account", zap.Error(dbRes.Error), zap.String("address", addressStr))
		return 0, dbRes.Error
	}

	log.InfoZ(ctx, "created new account", zap.Uint64("accountId", newAccount.ID), zap.String("address", addressStr))
	return newAccount.ID, nil
}

// convertUint8ToOrderSide converts uint8 to OrderSide
// 0 = Buy, 1 = Sell
func convertUint8ToOrderSide(side uint8) rwa.OrderSide {
	if side == 0 {
		return rwa.OrderSideBuy
	}
	return rwa.OrderSideSell
}

// convertUint8ToOrderType converts uint8 to OrderType
// 0 = Market, 1 = Limit, 2 = Stop, 3 = StopLimit
func convertUint8ToOrderType(orderType uint8) rwa.OrderType {
	switch orderType {
	case 0:
		return rwa.OrderTypeMarket
	case 1:
		return rwa.OrderTypeLimit
	case 2:
		return rwa.OrderTypeStop
	case 3:
		return rwa.OrderTypeStopLimit
	default:
		return rwa.OrderTypeMarket
	}
}

// convertUint8ToOrderStatus converts uint8 to OrderStatus
// Based on contract enum: 0 = New, 1 = Pending, 2 = Accepted, 3 = Rejected, 4 = Filled, 5 = PartiallyFilled, 6 = Cancelled, 7 = Expired
// NOTE: Currently unused but will be needed for OrderExecuted handler.
func convertUint8ToOrderStatus(status uint8) rwa.OrderStatus {
	switch status {
	case 0:
		return rwa.OrderStatusNew
	case 1:
		return rwa.OrderStatusPending
	case 2:
		return rwa.OrderStatusAccepted
	case 3:
		return rwa.OrderStatusRejected
	case 4:
		return rwa.OrderStatusFilled
	case 5:
		return rwa.OrderStatusPartiallyFilled
	case 6:
		return rwa.OrderStatusCancelled
	case 7:
		return rwa.OrderStatusExpired
	default:
		return rwa.OrderStatusNew
	}
}

// saveEventLog saves an event log to the database
func saveEventLog(ctx context.Context, db *gorm.DB, event *types.EventLogWithId, eventName string, accountID uint64, orderID uint64) (uint64, error) {
	blockNumber, err := hexutil.DecodeUint64(event.BlockNumber)
	if err != nil {
		return 0, err
	}

	logIndex, err := hexutil.DecodeUint64(event.Index)
	if err != nil {
		return 0, err
	}

	// Serialize event data to JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	eventLog := rwa.EventLog{
		TxHash:          event.TxHash,
		BlockNumber:     int64(blockNumber),
		BlockHash:       event.BlockHash,
		LogIndex:        int(logIndex),
		ContractAddress: event.Address.Hex(),
		EventName:       eventName,
		EventData:       string(eventData),
		AccountID:       accountID,
		OrderID:         orderID,
		Status:          "processed",
		ProcessedAt:     &now,
	}

	dbRes := gplus.Insert(&eventLog, gplus.Db(db.WithContext(ctx)))
	if dbRes.Error != nil {
		log.ErrorZ(ctx, "failed to save event log", zap.Error(dbRes.Error), zap.String("eventName", eventName))
		return 0, dbRes.Error
	}

	return eventLog.ID, nil
}

// bigIntToDecimalWithPrecision converts *big.Int to decimal.Decimal, dividing by 10^decimals.
// Use this when the contract stores values in fixed-point representation (e.g., 18 decimals for
// wei-based amounts or 6 decimals for USDC).
func bigIntToDecimalWithPrecision(bi *big.Int, decimals int32) decimal.Decimal {
	if bi == nil {
		return decimal.Zero
	}
	return decimal.NewFromBigInt(bi, -decimals)
}

// parseBlockTimestamp parses block timestamp from string to time.Time
func parseBlockTimestamp(timestampStr string) (time.Time, error) {
	timestamp, err := hexutil.DecodeUint64(timestampStr)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(timestamp), 0), nil
}

// convertUint8ToTimeInForce converts uint8 to alpaca.TimeInForce
// Based on Order.sol enum: DAY(0), GTC(1), IOC(2), FOK(3), OPG(4), GTX(5), GTD(6), CLS(7)
// Alpaca supports: day, gtc, opg, cls, ioc, fok, gtx, gtd
// GTX and GTD are not commonly supported by Alpaca so we fall back to Day and GTC respectively.
func convertUint8ToTimeInForce(tif uint8) alpaca.TimeInForce {
	switch tif {
	case 0:
		return alpaca.Day
	case 1:
		return alpaca.GTC
	case 2:
		return alpaca.IOC
	case 3:
		return alpaca.FOK
	case 4:
		return alpaca.OPG
	case 5:
		// GTX is not commonly supported by Alpaca, fall back to Day
		return alpaca.Day
	case 6:
		// GTD is not commonly supported by Alpaca, fall back to GTC
		return alpaca.GTC
	case 7:
		return alpaca.CLS
	default:
		return alpaca.Day // default to Day
	}
}

