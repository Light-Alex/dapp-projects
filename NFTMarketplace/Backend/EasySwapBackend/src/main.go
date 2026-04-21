package main

import (
	"flag"
	_ "net/http/pprof"

	"github.com/ProjectsTask/EasySwapBackend/src/api/router"
	"github.com/ProjectsTask/EasySwapBackend/src/app"
	"github.com/ProjectsTask/EasySwapBackend/src/config"
	"github.com/ProjectsTask/EasySwapBackend/src/service/svc"
)

const (
	// port       = ":9000"// 服务器端口号
	repoRoot          = ""                     // 项目根目录路径
	defaultConfigPath = "./config/config.toml" // 默认配置文件路径
)

func main() {
	// 解析命令行参数，获取配置文件路径
	conf := flag.String("conf", defaultConfigPath, "conf file path")
	flag.Parse()
	// 读取并解析配置文件
	c, err := config.UnmarshalConfig(*conf)
	if err != nil {
		panic(err)
	}
	// 校验配置文件中的链配置
	for _, chain := range c.ChainSupported {
		if chain.ChainID == 0 || chain.Name == "" {
			panic("invalid chain_suffix config")
		}
	}
	// 创建服务上下文
	serverCtx, err := svc.NewServiceContext(c)
	if err != nil {
		panic(err)
	}
	// Initialize router初始化路由
	r := router.NewRouter(serverCtx)
	// 初始化应用程序
	app, err := app.NewPlatform(c, r, serverCtx)
	if err != nil {
		panic(err)
	}
	// 启动应用
	app.Start()
}
