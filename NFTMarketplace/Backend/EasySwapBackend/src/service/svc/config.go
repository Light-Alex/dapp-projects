package svc

import (
	"github.com/ProjectsTask/EasySwapBase/evm/erc"
	//"github.com/ProjectsTask/EasySwapBase/image"
	"github.com/ProjectsTask/EasySwapBase/stores/xkv"
	"gorm.io/gorm"

	"github.com/ProjectsTask/EasySwapBackend/src/dao"
)

type CtxConfig struct {
	db *gorm.DB
	//imageMgr image.ImageManager
	dao     *dao.Dao
	KvStore *xkv.Store
	Evm     erc.Erc
}

type CtxOption func(conf *CtxConfig)

func NewServerCtx(options ...CtxOption) *ServerCtx {
	// 初始化 CtxConfig 对象
	c := &CtxConfig{}

	// 遍历所有传入的选项
	for _, opt := range options {
		// 调用选项函数并传入 c 对象
		opt(c)
	}

	return &ServerCtx{
		// 返回数据库连接
		DB: c.db,
		//ImageMgr: c.imageMgr,
		// 返回键值存储
		KvStore: c.KvStore,
		// 返回数据访问对象
		Dao:     c.dao,
	}
}

// WithKv 函数接收一个指向 xkv.Store 的指针作为参数，并返回一个 CtxOption 类型。
// CtxOption 是一个函数类型，它接收一个指向 CtxConfig 的指针作为参数，用于配置 CtxConfig 的相关属性。
// WithKv 返回的 CtxOption 函数会将传入的 kv 参数赋值给 CtxConfig 的 KvStore 字段。
func WithKv(kv *xkv.Store) CtxOption {
	// 返回一个匿名函数，该函数接收一个指向CtxConfig的指针作为参数
	return func(conf *CtxConfig) {
		// 将传入的kv赋值给conf的KvStore字段
		conf.KvStore = kv
	}
}

func WithDB(db *gorm.DB) CtxOption {
	// 返回一个闭包函数
	return func(conf *CtxConfig) {
		// 将传入的db赋值给CtxConfig结构体中的db字段
		conf.db = db
	}
}

func WithDao(dao *dao.Dao) CtxOption {
	// 返回一个函数，该函数接受一个指向CtxConfig的指针作为参数
	return func(conf *CtxConfig) {
		// 将传入的dao赋值给CtxConfig的dao字段
		conf.dao = dao
	}
}
