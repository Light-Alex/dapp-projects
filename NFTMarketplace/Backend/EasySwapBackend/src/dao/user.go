package dao

import (
	"context"

	"github.com/ProjectsTask/EasySwapBase/stores/gdb/orderbookmodel/base"
	"github.com/ProjectsTask/EasySwapBase/stores/gdb/orderbookmodel/multi"
	"github.com/pkg/errors"
)

// GetUserSigStatus 根据用户地址获取用户的签名状态。
// 参数 ctx 是上下文，用于控制请求的生命周期和传递请求范围的数据。
// 参数 userAddr 是用户的地址，用于唯一标识用户。
// 返回值为一个布尔值，表示用户的签名状态，以及一个错误对象。
func (d *Dao) GetUserSigStatus(ctx context.Context, userAddr string) (bool, error) {
	// 声明一个 base.User 结构体变量，用于存储从数据库中查询到的用户信息
	var userInfo base.User
	// 构建一个数据库查询，使用传入的上下文，从 base.UserTableName() 表中
	// 查询地址等于 userAddr 的用户信息，并将结果存储到 userInfo 变量中
	db := d.DB.WithContext(ctx).Table(base.UserTableName()).
		Where("address = ?", userAddr).
		Find(&userInfo)
	// 检查数据库查询是否出错
	if db.Error != nil {
		// 若出错，返回 false 和一个包装后的错误信息，提示获取用户信息失败
		return false, errors.Wrap(db.Error, "failed on get user info")
	}

	// 若没有错误，返回用户的签名状态和 nil 错误
	return userInfo.IsSigned, nil
}

// QueryUserBids 查询用户的出价订单信息
func (d *Dao) QueryUserBids(ctx context.Context, chain string, userAddrs []string, contractAddrs []string) ([]multi.Order, error) {
	var userBids []multi.Order

	// SQL解释:
	// 1. 从订单表中查询订单详细信息
	// 2. 选择字段包括:集合地址、代币ID、订单ID、订单类型、剩余数量等
	// 3. WHERE条件:
	//    - maker在给定用户地址列表中
	//    - 订单类型为Item出价或集合出价
	//    - 订单状态为活跃
	//    - 剩余数量大于0
	db := d.DB.WithContext(ctx).
		Table(multi.OrderTableName(chain)).
		Select("collection_address, token_id, order_id, token_id,order_type,"+
			"quantity_remaining, size, event_time, price, salt, expire_time").
		Where("maker in (?) and order_type in (?,?) and order_status = ? and quantity_remaining > 0",
			userAddrs, multi.ItemBidOrder, multi.CollectionBidOrder, multi.OrderStatusActive)

	// 如果指定了合约地址列表,添加集合地址过滤条件
	if len(contractAddrs) != 0 {
		db.Where("collection_address in (?)", contractAddrs)
	}

	if err := db.Scan(&userBids).Error; err != nil {
		return nil, errors.Wrap(err, "failed on get user bids")
	}

	return userBids, nil
}
