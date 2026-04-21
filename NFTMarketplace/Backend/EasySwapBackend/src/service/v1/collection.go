package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/ProjectsTask/EasySwapBase/errcode"
	"github.com/ProjectsTask/EasySwapBase/evm/eip"
	"github.com/ProjectsTask/EasySwapBase/logger/xzap"
	"github.com/ProjectsTask/EasySwapBase/ordermanager"
	"github.com/ProjectsTask/EasySwapBase/stores/gdb/orderbookmodel/multi"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/ProjectsTask/EasySwapBackend/src/dao"
	"github.com/ProjectsTask/EasySwapBackend/src/service/mq"
	"github.com/ProjectsTask/EasySwapBackend/src/service/svc"
	"github.com/ProjectsTask/EasySwapBackend/src/types/v1"
)

// GetBids 获取指定集合的出价信息列表
// 参数:
// - ctx context.Context: 上下文，用于控制请求的生命周期。
// - svcCtx *svc.ServerCtx: 服务上下文，包含数据库访问和其他服务的实例。
// - chain string: 链名称，指定要查询的链。
// - collectionAddr string: 集合地址，指定要查询的NFT集合。
// - page int: 页码，指定查询的分页页码。
// - pageSize int: 每页数量，指定每页返回的出价信息数量。
// 返回:
// - *types.CollectionBidsResp: 查询到的集合出价信息响应结构体指针，包含出价信息列表和总数。
// - error: 若查询过程中出现错误，返回错误信息；否则返回 nil。
func GetBids(ctx context.Context, svcCtx *svc.ServerCtx, chain string, collectionAddr string, page, pageSize int) (*types.CollectionBidsResp, error) {
	// 调用 Dao 层的 QueryCollectionBids 方法查询指定集合的出价信息列表
	// 传入上下文、链名称、集合地址、页码和每页数量作为参数
	bids, count, err := svcCtx.Dao.QueryCollectionBids(ctx, chain, collectionAddr, page, pageSize)
	// 检查查询过程中是否出现错误
	if err != nil {
		// 若出现错误，返回 nil 和包装后的错误信息，提示获取出价信息失败
		return nil, errors.Wrap(err, "failed on get item info")
	}

	// 若查询成功，创建并返回包含出价信息列表和总数的响应结构体指针
	return &types.CollectionBidsResp{
		Result: bids,  // 出价信息列表
		Count:  count, // 出价信息总数
	}, nil
}

