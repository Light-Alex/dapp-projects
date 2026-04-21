package v1

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ProjectsTask/EasySwapBase/errcode"
	"github.com/ProjectsTask/EasySwapBase/logger/xzap"
	"github.com/ProjectsTask/EasySwapBase/xhttp"

	"github.com/ProjectsTask/EasySwapBackend/src/service/svc"
	"github.com/ProjectsTask/EasySwapBackend/src/service/v1"
	"github.com/ProjectsTask/EasySwapBackend/src/types/v1"
)

// CollectionItemsHandler 处理获取集合中项目信息的请求。
// 它接收一个服务上下文对象 svcCtx，并返回一个 gin.HandlerFunc。
// 该处理函数会从请求中获取过滤参数、集合地址和链信息，验证参数的有效性，
// 然后调用 service.GetItems 函数获取集合中项目的信息。
// 如果发生错误，会返回相应的错误响应；如果成功，会以JSON格式返回结果。
func CollectionItemsHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求中获取过滤参数
		filterParam := c.Query("filters")
		// 检查过滤参数是否为空
		if filterParam == "" {
			// 如果为空，返回自定义错误响应，提示过滤参数为空
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			// 结束处理
			return
		}

		// 定义一个 CollectionItemFilterParams 类型的变量，用于存储解析后的过滤参数
		var filter types.CollectionItemFilterParams
		// 将过滤参数字符串解析为 CollectionItemFilterParams 类型
		err := json.Unmarshal([]byte(filterParam), &filter)
		// 检查解析过程中是否发生错误
		if err != nil {
			// 如果发生错误，返回自定义错误响应，提示过滤参数无效
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			// 结束处理
			return
		}

		// 从请求参数中获取集合地址
		collectionAddr := c.Params.ByName("address")
		// 检查集合地址是否为空
		if collectionAddr == "" {
			// 如果为空，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 根据过滤参数中的链ID查找对应的链
		chain, ok := chainIDToChain[filter.ChainID]
		// 检查是否找到对应的链
		if !ok {
			// 如果未找到，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 调用服务层的 GetItems 函数获取集合中项目的信息
		res, err := service.GetItems(c.Request.Context(), svcCtx, chain, filter, collectionAddr)
		// 检查是否发生错误
		if err != nil {
			// 如果发生错误，返回意外错误的响应
			xhttp.Error(c, errcode.ErrUnexpected)
			// 结束处理
			return
		}

		// 若未发生错误，将结果以JSON格式返回给客户端
		xhttp.OkJson(c, res)
	}
}

// CollectionBidsHandler 处理获取集合出价信息的请求。
// 它接收一个服务上下文对象 svcCtx，并返回一个 gin.HandlerFunc。
// 该处理函数会从请求中获取过滤参数、集合地址和链信息，验证参数的有效性，
// 然后调用 service.GetBids 函数获取集合的出价信息。
// 如果发生错误，会返回相应的错误响应；如果成功，会以JSON格式返回结果。
func CollectionBidsHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求中获取过滤参数
		filterParam := c.Query("filters")
		// 检查过滤参数是否为空
		if filterParam == "" {
			// 如果为空，返回自定义错误响应，提示过滤参数为空
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			// 结束处理
			return
		}

		// 定义一个 CollectionBidFilterParams 类型的变量，用于存储解析后的过滤参数
		var filter types.CollectionBidFilterParams
		// 将过滤参数字符串解析为 CollectionBidFilterParams 类型
		err := json.Unmarshal([]byte(filterParam), &filter)
		// 检查解析过程中是否发生错误
		if err != nil {
			// 如果发生错误，返回自定义错误响应，提示过滤参数无效
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			// 结束处理
			return
		}

		// 从请求参数中获取集合地址
		collectionAddr := c.Params.ByName("address")
		// 检查集合地址是否为空
		if collectionAddr == "" {
			// 如果为空，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 根据过滤参数中的链ID查找对应的链
		chain, ok := chainIDToChain[int(filter.ChainID)]
		// 检查是否找到对应的链
		if !ok {
			// 如果未找到，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 调用服务层的 GetBids 函数获取集合的出价信息
		res, err := service.GetBids(c.Request.Context(), svcCtx, chain, collectionAddr, filter.Page, filter.PageSize)
		// 检查是否发生错误
		if err != nil {
			// 如果发生错误，返回意外错误的响应
			xhttp.Error(c, errcode.ErrUnexpected)
			// 结束处理
			return
		}

		// 若未发生错误，将结果以JSON格式返回给客户端
		xhttp.OkJson(c, res)
	}
}

