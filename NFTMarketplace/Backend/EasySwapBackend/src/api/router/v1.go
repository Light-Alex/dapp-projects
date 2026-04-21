package router

import (
	"github.com/gin-gonic/gin"

	"github.com/ProjectsTask/EasySwapBackend/src/api/middleware"
	v1 "github.com/ProjectsTask/EasySwapBackend/src/api/v1"
	"github.com/ProjectsTask/EasySwapBackend/src/service/svc"
)

// loadV1 函数用于加载API的v1版本路由。
// 参数 r 是一个指向Gin引擎实例的指针，用于注册路由。
// 参数 svcCtx 是一个指向服务上下文的指针，包含了应用程序所需的服务和配置。
func loadV1(r *gin.Engine, svcCtx *svc.ServerCtx) {
	// 创建一个名为 /api/v1 的路由组
	apiV1 := r.Group("/api/v1")

	// 创建一个名为 /user 的子路由组
	user := apiV1.Group("/user")
	{
		// 生成login签名信息
		user.GET("/:address/login-message", v1.GetLoginMessageHandler(svcCtx))
		// 登陆
		user.POST("/login", v1.UserLoginHandler(svcCtx))
		// 获取用户签名状态
		user.GET("/:address/sig-status", v1.GetSigStatusHandler(svcCtx))
	}

	// 创建一个名为 /collections 的子路由组
	collections := apiV1.Group("/collections")
	{
		// 接口定义： 路由 + 中间件 + 处理函数
		// 指定Collection详情
		collections.GET("/:address", v1.CollectionDetailHandler(svcCtx))
		// 指定Collection的bids信息
		collections.GET("/:address/bids", v1.CollectionBidsHandler(svcCtx))
		// 指定Item的bid信息
		collections.GET("/:address/:token_id/bids", v1.CollectionItemBidsHandler(svcCtx))
		// 指定Collection的items信息
		collections.GET("/:address/items", v1.CollectionItemsHandler(svcCtx))

		// 获取NFT Item的详细信息
		collections.GET("/:address/:token_id", v1.ItemDetailHandler(svcCtx))
		// 获取NFT Item的Attribute信息
		collections.GET("/:address/:token_id/traits", v1.ItemTraitsHandler(svcCtx))
		// 获取NFT Item的Trait的最高价格信息
		collections.GET("/:address/top-trait", v1.ItemTopTraitPriceHandler(svcCtx))
		// 获取NFT Item的图片信息，使用缓存中间件，缓存时间为60秒
		collections.GET("/:address/:token_id/image", middleware.CacheApi(svcCtx.KvStore, 60), v1.GetItemImageHandler(svcCtx))
		// NFT销售历史价格信息
		collections.GET("/:address/history-sales", v1.HistorySalesHandler(svcCtx))
		// 获取NFT Item的owner信息
		collections.GET("/:address/:token_id/owner", v1.ItemOwnerHandler(svcCtx))
		// 刷新NFT Item的metadata
		collections.POST("/:address/:token_id/metadata", v1.ItemMetadataRefreshHandler(svcCtx))

		// 获取NFT集合排名信息，使用缓存中间件，缓存时间为60秒
		collections.GET("/ranking", middleware.CacheApi(svcCtx.KvStore, 60), v1.TopRankingHandler(svcCtx))
	}

	// 创建一个名为 /activities 的子路由组
	activities := apiV1.Group("/activities")
	{
		// 批量获取activity信息
		activities.GET("", v1.ActivityMultiChainHandler(svcCtx))
	}

	// 创建一个名为 /portfolio 的子路由组
	portfolio := apiV1.Group("/portfolio")
	{
		// 获取用户拥有Collection信息
		portfolio.GET("/collections", v1.UserMultiChainCollectionsHandler(svcCtx))
		// 查询用户拥有nft的Item基本信息
		portfolio.GET("/items", v1.UserMultiChainItemsHandler(svcCtx))
		// 查询用户挂单的Listing信息
		portfolio.GET("/listings", v1.UserMultiChainListingsHandler(svcCtx))
		// 查询用户挂单的Bids信息
		portfolio.GET("/bids", v1.UserMultiChainBidsHandler(svcCtx))
	}

	// 创建一个名为 /bid-orders 的子路由组
	orders := apiV1.Group("/bid-orders")
	{
		// 批量查询出价信息
		orders.GET("", v1.OrderInfosHandler(svcCtx))
	}
}
