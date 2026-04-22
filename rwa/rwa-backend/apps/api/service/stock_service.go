package service

import (
	"context"
	"errors"

	"github.com/AnchoredLabs/rwa-backend/apps/api/dto"
	"github.com/AnchoredLabs/rwa-backend/libs/core/models/rwa"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/acmestack/gorm-plus/gplus"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type StockService struct {
	db *gorm.DB
}

func NewStockService(db *gorm.DB) *StockService {
	return &StockService{
		db: db,
	}
}

func (s *StockService) GetStockList(ctx context.Context, page, pageSize int) (*dto.GetStockListResponse, error) {
	// Count
	countQuery, cu := gplus.NewQuery[rwa.Stock]()
	countQuery.Eq(&cu.Status, rwa.StockStatusActive)
	total, dbRes := gplus.SelectCount[rwa.Stock](countQuery, gplus.Db(s.db))
	if dbRes.Error != nil {
		log.ErrorZ(ctx, "failed to count stocks", zap.Error(dbRes.Error))
		return nil, dbRes.Error
	}

	// Query with pagination
	q, u := gplus.NewQuery[rwa.Stock]()
	q.Eq(&u.Status, rwa.StockStatusActive)

	db := s.db.WithContext(ctx).Order("id ASC")
	if pageSize > 0 {
		db = db.Limit(pageSize)
	}
	if page > 1 && pageSize > 0 {
		db = db.Offset((page - 1) * pageSize)
	}

	list, dbRes := gplus.SelectList(q, gplus.Db(db))
	if dbRes.Error != nil {
		log.ErrorZ(ctx, "failed to get stock list from database", zap.Error(dbRes.Error))
		return nil, dbRes.Error
	}

	stockList := make([]dto.StockInfo, 0, len(list))
	for _, stock := range list {
		stockList = append(stockList, s.toStockInfo(*stock))
	}

	return &dto.GetStockListResponse{
		List:  stockList,
		Total: total,
	}, nil
}

func (s *StockService) GetStockDetail(ctx context.Context, symbol string) (*dto.StockInfo, error) {
	if symbol == "" {
		return nil, errors.New("symbol is required")
	}

	q, u := gplus.NewQuery[rwa.Stock]()
	q.Eq(&u.Symbol, symbol)
	stock, dbRes := gplus.SelectOne(q, gplus.Db(s.db))
	if dbRes.Error != nil {
		if errors.Is(dbRes.Error, gorm.ErrRecordNotFound) {
			log.WarnZ(ctx, "stock not found", zap.String("symbol", symbol))
			return nil, dbRes.Error
		}
		log.ErrorZ(ctx, "failed to get stock detail from database", zap.Error(dbRes.Error), zap.String("symbol", symbol))
		return nil, dbRes.Error
	}

	stockInfo := s.toStockInfo(*stock)
	stockInfo.About = stock.About
	stockInfo.Status = string(stock.Status)

	return &stockInfo, nil
}

func (s *StockService) toStockInfo(stock rwa.Stock) dto.StockInfo {
	return dto.StockInfo{
		ID:        stock.ID,
		Symbol:    stock.Symbol,
		Name:      stock.Name,
		Exchange:  stock.Exchange,
		Contract:  stock.Contract,
		CreatedAt: stock.CreatedAt.Unix(),
		UpdatedAt: stock.UpdatedAt.Unix(),
	}
}
