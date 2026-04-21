package xhttp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/pkg/errors"
)

const (
	// jsonrpcVersion 默认 JSON-RPC 默认版本
	jsonrpcVersion = "2.0"
)

// RPCClient 通用 JSON-RPC 客户端接口
type RPCClient interface {
	// Call 进行 JSON-RPC 调用
	Call(method string, params ...interface{}) (*RPCResponse, error)
	// CallRaw 基于所给请求体进行 JSON-RPC 调用
	CallRaw(request *RPCRequest) (*RPCResponse, error)
	// CallFor 进行 JSON-RPC 调用并将响应结果反序列化到所给类型对象中
	CallFor(out interface{}, method string, params ...interface{}) error
}

// RPCOption JSON-RPC 客户端可选配置
type RPCOption func(server *rpcClient)

// WithHTTPClient 使用配置的 HTTP 客户端
func WithHTTPClient(hc *http.Client) RPCOption {
	// 返回一个匿名函数，该函数接收一个rpcClient类型的指针参数
	return func(c *rpcClient) {
		// 将传入的http.Client对象赋值给rpcClient的httpClient字段
		c.httpClient = hc
	}
}

// WithCustomHeaders 使用配置的 HTTP 请求头
func WithCustomHeaders(m map[string]string) RPCOption {
	// 返回一个闭包函数
	return func(c *rpcClient) {
		// 初始化 customHeaders 为一个新的 map
		c.customHeaders = make(map[string]string)
		// 遍历传入的 map，将键值对复制到 customHeaders 中
		for k, v := range m {
			c.customHeaders[k] = v
		}
	}
}

// NewRPCClient 新建通用 JSON-RPC 客户端
func NewRPCClient(endpoint string, opts ...RPCOption) RPCClient {
	// 初始化 rpcClient 实例，并设置 endpoint
	c := &rpcClient{endpoint: endpoint}

	// 遍历 opts，并调用每个选项函数对 c 进行配置
	for _, opt := range opts {
		opt(c)
	}

	// 如果 c 的 httpClient 字段为空，则为其设置默认 HTTP 客户端
	if c.httpClient == nil {
		c.httpClient = NewDefaultHTTPClient()
	}

	// 返回配置好的 rpcClient 实例
	return c
}

// rpcClient 默认 JSON-RPC 客户端
type rpcClient struct {
	endpoint      string
	httpClient    *http.Client
	customHeaders map[string]string
}

// newRequest 新建 HTTP 请求体
func (c *rpcClient) newRequest(req interface{}) (*http.Request, error) {
	// 将请求体转换为JSON格式的字节数组
	body, err := json.Marshal(req)
	if err != nil {
		// 如果转换失败，返回错误
		return nil, errors.WithMessagef(err, "json marshal %v err", req)
	}
	// fmt.Println(string(body))

	// 创建一个新的HTTP请求
	request, err := http.NewRequest(http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		// 如果创建请求失败，返回错误
		return nil, errors.WithMessage(err, "new http request err")
	}

	// 设置请求头，接受的内容类型为JSON
	request.Header.Set(HeaderAccept, ApplicationJSON)
	// 设置请求头，发送的内容类型为JSON
	request.Header.Set(HeaderContentType, ApplicationJSON)

	// 遍历自定义的HTTP头部，并设置到请求头中
	for k, v := range c.customHeaders {
		request.Header.Set(k, v)
	}

	return request, nil
}

// doCall 执行 JSON-RPC 调用
func (c *rpcClient) doCall(req *RPCRequest) (*RPCResponse, error) {
	// 创建HTTP请求
	httpReq, err := c.newRequest(req)
	if err != nil {
		// 如果创建请求失败，则返回错误信息
		return nil, errors.WithMessagef(err, "call %s method on %s err",
			req.Method, c.endpoint)
	}

	// 发送HTTP请求
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		// 如果发送请求失败，则返回错误信息
		return nil, errors.WithMessagef(err, "call %s method on %s err",
			req.Method, httpReq.URL.String())
	}
	defer httpResp.Body.Close() // 确保在函数结束时关闭HTTP响应体

	// 创建一个JSON解码器
	d := json.NewDecoder(httpResp.Body)
	d.DisallowUnknownFields() // 禁止解析未知字段
	d.UseNumber() // 使用json.Number代替float64进行解码

	// 定义RPC响应结构体指针
	var rpcResp *RPCResponse
	// 解码HTTP响应体到RPC响应结构体
	err = d.Decode(&rpcResp)
	if err != nil {
		// 如果解码失败，则返回错误信息
		return nil, errors.WithMessagef(err, "call %s method on %s status code: %d, decode body err",
			req.Method, httpReq.URL.String(), httpResp.StatusCode)
	}
	if rpcResp == nil {
		// 如果RPC响应为空，则返回错误信息
		return nil, errors.WithMessagef(err, "call %s method on %s status code: %d, rpc response missing err.",
			req.Method, httpReq.URL.String(), httpResp.StatusCode)
	}

	// 返回RPC响应
	return rpcResp, nil
}

