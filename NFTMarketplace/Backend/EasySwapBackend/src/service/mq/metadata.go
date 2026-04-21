package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ProjectsTask/EasySwapBase/logger/xzap"
	"github.com/ProjectsTask/EasySwapBase/stores/xkv"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ProjectsTask/EasySwapBackend/src/types/v1"
)

const CacheRefreshSingleItemMetadataKey = "cache:%s:%s:item:refresh:metadata"

// GetRefreshSingleItemMetadataKey 根据传入的项目名和链名生成用于刷新单个物品元数据的缓存键。
// 此函数会将项目名和链名转换为小写，然后按照 CacheRefreshSingleItemMetadataKey 定义的格式生成键。
//
// 参数:
//   - project: 项目名称，函数会将其转换为小写后使用。
//   - chain: 链的名称，函数会将其转换为小写后使用。
//
// 返回值:
//   - 返回一个格式化后的字符串，作为用于刷新单个物品元数据的缓存键。
func GetRefreshSingleItemMetadataKey(project, chain string) string {
	// 将项目名和链名转换为小写，并按照 CacheRefreshSingleItemMetadataKey 格式生成键
	return fmt.Sprintf(CacheRefreshSingleItemMetadataKey, strings.ToLower(project), strings.ToLower(chain))
}

const CacheRefreshPreventReentrancyKeyPrefix = "cache:es:item:refresh:prevent:reentrancy:%d:%s:%s"
const PreventReentrancyPeriod = 10 //second

// AddSingleItemToRefreshMetadataQueue 将单个物品添加到刷新元数据队列中。
// 此函数会检查是否在防重入周期内已经刷新过该物品的元数据，若未刷新，则将物品信息添加到队列，并设置防重入标记。
//
// 参数:
//   - kvStore: 用于存储和检索数据的键值存储实例。
//   - project: 项目名称。
//   - chainName: 链的名称。
//   - chainID: 链的ID。
//   - collectionAddr: 物品所属集合的地址。
//   - tokenID: 物品的令牌ID。
//
// 返回值:
//   - 如果操作成功，返回 nil。
//   - 如果在检查重入状态、序列化物品信息或添加到队列时发生错误，返回一个包含错误信息的错误对象。
func AddSingleItemToRefreshMetadataQueue(kvStore *xkv.Store, project, chainName string, chainID int64, collectionAddr, tokenID string) error {
	// 检查是否在防重入周期内已经刷新过该物品的元数据
	isRefreshed, err := kvStore.Get(fmt.Sprintf(CacheRefreshPreventReentrancyKeyPrefix, chainID, collectionAddr, tokenID))
	if err != nil {
		// 若获取重入状态失败，包装错误信息并返回
		return errors.Wrap(err, "failed on check reentrancy status")
	}

	// 如果在防重入周期内已经刷新过，记录日志并返回 nil
	if isRefreshed != "" {
		xzap.WithContext(context.Background()).Info("refresh within 10s", zap.String("collection_addr", collectionAddr), zap.String("token_id", tokenID))
		return nil
	}

	// 创建要刷新的物品信息
	item := types.RefreshItem{
		ChainID:        chainID,
		CollectionAddr: collectionAddr,
		TokenID:        tokenID,
	}

	// 将物品信息序列化为JSON格式
	rawInfo, err := json.Marshal(&item)
	if err != nil {
		// 若序列化失败，包装错误信息并返回
		return errors.Wrap(err, "failed on marshal item info")
	}

	// 将序列化后的物品信息添加到刷新元数据队列中
	_, err = kvStore.Sadd(GetRefreshSingleItemMetadataKey(project, chainName), string(rawInfo))
	if err != nil {
		// 若添加到队列失败，包装错误信息并返回
		return errors.Wrap(err, "failed on push item to refresh metadata queue")
	}

	// 设置防重入标记，防止在指定周期内重复刷新
	_ = kvStore.Setex(fmt.Sprintf(CacheRefreshPreventReentrancyKeyPrefix, chainID, collectionAddr, tokenID), "true", PreventReentrancyPeriod)

	// 若上述操作均成功，返回 nil 表示操作成功
	return nil
}
