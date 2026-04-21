package dao

import (
	"context"

	"github.com/ProjectsTask/EasySwapBase/stores/gdb/orderbookmodel/multi"
	"github.com/pkg/errors"

	"github.com/ProjectsTask/EasySwapBackend/src/types/v1"
)

// QueryItemTraits 查询单个NFT Item的 Trait信息
// chain: 区块链类型
// collectionAddr: 集合地址
// tokenID: 代币ID
// 返回值: []multi.ItemTrait 物品特性信息列表，error 错误信息
func (d *Dao) QueryItemTraits(ctx context.Context, chain string, collectionAddr string, tokenID string) ([]multi.ItemTrait, error) {
	var itemTraits []multi.ItemTrait
	// 从数据库中查询物品特性信息
	if err := d.DB.WithContext(ctx).Table(multi.ItemTraitTableName(chain)).
		// 选择需要查询的字段
		Select("collection_address, token_id, trait, trait_value").
		// 设置查询条件
		Where("collection_address = ? and token_id = ?", collectionAddr, tokenID).
		// 执行查询并将结果存储在itemTraits中
		Scan(&itemTraits).Error; err != nil {
		return nil, errors.Wrap(err, "failed on query items trait info")
	}

	return itemTraits, nil
}

// QueryItemsTraits 查询多个NFT Item的 Trait信息
func (d *Dao) QueryItemsTraits(ctx context.Context, chain string, collectionAddr string, tokenIds []string) ([]multi.ItemTrait, error) {
	var itemsTraits []multi.ItemTrait

	// 从数据库中查询满足条件的记录
	if err := d.DB.WithContext(ctx).Table(multi.ItemTraitTableName(chain)).
		// 选择要查询的字段
		Select("collection_address, token_id, trait, trait_value").
		// 设置查询条件
		Where("collection_address = ? and token_id in (?)", collectionAddr, tokenIds).
		// 执行查询并将结果扫描到itemsTraits变量中
		Scan(&itemsTraits).Error; err != nil {
		// 如果查询过程中发生错误，则返回错误
		return nil, errors.Wrap(err, "failed on query items trait info")
	}

	// 查询成功，返回查询结果
	return itemsTraits, nil
}

// QueryCollectionTraits 查询NFT合集的 Trait信息统计
func (d *Dao) QueryCollectionTraits(ctx context.Context, chain string, collectionAddr string) ([]types.TraitCount, error) {
	var traitCounts []types.TraitCount
	// 查询集合的特征数量
	if err := d.DB.WithContext(ctx).Table(multi.ItemTraitTableName(chain)).
		// 选择字段：特征、特征值、以及特征值的计数
		Select("`trait`,`trait_value`,count(*) as count").
		// 根据集合地址过滤
		Where("collection_address=?", collectionAddr).
		// 按特征和特征值分组
		Group("`trait`,`trait_value`").
		// 将结果扫描到 traitCounts 切片中
		Scan(&traitCounts).Error; err != nil {
		// 如果查询出错，返回错误
		return nil, errors.Wrap(err, "failed on query collection trait amount")
	}

	// 返回特征数量切片和 nil 错误
	return traitCounts, nil
}
