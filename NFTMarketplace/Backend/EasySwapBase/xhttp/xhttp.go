package xhttp

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"

	"github.com/ProjectsTask/EasySwapBase/kit/convert"
	"github.com/ProjectsTask/EasySwapBase/kit/validator"
	"github.com/ProjectsTask/EasySwapBase/logger/xzap"

	"github.com/ProjectsTask/EasySwapBase/errcode"
)

const (
	halfShowLen            = 100
	defaultMultipartMemory = 32 << 20 // 32 MB
)

// Response 业务通用响应体
type Response struct {
	TraceId string      `json:"trace_id" example:"a1b2c3d4e5f6g7h8" extensions:"x-order=000"` // 链路追踪id
	Code    uint32      `json:"code" example:"200" extensions:"x-order=001"`                  // 状态码
	Msg     string      `json:"msg" example:"OK" extensions:"x-order=002"`                    // 消息
	Data    interface{} `json:"data" extensions:"x-order=003"`                                // 数据
}

// GetTraceId 获取链路追踪id
func GetTraceId(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}

	return ""
}

// WriteHeader 写入自定义响应header
func WriteHeader(w http.ResponseWriter, err ...error) {
	// 声明一个 error 类型的变量 ee
	var ee error
	// 判断 err 参数的长度是否大于 0
	if len(err) > 0 {
		// 将 err 参数中的第一个错误赋值给 ee
		ee = err[0]
	}

	// 使用 errcode.ParseErr 函数解析 ee，得到 e
	e := errcode.ParseErr(ee)
	// 设置响应头 HeaderGWErrorCode 的值为 e.Code() 的字符串表示
	w.Header().Set(HeaderGWErrorCode, convert.ToString(e.Code()))
	// 使用 url.QueryEscape 对 e.Error() 进行 URL 编码，并将编码后的结果设置为响应头 HeaderGWErrorMessage 的值
	w.Header().Set(HeaderGWErrorMessage, url.QueryEscape(e.Error()))
}

// OkJson 成功json响应返回
func OkJson(c *gin.Context, v interface{}) {
	// 设置响应头
	WriteHeader(c.Writer)

	// 设置JSON响应
	c.JSON(http.StatusOK, &Response{
		// 设置TraceId
		TraceId: GetTraceId(c.Request.Context()),
		// 设置状态码
		Code:    errcode.CodeOK,
		// 设置状态信息
		Msg:     errcode.MsgOK,
		// 设置响应数据
		Data:    v,
	})
}

// Error 错误响应返回
func Error(c *gin.Context, err error) {
	// 获取请求的上下文
	ctx := c.Request.Context()
	// 解析错误码
	e := errcode.ParseErr(err)
	// 判断错误码是否为非预期错误或自定义错误
	if e == errcode.ErrUnexpected || e == errcode.ErrCustom {
		// 记录错误日志
		xzap.WithContext(ctx).Error("request handle err",
			zap.Error(err),
			zap.Uint32("code", e.Code()),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery))
	}

	// 写入HTTP头信息
	WriteHeader(c.Writer, e)
	// 返回JSON响应
	c.JSON(e.HTTPCode(), &Response{
		TraceId: GetTraceId(ctx),
		Code:    e.Code(),
		Msg:     e.Error(),
		Data:    nil,
	})
}

// CustomError :custom error http code
func CustomError(c *gin.Context, err error, httpCode int) {
	// 获取请求上下文
	ctx := c.Request.Context()

	// 解析错误代码
	e := errcode.ParseErr(err)

	// 如果错误是未预期的错误或自定义错误
	if e == errcode.ErrUnexpected || e == errcode.ErrCustom {
		// 使用zap记录错误日志
		xzap.WithContext(ctx).Error("Request handle custom err",
			zap.Error(err),
			zap.Uint32("code", e.Code()),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery))
	}

	// 设置HTTP响应头
	WriteHeader(c.Writer, e)

	// 返回JSON响应
	c.JSON(httpCode, &Response{
		TraceId: GetTraceId(ctx), // 获取跟踪ID
		Code:    e.Code(),         // 错误代码
		Msg:     e.Error(),        // 错误信息
		Data:    nil,              // 数据内容
	})
}

// Parse 请求体解析
func Parse(r *http.Request, v interface{}) error {
	// if err := httpx.Parse(r, v); err != nil {
	// 	xzap.WithContext(r.Context()).Errorf("request parse err, err: %s", formatStr(err.Error(), halfShowLen))
	// 	return errcode.ErrInvalidParams
	// }

	if err := validator.Verify(v); err != nil {
		return errcode.NewCustomErr(err.Error())
	}

	return nil
}

// ParseForm 请求表单解析
func ParseForm(r *http.Request, v interface{}) error {
	// if err := httpx.ParseForm(r, v); err != nil {
	// 	xzap.WithContext(r.Context()).Errorf("request parse form err, err: %s",
	// 		formatStr(err.Error(), halfShowLen))
	// 	return errcode.ErrInvalidParams
	// }

	if err := validator.Verify(v); err != nil {
		return errcode.NewCustomErr(err.Error())
	}

	return nil
}

