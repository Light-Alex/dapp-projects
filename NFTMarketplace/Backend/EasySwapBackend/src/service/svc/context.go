// 定义包名为 svc，用于服务相关的上下文管理
package svc

import (
	"context"

	// 引入 NFT 链服务相关的包
	"github.com/ProjectsTask/EasySwapBase/chain/nftchainservice"
	// 引入日志相关的包
	"github.com/ProjectsTask/EasySwapBase/logger/xzap"
	// 引入数据库相关的包
	"github.com/ProjectsTask/EasySwapBase/stores/gdb"
	// 引入 KV 存储相关的包
	"github.com/ProjectsTask/EasySwapBase/stores/xkv"
	// 引入错误处理相关的包
	"github.com/pkg/errors"
	// 引入 go-zero 框架的缓存相关包
	"github.com/zeromicro/go-zero/core/stores/cache"
	// 引入 go-zero 框架的 KV 存储相关包
	"github.com/zeromicro/go-zero/core/stores/kv"
	// 引入 go-zero 框架的 Redis 相关包
	"github.com/zeromicro/go-zero/core/stores/redis"
	// 引入 GORM 数据库操作包
	"gorm.io/gorm"

	// 引入项目配置相关的包
	"github.com/ProjectsTask/EasySwapBackend/src/config"
	// 引入数据访问对象相关的包
	"github.com/ProjectsTask/EasySwapBackend/src/dao"
)

// ServerCtx 结构体用于保存服务的上下文信息
type ServerCtx struct {
	// 配置对象，包含服务的各种配置信息
	C *config.Config
	// GORM 数据库连接对象
	DB *gorm.DB
	// 图片管理器，用于处理图片相关操作
	//ImageMgr image.ImageManager
	// 数据访问对象，用于操作数据库
	Dao *dao.Dao
	// KV 存储对象，用于缓存数据
	KvStore *xkv.Store
	// 排名相关的键，可能用于排序或统计
	RankKey string
	// 链服务映射，键为链 ID，值为对应的链服务实例
	NodeSrvs map[int64]*nftchainservice.Service
}

// NewServiceContext 函数用于创建并初始化服务上下文
func NewServiceContext(c *config.Config) (*ServerCtx, error) {
	var err error
	// 图片管理器初始化
	//imageMgr, err = image.NewManager(c.ImageCfg)
	//if err != nil {
	//    return nil, errors.Wrap(err, "failed on create image manager")
	//}

	// 根据配置设置日志
	_, err = xzap.SetUp(c.Log)
	if err != nil {
		// 如果日志设置失败，返回错误
		return nil, err
	}

	// 初始化 KV 配置
	var kvConf kv.KvConf
	// 遍历配置中的 Redis 配置
	for _, con := range c.Kv.Redis {
		// 将每个 Redis 配置添加到 KV 配置中
		kvConf = append(kvConf, cache.NodeConf{
			RedisConf: redis.RedisConf{
				Host: con.Host,
				Type: con.Type,
				Pass: con.Pass,
			},
			Weight: 1,
		})
	}

	// 使用配置好的 KV 配置创建 KV 存储实例
	store := xkv.NewStore(kvConf)
	// 根据配置创建数据库连接
	db, err := gdb.NewDB(&c.DB)
	if err != nil {
		// 如果数据库连接创建失败，返回错误
		return nil, err
	}

	// 初始化链服务映射
	nodeSrvs := make(map[int64]*nftchainservice.Service)
	// 遍历支持的链配置
	for _, supported := range c.ChainSupported {
		// 为每个支持的链创建链服务实例
		nodeSrvs[int64(supported.ChainID)], err = nftchainservice.New(context.Background(), supported.Endpoint, supported.Name, supported.ChainID,
			c.MetadataParse.NameTags, c.MetadataParse.ImageTags, c.MetadataParse.AttributesTags,
			c.MetadataParse.TraitNameTags, c.MetadataParse.TraitValueTags)

		if err != nil {
			// 如果链服务创建失败，返回错误
			return nil, errors.Wrap(err, "failed on start onchain sync service")
		}
	}

	// 创建数据访问对象实例
	dao := dao.New(context.Background(), db, store)
	// 创建服务上下文实例
	serverCtx := NewServerCtx(
		WithDB(db),
		WithKv(store),
		// 注释掉的图片管理器设置，可能后续会启用
		//WithImageMgr(imageMgr),
		WithDao(dao),
	)
	// 设置服务上下文的配置
	serverCtx.C = c
	// 设置服务上下文的链服务映射
	serverCtx.NodeSrvs = nodeSrvs

	// 返回服务上下文实例和 nil 错误
	return serverCtx, nil
}