// CollectionItemBidsHandler 处理获取集合中单个NFT项目出价信息的请求。
// 它接收一个服务上下文对象 svcCtx，并返回一个 gin.HandlerFunc。
// 该处理函数会从请求中获取过滤参数、集合地址、NFT的token ID和链信息，验证参数的有效性，
// 然后调用 service.GetItemBidsInfo 函数获取指定集合中单个NFT项目的出价信息。
// 如果发生错误，会返回相应的错误响应；如果成功，会以JSON格式返回结果。
func CollectionItemBidsHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求中获取过滤参数
		filterParam := c.Query("filters")
		// 检查过滤参数是否为空
		if filterParam == "" {
			// 如果为空，返回自定义错误响应，提示过滤参数为空
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			// 结束处理
			return
		}

		// 定义一个 CollectionBidFilterParams 类型的变量，用于存储解析后的过滤参数
		var filter types.CollectionBidFilterParams
		// 将过滤参数字符串解析为 CollectionBidFilterParams 类型
		err := json.Unmarshal([]byte(filterParam), &filter)
		// 检查解析过程中是否发生错误
		if err != nil {
			// 如果发生错误，返回自定义错误响应，提示过滤参数无效
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			// 结束处理
			return
		}

		// 从请求参数中获取集合地址
		collectionAddr := c.Params.ByName("address")
		// 检查集合地址是否为空
		if collectionAddr == "" {
			// 如果为空，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 从请求参数中获取NFT的token ID
		tokenID := c.Params.ByName("token_id")
		// 检查token ID是否为空
		if tokenID == "" {
			// 如果为空，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 根据过滤参数中的链ID查找对应的链
		chain, ok := chainIDToChain[int(filter.ChainID)]
		// 检查是否找到对应的链
		if !ok {
			// 如果未找到，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 调用服务层的 GetItemBidsInfo 函数获取指定集合中单个NFT项目的出价信息
		res, err := service.GetItemBidsInfo(c.Request.Context(), svcCtx, chain, collectionAddr, tokenID, filter.Page, filter.PageSize)
		// 检查是否发生错误
		if err != nil {
			// 如果发生错误，返回意外错误的响应
			xhttp.Error(c, errcode.ErrUnexpected)
			// 结束处理
			return
		}

		// 若未发生错误，将结果以JSON格式返回给客户端
		xhttp.OkJson(c, res)
	}
}

// ItemDetailHandler 处理获取单个NFT项目详情的请求。
// 它接收一个服务上下文对象 svcCtx，并返回一个 gin.HandlerFunc。
// 该处理函数会从请求中获取链ID、集合地址和NFT的token ID，验证参数的有效性，
// 然后调用 service.GetItem 函数获取指定NFT项目的详情。
// 如果发生错误，会返回相应的错误响应；如果成功，会以JSON格式返回结果。
func ItemDetailHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求参数中获取集合地址
		collectionAddr := c.Params.ByName("address")
		// 检查集合地址是否为空
		if collectionAddr == "" {
			// 如果为空，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 从请求参数中获取NFT的token ID
		tokenID := c.Params.ByName("token_id")
		// 检查token ID是否为空
		if tokenID == "" {
			// 如果为空，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 从请求参数中获取链ID并将其转换为 int64 类型
		chainID, err := strconv.ParseInt(c.Query("chain_id"), 10, 64)
		// 检查转换过程中是否发生错误
		if err != nil {
			// 如果发生错误，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 根据链ID查找对应的链
		chain, ok := chainIDToChain[int(chainID)]
		// 检查是否找到对应的链
		if !ok {
			// 如果未找到，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 调用服务层的 GetItem 函数获取指定NFT项目的详情
		res, err := service.GetItem(c.Request.Context(), svcCtx, chain, int(chainID), collectionAddr, tokenID)
		// 检查是否发生错误
		if err != nil {
			// 如果发生错误，返回自定义错误响应，提示获取项目信息失败
			xhttp.Error(c, errcode.NewCustomErr("get item error"))
			// 结束处理
			return
		}

		// 若未发生错误，将结果以JSON格式返回给客户端
		xhttp.OkJson(c, res)
	}
}

// ItemTopTraitPriceHandler 处理获取指定集合中，特定NFT的顶级特征价格信息的请求。
// 它接收一个服务上下文对象 svcCtx，并返回一个 gin.HandlerFunc。
// 该处理函数会从请求中获取过滤参数、集合地址和链信息，验证参数的有效性，
// 然后调用 service.GetItemTopTraitPrice 函数获取指定集合中，特定NFT的顶级特征价格信息。
// 如果发生错误，会返回相应的错误响应；如果成功，会以JSON格式返回结果。
func ItemTopTraitPriceHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求参数中获取集合地址
		collectionAddr := c.Params.ByName("address")
		// 检查集合地址是否为空
		if collectionAddr == "" {
			// 如果为空，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 从请求中获取过滤参数
		filterParam := c.Query("filters")
		// 检查过滤参数是否为空
		if filterParam == "" {
			// 如果为空，返回自定义错误响应，提示过滤参数为空
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			// 结束处理
			return
		}

		// 定义一个 TopTraitFilterParams 类型的变量，用于存储解析后的过滤参数
		var filter types.TopTraitFilterParams
		// 将过滤参数字符串解析为 TopTraitFilterParams 类型
		err := json.Unmarshal([]byte(filterParam), &filter)
		// 检查解析过程中是否发生错误
		if err != nil {
			// 如果发生错误，返回自定义错误响应，提示过滤参数无效
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			// 结束处理
			return
		}

		// 根据过滤参数中的链ID查找对应的链
		chain, ok := chainIDToChain[filter.ChainID]
		// 检查是否找到对应的链
		if !ok {
			// 如果未找到，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 调用服务层的 GetItemTopTraitPrice 函数获取指定集合中，特定NFT的顶级特征价格信息
		res, err := service.GetItemTopTraitPrice(c.Request.Context(), svcCtx, chain, collectionAddr, filter.TokenIds)
		// 检查是否发生错误
		if err != nil {
			// 如果发生错误，返回自定义错误响应，提示获取项目信息失败
			xhttp.Error(c, errcode.NewCustomErr("get item error"))
			// 结束处理
			return
		}
		// 若未发生错误，将结果以JSON格式返回给客户端
		xhttp.OkJson(c, res)
	}
}

// HistorySalesHandler 处理获取指定集合历史销售价格信息的请求。
// 它接收一个服务上下文对象 svcCtx，并返回一个 gin.HandlerFunc。
// 该处理函数会从请求中获取集合地址、链ID和时间范围，验证参数的有效性，
// 然后调用 service.GetHistorySalesPrice 函数获取指定集合的历史销售价格信息。
// 如果发生错误，会返回相应的错误响应；如果成功，会以JSON格式返回结果。
func HistorySalesHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求参数中获取集合地址
		collectionAddr := c.Params.ByName("address")
		// 检查集合地址是否为空
		if collectionAddr == "" {
			// 如果为空，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 从请求参数中获取链ID并将其转换为 int64 类型
		chainID, err := strconv.ParseInt(c.Query("chain_id"), 10, 64)
		// 检查转换过程中是否发生错误
		if err != nil {
			// 如果发生错误，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 根据链ID查找对应的链
		chain, ok := chainIDToChain[int(chainID)]
		// 检查是否找到对应的链
		if !ok {
			// 如果未找到，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 从请求中获取时间范围参数
		duration := c.Query("duration")
		// 检查时间范围参数是否为空
		if duration != "" {
			// 定义有效的时间范围参数
			validParams := map[string]bool{
				"24h": true,
				"7d":  true,
				"30d": true,
			}
			// 检查传入的时间范围是否有效
			if ok := validParams[duration]; !ok {
				// 若无效，记录错误日志
				xzap.WithContext(c).Error("duration parse error: ", zap.String("duration", duration))
				// 返回无效参数的错误响应
				xhttp.Error(c, errcode.ErrInvalidParams)
				// 结束处理
				return
			}
		} else {
			// 若时间范围参数为空，默认设置为 7 天
			duration = "7d"
		}

		// 调用服务层的 GetHistorySalesPrice 函数获取指定集合的历史销售价格信息
		res, err := service.GetHistorySalesPrice(c.Request.Context(), svcCtx, chain, collectionAddr, duration)
		// 检查是否发生错误
		if err != nil {
			// 如果发生错误，返回自定义错误响应，提示获取历史销售价格信息失败
			xhttp.Error(c, errcode.NewCustomErr("get history sales price error"))
			// 结束处理
			return
		}

		// 若未发生错误，将结果以JSON格式返回给客户端
		xhttp.OkJson(c, struct {
			// Result 表示获取到的历史销售价格信息
			Result interface{} `json:"result"`
		}{
			Result: res,
		})
	}
}

// ItemTraitsHandler 处理获取单个NFT项目特征信息的请求。
// 它接收一个服务上下文对象 svcCtx，并返回一个 gin.HandlerFunc。
// 该处理函数会从请求中获取链ID、集合地址和NFT的token ID，验证参数的有效性，
// 然后调用 service.GetItemTraits 函数获取指定NFT项目的特征信息。
// 如果发生错误，会返回相应的错误响应；如果成功，会以JSON格式返回结果。
func ItemTraitsHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求参数中获取集合地址
		collectionAddr := c.Params.ByName("address")
		// 检查集合地址是否为空
		if collectionAddr == "" {
			// 如果为空，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 从请求参数中获取NFT的token ID
		tokenID := c.Params.ByName("token_id")
		// 检查token ID是否为空
		if tokenID == "" {
			// 如果为空，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 从请求参数中获取链ID并将其转换为 int64 类型
		chainID, err := strconv.ParseInt(c.Query("chain_id"), 10, 64)
		// 检查转换过程中是否发生错误
		if err != nil {
			// 如果发生错误，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 根据链ID查找对应的链
		chain, ok := chainIDToChain[int(chainID)]
		// 检查是否找到对应的链
		if !ok {
			// 如果未找到，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 调用服务层的 GetItemTraits 函数获取指定NFT项目的特征信息
		itemTraits, err := service.GetItemTraits(c.Request.Context(), svcCtx, chain, collectionAddr, tokenID)
		// 检查是否发生错误
		if err != nil {
			// 如果发生错误，返回自定义错误响应，提示获取项目特征信息失败
			xhttp.Error(c, errcode.NewCustomErr("get item traits error"))
			// 结束处理
			return
		}

		// 若未发生错误，将结果以JSON格式返回给客户端
		xhttp.OkJson(c, types.ItemTraitsResp{Result: itemTraits})
	}
}

// GetItemImageHandler 处理获取单个NFT项目图片信息的请求。
// 它接收一个服务上下文对象 svcCtx，并返回一个 gin.HandlerFunc。
// 该处理函数会从请求中获取链ID、集合地址和NFT的token ID，验证参数的有效性，
// 然后调用 service.GetItemImage 函数获取指定NFT项目的图片信息。
// 如果发生错误，会返回相应的错误响应；如果成功，会以JSON格式返回结果。
func ItemOwnerHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求参数中获取集合地址
		collectionAddr := c.Params.ByName("address")
		// 检查集合地址是否为空
		if collectionAddr == "" {
			// 如果为空，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}
		// 从请求参数中获取NFT的token ID
		tokenID := c.Params.ByName("token_id")
		// 检查token ID是否为空
		if tokenID == "" {
			// 如果为空，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}
		// 从请求参数中获取链ID并将其转换为 int64 类型
		chainID, err := strconv.ParseInt(c.Query("chain_id"), 10, 64)
		// 检查转换过程中是否发生错误
		if err != nil {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}
		// 根据链ID查找对应的链
		chain, ok := chainIDToChain[int(chainID)]
		if !ok {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}
		// 获取NFT Item的所有者信息
		owner, err := service.GetItemOwner(c.Request.Context(), svcCtx, chainID, chain, collectionAddr, tokenID)
		if err != nil {
			xhttp.Error(c, errcode.NewCustomErr("get item owner error"))
			return
		}

		xhttp.OkJson(c, struct {
			Result interface{} `json:"result"`
		}{
			Result: owner,
		})
	}
}

// GetItemImageHandler 处理获取单个NFT项目图片信息的请求。
// 它接收一个服务上下文对象 svcCtx，并返回一个 gin.HandlerFunc。
// 该处理函数会从请求中获取链ID、集合地址和NFT的token ID，验证参数的有效性，
// 然后调用 service.GetItemImage 函数获取指定NFT项目的图片信息。
// 如果发生错误，会返回相应的错误响应；如果成功，会以JSON格式返回结果。
func GetItemImageHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求参数中获取集合地址
		collectionAddr := c.Params.ByName("address")
		// 检查集合地址是否为空
		if collectionAddr == "" {
			// 如果为空，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 从请求参数中获取NFT的token ID
		tokenID := c.Params.ByName("token_id")
		// 检查token ID是否为空
		if tokenID == "" {
			// 如果为空，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 从请求参数中获取链ID并将其转换为 int64 类型
		chainID, err := strconv.ParseInt(c.Query("chain_id"), 10, 64)
		// 检查转换过程中是否发生错误
		if err != nil {
			// 如果发生错误，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 根据链ID查找对应的链
		chain, ok := chainIDToChain[int(chainID)]
		// 检查是否找到对应的链
		if !ok {
			// 如果未找到，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 调用服务层的 GetItemImage 函数获取指定NFT项目的图片信息
		result, err := service.GetItemImage(c.Request.Context(), svcCtx, chain, collectionAddr, tokenID)
		// 检查是否发生错误
		if err != nil {
			// 如果发生错误，返回自定义错误响应，提示获取项目图片信息失败
			xhttp.Error(c, errcode.NewCustomErr("failed on get item image"))
			// 结束处理
			return
		}

		// 若未发生错误，将结果以JSON格式返回给客户端
		xhttp.OkJson(c, struct {
			// Result 表示获取到的项目图片信息，使用 any 类型替代 interface{} 以提高代码可读性
			Result any `json:"result"`
		}{Result: result})
	}
}

func ItemMetadataRefreshHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析chain_id参数
		chainId, err := strconv.ParseInt(c.Query("chain_id"), 10, 32)
		if err != nil {
			// 参数错误
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		// 根据chainId获取对应的chain
		chain, ok := chainIDToChain[int(chainId)]
		if !ok {
			// 参数错误
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		// 获取address参数
		collectionAddr := c.Params.ByName("address")
		if collectionAddr == "" {
			// 参数错误
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		// 获取token_id参数
		tokenId := c.Params.ByName("token_id")
		if tokenId == "" {
			// 参数错误
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		// 调用RefreshItemMetadata函数刷新元数据
		err = service.RefreshItemMetadata(c.Request.Context(), svcCtx, chain, chainId, collectionAddr, tokenId)
		if err != nil {
			// 刷新失败
			xhttp.Error(c, err)
			return
		}

		// 刷新成功，返回成功信息
		successStr := "Success to joined the refresh queue and waiting for refresh."
		xhttp.OkJson(c, types.CommonResp{Result: successStr})
	}
}

// CollectionDetailHandler 处理获取集合详情的请求。
// 它接收一个服务上下文对象 svcCtx，并返回一个 gin.HandlerFunc。
// 该处理函数会从请求参数中获取链ID和集合地址，验证参数的有效性，
// 然后调用 service.GetCollectionDetail 函数获取集合详情。
// 如果发生错误，会返回相应的错误响应；如果成功，会以JSON格式返回结果。
func CollectionDetailHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求参数中获取链ID并将其转换为 int32 类型
		chainID, err := strconv.ParseInt(c.Query("chain_id"), 10, 32)
		// 检查转换过程中是否发生错误
		if err != nil {
			// 如果发生错误，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 根据链ID查找对应的链
		chain, ok := chainIDToChain[int(chainID)]
		// 检查是否找到对应的链
		if !ok {
			// 如果未找到，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}

		// 从请求参数中获取集合地址
		collectionAddr := c.Params.ByName("address")
		// 检查集合地址是否为空
		if collectionAddr == "" {
			// 如果为空，返回无效参数的错误响应
			xhttp.Error(c, errcode.ErrInvalidParams)
			// 结束处理
			return
		}
		// 调用服务层的 GetCollectionDetail 函数获取集合详情
		res, err := service.GetCollectionDetail(c.Request.Context(), svcCtx, chain, collectionAddr)
		// 检查是否发生错误
		if err != nil {
			// 如果发生错误，返回意外错误的响应
			xhttp.Error(c, errcode.ErrUnexpected)
			// 结束处理
			return
		}

		// 若未发生错误，将结果以JSON格式返回给客户端
		xhttp.OkJson(c, res)
	}
}
