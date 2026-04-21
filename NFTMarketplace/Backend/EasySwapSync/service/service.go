package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/ProjectsTask/EasySwapBase/chain"
	"github.com/ProjectsTask/EasySwapBase/chain/chainclient"
	"github.com/ProjectsTask/EasySwapBase/ordermanager"
	"github.com/ProjectsTask/EasySwapBase/stores/xkv"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/kv"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"

	"github.com/ProjectsTask/EasySwapSync/service/orderbookindexer"

	"github.com/ProjectsTask/EasySwapSync/model"
	"github.com/ProjectsTask/EasySwapSync/service/collectionfilter"
	"github.com/ProjectsTask/EasySwapSync/service/config"
)

type Service struct {
	ctx              context.Context
	config           *config.Config
	kvStore          *xkv.Store
	db               *gorm.DB
	wg               *sync.WaitGroup
	collectionFilter *collectionfilter.Filter
	orderbookIndexer *orderbookindexer.Service
	orderManager     *ordermanager.OrderManager
}

// New 函数用于创建一个新的 Service 实例。
// 参数 ctx 是一个上下文对象，用于控制 Service 的生命周期。
// 参数 cfg 是一个指向 config.Config 结构体的指针，包含了 Service 的配置信息。
// 返回值是一个指向 Service 结构体的指针和一个错误对象。如果创建成功，错误对象为 nil。
func New(ctx context.Context, cfg *config.Config) (*Service, error) {
	// 初始化一个 kv.KvConf 类型的变量，用于存储 Redis 节点配置。
	var kvConf kv.KvConf
	// 遍历配置文件中的 Redis 节点配置。
	for _, con := range cfg.Kv.Redis {
		// 将每个 Redis 节点的配置添加到 kvConf 中。
		kvConf = append(kvConf, cache.NodeConf{
			RedisConf: redis.RedisConf{
				// Redis 节点的主机地址
				Host: con.Host,
				// Redis 节点的类型
				Type: con.Type,
				// Redis 节点的密码
				Pass: con.Pass,
			},
			// 节点的权重，用于负载均衡
			Weight: 2,
		})
	}

	// 使用配置好的 Redis 节点信息创建一个新的 xkv.Store 实例。
	kvStore := xkv.NewStore(kvConf)

	// 声明一个错误变量，用于存储可能出现的错误。
	var err error
	// 根据配置文件中的数据库信息创建一个新的数据库实例。
	db := model.NewDB(cfg.DB)
	// 创建一个新的 collectionfilter.Filter 实例，用于过滤集合信息。
	collectionFilter := collectionfilter.New(ctx, db, cfg.ChainCfg.Name, cfg.ProjectCfg.Name)
	// 创建一个新的 ordermanager.OrderManager 实例，用于管理订单信息。
	orderManager := ordermanager.New(ctx, db, kvStore, cfg.ChainCfg.Name, cfg.ProjectCfg.Name)
	// 声明一个指向 orderbookindexer.Service 结构体的指针，用于同步订单簿信息。
	var orderbookSyncer *orderbookindexer.Service
	// 声明一个 chainclient.ChainClient 接口类型的变量，用于与区块链节点通信。
	var chainClient chainclient.ChainClient
	// 打印链客户端的 URL 信息，方便调试。
	fmt.Println("chainClient url:" + cfg.AnkrCfg.HttpsUrl + cfg.AnkrCfg.ApiKey)

	// 根据配置文件中的链 ID 和 API 密钥创建一个新的链客户端实例。
	chainClient, err = chainclient.New(int(cfg.ChainCfg.ID), cfg.AnkrCfg.HttpsUrl+cfg.AnkrCfg.ApiKey)
	// 检查创建链客户端时是否出现错误。
	if err != nil {
		// 如果出现错误，返回 nil 和一个包装后的错误信息。
		return nil, errors.Wrap(err, "failed on create evm client")
	}

	// 根据配置文件中的链 ID 选择不同的同步器。
	switch cfg.ChainCfg.ID {
	case chain.EthChainID, chain.OptimismChainID, chain.SepoliaChainID:
		// 创建一个新的 orderbookindexer.Service 实例，用于同步订单簿信息。
		orderbookSyncer = orderbookindexer.New(ctx, cfg, db, kvStore, chainClient, cfg.ChainCfg.ID, cfg.ChainCfg.Name, orderManager)
	}
	// 再次检查是否出现错误，这一步可能是多余的，需要确认逻辑。
	if err != nil {
		// 如果出现错误，返回 nil 和一个包装后的错误信息。
		return nil, errors.Wrap(err, "failed on create trade info server")
	}
	// 初始化一个 Service 结构体实例。
	manager := Service{
		// 上下文对象，用于控制 Service 的生命周期
		ctx: ctx,
		// 配置信息
		config: cfg,
		// 数据库实例
		db: db,
		// 键值存储实例
		kvStore: kvStore,
		// 集合过滤器实例
		collectionFilter: collectionFilter,
		// 订单簿同步器实例
		orderbookIndexer: orderbookSyncer,
		// 订单管理器实例
		orderManager: orderManager,
		// 同步等待组，用于等待所有 goroutine 完成
		wg: &sync.WaitGroup{},
	}
	// 返回 Service 实例的指针和 nil 错误。
	return &manager, nil
}

// Start 方法用于启动 Service 实例中的各个组件。
// 该方法会按顺序调用集合过滤器的预加载方法，以及订单簿同步器和订单管理器的启动方法。
// 如果在预加载集合时出现错误，该方法会返回一个包装后的错误信息。
// 返回值为错误对象，如果启动过程中没有出现错误，返回 nil。
func (s *Service) Start() error {
	// 调用集合过滤器的 PreloadCollections 方法，预先加载集合信息到过滤器中。
	// 如果预加载过程中出现错误，将错误信息包装并返回。
	if err := s.collectionFilter.PreloadCollections(); err != nil {
		return errors.Wrap(err, "failed on preload collection to filter")
	}

	// 启动订单簿同步器，开始同步订单簿信息。
	s.orderbookIndexer.Start()
	// 启动订单管理器，开始管理订单信息。
	s.orderManager.Start()
	// 如果所有组件都成功启动，返回 nil 表示没有错误。
	return nil
}
