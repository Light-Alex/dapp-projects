package dao

import (
	"context"
	"fmt"
	"strings"

	"github.com/ProjectsTask/EasySwapBase/stores/gdb/orderbookmodel/multi"
	"github.com/pkg/errors"
)

// QueryCollectionItemsImage 查询集合内NFT Item的图片和视频信息
// 参数:
// - ctx context.Context: 上下文，用于控制请求的生命周期。
// - chain string: 链名称，指定要查询的链。
// - collectionAddr string: 集合地址，指定要查询的NFT集合。
// - tokenIds []string: 一组NFT的token ID，指定要查询的具体NFT。
// 返回:
// - []multi.ItemExternal: 查询到的NFT Item的图片和视频信息列表。
// - error: 若查询过程中出现错误，返回错误信息；否则返回 nil。
func (d *Dao) QueryCollectionItemsImage(ctx context.Context, chain string,
	collectionAddr string, tokenIds []string) ([]multi.ItemExternal, error) {
	// 定义一个切片，用于存储查询到的NFT Item的图片和视频信息
	var itemsExternal []multi.ItemExternal

	// 构建并执行SQL查询
	// 使用DB对象的WithContext方法设置上下文，确保查询操作受上下文控制
	// 使用Table方法指定要查询的表名，通过multi.ItemExternalTableName函数根据链名称生成表名
	// 使用Select方法指定要查询的字段，包括集合地址、token ID、图片和视频的相关信息
	// 使用Where方法添加查询条件，要求集合地址等于指定的集合地址，并且token ID在指定的token ID列表中
	// 使用Scan方法将查询结果扫描到itemsExternal切片中
	if err := d.DB.WithContext(ctx).
		Table(multi.ItemExternalTableName(chain)).
		Select("collection_address, token_id, is_uploaded_oss, "+
			"image_uri, oss_uri, video_type, is_video_uploaded, "+
			"video_uri, video_oss_uri").
		Where("collection_address = ? and token_id in (?)",
			collectionAddr, tokenIds).
		Scan(&itemsExternal).Error; err != nil {
		// 如果查询过程中出现错误，使用errors.Wrap方法包装错误信息，并返回nil和错误信息
		return nil, errors.Wrap(err, "failed on query items external info")
	}

	// 若查询成功，返回查询到的NFT Item的图片和视频信息列表和nil
	return itemsExternal, nil
}

// QueryMultiChainCollectionsItemsImage 查询多条链上NFT Item的图片信息
// 主要功能:
// 1. 按链名称对输入的Item信息进行分组
// 2. 构建多条链的联合查询SQL
// 3. 返回所有链上Item的图片信息
func (d *Dao) QueryMultiChainCollectionsItemsImage(ctx context.Context, itemInfos []MultiChainItemInfo) ([]multi.ItemExternal, error) {
	var itemsExternal []multi.ItemExternal

	// SQL语句组成部分
	sqlHead := "SELECT * FROM (" // 外层查询开始
	sqlTail := ") as combined"   // 外层查询结束
	var sqlMids []string         // 存储每条链的子查询

	// 按链名称对Item信息分组
	chainItems := make(map[string][]MultiChainItemInfo)
	for _, itemInfo := range itemInfos {
		items, ok := chainItems[strings.ToLower(itemInfo.ChainName)]
		if ok {
			items = append(items, itemInfo)
			chainItems[strings.ToLower(itemInfo.ChainName)] = items
		} else {
			chainItems[strings.ToLower(itemInfo.ChainName)] = []MultiChainItemInfo{itemInfo}
		}
	}

	// 遍历每条链构建子查询
	for chainName, items := range chainItems {
		// 构建IN查询条件: (('addr1','id1'),('addr2','id2'),...)
		tmpStat := fmt.Sprintf("(('%s','%s')", items[0].CollectionAddress, items[0].TokenID)
		for i := 1; i < len(items); i++ {
			tmpStat += fmt.Sprintf(",('%s','%s')", items[i].CollectionAddress, items[i].TokenID)
		}
		tmpStat += ") "

		// 构建子查询SQL:
		// 1. 选择Item的图片相关字段
		// 2. 从对应链的external表查询
		// 3. 匹配集合地址和tokenID
		sqlMid := "("
		sqlMid += "select collection_address, token_id, is_uploaded_oss, image_uri, oss_uri "
		sqlMid += fmt.Sprintf("from %s ", multi.ItemExternalTableName(chainName))
		sqlMid += "where (collection_address,token_id) in "
		sqlMid += tmpStat
		sqlMid += ")"

		sqlMids = append(sqlMids, sqlMid)
	}

	// 使用UNION ALL组合所有子查询
	sql := sqlHead
	for i := 0; i < len(sqlMids); i++ {
		if i != 0 {
			sql += " UNION ALL " // 使用UNION ALL合并结果集
		}
		sql += sqlMids[i]
	}
	sql += sqlTail

	// 执行SQL查询
	if err := d.DB.WithContext(ctx).Raw(sql).Scan(&itemsExternal).Error; err != nil {
		return nil, errors.Wrap(err, "failed on query multi chain items external info")
	}

	return itemsExternal, nil
}
