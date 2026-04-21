package service

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ProjectsTask/EasySwapBackend/src/service/svc"
	"github.com/ProjectsTask/EasySwapBackend/src/types/v1"
)

// GetMultiChainActivities 函数用于查询多链活动
//
// 参数：
//     ctx: 上下文对象
//     svcCtx: 服务上下文对象
//     chainID: 链ID数组
//     chainName: 链名称数组
//     collectionAddrs: 集合地址数组
//     tokenID: 代币ID
//     userAddrs: 用户地址数组
//     eventTypes: 事件类型数组
//     page: 页码
//     pageSize: 每页条数
//
// 返回值：
//     *types.ActivityResp: 活动响应对象，包含活动结果和总数
//     error: 错误信息，如果发生错误则返回
func GetMultiChainActivities(ctx context.Context, svcCtx *svc.ServerCtx, chainID []int, chainName []string, collectionAddrs []string, tokenID string, userAddrs []string, eventTypes []string, page, pageSize int) (*types.ActivityResp, error) {
	// 查询多链活动
	activities, total, err := svcCtx.Dao.QueryMultiChainActivities(ctx, chainName, collectionAddrs, tokenID, userAddrs, eventTypes, page, pageSize)
	if err != nil {
		// 如果查询失败，返回错误信息
		return nil, errors.Wrap(err, "failed on query multi-chain activity")
	}

	// 如果没有活动或总数为0，返回空结果
	if total == 0 || len(activities) == 0 {
		return &types.ActivityResp{
			Result: nil,
			Count:  0,
		}, nil
	}

	// 查询外部信息
	//external info query
	results, err := svcCtx.Dao.QueryMultiChainActivityExternalInfo(ctx, chainID, chainName, activities)
	if err != nil {
		// 如果查询外部信息失败，返回错误信息
		return nil, errors.Wrap(err, "failed on query activity external info")
	}

	// 返回活动响应
	return &types.ActivityResp{
		Result: results,
		Count:  total,
	}, nil
}