// FromFile 请求表单文件获取
func FromFile(r *http.Request, name string, size int64) (*multipart.FileHeader, error) {
	// 如果请求中未解析MultipartForm
	if r.MultipartForm == nil {
		// 解析MultipartForm
		if err := r.ParseMultipartForm(size); err != nil {
			// 如果解析失败，返回错误
			return nil, err
		}
	}

	// 从请求中获取文件
	f, fh, err := r.FormFile(name)
	// 如果获取文件时出错
	if err != nil {
		// 如果错误是因为缺少文件
		if err == http.ErrMissingFile {
			// 返回参数错误
			return nil, errcode.ErrInvalidParams
		}
		// 返回错误
		return nil, err
	}
	// 关闭文件
	f.Close()
	// 返回文件头信息和nil错误
	return fh, nil
}

// Query 返回给定请求查询参数键的字符串值
func Query(r *http.Request, key string) string {
	value, _ := GetQuery(r, key)
	return value
}

// GetQuery 返回给定请求查询参数键的字符串值并判断其是否存在
func GetQuery(r *http.Request, key string) (string, bool) {
	if values, ok := GetQueryArray(r, key); ok {
		return values[0], ok
	}
	return "", false
}

// QueryArray 返回给定请求查询参数键的字符串切片值
func QueryArray(r *http.Request, key string) []string {
	values, _ := GetQueryArray(r, key)
	return values
}

// GetQueryArray 返回给定请求查询参数键的字符串切片值并判断其是否存在
func GetQueryArray(r *http.Request, key string) ([]string, bool) {
	query := r.URL.Query()
	if values, ok := query[key]; ok && len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

// GetClientIP 获取客户端的IP
func GetClientIP(r *http.Request) string {
	// 获取X-Forwarded-For请求头中的IP地址，并去除空白字符
	ip := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0])
	if ip != "" {
		// 如果获取到IP地址，则返回该IP地址
		return ip
	}

	// 获取X-Real-Ip请求头中的IP地址，并去除空白字符
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		// 如果获取到IP地址，则返回该IP地址
		return ip
	}

	// 获取X-Appengine-Remote-Addr请求头中的IP地址
	if addr := r.Header.Get("X-Appengine-Remote-Addr"); addr != "" {
		// 如果获取到IP地址，则返回该IP地址
		return addr
	}

	// 获取RemoteAddr请求头中的IP地址，并去除空白字符
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		// 如果获取到IP地址，则返回该IP地址
		return ip
	}

	// 如果以上方法都没有获取到IP地址，则返回空字符串
	return ""
}

// GetExternalIP 通过API获取服务端的外部IP
func GetExternalIP() (string, error) {
	// 设置要访问的API地址
	api := "http://pv.sohu.com/cityjson?ie=utf-8"

	// 发起HTTP GET请求
	resp, err := http.Get(api)
	if err != nil {
		// 如果请求失败，返回错误信息
		return "", errors.WithMessagef(err, "http get api = %v err", api)
	}
	// 确保响应体在函数结束时关闭
	defer resp.Body.Close()

	// 读取响应体的全部内容
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// 如果读取失败，返回错误信息
		return "", errors.WithMessage(err, "read all response body err")
	}
	// 将读取到的字节数据转换为字符串
	s := string(b)

	// 查找字符串中cip字段的位置
	i := strings.Index(s, `"cip": "`)
	// 截取cip字段的值
	s = s[i+len(`"cip": "`):]
	// 查找cip字段值末尾的位置
	i = strings.Index(s, `"`)
	// 截取完整的cip字段值
	s = s[:i]

	// 返回获取到的外部IP地址
	return s, nil
}

// GetInternalIP 获取服务端的内部IP
func GetInternalIP() string {
	infs, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, inf := range infs {
		if isEthDown(inf.Flags) || isLoopback(inf.Flags) {
			continue
		}

		addrs, err := inf.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}

	return ""
}

func isEthDown(f net.Flags) bool {
	return f&net.FlagUp != net.FlagUp
}

func isLoopback(f net.Flags) bool {
	return f&net.FlagLoopback == net.FlagLoopback
}

func formatStr(s string, halfShowLen int) string {
	if length := len(s); length > halfShowLen*2 {
		return s[:halfShowLen] + " ...... " + s[length-halfShowLen-1:]
	}

	return s
}

// CopyHttpRequest 复制请求体
func CopyHttpRequest(r *http.Request) (*http.Request, error) {
	rClone := r.Clone(context.Background())
	// 克隆请求体
	if r.Body != nil {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		r.Body = ioutil.NopCloser(bytes.NewReader(body))
		rClone.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	return rClone, nil
}