// GetItems 获取NFT Item列表信息：Item基本信息、订单信息、图片信息、用户持有数量、最近成交价格、最高出价信息
// 参数:
// - ctx context.Context: 上下文，用于控制请求的生命周期。
// - svcCtx *svc.ServerCtx: 服务上下文，包含数据库访问和其他服务的实例。
// - chain string: 链名称，指定要查询的链。
// - filter types.CollectionItemFilterParams: 过滤参数，用于筛选查询结果。
// - collectionAddr string: 集合地址，指定要查询的NFT集合。
// 返回:
// - *types.NFTListingInfoResp: 查询到的NFT列表信息响应结构体指针，包含NFT列表信息和总数。
// - error: 若查询过程中出现错误，返回错误信息；否则返回 nil。
func GetItems(ctx context.Context, svcCtx *svc.ServerCtx, chain string, filter types.CollectionItemFilterParams, collectionAddr string) (*types.NFTListingInfoResp, error) {
	// 1. 查询基础Item信息和订单信息
	// 调用Dao层的QueryCollectionItemOrder方法查询指定集合中符合过滤条件的Item信息和订单信息
	// 传入上下文、链名称、过滤参数和集合地址作为参数
	items, count, err := svcCtx.Dao.QueryCollectionItemOrder(ctx, chain, filter, collectionAddr)
	// 检查查询过程中是否出现错误
	if err != nil {
		// 若出现错误，返回 nil 和包装后的错误信息，提示获取Item信息失败
		return nil, errors.Wrap(err, "failed on get item info")
	}

	// 2. 提取需要查询的ItemID和所有者地址
	var ItemIds []string
	var ItemOwners []string
	var itemPrice []types.ItemPriceInfo
	// 遍历查询到的Item信息
	for _, item := range items {
		// 如果Item的TokenID不为空，将其添加到ItemIds列表中
		if item.TokenId != "" {
			ItemIds = append(ItemIds, item.TokenId)
		}
		// 如果Item的所有者地址不为空，将其添加到ItemOwners列表中
		if item.Owner != "" {
			ItemOwners = append(ItemOwners, item.Owner)
		}
		// 记录已上架Item的价格信息
		if item.Listing {
			itemPrice = append(itemPrice, types.ItemPriceInfo{
				CollectionAddress: item.CollectionAddress,
				TokenID:           item.TokenId,
				Maker:             item.Owner,
				Price:             item.ListPrice,
				OrderStatus:       multi.OrderStatusActive,
			})
		}
	}

	// 3. 并发查询各类扩展信息
	var queryErr error
	var wg sync.WaitGroup

	// 3.1 查询订单详情
	// 用于存储订单详情信息，键为CollectionAddress和TokenId拼接的字符串
	ordersInfo := make(map[string]multi.Order)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 如果有已上架Item的价格信息
		if len(itemPrice) > 0 {
			// 调用Dao层的QueryListingInfo方法查询订单详情信息
			orders, err := svcCtx.Dao.QueryListingInfo(ctx, chain, itemPrice)
			if err != nil {
				// 若查询过程中出现错误，记录错误信息
				queryErr = errors.Wrap(err, "failed on get orders time info")
				return
			}
			// 将查询到的订单信息存储到ordersInfo映射中
			for _, order := range orders {
				ordersInfo[strings.ToLower(order.CollectionAddress+order.TokenId)] = order
			}
		}
	}()

	// 3.2 查询Item图片信息
	// 用于存储Item的图片和视频信息，键为TokenId
	ItemsExternal := make(map[string]multi.ItemExternal)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 如果有需要查询的ItemID
		if len(ItemIds) != 0 {
			// 调用Dao层的QueryCollectionItemsImage方法查询Item的图片和视频信息
			items, err := svcCtx.Dao.QueryCollectionItemsImage(ctx, chain, collectionAddr, ItemIds)
			if err != nil {
				// 若查询过程中出现错误，记录错误信息
				queryErr = errors.Wrap(err, "failed on get items image info")
				return
			}
			// 将查询到的图片和视频信息存储到ItemsExternal映射中
			for _, item := range items {
				ItemsExternal[strings.ToLower(item.TokenId)] = item
			}
		}
	}()

	// 3.3 查询用户持有数量
	// 用于存储用户持有Item的数量，键为用户地址
	userItemCount := make(map[string]int64)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 如果有需要查询的ItemID
		if len(ItemIds) != 0 {
			// 调用Dao层的QueryUsersItemCount方法查询用户持有Item的数量
			itemCount, err := svcCtx.Dao.QueryUsersItemCount(ctx, chain, collectionAddr, ItemOwners)
			if err != nil {
				// 若查询过程中出现错误，记录错误信息
				queryErr = errors.Wrap(err, "failed on get items image info")
				return
			}
			// 将查询到的用户持有数量存储到userItemCount映射中
			for _, v := range itemCount {
				userItemCount[strings.ToLower(v.Owner)] = v.Counts
			}
		}
	}()

	// 3.4 查询最近成交价格
	// 用于存储Item的最近成交价格，键为TokenId
	lastSales := make(map[string]decimal.Decimal)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 如果有需要查询的ItemID
		if len(ItemIds) != 0 {
			// 调用Dao层的QueryLastSalePrice方法查询Item的最近成交价格
			lastSale, err := svcCtx.Dao.QueryLastSalePrice(ctx, chain, collectionAddr, ItemIds)
			if err != nil {
				// 若查询过程中出现错误，记录错误信息
				queryErr = errors.Wrap(err, "failed on get items last sale info")
				return
			}
			// 将查询到的最近成交价格存储到lastSales映射中
			for _, v := range lastSale {
				lastSales[strings.ToLower(v.TokenId)] = v.Price
			}
		}
	}()

	// 3.5 查询Item级别最高出价
	// 用于存储Item级别的最高出价信息，键为TokenId
	bestBids := make(map[string]multi.Order)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 如果有需要查询的ItemID
		if len(ItemIds) != 0 {
			// 调用Dao层的QueryBestBids方法查询Item级别的最高出价信息
			bids, err := svcCtx.Dao.QueryBestBids(ctx, chain, filter.UserAddress, collectionAddr, ItemIds)
			if err != nil {
				// 若查询过程中出现错误，记录错误信息
				queryErr = errors.Wrap(err, "failed on get items last sale info")
				return
			}
			// 遍历查询到的出价信息，找出每个Item的最高出价
			for _, bid := range bids {
				order, ok := bestBids[strings.ToLower(bid.TokenId)]
				if !ok {
					bestBids[strings.ToLower(bid.TokenId)] = bid
					continue
				}
				if bid.Price.GreaterThan(order.Price) {
					bestBids[strings.ToLower(bid.TokenId)] = bid
				}
			}
		}
	}()

	// 3.6 查询集合级别最高出价
	var collectionBestBid multi.Order
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 调用Dao层的QueryCollectionBestBid方法查询集合级别的最高出价信息
		collectionBestBid, err = svcCtx.Dao.QueryCollectionBestBid(ctx, chain, filter.UserAddress, collectionAddr)
		if err != nil {
			// 若查询过程中出现错误，记录错误信息
			queryErr = errors.Wrap(err, "failed on get items last sale info")
			return
		}
	}()

	// 4. 等待所有查询完成
	wg.Wait()
	if queryErr != nil {
		// 若查询过程中出现错误，返回 nil 和包装后的错误信息，提示获取Items信息失败
		return nil, errors.Wrap(queryErr, "failed on get items info")
	}

	// 5. 整合所有信息
	var respItems []*types.NFTListingInfo
	// 遍历查询到的Item信息
	for _, item := range items {
		// 设置Item名称
		nameStr := item.Name
		if nameStr == "" {
			nameStr = fmt.Sprintf("#%s", item.TokenId)
		}

		// 构建返回结构
		respItem := &types.NFTListingInfo{
			Name:              nameStr,
			CollectionAddress: item.CollectionAddress,
			TokenID:           item.TokenId,
			OwnerAddress:      item.Owner,
			ListPrice:         item.ListPrice,
			MarketID:          item.MarketID,
			// 初始设置为集合级别的最高出价信息
			BidOrderID:    collectionBestBid.OrderID,
			BidExpireTime: collectionBestBid.ExpireTime,
			BidPrice:      collectionBestBid.Price,
			BidTime:       collectionBestBid.EventTime,
			BidSalt:       collectionBestBid.Salt,
			BidMaker:      collectionBestBid.Maker,
			BidType:       getBidType(collectionBestBid.OrderType),
			BidSize:       collectionBestBid.Size,
			BidUnfilled:   collectionBestBid.QuantityRemaining,
		}

		// 添加订单信息
		listOrder, ok := ordersInfo[strings.ToLower(item.CollectionAddress+item.TokenId)]
		if ok {
			respItem.ListTime = listOrder.EventTime
			respItem.ListOrderID = listOrder.OrderID
			respItem.ListExpireTime = listOrder.ExpireTime
			respItem.ListSalt = listOrder.Salt
		}

		// 添加最高出价信息
		bidOrder, ok := bestBids[strings.ToLower(item.TokenId)]
		if ok {
			if bidOrder.Price.GreaterThan(collectionBestBid.Price) {
				// 如果Item级别的最高出价大于集合级别的最高出价，则使用Item级别的出价信息
				respItem.BidOrderID = bidOrder.OrderID
				respItem.BidExpireTime = bidOrder.ExpireTime
				respItem.BidPrice = bidOrder.Price
				respItem.BidTime = bidOrder.EventTime
				respItem.BidSalt = bidOrder.Salt
				respItem.BidMaker = bidOrder.Maker
				respItem.BidType = getBidType(bidOrder.OrderType)
				respItem.BidSize = bidOrder.Size
				respItem.BidUnfilled = bidOrder.QuantityRemaining
			}
		}

		// 添加图片和视频信息
		itemExternal, ok := ItemsExternal[strings.ToLower(item.TokenId)]
		if ok {
			if itemExternal.IsUploadedOss {
				respItem.ImageURI = itemExternal.OssUri
			} else {
				respItem.ImageURI = itemExternal.ImageUri
			}
			if len(itemExternal.VideoUri) > 0 {
				respItem.VideoType = itemExternal.VideoType
				if itemExternal.IsVideoUploaded {
					respItem.VideoURI = itemExternal.VideoOssUri
				} else {
					respItem.VideoURI = itemExternal.VideoUri
				}
			}
		}

		// 添加用户持有数量
		count, ok := userItemCount[strings.ToLower(item.Owner)]
		if ok {
			respItem.OwnerOwnedAmount = count
		}

		// 添加最近成交价格
		price, ok := lastSales[strings.ToLower(item.TokenId)]
		if ok {
			respItem.LastSellPrice = price
		}

		respItems = append(respItems, respItem)
	}

	return &types.NFTListingInfoResp{
		Result: respItems,
		Count:  count,
	}, nil
}

