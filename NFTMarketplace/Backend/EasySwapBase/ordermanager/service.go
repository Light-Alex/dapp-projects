package ordermanager

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/threading"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/ProjectsTask/EasySwapBase/logger/xzap"
	"github.com/ProjectsTask/EasySwapBase/stores/gdb/orderbookmodel/multi"
	"github.com/ProjectsTask/EasySwapBase/stores/xkv"
)

const (
	// WheelSize : length of time wheel, 1s for a unit, 1h for a cycle
	WheelSize           = 3600
	List                = 3
	CacheOrdersQueuePre = "cache:es:orders:%s"
)

func GenOrdersCacheKey(chain string) string {
	return fmt.Sprintf(CacheOrdersQueuePre, chain)
}

// Activity ：collection order-activity
type Order struct {
	// order Id
	orderID        string
	CollectionAddr string
	// chain suffix name: ethw/bsc
	ChainSuffix string
	// expireIn - createIn (unit: s)
	CycleCount int64
	// position of the task on the time wheel
	WheelPosition int64

	Next *Order
}

type wheel struct {
	// linked list
	NotifyActivities *Order
}

type OrderManager struct {
	chain string

	// cycle time wheel
	TimeWheel [WheelSize]wheel
	// current time wheel index
	CurrentIndex int64

	collectionOrders map[string]*collectionTradeInfo

	collectionListedCh chan string
	project            string

	Xkv *xkv.Store
	DB  *gorm.DB
	Ctx context.Context
	Mux *sync.RWMutex
}

// NewDelayQueue : create func instance entrance
func New(ctx context.Context, db *gorm.DB, xkv *xkv.Store, chain string, project string) *OrderManager {
	// 初始化 OrderManager 实例
	return &OrderManager{
		// 设置链的名称
		chain:              chain,
		// 设置 xkv 存储
		Xkv:                xkv,
		// 设置数据库连接
		DB:                 db,
		// 设置上下文
		Ctx:                ctx,
		// 初始化读写互斥锁
		Mux:                new(sync.RWMutex),
		// 初始化集合订单映射
		collectionOrders:   make(map[string]*collectionTradeInfo),
		// 初始化集合上市通道
		collectionListedCh: make(chan string, 1000),
		// 设置项目名称
		project:            project,
	}
}

func (om *OrderManager) Start() {
	// listen redis cache
	threading.GoSafe(om.ListenNewListingLoop) // 处理新订单
	threading.GoSafe(om.orderExpiryProcess)   // 处理订单过期状态
	threading.GoSafe(om.floorPriceProcess)    // 处理floorprice更新
	threading.GoSafe(om.listCountProcess)     // 处理listCount更新
}

func (om *OrderManager) Stop() {
}

type ListingInfo struct {
	ExpireIn       int64           `json:"expire_in"`
	OrderId        string          `json:"order_id"`
	CollectionAddr string          `json:"collection_addr"`
	TokenID        string          `json:"token_id"`
	Price          decimal.Decimal `json:"price"`
	Maker          string          `json:"maker"`
}

