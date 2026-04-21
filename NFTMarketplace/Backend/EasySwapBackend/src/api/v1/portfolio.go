package v1

import (
	"encoding/json"

	"github.com/ProjectsTask/EasySwapBase/errcode"
	"github.com/ProjectsTask/EasySwapBase/xhttp"
	"github.com/gin-gonic/gin"

	"github.com/ProjectsTask/EasySwapBackend/src/service/svc"
	"github.com/ProjectsTask/EasySwapBackend/src/service/v1"
	"github.com/ProjectsTask/EasySwapBackend/src/types/v1"
)

func UserMultiChainCollectionsHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		filterParam := c.Query("filters")
		if filterParam == "" {
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		var filter types.UserCollectionsParams
		err := json.Unmarshal([]byte(filterParam), &filter)
		if err != nil {
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		var chainNames []string
		var chainIDs []int
		for _, chain := range svcCtx.C.ChainSupported {
			chainIDs = append(chainIDs, chain.ChainID)
			chainNames = append(chainNames, chain.Name)
		}

		res, err := service.GetMultiChainUserCollections(c.Request.Context(), svcCtx, chainIDs, chainNames, filter.UserAddresses)
		if err != nil {
			xhttp.Error(c, errcode.NewCustomErr("query user multi chain collections err."))
			return
		}

		xhttp.OkJson(c, res)
	}
}

// UserMultiChainItemsHandler 返回一个处理多链用户项目请求的 gin.HandlerFunc
// 该函数解析请求参数，获取多链用户项目信息，并返回查询结果。
func UserMultiChainItemsHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取查询参数
		filterParam := c.Query("filters")
		if filterParam == "" {
			// 如果没有提供过滤参数，返回错误
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		var filter types.PortfolioMultiChainItemFilterParams
		// 将查询参数解析为结构体
		err := json.Unmarshal([]byte(filterParam), &filter)
		if err != nil {
			// 解析失败，返回错误
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		// 如果过滤条件中的ChainID为空，则显示所有链的信息
		if len(filter.ChainID) == 0 {
			for _, chain := range svcCtx.C.ChainSupported {
				filter.ChainID = append(filter.ChainID, chain.ChainID)
			}
		}

		var chainNames []string
		for _, chainID := range filter.ChainID {
			chain, ok := chainIDToChain[chainID]
			if !ok {
				// 链ID不存在，返回错误
				xhttp.Error(c, errcode.ErrInvalidParams)
				return
			}
			chainNames = append(chainNames, chain)
		}

		// 获取多链用户项目信息
		res, err := service.GetMultiChainUserItems(c.Request.Context(), svcCtx, filter.ChainID, chainNames, filter.UserAddresses, filter.CollectionAddresses, filter.Page, filter.PageSize)
		if err != nil {
			// 查询失败，返回错误
			xhttp.Error(c, errcode.NewCustomErr("query user multi chain items err."))
			return
		}

		// 返回查询结果
		xhttp.OkJson(c, res)
	}
}

func UserMultiChainListingsHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取查询参数中的过滤器
		filterParam := c.Query("filters")
		if filterParam == "" {
			// 如果过滤器参数为空，则返回错误
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		var filter types.PortfolioMultiChainListingFilterParams
		// 将过滤器参数解析为结构体
		err := json.Unmarshal([]byte(filterParam), &filter)
		if err != nil {
			// 如果解析出错，则返回错误
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		// 如果过滤器的链ID为空，则显示所有链的信息
		// if filter.ChainID is empty, show all chain info
		if len(filter.ChainID) == 0 {
			for _, chain := range svcCtx.C.ChainSupported {
				filter.ChainID = append(filter.ChainID, chain.ChainID)
			}
		}

		var chainNames []string
		// 获取链名称
		for _, chainID := range filter.ChainID {
			chain, ok := chainIDToChain[chainID]
			if !ok {
				// 如果链ID不存在，则返回错误
				xhttp.Error(c, errcode.ErrInvalidParams)
				return
			}
			chainNames = append(chainNames, chain)
		}

		// 获取多链用户列表
		res, err := service.GetMultiChainUserListings(c.Request.Context(), svcCtx, filter.ChainID, chainNames, filter.UserAddresses, filter.CollectionAddresses, filter.Page, filter.PageSize)
		if err != nil {
			// 如果查询出错，则返回错误
			xhttp.Error(c, errcode.NewCustomErr("query user multi chain items err."))
			return
		}

		// 返回查询结果
		xhttp.OkJson(c, res)
	}
}

func UserMultiChainBidsHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取查询参数中的过滤器参数
		filterParam := c.Query("filters")
		if filterParam == "" {
			// 如果过滤器参数为空，则返回错误
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		// 将过滤器参数转换为结构体
		var filter types.PortfolioMultiChainBidFilterParams
		err := json.Unmarshal([]byte(filterParam), &filter)
		if err != nil {
			// 如果解析过滤器参数失败，则返回错误
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		// 如果过滤器中的链ID为空，则显示所有链的信息
		// if filter.ChainID is empty, show all chain info
		if len(filter.ChainID) == 0 {
			for _, chain := range svcCtx.C.ChainSupported {
				filter.ChainID = append(filter.ChainID, chain.ChainID)
			}
		}

		var chainNames []string
		for _, chainID := range filter.ChainID {
			// 获取链名称
			chain, ok := chainIDToChain[chainID]
			if !ok {
				// 如果链ID不存在，则返回错误
				xhttp.Error(c, errcode.ErrInvalidParams)
				return
			}
			chainNames = append(chainNames, chain)
		}

		// 获取用户跨链投标信息
		res, err := service.GetMultiChainUserBids(c.Request.Context(), svcCtx, filter.ChainID, chainNames, filter.UserAddresses, filter.CollectionAddresses, filter.Page, filter.PageSize)
		if err != nil {
			// 如果查询失败，则返回错误
			xhttp.Error(c, errcode.NewCustomErr("query user multi chain items err."))
			return
		}

		// 返回查询结果
		xhttp.OkJson(c, res)
	}
}