// Call 进行 JSON-RPC 调用
func (c *rpcClient) Call(method string, params ...interface{}) (*RPCResponse, error) {
	req := &RPCRequest{
		Method:  method,
		Params:  Params(params...),
		JSONRPC: jsonrpcVersion,
	}

	return c.doCall(req)
}

// CallRaw 基于所给请求体进行 JSON-RPC 调用
func (c *rpcClient) CallRaw(request *RPCRequest) (*RPCResponse, error) {
	return c.doCall(request)
}

// CallFor 进行 JSON-RPC 调用并将响应结果反序列化到所给类型对象中
func (c *rpcClient) CallFor(out interface{}, method string, params ...interface{}) error {
	rpcResp, err := c.Call(method, params...)
	if err != nil {
		return err
	}

	return rpcResp.ReadToObject(out)
}

// RPCRequest 通用 JSON-RPC 请求体
type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// NewRPCRequest 新建通用 JSON-RPC 请求体
func NewRPCRequest(method string, params ...interface{}) *RPCRequest {
	req := &RPCRequest{
		Method:  method,
		Params:  Params(params...),
		JSONRPC: jsonrpcVersion,
	}

	return req
}

// Params 构建请求参数
func Params(params ...interface{}) interface{} {
	var ps interface{}

	if params != nil {
		// 判断params的长度
		switch len(params) {
		case 0:
		case 1:
			// 当params的长度为1时
			if param := params[0]; param != nil {
				// 获取参数的类型
				typeOf := reflect.TypeOf(param)
				// 如果param的类型是指针类型，则逐层解引用，直到找到非指针类型
				for typeOf != nil && typeOf.Kind() == reflect.Ptr {
					typeOf = typeOf.Elem()
				}

				// array、slice、interface 和 map 不改变其参数方式，其余类型都包装在数组中
				if typeOf != nil {
					switch typeOf.Kind() {
					case reflect.Array:
						// 如果参数是数组类型，则直接赋值给ps
						ps = param
					case reflect.Slice:
						// 如果参数是切片类型，则直接赋值给ps
						ps = param
					case reflect.Interface:
						// 如果参数是接口类型，则直接赋值给ps
						ps = param
					case reflect.Map:
						// 如果参数是映射类型，则直接赋值给ps
						ps = param
					default:
						// 其他类型将params数组赋值给ps
						ps = params
					}
				}
			} else {
				// 如果param为nil，则将params数组赋值给ps
				ps = params
			}
		default:
			// 当params的长度大于1时，将params数组赋值给ps
			ps = params
		}
	}

	return ps
}

// RPCResponse 通用 JSON-RPC 响应体
type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// GetInt64 获取响应结果的 int64 类型值
func (resp *RPCResponse) GetInt64() (int64, error) {
	if resp.Error != nil {
		return 0, errors.Errorf("%v", resp.Error)
	}

	val, ok := resp.Result.(json.Number)
	if !ok {
		return 0, errors.Errorf("parse int64 from %v err", resp.Result)
	}

	i, err := val.Int64()
	if err != nil {
		return 0, err
	}

	return i, nil
}

// GetFloat64 获取响应结果的 float64 类型值
func (resp *RPCResponse) GetFloat64() (float64, error) {
	if resp.Error != nil {
		return 0, errors.Errorf("%v", resp.Error)
	}

	val, ok := resp.Result.(json.Number)
	if !ok {
		return 0, errors.Errorf("parse float64 from %v err", resp.Result)
	}

	f, err := val.Float64()
	if err != nil {
		return 0, err
	}

	return f, nil
}

// GetBool 获取响应结果的 bool 类型值
func (resp *RPCResponse) GetBool() (bool, error) {
	if resp.Error != nil {
		return false, errors.Errorf("%v", resp.Error)
	}

	val, ok := resp.Result.(bool)
	if !ok {
		return false, errors.Errorf("parse bool from %v err", resp.Result)
	}

	return val, nil
}

// GetString 获取响应结果的 string 类型值
func (resp *RPCResponse) GetString() (string, error) {
	if resp.Error != nil {
		return "", errors.Errorf("%v", resp.Error)
	}

	val, ok := resp.Result.(string)
	if !ok {
		return "", errors.Errorf("parse string from %v err", resp.Result)
	}

	return val, nil
}

// ReadToObject 将响应结果反序列化到所给类型对象中
func (resp *RPCResponse) ReadToObject(to interface{}) error {
	if resp.Error != nil {
		return errors.Errorf("%v", resp.Error)
	}

	from, err := json.Marshal(resp.Result)
	if err != nil {
		return errors.WithMessagef(err, "json marshal %v err", resp.Result)
	}

	err = json.Unmarshal(from, to)
	if err != nil {
		return errors.WithMessagef(err, "json unmarshal %s err", from)
	}

	return nil
}
