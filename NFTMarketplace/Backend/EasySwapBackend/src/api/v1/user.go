package v1

import (
	"github.com/ProjectsTask/EasySwapBase/errcode"
	"github.com/ProjectsTask/EasySwapBase/kit/validator"
	"github.com/ProjectsTask/EasySwapBase/xhttp"
	"github.com/gin-gonic/gin"

	"github.com/ProjectsTask/EasySwapBackend/src/service/svc"
	"github.com/ProjectsTask/EasySwapBackend/src/service/v1"
	"github.com/ProjectsTask/EasySwapBackend/src/types/v1"
)

// UserLoginHandler 返回一个Gin处理函数，用于处理用户登录请求。
// 参数 svcCtx 是一个指向服务上下文的指针，包含了应用程序所需的服务和配置。
// 返回一个 gin.HandlerFunc 类型的函数，用于处理HTTP请求。
func UserLoginHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建一个 LoginReq 结构体实例，用于存储解析后的请求数据
		req := types.LoginReq{}
		// 尝试将请求的JSON数据绑定到 req 结构体中
		if err := c.BindJSON(&req); err != nil {
			// 如果绑定失败，返回错误响应
			xhttp.Error(c, err)
			return
		}

		// 使用 validator 包验证请求数据
		if err := validator.Verify(&req); err != nil {
			// 如果验证失败，返回自定义错误响应
			xhttp.Error(c, errcode.NewCustomErr(err.Error()))
			return
		}

		// 调用服务层的 UserLogin 函数进行用户登录操作
		res, err := service.UserLogin(c.Request.Context(), svcCtx, req)
		if err != nil {
			// 如果登录过程中出现错误，返回自定义错误响应
			xhttp.Error(c, errcode.NewCustomErr(err.Error()))
			return
		}

		// 登录成功，返回包含登录结果的JSON响应
		xhttp.OkJson(c, types.UserLoginResp{
			Result: res,
		})
	}
}

// GetSigStatusHandler 处理获取签名状态消息的请求。
// 它接收一个服务上下文对象 svcCtx，并返回一个 gin.HandlerFunc。
// 该处理函数会从请求参数中获取用户地址，验证地址是否为空，
// 然后调用 service.GetSigStatusMsg 函数获取签名状态消息。
// 如果发生错误，会返回自定义错误响应；如果成功，会以JSON格式返回结果。
func GetLoginMessageHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求参数中获取用户地址
		address := c.Params.ByName("address")
		// 检查用户地址是否为空
		if address == "" {
			// 如果为空，返回自定义错误响应
			xhttp.Error(c, errcode.NewCustomErr("user addr is null"))
			// 结束处理
			return
		}
		// 调用服务层的 GetSigStatusMsg 函数获取签名状态消息
		res, err := service.GetUserLoginMsg(c.Request.Context(), svcCtx, address)
		// 检查是否发生错误
		if err != nil {
			// 返回自定义错误响应，结束处理
			xhttp.Error(c, errcode.NewCustomErr(err.Error()))
			return
		}
		// 若未发生错误，将结果以JSON格式返回给客户端
		xhttp.OkJson(c, res)
	}
}

// GetSigStatusHandler 返回一个Gin处理函数，用于处理获取用户签名状态的请求。
// 参数 svcCtx 是一个指向服务上下文的指针，包含了应用程序所需的服务和配置。
// 返回一个 gin.HandlerFunc 类型的函数，用于处理HTTP请求。
func GetSigStatusHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求的URL参数中获取用户地址
		userAddr := c.Params.ByName("address")
		// 检查获取的用户地址是否为空
		if userAddr == "" {
			// 若为空，返回一个自定义错误响应，提示用户地址为空
			xhttp.Error(c, errcode.NewCustomErr("user addr is null"))
			// 结束当前请求处理
			return
		}

		// 调用服务层的 GetSigStatusMsg 函数，传入请求上下文、服务上下文和用户地址，获取签名状态信息
		res, err := service.GetSigStatusMsg(c.Request.Context(), svcCtx, userAddr)
		// 检查调用服务层函数时是否出现错误
		if err != nil {
			// 若出现错误，返回一个包含错误信息的自定义错误响应
			xhttp.Error(c, errcode.NewCustomErr(err.Error()))
			// 若出现错误，返回错误响应并结束当前请求处理
			return
		}

		// 若没有错误，以JSON格式返回成功响应，包含签名状态信息
		xhttp.OkJson(c, res)
	}
}