// GetItem 获取单个NFT的详细信息
// 参数:
// - ctx context.Context: 上下文，用于控制请求的生命周期。
// - svcCtx *svc.ServerCtx: 服务上下文，包含数据库访问和其他服务的实例。
// - chain string: 链名称，指定要查询的链。
// - chainID int: 链ID，用于标识链。
// - collectionAddr string: 集合地址，指定要查询的NFT集合。
// - tokenID string: NFT的token ID，指定要查询的单个NFT。
// 返回:
// - *types.ItemDetailInfoResp: 查询到的单个NFT详细信息响应结构体指针，包含详细信息。
// - error: 若查询过程中出现错误，返回错误信息；否则返回 nil。
func GetItem(ctx context.Context, svcCtx *svc.ServerCtx, chain string, chainID int, collectionAddr, tokenID string) (*types.ItemDetailInfoResp, error) {
	var queryErr error
	var wg sync.WaitGroup

	// 并发查询以下信息:
	// 1. 查询collection信息
	var collection *multi.Collection
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 调用Dao层的QueryCollectionInfo方法查询指定集合的信息
		collection, queryErr = svcCtx.Dao.QueryCollectionInfo(ctx, chain, collectionAddr)
		if queryErr != nil {
			return
		}
	}()

	// 2. 查询item基本信息
	var item *multi.Item
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 调用Dao层的QueryItemInfo方法查询指定NFT的基本信息
		item, queryErr = svcCtx.Dao.QueryItemInfo(ctx, chain, collectionAddr, tokenID)
		if queryErr != nil {
			return
		}
	}()

	// 3. 查询item挂单信息
	var itemListInfo *dao.CollectionItem
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 调用Dao层的QueryItemListInfo方法查询指定NFT的挂单信息
		itemListInfo, queryErr = svcCtx.Dao.QueryItemListInfo(ctx, chain, collectionAddr, tokenID)
		if queryErr != nil {
			return
		}
	}()

	// 4. 查询item图片和视频信息
	ItemExternals := make(map[string]multi.ItemExternal)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 调用Dao层的QueryCollectionItemsImage方法查询指定NFT的图片和视频信息
		items, err := svcCtx.Dao.QueryCollectionItemsImage(ctx, chain, collectionAddr, []string{tokenID})
		if err != nil {
			queryErr = errors.Wrap(err, "failed on get items image info")
			return
		}

		// 将查询到的图片和视频信息存储到ItemExternals映射中
		for _, item := range items {
			ItemExternals[strings.ToLower(item.TokenId)] = item
		}
	}()

	// 5. 查询最近成交价格
	lastSales := make(map[string]decimal.Decimal)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 调用Dao层的QueryLastSalePrice方法查询指定NFT的最近成交价格
		lastSale, err := svcCtx.Dao.QueryLastSalePrice(ctx, chain, collectionAddr, []string{tokenID})
		if err != nil {
			queryErr = errors.Wrap(err, "failed on get items last sale info")
			return
		}

		// 将查询到的最近成交价格存储到lastSales映射中
		for _, v := range lastSale {
			lastSales[strings.ToLower(v.TokenId)] = v.Price
		}
	}()

	// 6. 查询最高出价信息
	bestBids := make(map[string]multi.Order)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 调用Dao层的QueryBestBids方法查询指定NFT的最高出价信息
		bids, err := svcCtx.Dao.QueryBestBids(ctx, chain, "", collectionAddr, []string{tokenID})
		if err != nil {
			queryErr = errors.Wrap(err, "failed on get items last sale info")
			return
		}

		// 遍历查询到的出价信息，找出最高出价
		for _, bid := range bids {
			order, ok := bestBids[strings.ToLower(bid.TokenId)]
			if !ok {
				bestBids[strings.ToLower(bid.TokenId)] = bid
				continue
			}
			if bid.Price.GreaterThan(order.Price) {
				bestBids[strings.ToLower(bid.TokenId)] = bid
			}
		}
	}()

	// 7. 查询collection最高出价信息
	var collectionBestBid multi.Order
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 调用Dao层的QueryCollectionBestBid方法查询指定集合的最高出价信息
		bid, err := svcCtx.Dao.QueryCollectionBestBid(ctx, chain, "", collectionAddr)
		if err != nil {
			queryErr = errors.Wrap(err, "failed on get items last sale info")
			return
		}
		collectionBestBid = bid
	}()

	// 等待所有查询完成
	wg.Wait()
	if queryErr != nil {
		return nil, errors.Wrap(queryErr, "failed on get items info")
	}

	// 组装返回数据
	var itemDetail types.ItemDetailInfo
	itemDetail.ChainID = chainID

	// 设置item基本信息
	if item != nil {
		itemDetail.Name = item.Name
		itemDetail.CollectionAddress = item.CollectionAddress
		itemDetail.TokenID = item.TokenId
		itemDetail.OwnerAddress = item.Owner
		// 设置collection级别的最高出价信息
		itemDetail.BidOrderID = collectionBestBid.OrderID
		itemDetail.BidExpireTime = collectionBestBid.ExpireTime
		itemDetail.BidPrice = collectionBestBid.Price
		itemDetail.BidTime = collectionBestBid.EventTime
		itemDetail.BidSalt = collectionBestBid.Salt
		itemDetail.BidMaker = collectionBestBid.Maker
		itemDetail.BidType = getBidType(collectionBestBid.OrderType)
		itemDetail.BidSize = collectionBestBid.Size
		itemDetail.BidUnfilled = collectionBestBid.QuantityRemaining
	}

	// 如果item级别的最高出价大于collection级别的最高出价,则使用item级别的出价信息
	bidOrder, ok := bestBids[strings.ToLower(item.TokenId)]
	if ok {
		if bidOrder.Price.GreaterThan(collectionBestBid.Price) {
			itemDetail.BidOrderID = bidOrder.OrderID
			itemDetail.BidExpireTime = bidOrder.ExpireTime
			itemDetail.BidPrice = bidOrder.Price
			itemDetail.BidTime = bidOrder.EventTime
			itemDetail.BidSalt = bidOrder.Salt
			itemDetail.BidMaker = bidOrder.Maker
			itemDetail.BidType = getBidType(bidOrder.OrderType)
			itemDetail.BidSize = bidOrder.Size
			itemDetail.BidUnfilled = bidOrder.QuantityRemaining
		}
	}

	// 设置挂单信息
	if itemListInfo != nil {
		itemDetail.ListPrice = itemListInfo.ListPrice
		itemDetail.MarketplaceID = itemListInfo.MarketID
		itemDetail.ListOrderID = itemListInfo.OrderID
		itemDetail.ListTime = itemListInfo.ListTime
		itemDetail.ListExpireTime = itemListInfo.ListExpireTime
		itemDetail.ListSalt = itemListInfo.ListSalt
		itemDetail.ListMaker = itemListInfo.ListMaker
	}

	// 设置collection信息
	if collection != nil {
		itemDetail.CollectionName = collection.Name
		itemDetail.FloorPrice = collection.FloorPrice
		itemDetail.CollectionImageURI = collection.ImageUri
		if itemDetail.Name == "" {
			itemDetail.Name = fmt.Sprintf("%s #%s", collection.Name, tokenID)
		}
	}

	// 设置最近成交价格
	price, ok := lastSales[strings.ToLower(tokenID)]
	if ok {
		itemDetail.LastSellPrice = price
	}

	// 设置图片和视频信息
	itemExternal, ok := ItemExternals[strings.ToLower(tokenID)]
	if ok {
		itemDetail.ImageURI = itemExternal.ImageUri
		if itemExternal.IsUploadedOss {
			itemDetail.ImageURI = itemExternal.OssUri
		}
		if len(itemExternal.VideoUri) > 0 {
			itemDetail.VideoType = itemExternal.VideoType
			if itemExternal.IsVideoUploaded {
				itemDetail.VideoURI = itemExternal.VideoOssUri
			} else {
				itemDetail.VideoURI = itemExternal.VideoUri
			}
		}
	}

	return &types.ItemDetailInfoResp{
		Result: itemDetail,
	}, nil
}

