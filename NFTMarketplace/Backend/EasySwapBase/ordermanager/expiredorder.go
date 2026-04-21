package ordermanager

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ProjectsTask/EasySwapBase/stores/gdb"

	"github.com/ProjectsTask/EasySwapBase/logger/xzap"
	"github.com/ProjectsTask/EasySwapBase/stores/gdb/orderbookmodel/multi"
)

// orderExpiryProcess 函数负责处理订单过期的逻辑,主要包含以下功能:
// 1. 使用defer recover防止panic导致主协程退出
// 2. 启动时从数据库加载所有活跃订单到时间轮队列
// 3. 每秒检查一次时间轮,处理到期的订单
func (om *OrderManager) orderExpiryProcess() {
	// 1. 使用 defer recover 来捕获可能的 panic,防止主协程死掉
	defer func() {
		if r := recover(); r != nil {
			xzap.WithContext(om.Ctx).Error("[Order Manage] dq process recovered: " + fmt.Sprintf("%v", r))
		}
	}()

	// 2. 启动时从数据库加载所有活跃订单到时间轮队列中
	if err := om.loadOrdersToQueue(); err != nil {
		xzap.WithContext(om.Ctx).Error("[Order Manage] load orders to queue", zap.Error(err))
		return
	}

	// 3. 每秒检查一次时间轮
	for {
		select {
		case <-time.After(time.Second * 1): // 每秒执行一次检查
			// 如果当前索引超过时间轮大小,则取模重置
			if om.CurrentIndex >= WheelSize {
				om.CurrentIndex = om.CurrentIndex % WheelSize
			}

			// 获取当前时间槽的任务链表头
			taskLinkHead := om.TimeWheel[om.CurrentIndex].NotifyActivities
			headIndex := om.CurrentIndex
			om.CurrentIndex++

			// prev指向前一个节点,p指向当前节点
			prev := taskLinkHead
			p := taskLinkHead

			// 遍历当前时间槽的所有任务
			for p != nil {
				// 如果任务的循环计数为0,说明到期需要处理
				if p.CycleCount == 0 {
					// 异步更新订单状态,避免阻塞主循环
					go func(chain string, orderId string, collectionAddr string) {
						if err := om.updateOrderState(orderId, collectionAddr); err != nil {
							xzap.WithContext(om.Ctx).Error("failed on update order status", zap.Error(err), zap.String("chain", chain), zap.String("order_id", orderId))
						}
					}(p.ChainSuffix, p.orderID, p.CollectionAddr)

					// 从链表中删除该节点
					if prev == p { // 如果是头节点
						om.TimeWheel[headIndex].NotifyActivities = p.Next
						prev = p.Next
						p = p.Next
					} else { // 如果不是头节点
						prev.Next = p.Next
						p = p.Next
					}
				} else {
					// 如果任务未到期,循环计数减1
					p.CycleCount--
					prev = p
					p = p.Next
				}
			}
		}
	}
}

// loadOrdersToQueue 函数负责在系统启动时加载所有活跃订单并处理它们的过期状态
// 主要功能包括:
// 1. 从数据库分批加载所有活跃订单
// 2. 检查每个订单是否已过期:
//   - 已过期的订单:更新状态为过期并触发地板价更新事件
//   - 未过期的订单:添加到延迟队列等待过期检查
func (om *OrderManager) loadOrdersToQueue() error {
	// 分批加载所有活跃订单
	var totalOrders []*multi.Order
	var id int64
	for {
		var orders []*multi.Order
		if err := om.DB.WithContext(om.Ctx).Table(gdb.GetMultiProjectOrderTableName(om.project, om.chain)).
			Select("id, order_id, collection_address, price, expire_time").
			Where("order_status = ? and id > ?", multi.OrderStatusActive, id).
			Order("id asc").Limit(1000).
			Scan(&orders).Error; err != nil {
			return errors.Wrap(err, "failed on get collection orders")
		}
		totalOrders = append(totalOrders, orders...)
		if len(orders) < 1000 {
			break
		}

		id = orders[len(orders)-1].ID
	}

	if totalOrders == nil || len(totalOrders) == 0 {
		return nil
	}

	// 分别记录已过期和未过期的订单
	var expiredOrderIDs []int64
	var expiredOrders []*multi.Order
	for _, order := range totalOrders {
		if order.ExpireTime < time.Now().Unix() { // 已过期
			expiredOrderIDs = append(expiredOrderIDs, order.ID)
			expiredOrders = append(expiredOrders, order)
		} else { // 未过期,添加到延迟队列
			delaySeconds := order.ExpireTime - time.Now().Unix()
			if err := om.addToOrderExpiryCheckQueue(delaySeconds, om.chain, order.OrderID, order.CollectionAddress); err != nil {
				xzap.WithContext(om.Ctx).Error("[Order Manage] failed on add order to delay queue",
					zap.Error(err))
			}
		}
	}

	// 批量更新已过期订单的状态
	for i := 0; i < len(expiredOrderIDs); i += 100 {
		end := i + MaxBatchReqNum
		if i+MaxBatchReqNum >= len(expiredOrderIDs) {
			end = len(expiredOrderIDs)
		}

		if err := om.DB.WithContext(om.Ctx).Table(gdb.GetMultiProjectOrderTableName(om.project, om.chain)).
			Where("id in (?)", expiredOrderIDs[i:end]).Update("order_status", multi.OrderStatusExpired).Error; err != nil {
			return errors.Wrap(err, "failed on update expired orders status")
		}
	}

	// 为每个过期订单触发地板价更新事件
	for _, order := range expiredOrders {
		if err := om.addUpdateFloorPriceEvent(&TradeEvent{
			EventType:      Expired,
			CollectionAddr: order.CollectionAddress,
			OrderId:        order.OrderID,
		}); err != nil {
			return errors.Wrap(err, "failed on add update floor price event")
		}
	}

	return nil
}

