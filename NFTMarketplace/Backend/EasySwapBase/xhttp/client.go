package xhttp

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const MB = 1 << (10 * 2)

var ErrBodySizeLimit = errors.New("body size too large")

// Config HTTP客户端相关配置
type Config struct {
	HTTPTimeout           time.Duration // HTTP请求超时时间
	DialTimeout           time.Duration // 拨号超时时间
	DialKeepAlive         time.Duration // 拨号保持连接时间
	MaxIdleConns          int           // 最大空闲连接数
	MaxIdleConnsPerHost   int           // 每个主机最大空闲连接数
	MaxConnsPerHost       int           // 每个主机最大连接数
	IdleConnTimeout       time.Duration // 空闲连接超时时间
	ResponseHeaderTimeout time.Duration // 读取响应头超时时间
	ExpectContinueTimeout time.Duration // 期望继续超时时间
	TLSHandshakeTimeout   time.Duration // TLS握手超时时间
	ForceAttemptHTTP2     bool          // 允许尝试启用HTTP/2
}

// GetDefaultConfig 获取默认HTTP客户端相关配置
func GetDefaultConfig() *Config {
	return &Config{
		HTTPTimeout:           20 * time.Second,
		DialTimeout:           15 * time.Second,
		DialKeepAlive:         30 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		MaxConnsPerHost:       100,
		IdleConnTimeout:       60 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ForceAttemptHTTP2:     true,
	}
}

// NewHTTPClient 新建HTTP客户端
func NewHTTPClient(c *Config) *http.Client {
	if c == nil {
		c = GetDefaultConfig()
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   c.DialTimeout,
			KeepAlive: c.DialKeepAlive,
		}).DialContext,
		MaxIdleConns:          c.MaxIdleConns,
		MaxIdleConnsPerHost:   c.MaxIdleConnsPerHost,
		MaxConnsPerHost:       c.MaxConnsPerHost,
		IdleConnTimeout:       c.IdleConnTimeout,
		ResponseHeaderTimeout: c.ResponseHeaderTimeout,
		ExpectContinueTimeout: c.ExpectContinueTimeout,
		TLSHandshakeTimeout:   c.TLSHandshakeTimeout,
		ForceAttemptHTTP2:     c.ForceAttemptHTTP2,
	}

	client := &http.Client{
		Timeout:   c.HTTPTimeout,
		Transport: tr,
	}

	return client
}

// NewDefaultHTTPClient 新建默认HTTP客户端
func NewDefaultHTTPClient() *http.Client {
	return NewHTTPClient(nil)
}

// Client HTTP拓展客户端结构详情
type Client struct {
	*http.Client
}

// NewClient 新建HTTP拓展客户端
func NewClient(c *Config) *Client {
	return &Client{Client: NewHTTPClient(c)}
}

// NewDefaultClient 新建默认HTTP拓展客户端
func NewDefaultClient() *Client {
	return &Client{Client: NewDefaultHTTPClient()}
}

// NewClientWithHTTPClient 使用HTTP客户端新建HTTP拓展客户端
func NewClientWithHTTPClient(client *http.Client) *Client {
	return &Client{Client: client}
}

// GetRequest 获取HTTP请求
func (c *Client) GetRequest(method, rawurl string, header map[string]string, data io.Reader) (*http.Request, error) {
	// 创建一个新的HTTP请求
	req, err := http.NewRequest(method, rawurl, data)
	if err != nil {
		// 如果创建请求失败，则返回错误
		return nil, errors.WithMessagef(err, "new http request err, method = %v, rawurl = %v, header = %v",
			method, rawurl, header)
	}

	// 遍历header，将每个键值对添加到请求头中
	for k, v := range header {
		req.Header.Add(k, v)
	}

	return req, nil
}

// GetResponse 获取HTTP响应及其响应体内容
func (c *Client) GetResponse(req *http.Request) (*http.Response, []byte, error) {
	// 发送HTTP请求并获取响应
	response, err := c.Do(req)
	if err != nil {
		// 如果发送请求失败，则返回错误信息
		return nil, nil, errors.WithMessage(err, "http client do request err")
	}
	defer response.Body.Close()  // 确保响应体被关闭

	// 读取响应体的全部内容
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// 如果读取响应体失败，则返回错误信息
		return nil, nil, errors.WithMessage(err, "read all response body err")
	}

	// 返回响应、响应体和错误信息（为nil）
	return response, body, nil
}

// GetResponse 获取HTTP响应及其响应体内容
func (c *Client) GetResponseWithSizeLimit(req *http.Request, sizeLimit int64) (*http.Response, []byte, error) {
	// 发送HTTP请求
	response, err := c.Do(req)
	if err != nil {
		// 如果请求失败，返回错误信息
		return nil, nil, errors.WithMessage(err, "http client do request err")
	}
	defer response.Body.Close()

	// 检查响应体大小是否超过限制
	if sizeLimit != 0 && response.ContentLength > sizeLimit {
		// 如果超过限制，返回错误信息
		return nil, nil, errors.Wrapf(ErrBodySizeLimit, "length(MB):%d", response.ContentLength/MB)
	}
	// 读取响应体
	body, err := io.ReadAll(response.Body)
	if err != nil {
		// 如果读取响应体失败，返回错误信息
		return nil, nil, errors.WithMessage(err, "read all response body err")
	}

	// 返回响应对象和响应体
	return response, body, nil
}

// CallWithRequest 利用HTTP请求进行HTTP调用
func (c *Client) CallWithRequest(req *http.Request, resp interface{}) (*http.Response, error) {
	// 发送请求并获取响应
	response, body, err := c.GetResponse(req)
	if err != nil {
		// 如果发生错误，返回错误信息和错误信息
		return nil, errors.WithMessage(err, "get response err")
	}

	// 如果响应体不为空
	if len(body) > 0 {
		// 将响应体解析为JSON格式并存储到resp中
		err = json.Unmarshal(body, resp)
		if err != nil {
			// 如果解析JSON时发生错误，返回错误信息和错误信息
			return nil, errors.Wrap(err, fmt.Sprintf("json unmarshal response body err.content:%s", body))
		}
	}

	// 返回响应
	return response, nil
}

// Call HTTP调用
func (c *Client) Call(method, rawurl string, header map[string]string, data io.Reader, resp interface{}) error {
	// 创建请求
	req, err := c.GetRequest(method, rawurl, header, data)
	if err != nil {
		// 如果请求创建失败，返回错误信息
		return errors.WithMessage(err, "get request err")
	}

	// 使用请求调用 CallWithRequest 方法
	_, err = c.CallWithRequest(req, resp)
	if err != nil {
		// 如果调用失败，返回错误信息
		return errors.WithMessage(err, "call with request err")
	}

	return nil
}