// GetItemTopTraitPrice 获取指定 token ids的Trait的最高价格信息
func GetItemTopTraitPrice(ctx context.Context, svcCtx *svc.ServerCtx, chain, collectionAddr string, tokenIDs []string) (*types.ItemTopTraitResp, error) {
	// 1. 查询Trait对应的最低挂单价格
	traitsPrice, err := svcCtx.Dao.QueryTraitsPrice(ctx, chain, collectionAddr, tokenIDs)
	if err != nil {
		return nil, errors.Wrap(err, "failed on calc top trait")
	}

	// 2. 空结果处理
	if len(traitsPrice) == 0 {
		return &types.ItemTopTraitResp{
			Result: []types.TraitPrice{},
		}, nil
	}

	// 3. 构建 Trait -> 最低挂单价格映射
	traitsPrices := make(map[string]decimal.Decimal)
	for _, traitPrice := range traitsPrice {
		traitsPrices[strings.ToLower(fmt.Sprintf("%s:%s", traitPrice.Trait, traitPrice.TraitValue))] = traitPrice.Price
	}

	// 4. 查询指定 token ids的 所有Trait
	traits, err := svcCtx.Dao.QueryItemsTraits(ctx, chain, collectionAddr, tokenIDs)
	if err != nil {
		return nil, errors.Wrap(err, "failed on query items trait")
	}

	// 5. 计算指定 token ids的 最高价值 Trait
	topTraits := make(map[string]types.TraitPrice)
	for _, trait := range traits {
		key := strings.ToLower(fmt.Sprintf("%s:%s", trait.Trait, trait.TraitValue))
		price, ok := traitsPrices[key]
		if ok {
			topPrice, ok := topTraits[trait.TokenId]
			// 如果已有最高价且当前价格不高于最高价,跳过
			if ok {
				if price.LessThanOrEqual(topPrice.Price) {
					continue
				}
			}

			// 更新最高价值 Trait
			topTraits[trait.TokenId] = types.TraitPrice{
				CollectionAddress: collectionAddr,
				TokenID:           trait.TokenId,
				Trait:             trait.Trait,
				TraitValue:        trait.TraitValue,
				Price:             price,
			}
		}
	}

	// 6. 整理返回结果
	var results []types.TraitPrice
	for _, topTrait := range topTraits {
		results = append(results, topTrait)
	}

	return &types.ItemTopTraitResp{
		Result: results,
	}, nil
}

