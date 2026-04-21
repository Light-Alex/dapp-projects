package dao

import (
	"context"

	"github.com/ProjectsTask/EasySwapBase/stores/xkv"
	"gorm.io/gorm"
)

// Dao is show dao.
type Dao struct {
	ctx context.Context

	DB      *gorm.DB
	KvStore *xkv.Store
}

// New 函数用于创建一个新的 Dao 实例
//
// 参数：
//     ctx context.Context: 上下文对象，用于传递请求范围内的值、取消信号等
//     db *gorm.DB: GORM 数据库连接对象，用于数据库操作
//     kvStore *xkv.Store: 键值存储对象，用于键值对存储操作
//
// 返回值：
//     *Dao: 初始化后的 Dao 实例指针
func New(ctx context.Context, db *gorm.DB, kvStore *xkv.Store) *Dao {
	// 初始化 Dao 结构体并返回
	return &Dao{
		// 设置上下文
		ctx:     ctx,
		// 设置数据库连接
		DB:      db,
		// 设置键值存储
		KvStore: kvStore,
	}
}
