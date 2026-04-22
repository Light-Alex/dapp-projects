package controller

import (
	"strconv"

	"github.com/AnchoredLabs/rwa-backend/apps/api/dto"
	"github.com/AnchoredLabs/rwa-backend/apps/api/service"
	"github.com/AnchoredLabs/rwa-backend/libs/core/error_msg"
	"github.com/AnchoredLabs/rwa-backend/libs/core/models/rwa"
	"github.com/AnchoredLabs/rwa-backend/libs/core/web"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrderController struct {
	orderService *service.OrderService
}

func NewOrderController(orderService *service.OrderService) *OrderController {
	return &OrderController{
		orderService: orderService,
	}
}

// GetOrders godoc
// @Summary Get Orders
// @Description Get list of orders with filters and pagination
// @Tags Order
// @Accept json
// @Produce json
// @Param account_id query int false "Account ID" example(1)
// @Param symbol query string false "Stock symbol" example(AAPL)
// @Param side query string false "Order side (buy/sell)" example(buy)
// @Param status query string false "Order status" example(filled)
// @Param page query int false "Page number" default(1) example(1)
// @Param page_size query int false "Page size" default(20) example(20)
// @Success 200 {object} web.Response{data=dto.GetOrdersResponse}
// @Failure 400 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /order/list [get]
func (ctl *OrderController) GetOrders(g *gin.Context) {
	ctx := g.Request.Context()

	var req dto.GetOrdersRequest
	if err := g.ShouldBindQuery(&req); err != nil {
		log.ErrorZ(ctx, "failed to bind query parameters", zap.Error(err))
		web.ResponseError(error_msg.ErrInvalidRequestParams, g)
		return
	}

	// Default pagination
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

	filter := rwa.OrderFilter{
		Symbol: req.Symbol,
		Side:   req.Side,
		Status: req.Status,
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}
	if req.AccountID > 0 {
		filter.AccountID = strconv.FormatUint(req.AccountID, 10)
	}

	log.InfoZ(ctx, "Getting orders",
		zap.Uint64("accountId", req.AccountID),
		zap.String("symbol", req.Symbol),
		zap.String("side", req.Side),
		zap.String("status", req.Status),
		zap.Int("page", page),
		zap.Int("pageSize", pageSize))

	orders, total, err := ctl.orderService.GetOrders(ctx, filter)
	if err != nil {
		log.ErrorZ(ctx, "failed to get orders", zap.Error(err))
		web.ResponseError(error_msg.ErrFailedToGetOrders, g)
		return
	}

	orderDTOs := make([]dto.OrderDTO, 0, len(orders))
	for _, order := range orders {
		orderDTOs = append(orderDTOs, toOrderDTO(order))
	}

	response := dto.GetOrdersResponse{
		List:  orderDTOs,
		Total: total,
	}

	log.InfoZ(ctx, "Successfully retrieved orders",
		zap.Int("count", len(orderDTOs)),
		zap.Int64("total", total))

	web.ResponseOk(response, g)
}

// GetOrderDetail godoc
// @Summary Get Order Detail
// @Description Get single order by ID or clientOrderId
// @Tags Order
// @Accept json
// @Produce json
// @Param id query int false "Order ID" example(1)
// @Param client_order_id query string false "Client Order ID"
// @Success 200 {object} web.Response{data=dto.GetOrderDetailResponse}
// @Failure 400 {object} web.Response
// @Failure 404 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /order/detail [get]
func (ctl *OrderController) GetOrderDetail(g *gin.Context) {
	ctx := g.Request.Context()

	var req dto.GetOrderDetailRequest
	if err := g.ShouldBindQuery(&req); err != nil {
		log.ErrorZ(ctx, "failed to bind query parameters", zap.Error(err))
		web.ResponseError(error_msg.ErrInvalidRequestParams, g)
		return
	}

	if req.ID == 0 && req.ClientOrderID == "" {
		log.ErrorZ(ctx, "either id or client_order_id is required")
		web.ResponseError(error_msg.ErrInvalidRequestParams, g)
		return
	}

	var (
		order *rwa.Order
		err   error
	)

	if req.ID > 0 {
		log.InfoZ(ctx, "Getting order by ID", zap.Uint64("id", req.ID))
		order, err = ctl.orderService.GetOrderByID(ctx, req.ID)
	} else {
		log.InfoZ(ctx, "Getting order by client order ID", zap.String("clientOrderId", req.ClientOrderID))
		order, err = ctl.orderService.GetOrderByClientOrderID(ctx, req.ClientOrderID)
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.WarnZ(ctx, "order not found",
				zap.Uint64("id", req.ID),
				zap.String("clientOrderId", req.ClientOrderID))
			web.ResponseError(error_msg.ErrOrderNotFound, g)
			return
		}
		log.ErrorZ(ctx, "failed to get order detail", zap.Error(err))
		web.ResponseError(error_msg.ErrFailedToGetOrderDetail, g)
		return
	}

	response := dto.GetOrderDetailResponse{
		Order: toOrderDTO(order),
	}

	web.ResponseOk(response, g)
}