// GetHistorySalesPrice 获取指定集合在特定时间范围内的历史销售价格信息。
// 参数:
// - ctx context.Context: 上下文，用于控制请求的生命周期。
// - svcCtx *svc.ServerCtx: 服务上下文，包含数据库访问和其他服务的实例。
// - chain string: 链名称，指定要查询的链。
// - collectionAddr string: 集合地址，指定要查询的NFT集合。
// - duration string: 时间范围，支持 "24h"、"7d"、"30d"。
// 返回:
// - []types.HistorySalesPriceInfo: 查询到的历史销售价格信息列表。
// - error: 若查询过程中出现错误，返回错误信息；否则返回 nil。
func GetHistorySalesPrice(ctx context.Context, svcCtx *svc.ServerCtx, chain, collectionAddr, duration string) ([]types.HistorySalesPriceInfo, error) {
	// 定义一个变量，用于存储时间范围对应的时间戳
	var durationTimeStamp int64
	// 根据传入的时间范围，计算对应的时间戳
	if duration == "24h" {
		// 24小时对应的秒数
		durationTimeStamp = 24 * 60 * 60
	} else if duration == "7d" {
		// 7天对应的秒数
		durationTimeStamp = 7 * 24 * 60 * 60
	} else if duration == "30d" {
		// 30天对应的秒数
		durationTimeStamp = 30 * 24 * 60 * 60
	} else {
		// 若传入的时间范围不支持，返回错误信息
		return nil, errors.New("only support 24h/7d/30d")
	}

	// 调用Dao层的QueryHistorySalesPriceInfo方法，查询指定集合在特定时间范围内的历史销售价格信息
	// 传入上下文、链名称、集合地址和时间戳作为参数
	historySalesPriceInfo, err := svcCtx.Dao.QueryHistorySalesPriceInfo(ctx, chain, collectionAddr, durationTimeStamp)
	// 检查查询过程中是否出现错误
	if err != nil {
		// 若出现错误，返回 nil 和包装后的错误信息，提示获取历史销售价格信息失败
		return nil, errors.Wrap(err, "failed on get history sales price info")
	}

	// 创建一个切片，用于存储最终的历史销售价格信息
	res := make([]types.HistorySalesPriceInfo, len(historySalesPriceInfo))

	// 遍历查询到的历史销售价格信息
	for i, ele := range historySalesPriceInfo {
		// 将查询到的信息转换为 HistorySalesPriceInfo 结构体，并存储到结果切片中
		res[i] = types.HistorySalesPriceInfo{
			Price:     ele.Price,     // 销售价格
			TokenID:   ele.TokenId,   // NFT的token ID
			TimeStamp: ele.EventTime, // 销售时间戳
		}
	}

	// 若查询成功，返回历史销售价格信息列表和 nil
	return res, nil
}