// ListenNewListingLoop 监听新上架的订单，并处理订单逻辑
func (om *OrderManager) ListenNewListingLoop() {
	key := GenOrdersCacheKey(om.chain)
	for {
		// 从缓存中获取订单
		result, err := om.Xkv.Lpop(key)
		if err != nil || result == "" {
			// 如果获取订单失败或结果为空
			if err != nil && err != redis.Nil {
				// 记录获取订单失败日志
				xzap.WithContext(context.Background()).Warn("failed on get order from cache", zap.Error(err), zap.String("result", result))
			}
			// 休眠1秒
			time.Sleep(1 * time.Second)
			// 继续循环
			continue
		}

		// 记录从缓存中获取到订单
		xzap.WithContext(om.Ctx).Info("get listing from cache", zap.String("result", result))

		// 定义订单结构体
		var listing ListingInfo
		// 将结果反序列化为订单结构体
		if err := json.Unmarshal([]byte(result), &listing); err != nil {
			// 记录反序列化订单失败日志
			xzap.WithContext(om.Ctx).Warn("failed on Unmarshal order info", zap.Error(err))
			// 继续循环
			continue
		}
		// 检查订单ID是否为空
		if listing.OrderId == "" {
			// 记录订单ID为空错误日志
			xzap.WithContext(om.Ctx).Error("invalid null order id")
			// 继续循环
			continue
		}

		// 判断订单是否过期
		if listing.ExpireIn < time.Now().Unix() { // 订单已经过期
			// 记录订单过期日志
			xzap.WithContext(om.Ctx).Info("expired activity order", zap.String("order_id", listing.OrderId))

			// 更新订单状态
			if err := om.updateOrdersStatus(listing.OrderId, multi.OrderStatusExpired); err != nil {
				// 记录更新订单状态失败日志
				xzap.WithContext(om.Ctx).Error("failed on update activity status", zap.String("order_id", listing.OrderId), zap.Error(err))
			}

			// 添加更新floorprice事件
			if err := om.addUpdateFloorPriceEvent(&TradeEvent{
				EventType:      Expired,
				CollectionAddr: listing.CollectionAddr,
				TokenID:        listing.TokenID,
				OrderId:        listing.OrderId,
				From:           listing.Maker,
			}); err != nil {
				// 记录添加更新floorprice事件失败日志
				xzap.WithContext(om.Ctx).Error("failed on add update floor price event", zap.String("order_id", listing.OrderId), zap.Error(err))
			}
			// 继续循环
			continue
		} else { // 订单未过期
			// 添加更新floorprice事件
			if err := om.addUpdateFloorPriceEvent(&TradeEvent{ // 添加更新floorprice事件
				EventType:      Listing,
				CollectionAddr: listing.CollectionAddr,
				TokenID:        listing.TokenID,
				OrderId:        listing.OrderId,
				Price:          listing.Price,
				From:           listing.Maker,
			}); err != nil {
				// 记录推送订单到更新价格队列失败日志
				xzap.WithContext(om.Ctx).Error("failed on push order to update price queue", zap.Error(err), zap.String("order_id", listing.OrderId),
					zap.String("order_id", listing.OrderId),
					zap.String("price", listing.Price.String()),
					zap.String("chain", om.chain))
			}

			// 添加到订单过期检查队列
			delaySeconds := listing.ExpireIn - time.Now().Unix()
			if err := om.addToOrderExpiryCheckQueue(delaySeconds, om.chain, listing.OrderId, listing.CollectionAddr); err != nil {
				// 记录推送订单到过期检查队列失败日志
				xzap.WithContext(om.Ctx).Error("failed on push order to expired check queue", zap.Error(err), zap.String("order_id", listing.OrderId),
					zap.String("chain", om.chain))
			}
		}
	}
}

// AddToOrderManagerQueue 将订单添加到订单管理器的队列中
//
// 参数:
//   order: 要添加到队列中的订单
//
// 返回值:
//   error: 如果出现错误，则返回错误信息；否则返回nil
func (om *OrderManager) AddToOrderManagerQueue(order *multi.Order) error {
	// 检查订单是否有有效的 TokenId
	if order.TokenId == "" {
		return errors.New("order manger need token id")
	}

	// 将订单信息序列化为 JSON 字符串
	rawInfo, err := json.Marshal(ListingInfo{
		ExpireIn:       order.ExpireTime,  // 订单过期时间
		OrderId:        order.OrderID,     // 订单ID
		CollectionAddr: order.CollectionAddress, // 收款地址
		TokenID:        order.TokenId,     // Token ID
		Price:          order.Price,       // 价格
		Maker:          order.Maker,       // 订单创建者
	})
	if err != nil {
		return errors.Wrap(err, "failed on marshal listing info")  // 包装错误，序列化订单信息失败
	}

	// 将序列化后的订单信息添加到队列中
	if _, err := om.Xkv.Lpush(GenOrdersCacheKey(om.chain), string(rawInfo)); err != nil {
		return errors.Wrap(err, "failed on add to queue")  // 包装错误，添加到队列失败
	}

	return nil
}
