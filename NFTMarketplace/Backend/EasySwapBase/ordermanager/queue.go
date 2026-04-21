package ordermanager

import (
	"sort"
	"strings"

	"github.com/shopspring/decimal"
)

type Entry struct {
	orderID  string          //order ID
	priority decimal.Decimal // price
	maker    string
	tokenID  string
}

type PriorityQueue []*Entry

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// 比较优先级
	return pq[i].priority.LessThan(pq[j].priority)
}

// Swap 方法用于交换索引i和索引j处的元素。
// 参数i和j分别代表需要交换的两个元素的索引位置。
func (pq PriorityQueue) Swap(i, j int) {
	// 交换索引i和索引j处的元素
	pq[i], pq[j] = pq[j], pq[i]
}

type PriorityQueueMap struct {
	pq     PriorityQueue
	orders map[string]*Entry
	maxLen int
}

// NewPriorityQueueMap 创建并返回一个新的 PriorityQueueMap 实例。
// 参数 maxLen 表示优先级队列的最大长度。
func NewPriorityQueueMap(maxLen int) *PriorityQueueMap {
	// 创建一个新的 PriorityQueueMap 实例
	return &PriorityQueueMap{
		// 初始化 orders 映射，用于存储优先级队列中的条目
		orders: make(map[string]*Entry),
		// 设置最大长度
		maxLen: maxLen,
	}
}

func (pqm *PriorityQueueMap) Len() int {
	// 返回 pqm.orders 的长度
	return len(pqm.orders)
}

// Add 向优先级队列中添加订单
//
// 参数:
//     orderID: 订单ID
//     price: 订单价格
//     maker: 订单制造者
//     tokenID: 订单代币ID
func (pqm *PriorityQueueMap) Add(orderID string, price decimal.Decimal, maker, tokenID string) {
	// 如果优先级队列的长度超过了最大长度
	if len(pqm.pq) > pqm.maxLen {
		// 从orders映射中删除最后一个元素的orderID
		delete(pqm.orders, pqm.pq[len(pqm.pq)-1].orderID)
		// 截断优先级队列，移除最后一个元素
		pqm.pq = pqm.pq[0 : len(pqm.pq)-1]
	}

	// 创建Entry实例，并设置其属性
	entry := &Entry{
		orderID: orderID,
		priority: price,
		maker:    strings.ToLower(maker),
		tokenID:  strings.ToLower(tokenID),
	}
	// 将Entry实例添加到优先级队列中
	pqm.pq = append(pqm.pq, entry)
	// 对优先级队列进行排序
	sort.Sort(pqm.pq)
	// 将Entry实例添加到orders映射中，键为orderID
	pqm.orders[orderID] = entry
}

// GetMin 方法从 PriorityQueueMap 结构体中获取并返回具有最低优先级的订单ID和对应的优先级。
//
// 参数:
//     无
//
// 返回值:
//     string: 具有最低优先级的订单ID
//     decimal.Decimal: 具有最低优先级的订单的优先级
//
// 说明:
//     如果优先队列为空，则返回空字符串("")和值为0的Decimal对象。
func (pqm *PriorityQueueMap) GetMin() (string, decimal.Decimal) {
	// 如果优先队列为空
	if len(pqm.pq) == 0 {
		// 返回空字符串和0的Decimal值
		return "", decimal.Zero
	}
	// 获取优先队列的第一个元素
	entry := pqm.pq[0]
	// 返回该元素的订单ID和优先级
	return entry.orderID, entry.priority
}

func (pqm *PriorityQueueMap) GetMax() (string, decimal.Decimal) {
	if len(pqm.pq) == 0 {
		return "", decimal.Zero
	}
	entry := pqm.pq[len(pqm.pq)-1]
	return entry.orderID, entry.priority
}

func (pqm *PriorityQueueMap) Remove(orderID string) {
	_, ok := pqm.orders[orderID]
	if !ok {
		return
	}

	var newPQ []*Entry
	for i, v := range pqm.pq {
		if v.orderID != orderID {
			newPQ = append(newPQ, pqm.pq[i])
		}
	}

	pqm.pq = newPQ
	delete(pqm.orders, orderID)
}

func (pqm *PriorityQueueMap) RemoveMakerOrders(maker, tokenID string) {
	maker = strings.ToLower(maker)
	tokenID = strings.ToLower(tokenID)
	var newPQ []*Entry

	for i, v := range pqm.pq {
		if v.maker == maker && v.tokenID == tokenID {
			delete(pqm.orders, v.orderID)
		} else {
			newPQ = append(newPQ, pqm.pq[i])
		}
	}

	pqm.pq = newPQ
}
