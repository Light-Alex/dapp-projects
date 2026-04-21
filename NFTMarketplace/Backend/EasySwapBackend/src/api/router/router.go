package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/ProjectsTask/EasySwapBackend/src/api/middleware"

	"github.com/ProjectsTask/EasySwapBackend/src/service/svc"
)

// NewRouter 创建一个新的Gin引擎实例，并配置必要的中间件和路由。
// 参数 svcCtx 是一个指向服务上下文的指针，包含了应用程序所需的服务和配置。
// 返回一个指向Gin引擎实例的指针。
func NewRouter(svcCtx *svc.ServerCtx) *gin.Engine {
	// 强制Gin在控制台输出彩色日志
	gin.ForceConsoleColor()
	// 设置Gin的运行模式为发布模式
	gin.SetMode(gin.ReleaseMode)
	// 新建一个Gin引擎实例
	r := gin.New()
	// 使用恢复中间件，用于捕获并处理panic，防止应用崩溃
	r.Use(middleware.RecoverMiddleware())
	// 使用日志中间件，记录请求的相关信息
	r.Use(middleware.RLog())
	// 使用CORS中间件，处理跨域请求
	r.Use(cors.New(cors.Config{
		// 允许所有来源的请求
		AllowAllOrigins: true,
		// 允许的HTTP方法
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		// 允许的请求头
		AllowHeaders: []string{"Origin", "Content-Length", "Content-Type", "X-CSRF-Token", "Authorization", "AccessToken", "Token"},
		// 暴露给客户端的响应头
		ExposeHeaders: []string{"Content-Length", "Content-Type", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "X-GW-Error-Code", "X-GW-Error-Message"},
		// 允许携带凭证（如cookie）
		AllowCredentials: true,
		// 预检请求的最大缓存时间
		MaxAge: 1 * time.Hour,
	}))
	// 加载v1版本的路由
	loadV1(r, svcCtx)

	// 返回配置好的Gin引擎实例
	return r
}
