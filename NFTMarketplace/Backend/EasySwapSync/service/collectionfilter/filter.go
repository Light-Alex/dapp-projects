package collectionfilter

import (
	"context"
	"strings"
	"sync"

	"github.com/ProjectsTask/EasySwapBase/stores/gdb"
	"github.com/pkg/errors"

	"gorm.io/gorm"

	"github.com/ProjectsTask/EasySwapSync/service/comm"
)

// Filter is a thread-safe structure to store a set of strings.
type Filter struct {
	ctx     context.Context
	db      *gorm.DB
	chain   string
	set     map[string]bool // Set of strings
	lock    *sync.RWMutex   // Read/Write mutex for thread safety
	project string
}

// New 创建一个新的 Filter 实例。
// 该函数接收上下文、数据库连接、链标识和项目标识作为参数，
// 并返回一个初始化后的 Filter 指针。
// 参数:
// - ctx: 上下文环境，用于控制操作的生命周期。
// - db: 数据库连接，用于与数据库交互。
// - chain: 链标识，指定操作所在的区块链。
// - project: 项目标识，用于区分不同的项目。
// 返回值:
// - *Filter: 初始化后的 Filter 指针。
func New(ctx context.Context, db *gorm.DB, chain string, project string) *Filter {
	// 创建一个新的Filter结构体实例
	return &Filter{
		// 上下文环境
		ctx: ctx,
		// 数据库连接
		db: db,
		// 链标识
		chain: chain,
		// 集合，用于存储键值对
		set: make(map[string]bool),
		// 读写锁
		lock: &sync.RWMutex{},
		// 项目标识
		project: project,
	}
}

// Add 向过滤器中插入一个新元素。
// 在插入之前，元素会被转换为小写。
// 参数:
// - element: 要插入的元素。
func (f *Filter) Add(element string) {
	// 获取写锁，防止其他 goroutine 同时读写
	f.lock.Lock()
	// 延迟解锁，确保在函数返回时释放锁
	defer f.lock.Unlock()
	// 将转换为小写后的元素添加到过滤器中，标记为存在
	f.set[strings.ToLower(element)] = true
}

// Remove 从过滤器中删除指定的元素。
// 在删除之前，元素会被转换为小写。
// 参数:
// - element: 要删除的元素。
func (f *Filter) Remove(element string) {
	// 获取写锁，防止其他 goroutine 同时读写
	f.lock.Lock()
	// 延迟解锁，确保在函数返回时释放锁
	defer f.lock.Unlock()
	// 从过滤器中删除转换为小写后的元素
	delete(f.set, strings.ToLower(element))
}

// Contains 检查过滤器中是否包含指定的元素。
// 在检查之前，元素会被转换为小写。
// 参数:
// - element: 要检查的元素。
// 返回值:
// - bool: 如果过滤器中包含该元素，返回 true；否则返回 false。
func (f *Filter) Contains(element string) bool {
	// 获取读锁，允许多个 goroutine 同时读取
	f.lock.RLock()
	// 延迟解锁，确保在函数返回时释放锁
	defer f.lock.RUnlock()
	// 检查转换为小写后的元素是否存在于过滤器中
	_, exists := f.set[strings.ToLower(element)]
	// 返回检查结果
	return exists
}

// PreloadCollections 预加载符合条件的集合地址到过滤器中。
// 该函数从数据库中查询所有 floor_price_status 为 CollectionFloorPriceImported 的集合地址，
// 并将这些地址添加到过滤器中。
// 返回值：
// - 如果查询数据库或添加地址到过滤器过程中出现错误，返回包装后的错误信息；
// - 如果操作成功，返回 nil。
func (f *Filter) PreloadCollections() error {
	// 定义一个字符串切片，用于存储从数据库中查询到的集合地址
	var addresses []string
	// 定义一个错误变量，用于存储可能出现的错误信息
	var err error

	// 从数据库中直接查询集合地址
	// 使用 WithContext 方法设置查询的上下文
	// Table 方法指定要查询的表名，通过 gdb.GetMultiProjectCollectionTableName 函数动态生成表名
	// Select 方法指定要查询的字段为 address
	// Where 方法设置查询条件，过滤出 floor_price_status 等于 comm.CollectionFloorPriceImported 的记录
	// Scan 方法将查询结果扫描到 addresses 切片中
	err = f.db.WithContext(f.ctx).
		Table(gdb.GetMultiProjectCollectionTableName(f.project, f.chain)).
		Select("address").
		Where("floor_price_status = ?", comm.CollectionFloorPriceImported).
		Scan(&addresses).Error

	// 检查查询过程中是否出现错误
	if err != nil {
		// 如果出现错误，使用 errors.Wrap 函数包装错误信息，便于调试和定位问题
		return errors.Wrap(err, "failed on query collections from db")
	}

	// 遍历查询到的地址切片，将每个地址添加到过滤器中
	for _, address := range addresses {
		// 调用 Add 方法将地址添加到过滤器中
		f.Add(address)
	}

	// 如果操作成功，返回 nil 表示没有错误
	return nil
}
