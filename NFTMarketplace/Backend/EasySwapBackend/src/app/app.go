package app

import (
	"context"

	"github.com/ProjectsTask/EasySwapBase/logger/xzap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ProjectsTask/EasySwapBackend/src/config"
	"github.com/ProjectsTask/EasySwapBackend/src/service/svc"
)

type Platform struct {
	config    *config.Config
	router    *gin.Engine
	serverCtx *svc.ServerCtx
}

// NewPlatform 函数用于创建一个新的 Platform 结构体实例，并初始化其成员变量。
//
// 参数:
// - config: *config.Config 类型，指向配置信息的结构体指针。
// - router: *gin.Engine 类型，指向路由引擎的结构体指针。
// - serverCtx: *svc.ServerCtx 类型，指向服务器上下文的结构体指针。
//
// 返回值:
// - *Platform: 返回初始化后的 Platform 结构体指针。
// - error: 如果初始化失败，则返回错误信息；否则返回 nil。
func NewPlatform(config *config.Config, router *gin.Engine, serverCtx *svc.ServerCtx) (*Platform, error) {
	// 创建一个Platform结构体实例，并初始化其成员变量
	return &Platform{
		// 初始化config成员变量
		config:    config,
		// 初始化router成员变量
		router:    router,
		// 初始化serverCtx成员变量
		serverCtx: serverCtx,
	}, nil
}

// Start 启动Platform实例
// 使用zap日志库记录日志，并启动路由并监听指定端口
func (p *Platform) Start() {
	// 使用zap日志库记录日志
	xzap.WithContext(context.Background()).Info("EasySwap-End run", zap.String("port", p.config.Api.Port))

	// 启动路由并监听指定端口
	if err := p.router.Run(p.config.Api.Port); err != nil {
		// 如果启动失败，则抛出异常
		panic(err)
	}
}