// GetItemOwner 从链上获取NFT的所有者信息，并更新数据库中的所有者信息
// ctx: 上下文信息
// svcCtx: 服务上下文信息
// chainID: 链ID
// chain: 链名
// collectionAddr: NFT集合地址
// tokenID: NFT的ID
// 返回值:
// *types.ItemOwner: 包含NFT所有者信息的结构体指针
// error: 错误信息
func GetItemOwner(ctx context.Context, svcCtx *svc.ServerCtx, chainID int64, chain, collectionAddr, tokenID string) (*types.ItemOwner, error) {
	// 从链上获取NFT所有者地址
	address, err := svcCtx.NodeSrvs[chainID].FetchNftOwner(collectionAddr, tokenID)
	if err != nil {
		xzap.WithContext(ctx).Error("failed on fetch nft owner onchain", zap.Error(err))
		return nil, errcode.ErrUnexpected
	}

	// 将地址转换为校验和格式
	owner, err := eip.ToCheckSumAddress(address.String())
	if err != nil {
		xzap.WithContext(ctx).Error("invalid address", zap.Error(err), zap.String("address", address.String()))
		return nil, errcode.ErrUnexpected
	}

	// 更新数据库中的所有者信息
	if err := svcCtx.Dao.UpdateItemOwner(ctx, chain, collectionAddr, tokenID, owner); err != nil {
		xzap.WithContext(ctx).Error("failed on update item owner", zap.Error(err), zap.String("address", address.String()))
	}

	// 返回NFT所有者信息
	return &types.ItemOwner{
		CollectionAddress: collectionAddr,
		TokenID:           tokenID,
		Owner:             owner,
	}, nil
}