// addToOrderExpiryCheckQueue 函数用于将订单添加到过期检查队列中
// 主要功能:
// 1. 根据延迟时间计算订单在时间轮上的位置
// 2. 将订单添加到时间轮对应位置的链表中
// 参数说明:
// - delaySeconds: 延迟秒数,即订单过期前的剩余时间
// - chainSuffix: 链标识
// - orderId: 订单ID
// - collectionAddr: NFT集合地址
func (om *OrderManager) addToOrderExpiryCheckQueue(delaySeconds int64, chainSuffix string, orderId string, collectionAddr string) error {
	// 加锁保护并发安全
	om.Mux.Lock()
	defer om.Mux.Unlock()

	// 计算订单最终的时间轮位置
	// 从当前索引开始计数
	calculateValue := om.CurrentIndex + delaySeconds

	// 计算需要循环的圈数
	// 如果需要循环,则圈数减1(因为从0开始计数)
	cycle := calculateValue / WheelSize
	if cycle > 0 {
		cycle--
	}

	// 计算订单在时间轮上的实际位置
	index := calculateValue % WheelSize

	// 创建订单对象
	orderActivity := &Order{
		orderID:        orderId,
		CollectionAddr: collectionAddr,
		ChainSuffix:    chainSuffix,
		CycleCount:     cycle,
		WheelPosition:  index,
	}

	// 如果该位置为空,直接放入
	if om.TimeWheel[index].NotifyActivities == nil {
		om.TimeWheel[index].NotifyActivities = orderActivity
	} else {
		// 否则插入到链表头部
		// 因为不需要排序,所以直接插入到头部即可
		head := om.TimeWheel[index].NotifyActivities
		orderActivity.Next = head
		om.TimeWheel[index].NotifyActivities = orderActivity
	}
	return nil
}

// GetWheelTaskQuantity :check tasks number in the time wheel at the moment
func (om *OrderManager) GetWheelTaskQuantity(index int64) int64 {
	orders := om.TimeWheel[index].NotifyActivities
	if orders == nil {
		return 0
	}
	var orderNum int64
	for p := orders; p != nil; p = p.Next {
		orderNum++
	}
	return orderNum
}

// updateOrderState : handler function by sql transaction
// updateOrderState 是一个处理订单状态更新和底价更新的函数。
// 它将指定订单的状态更新为过期，并触发一个更新底价的事件。
//
// 参数:
//   - orderId: 要更新的订单的唯一标识符。
//   - collectionAddr: 订单所属的NFT集合的地址。
//
// 返回值:
//   - 如果操作成功，返回 nil。
//   - 如果在更新订单状态或触发底价更新事件时发生错误，返回一个包含错误信息的错误对象。
func (om *OrderManager) updateOrderState(orderId string, collectionAddr string) error {
	// 更新订单状态为过期
	// update orders status to expired
	if err := om.updateOrdersStatus(orderId, multi.OrderStatusExpired); err != nil {
		// 若更新订单状态失败，包装错误信息并返回
		return errors.Wrap(err, "failed on update activities status")
	}

	// 更新底价
	// update floor price
	if err := om.addUpdateFloorPriceEvent(&TradeEvent{
		EventType:      Expired,
		OrderId:        orderId,
		CollectionAddr: collectionAddr,
	}); err != nil {
		// 若添加更新底价事件失败，包装错误信息并返回
		return errors.Wrap(err, "failed on add update floor price event")
	}

	// 若上述操作均成功，返回 nil 表示操作成功
	return nil
}

// updateOrdersStatus 用于更新指定订单的状态。
// 该函数通过数据库操作，将指定订单ID的订单状态更新为传入的状态。
//
// 参数:
//   - orderID: 要更新状态的订单的唯一标识符。
//   - orderStatus: 要更新的订单状态，使用整数表示。
//
// 返回值:
//   - 如果更新操作成功，返回 nil。
//   - 如果更新操作失败，返回一个包含错误信息的错误对象。
func (om *OrderManager) updateOrdersStatus(orderID string, orderStatus int) error {
	// 使用数据库连接和上下文，指定要操作的表
	// 注意：原代码中 fmt.Sprintf("%s", ...) 可简化，这里已按原代码保留注释
	if err := om.DB.WithContext(om.Ctx).Table(fmt.Sprintf("%s", gdb.GetMultiProjectOrderTableName(om.project, om.chain))).
		// 筛选出指定订单ID的记录
		Where("order_id = ?", orderID).
		// 更新订单状态为传入的状态
		Update("order_status", orderStatus).Error; err != nil {
		// 若更新操作失败，包装错误信息并返回
		return errors.Wrap(err, "failed on update expired orders status")
	}

	// 若更新操作成功，返回 nil 表示操作成功
	return nil
}