// GetOrderExecutions godoc
// @Summary Get Order Executions
// @Description Get executions for an order
// @Tags Order
// @Accept json
// @Produce json
// @Param order_id query int true "Order ID" example(1)
// @Success 200 {object} web.Response{data=dto.GetOrderExecutionsResponse}
// @Failure 400 {object} web.Response
// @Failure 500 {object} web.Response
// @Router /order/executions [get]
func (ctl *OrderController) GetOrderExecutions(g *gin.Context) {
	ctx := g.Request.Context()

	var req dto.GetOrderExecutionsRequest
	if err := g.ShouldBindQuery(&req); err != nil {
		log.ErrorZ(ctx, "failed to bind query parameters", zap.Error(err))
		web.ResponseError(error_msg.ErrInvalidRequestParams, g)
		return
	}

	log.InfoZ(ctx, "Getting order executions", zap.Uint64("orderId", req.OrderID))

	executions, err := ctl.orderService.GetOrderExecutions(ctx, req.OrderID)
	if err != nil {
		log.ErrorZ(ctx, "failed to get order executions",
			zap.Uint64("orderId", req.OrderID),
			zap.Error(err))
		web.ResponseError(error_msg.ErrFailedToGetOrderExecutions, g)
		return
	}

	executionDTOs := make([]dto.OrderExecutionDTO, 0, len(executions))
	for _, exec := range executions {
		executionDTOs = append(executionDTOs, toOrderExecutionDTO(exec))
	}

	response := dto.GetOrderExecutionsResponse{
		List: executionDTOs,
	}

	log.InfoZ(ctx, "Successfully retrieved order executions",
		zap.Uint64("orderId", req.OrderID),
		zap.Int("count", len(executionDTOs)))

	web.ResponseOk(response, g)
}

func toOrderDTO(order *rwa.Order) dto.OrderDTO {
	d := dto.OrderDTO{
		ID:                order.ID,
		ClientOrderID:     order.ClientOrderID,
		AccountID:         order.AccountID,
		Symbol:            order.Symbol,
		AssetType:         string(order.AssetType),
		Side:              string(order.Side),
		Type:              string(order.Type),
		Quantity:          order.Quantity.String(),
		Price:             order.Price.String(),
		StopPrice:         order.StopPrice.String(),
		Status:            string(order.Status),
		FilledQuantity:    order.FilledQuantity.String(),
		FilledPrice:       order.FilledPrice.String(),
		RemainingQuantity: order.RemainingQuantity.String(),
		ContractTxHash:    order.ContractTxHash,
		ExternalOrderID:   order.ExternalOrderID,
		Provider:          order.Provider,
		Commission:        order.Commission,
		CommissionAsset:   order.CommissionAsset,
		Metadata:          order.Metadata,
		Notes:             order.Notes,
		CreatedAt:         order.CreatedAt.Unix(),
		UpdatedAt:         order.UpdatedAt.Unix(),
	}
	if order.SubmittedAt != nil {
		d.SubmittedAt = order.SubmittedAt.Unix()
	}
	if order.FilledAt != nil {
		d.FilledAt = order.FilledAt.Unix()
	}
	if order.CancelledAt != nil {
		d.CancelledAt = order.CancelledAt.Unix()
	}
	if order.ExpiredAt != nil {
		d.ExpiredAt = order.ExpiredAt.Unix()
	}
	return d
}

func toOrderExecutionDTO(exec *rwa.OrderExecution) dto.OrderExecutionDTO {
	d := dto.OrderExecutionDTO{
		ID:              exec.ID,
		OrderID:         exec.OrderID,
		ExecutionID:     exec.ExecutionID,
		Quantity:        exec.Quantity.String(),
		Price:           exec.Price.String(),
		Commission:      exec.Commission,
		CommissionAsset: exec.CommissionAsset,
		Provider:        exec.Provider,
		ExternalID:      exec.ExternalID,
		CreatedAt:       exec.CreatedAt.Unix(),
	}
	if !exec.ExecutedAt.IsZero() {
		d.ExecutedAt = exec.ExecutedAt.Unix()
	}
	return d
}