// GetItemTraits 获取NFT的 Trait信息
// 主要功能:
// 1. 并发查询三个信息:
//   - NFT的 Trait信息
//   - 集合中每个 Trait的数量统计
//   - 集合基本信息
//
// 2. 计算每个 Trait的百分比
// 3. 组装返回数据
func GetItemTraits(ctx context.Context, svcCtx *svc.ServerCtx, chain, collectionAddr, tokenID string) ([]types.TraitInfo, error) {
	var traitInfos []types.TraitInfo
	var itemTraits []multi.ItemTrait
	var collection *multi.Collection
	var traitCounts []types.TraitCount
	var queryErr error
	var wg sync.WaitGroup

	// 并发查询NFT Trait信息
	wg.Add(1)
	go func() {
		defer wg.Done()
		itemTraits, queryErr = svcCtx.Dao.QueryItemTraits(
			ctx,
			chain,
			collectionAddr,
			tokenID,
		)
		if queryErr != nil {
			return
		}
	}()

	// 并发查询集合 Trait统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		traitCounts, queryErr = svcCtx.Dao.QueryCollectionTraits(
			ctx,
			chain,
			collectionAddr,
		)
		if queryErr != nil {
			return
		}
	}()

	// 并发查询集合信息
	wg.Add(1)
	go func() {
		defer wg.Done()
		collection, queryErr = svcCtx.Dao.QueryCollectionInfo(
			ctx,
			chain,
			collectionAddr,
		)
		if queryErr != nil {
			return
		}
	}()

	// 等待所有查询完成
	wg.Wait()
	if queryErr != nil {
		return nil, queryErr
	}

	// 如果NFT没有 Trait信息,返回空数组
	if len(itemTraits) == 0 {
		return traitInfos, nil
	}

	// 构建 Trait数量映射
	traitCountMap := make(map[string]int64)
	for _, trait := range traitCounts {
		traitCountMap[fmt.Sprintf("%s-%s", trait.Trait, trait.TraitValue)] = trait.Count
	}

	// 计算每个 Trait的百分比并组装返回数据
	for _, trait := range itemTraits {
		key := fmt.Sprintf("%s-%s", trait.Trait, trait.TraitValue)
		if count, ok := traitCountMap[key]; ok {
			traitPercent := 0.0
			if collection.ItemAmount != 0 {
				traitPercent = decimal.NewFromInt(count).
					DivRound(decimal.NewFromInt(collection.ItemAmount), 4).
					Mul(decimal.NewFromInt(100)).
					InexactFloat64()
			}
			traitInfos = append(traitInfos, types.TraitInfo{
				Trait:        trait.Trait,
				TraitValue:   trait.TraitValue,
				TraitAmount:  count,
				TraitPercent: traitPercent,
			})
		}
	}

	return traitInfos, nil
}

// GetCollectionDetail 获取NFT集合的详细信息：基本信息、24小时交易信息、上架数量、地板价、卖单价格、总交易量
func GetCollectionDetail(ctx context.Context, svcCtx *svc.ServerCtx, chain string, collectionAddr string) (*types.CollectionDetailResp, error) {
	// 查询集合基本信息
	collection, err := svcCtx.Dao.QueryCollectionInfo(ctx, chain, collectionAddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed on get collection info")
	}

	// 获取集合24小时交易信息
	tradeInfos, err := svcCtx.Dao.GetTradeInfoByCollection(chain, collectionAddr, "1d")
	if err != nil {
		xzap.WithContext(ctx).Error("failed on get collection trade info", zap.Error(err))
		//return nil, errcode.NewCustomErr("cache error")
	}

	// 查询上架数量
	listed, err := svcCtx.Dao.QueryListedAmount(ctx, chain, collectionAddr)
	if err != nil {
		xzap.WithContext(ctx).Error("failed on get listed count", zap.Error(err))
		//return nil, errcode.NewCustomErr("cache error")
	} else {
		// 缓存上架数量
		if err := svcCtx.Dao.CacheCollectionsListed(ctx, chain, collectionAddr, int(listed)); err != nil {
			xzap.WithContext(ctx).Error("failed on cache collection listed", zap.Error(err))
		}
	}

	// 查询地板价
	floorPrice, err := svcCtx.Dao.QueryFloorPrice(ctx, chain, collectionAddr)
	if err != nil {
		xzap.WithContext(ctx).Error("failed on get floor price", zap.Error(err))
	}

	// 查询卖单价格
	collectionSell, err := svcCtx.Dao.QueryCollectionSellPrice(ctx, chain, collectionAddr)
	if err != nil {
		xzap.WithContext(ctx).Error("failed on get floor price", zap.Error(err))
	}

	// 如果地板价发生变化,更新价格事件
	if !floorPrice.Equal(collection.FloorPrice) {
		if err := ordermanager.AddUpdatePriceEvent(svcCtx.KvStore, &ordermanager.TradeEvent{
			EventType:      ordermanager.UpdateCollection,
			CollectionAddr: collectionAddr,
			Price:          floorPrice,
		}, chain); err != nil {
			xzap.WithContext(ctx).Error("failed on update floor price", zap.Error(err))
		}
	}

	// 获取24小时交易量和销售数量
	var volume24h decimal.Decimal
	var sold int64
	if tradeInfos != nil {
		volume24h = tradeInfos.Volume
		sold = tradeInfos.ItemCount
	}

	// 查询总交易量
	var allVol decimal.Decimal
	collectionVol, err := svcCtx.Dao.GetCollectionVolume(chain, collectionAddr)
	if err != nil {
		xzap.WithContext(ctx).Error("failed on query collection all volume", zap.Error(err))
	} else {
		allVol = collectionVol
	}

	// 构建返回结果
	detail := types.CollectionDetail{
		ImageUri:    collection.ImageUri, // svcCtx.ImageMgr.GetFileUrl(collection.ImageUri),
		Name:        collection.Name,
		Address:     collection.Address,
		ChainId:     collection.ChainId,
		FloorPrice:  floorPrice,
		SellPrice:   collectionSell.SalePrice.String(),
		VolumeTotal: allVol,
		Volume24h:   volume24h,
		Sold24h:     sold,
		ListAmount:  listed,
		TotalSupply: collection.ItemAmount,
		OwnerAmount: collection.OwnerAmount,
	}

	return &types.CollectionDetailResp{
		Result: detail,
	}, nil
}

