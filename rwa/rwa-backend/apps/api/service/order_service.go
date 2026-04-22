package service

import (
	"context"
	"errors"

	"github.com/AnchoredLabs/rwa-backend/libs/core/models/rwa"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/acmestack/gorm-plus/gplus"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{
		db: db,
	}
}

func (s *OrderService) GetOrders(ctx context.Context, filter rwa.OrderFilter) ([]*rwa.Order, int64, error) {
	q, u := gplus.NewQuery[rwa.Order]()

	if filter.AccountID != "" {
		q.Eq(&u.AccountID, filter.AccountID)
	}
	if filter.Symbol != "" {
		q.Eq(&u.Symbol, filter.Symbol)
	}
	if filter.Side != "" {
		q.Eq(&u.Side, filter.Side)
	}
	if filter.Status != "" {
		q.Eq(&u.Status, filter.Status)
	}
	if !filter.StartTime.IsZero() {
		q.Ge(&u.CreatedAt, filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		q.Le(&u.CreatedAt, filter.EndTime)
	}

	// Get total count
	countQuery, cu := gplus.NewQuery[rwa.Order]()
	if filter.AccountID != "" {
		countQuery.Eq(&cu.AccountID, filter.AccountID)
	}
	if filter.Symbol != "" {
		countQuery.Eq(&cu.Symbol, filter.Symbol)
	}
	if filter.Side != "" {
		countQuery.Eq(&cu.Side, filter.Side)
	}
	if filter.Status != "" {
		countQuery.Eq(&cu.Status, filter.Status)
	}
	if !filter.StartTime.IsZero() {
		countQuery.Ge(&cu.CreatedAt, filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		countQuery.Le(&cu.CreatedAt, filter.EndTime)
	}

	total, dbRes := gplus.SelectCount[rwa.Order](countQuery, gplus.Db(s.db))
	if dbRes.Error != nil {
		log.ErrorZ(ctx, "failed to count orders", zap.Error(dbRes.Error))
		return nil, 0, dbRes.Error
	}

	// Apply pagination and ordering via GORM scope
	db := s.db.WithContext(ctx).Order("created_at DESC")
	if filter.Limit > 0 {
		db = db.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		db = db.Offset(filter.Offset)
	}

	list, dbRes := gplus.SelectList(q, gplus.Db(db))
	if dbRes.Error != nil {
		log.ErrorZ(ctx, "failed to get orders from database", zap.Error(dbRes.Error))
		return nil, 0, dbRes.Error
	}

	return list, total, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, id uint64) (*rwa.Order, error) {
	q, u := gplus.NewQuery[rwa.Order]()
	q.Eq(&u.ID, id)

	order, dbRes := gplus.SelectOne(q, gplus.Db(s.db))
	if dbRes.Error != nil {
		if errors.Is(dbRes.Error, gorm.ErrRecordNotFound) {
			log.WarnZ(ctx, "order not found", zap.Uint64("id", id))
			return nil, dbRes.Error
		}
		log.ErrorZ(ctx, "failed to get order by id", zap.Error(dbRes.Error), zap.Uint64("id", id))
		return nil, dbRes.Error
	}

	return order, nil
}

func (s *OrderService) GetOrderByClientOrderID(ctx context.Context, clientOrderID string) (*rwa.Order, error) {
	q, u := gplus.NewQuery[rwa.Order]()
	q.Eq(&u.ClientOrderID, clientOrderID)

	order, dbRes := gplus.SelectOne(q, gplus.Db(s.db))
	if dbRes.Error != nil {
		if errors.Is(dbRes.Error, gorm.ErrRecordNotFound) {
			log.WarnZ(ctx, "order not found by client order id", zap.String("clientOrderId", clientOrderID))
			return nil, dbRes.Error
		}
		log.ErrorZ(ctx, "failed to get order by client order id", zap.Error(dbRes.Error), zap.String("clientOrderId", clientOrderID))
		return nil, dbRes.Error
	}

	return order, nil
}

func (s *OrderService) GetOrderExecutions(ctx context.Context, orderID uint64) ([]*rwa.OrderExecution, error) {
	q, u := gplus.NewQuery[rwa.OrderExecution]()
	q.Eq(&u.OrderID, orderID)

	db := s.db.WithContext(ctx).Order("created_at DESC")
	list, dbRes := gplus.SelectList(q, gplus.Db(db))
	if dbRes.Error != nil {
		log.ErrorZ(ctx, "failed to get order executions", zap.Error(dbRes.Error), zap.Uint64("orderId", orderID))
		return nil, dbRes.Error
	}

	return list, nil
}
