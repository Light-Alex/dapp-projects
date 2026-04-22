package server

import (
	"context"
	"strconv"

	"github.com/AnchoredLabs/rwa-backend/apps/api/config"
	"github.com/AnchoredLabs/rwa-backend/apps/api/controller"
	"github.com/AnchoredLabs/rwa-backend/apps/api/docs"
	"github.com/AnchoredLabs/rwa-backend/apps/api/server/middleware"
	"github.com/AnchoredLabs/rwa-backend/libs/core/redis_cache"
	"github.com/AnchoredLabs/rwa-backend/libs/core/web"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type Router struct {
	conf               *config.Config
	apiKeyMap          map[string]bool
	apiKeyCacheService *redis_cache.ApiKeyCacheService
}

const apiHeaderKey = "X-API-Key"

// NewRouter creates a new Gin router with the provided configuration and controllers.
// @title RWA API
// @version 1.0
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description Rwa common-api
func NewRouter(conf *config.Config, commonController *controller.CommonController, tradeController *controller.TradeController, stockController *controller.StockController, orderController *controller.OrderController, apiKeyCacheService *redis_cache.ApiKeyCacheService) (*gin.Engine, error) {
	docs.SwaggerInfo.BasePath = conf.Server.BasePath
	apiKeyMap := make(map[string]bool)
	for _, key := range conf.Server.ApiKeys {
		apiKeyMap[key] = true
	}
	if conf.Server.GinMode != "" {
		gin.SetMode(conf.Server.GinMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	ret := gin.New()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"*"}
	ret.Use(cors.New(corsConfig))
	ret.Use(gzip.Gzip(gzip.DefaultCompression), log.RequestLog())
	if conf.Server.Env != "prod" {
		ret.GET(conf.Server.BasePath+"/swagger-ui/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		log.InfoZ(context.Background(), "Swagger enabled", zap.String("url", "http://127.0.0.1:"+strconv.Itoa(conf.Server.Port)+conf.Server.BasePath+"/swagger-ui/index.html"))
	}
	g := ret.Group(conf.Server.BasePath)
	router := &Router{
		conf:               conf,
		apiKeyMap:          apiKeyMap,
		apiKeyCacheService: apiKeyCacheService,
	}
	router.initCommon(g, commonController)
	router.initTrade(g, tradeController)
	router.initStock(g, stockController)
	router.initOrder(g, orderController)
	return ret, nil
}

func (r *Router) auth(c *gin.Context) {
	if r.conf.Server.Env == "dev" || r.conf.Server.ApiKeys == nil || len(r.conf.Server.ApiKeys) == 0 {
		c.Next()
		return
	}
	key := c.GetHeader(apiHeaderKey)
	if key == "" || r.apiKeyMap[key] == false {
		web.ResponseUnAuthorizedError(c)
		c.Abort()
		return
	}
	c.Next()
}

func (r *Router) initCommon(group *gin.RouterGroup, c *controller.CommonController) {
	g := group.Group("/common").Use(middleware.ApiSignMiddleware(r.conf))
	g.GET("/health", c.HealthCheck)
}

func (r *Router) initTrade(group *gin.RouterGroup, c *controller.TradeController) {
	g := group.Group("/trade").Use(middleware.ApiSignMiddleware(r.conf))
	g.GET("/currentPrice", c.GetCurrentPrice)
	g.GET("/latestQuote", c.GetLatestQuote)
	g.GET("/snapshot", c.GetSnapshot)
	g.GET("/historicalData", c.GetHistoricalData)
	g.GET("/marketClock", c.GetMarketClock)
	g.GET("/assets", c.GetAssets)
	g.GET("/asset", c.GetAsset)
}

func (r *Router) initStock(group *gin.RouterGroup, c *controller.StockController) {
	g := group.Group("/stock").Use(middleware.ApiSignMiddleware(r.conf))
	g.GET("/list", c.GetStockList)
	g.GET("/detail", c.GetStockDetail)
}

func (r *Router) initOrder(group *gin.RouterGroup, c *controller.OrderController) {
	g := group.Group("/order").Use(middleware.ApiSignMiddleware(r.conf))
	g.GET("/list", c.GetOrders)
	g.GET("/detail", c.GetOrderDetail)
	g.GET("/executions", c.GetOrderExecutions)
}