// RefreshItemMetadata refresh item meta data.
func RefreshItemMetadata(ctx context.Context, svcCtx *svc.ServerCtx, chainName string, chainId int64, collectionAddress, tokenId string) error {
	if err := mq.AddSingleItemToRefreshMetadataQueue(svcCtx.KvStore, svcCtx.C.ProjectCfg.Name, chainName, chainId, collectionAddress, tokenId); err != nil {
		xzap.WithContext(ctx).Error("failed on add item to refresh queue", zap.Error(err), zap.String("collection address: ", collectionAddress), zap.String("item_id", tokenId))
		return errcode.ErrUnexpected
	}

	return nil

}

// GetItemImage 获取指定NFT项目的图片信息。
// 参数:
// - ctx context.Context: 上下文，用于控制请求的生命周期。
// - svcCtx *svc.ServerCtx: 服务上下文，包含数据库访问和其他服务的实例。
// - chain string: 链名称，指定要查询的链。
// - collectionAddress string: 集合地址，指定要查询的NFT集合。
// - tokenId string: NFT的token ID，指定要查询的具体NFT。
// 返回:
// - *types.ItemImage: 查询到的NFT项目图片信息结构体指针，包含集合地址、token ID和图片URI。
// - error: 若查询过程中出现错误，返回错误信息；否则返回 nil。
func GetItemImage(ctx context.Context, svcCtx *svc.ServerCtx, chain string, collectionAddress, tokenId string) (*types.ItemImage, error) {
	// 调用Dao层的QueryCollectionItemsImage方法，查询指定集合中特定NFT的图片信息
	// 传入上下文、链名称、集合地址和token ID列表作为参数
	items, err := svcCtx.Dao.QueryCollectionItemsImage(ctx, chain, collectionAddress, []string{tokenId})
	// 检查查询过程中是否出现错误，或者未查询到任何图片信息
	if err != nil || len(items) == 0 {
		// 若出现错误或未查询到信息，返回 nil 和包装后的错误信息，提示获取项目图片信息失败
		return nil, errors.Wrap(err, "failed on get item image")
	}
	// 定义一个字符串变量，用于存储图片的URI
	var imageUri string
	// 检查查询到的图片信息中，是否已上传到OSS（对象存储服务）
	if items[0].IsUploadedOss {
		// 如果已上传到OSS，使用OSS的URI作为图片的URI
		imageUri = items[0].OssUri // svcCtx.ImageMgr.GetSmallSizeImageUrl(items[0].OssUri)
	} else {
		// 如果未上传到OSS，使用原始的图片URI
		imageUri = items[0].ImageUri // svcCtx.ImageMgr.GetSmallSizeImageUrl(items[0].ImageUri)
	}

	// 若查询成功，创建并返回包含集合地址、token ID和图片URI的响应结构体指针
	return &types.ItemImage{
		CollectionAddress: collectionAddress, // NFT所属集合的地址
		TokenID:           tokenId,           // NFT的token ID
		ImageUri:          imageUri,          // NFT的图片URI
	}, nil
}
