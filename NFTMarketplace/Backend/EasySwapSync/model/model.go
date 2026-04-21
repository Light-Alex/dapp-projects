package model

import (
	"context"

	"gorm.io/gorm"

	"github.com/ProjectsTask/EasySwapBase/stores/gdb"
)

// NewDB 根据传入的数据库配置创建一个新的 gorm.DB 实例，并初始化服务模型信息。
// 参数 ndb 是一个指向 gdb.Config 结构体的指针，包含了数据库的配置信息。
// 返回值是一个指向 gorm.DB 结构体的指针，表示创建好的数据库实例。
func NewDB(ndb *gdb.Config) *gorm.DB {
	// 调用 gdb.MustNewDB 函数，根据传入的配置创建一个新的数据库实例。
	// 该函数会确保数据库连接成功，如果失败会抛出异常。
	db := gdb.MustNewDB(ndb)
	// 创建一个背景上下文，用于后续的数据库操作。
	ctx := context.Background()
	// 调用 InitModel 函数，初始化服务模型信息。
	// 该函数会设置数据库表的选项，如引擎、字符集和排序规则等。
	err := InitModel(ctx, db)
	// 检查初始化过程中是否出现错误。
	if err != nil {
		// 如果出现错误，使用 panic 函数抛出异常，终止程序运行。
		panic(err)
	}

	// 返回创建好的数据库实例。
	return db
}

// InitModel 初始化服务模型信息
func InitModel(ctx context.Context, db *gorm.DB) error {
	// 设置数据库表选项
	err := db.Set(
		"gorm:table_options",
		"ENGINE=InnoDB AUTO_INCREMENT=1 CHARACTER SET=utf8mb4 COLLATE=utf8mb4_general_ci",
	).Error
	if err != nil {
		// 如果设置数据库表选项失败，则返回错误
		return err
	}

	return nil
}
