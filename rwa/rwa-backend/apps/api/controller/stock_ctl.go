package controller

import (
	"github.com/AnchoredLabs/rwa-backend/apps/api/dto"
	"github.com/AnchoredLabs/rwa-backend/apps/api/service"
	"github.com/AnchoredLabs/rwa-backend/libs/core/error_msg"
	"github.com/AnchoredLabs/rwa-backend/libs/core/web"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type StockController struct {
	stockService *service.StockService
}

func NewStockController(stockService *service.StockService) *StockController {
	return &StockController{
		stockService: stockService,
	}
}

// GetStockList godoc
// @Summary Get Stock List
// @Description Get list of active stocks with pagination
// @Tags Stock
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1) example(1)
// @Param page_size query int false "Page size" default(20) example(20)
// @Success 200 {object} web.Response{data=dto.GetStockListResponse}
// @Failure 400 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /stock/list [get]
func (ctl *StockController) GetStockList(g *gin.Context) {
	ctx := g.Request.Context()

	var req dto.GetStockListRequest
	if err := g.ShouldBindQuery(&req); err != nil {
		log.ErrorZ(ctx, "failed to bind query parameters", zap.Error(err))
		web.ResponseError(error_msg.ErrInvalidRequestParams, g)
		return
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	response, err := ctl.stockService.GetStockList(ctx, page, pageSize)
	if err != nil {
		log.ErrorZ(ctx, "failed to get stock list", zap.Error(err))
		web.ResponseError(error_msg.ErrFailedToGetStockList, g)
		return
	}

	web.ResponseOk(response, g)
}

// GetStockDetail godoc
// @Summary Get Stock Detail
// @Description Get stock detail by symbol
// @Tags Stock
// @Accept json
// @Produce json
// @Param symbol query string true "Stock symbol"
// @Success 200 {object} web.Response{data=dto.StockInfo}
// @Failure 400 {object} web.Response
// @Failure 404 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /stock/detail [get]
func (ctl *StockController) GetStockDetail(g *gin.Context) {
	ctx := g.Request.Context()

	var req dto.GetStockDetailRequest
	if err := g.ShouldBindQuery(&req); err != nil {
		log.ErrorZ(ctx, "failed to bind query parameters", zap.Error(err))
		web.ResponseError(error_msg.ErrInvalidRequestParams, g)
		return
	}

	response, err := ctl.stockService.GetStockDetail(ctx, req.Symbol)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.WarnZ(ctx, "stock not found", zap.String("symbol", req.Symbol))
			web.ResponseError(error_msg.ErrNotFound, g)
			return
		}
		log.ErrorZ(ctx, "failed to get stock detail", zap.Error(err), zap.String("symbol", req.Symbol))
		web.ResponseError(error_msg.ErrFailedToGetStockDetail, g)
		return
	}

	web.ResponseOk(response, g)
}
