package service

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ProjectsTask/EasySwapBackend/src/service/svc"
	"github.com/ProjectsTask/EasySwapBackend/src/types/v1"
)

// GetItemBidsInfo 获取指定集合中单个NFT项目的出价信息列表
// 参数:
// - ctx context.Context: 上下文，用于控制请求的生命周期。
// - svcCtx *svc.ServerCtx: 服务上下文，包含数据库访问和其他服务的实例。
// - chain string: 链名称，指定要查询的链。
// - collectionAddr string: 集合地址，指定要查询的NFT集合。
// - tokenID string: NFT的token ID，指定要查询的单个NFT项目。
// - page int: 页码，指定查询的分页页码。
// - pageSize int: 每页数量，指定每页返回的出价信息数量。
// 返回:
// - *types.CollectionBidsResp: 查询到的集合出价信息响应结构体指针，包含出价信息列表和总数。
// - error: 若查询过程中出现错误，返回错误信息；否则返回 nil。
func GetItemBidsInfo(ctx context.Context, svcCtx *svc.ServerCtx, chain string, collectionAddr, tokenID string, page, pageSize int) (*types.CollectionBidsResp, error) {
	// 调用 Dao 层的 QueryItemBids 方法查询指定集合中单个NFT项目的出价信息列表
	// 传入上下文、链名称、集合地址、token ID、页码和每页数量作为参数
	bids, count, err := svcCtx.Dao.QueryItemBids(ctx, chain, collectionAddr, tokenID, page, pageSize)
	// 检查查询过程中是否出现错误
	if err != nil {
		// 若出现错误，返回 nil 和包装后的错误信息，提示获取项目信息失败
		return nil, errors.Wrap(err, "failed on get item info")
	}

	// 遍历查询到的出价信息列表
	for i := 0; i < len(bids); i++ {
		// 调用 getBidType 函数转换出价信息的订单类型
		bids[i].OrderType = getBidType(bids[i].OrderType)
	}
	// 若查询成功，创建并返回包含出价信息列表和总数的响应结构体指针
	return &types.CollectionBidsResp{
		Result: bids,  // 出价信息列表
		Count:  count, // 出价信息总数
	}, nil
}
