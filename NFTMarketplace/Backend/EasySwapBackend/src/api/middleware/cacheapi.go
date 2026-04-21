package middleware

import (
	"bytes"
	"crypto/sha512"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ProjectsTask/EasySwapBase/errcode"
	"github.com/ProjectsTask/EasySwapBase/stores/xkv"
	"github.com/ProjectsTask/EasySwapBase/xhttp"
)

const CacheApiPrefix = "apicache:"

type responseCache struct {
	Status int
	Header http.Header
	Data   []byte
}

// CacheApi 是一个缓存中间件函数,用于缓存API响应数据
// 主要功能包括:
// 1. 接收一个 xkv.Store 存储实例和过期时间作为参数
// 2. 检查请求是否有缓存,如果有且状态码为200则直接返回缓存数据
// 3. 如果没有缓存,则继续处理请求
// 4. 请求处理完成后,如果响应状态码为200,则将响应数据缓存起来
func CacheApi(store *xkv.Store, expireSeconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data xhttp.Response
		// 创建响应体写入器用于获取响应内容
		bodyLogWriter := &BodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bodyLogWriter

		// 生成缓存key
		cacheKey := CreateKey(c)
		if cacheKey == "" {
			xhttp.Error(c, errcode.NewCustomErr("cache error:no cache"))
			c.Abort()
		}

		// 尝试获取缓存数据
		cacheData, err := (*store).Get(cacheKey)
		if err == nil && cacheData != "" {
			cache := unserialize(cacheData)
			if cache != nil {
				// 如果有缓存,则直接返回缓存的响应
				bodyLogWriter.ResponseWriter.WriteHeader(cache.Status)
				for k, vals := range cache.Header {
					for _, v := range vals {
						bodyLogWriter.ResponseWriter.Header().Set(k, v)
					}
				}

				if err := json.Unmarshal(cache.Data, &data); err == nil {
					if data.Code == http.StatusOK {
						bodyLogWriter.ResponseWriter.Write(cache.Data)
						c.Abort()
					}
				}
			}
		}

		// 继续处理请求
		c.Next()

		// 获取响应数据
		responseBody := bodyLogWriter.body.Bytes()

		// 如果响应状态码为200,则缓存响应数据
		if err := json.Unmarshal(responseBody, &data); err == nil {
			if data.Code == http.StatusOK {
				storeCache := responseCache{
					Header: bodyLogWriter.Header().Clone(),
					Status: bodyLogWriter.ResponseWriter.Status(),
					Data:   responseBody,
				}
				store.SetnxEx(cacheKey, serialize(storeCache), expireSeconds)
			}
		}

	}
}

// CreateKey 生成缓存的key
// 主要功能:
// 1. 将路径、查询参数和请求体组合成缓存key
// 2. 如果key长度超过128,使用SHA512进行哈希
// 3. 添加缓存前缀并返回最终的key
func CreateKey(c *gin.Context) string {
	var buf bytes.Buffer
	// 使用io.TeeReader将请求体内容复制到buf中
	tee := io.TeeReader(c.Request.Body, &buf)
	requestBody, _ := ioutil.ReadAll(tee)
	// 将c.Request.Body重置为读取了请求体内容后的buf
	c.Request.Body = ioutil.NopCloser(&buf)

	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery

	// 组合缓存key
	cacheKey := path + "," + query + string(requestBody)

	// 如果key太长则进行哈希
	if len(cacheKey) > 128 {
		hash := sha512.New() // 512/8*2
		// 对cacheKey进行哈希处理
		hash.Write([]byte(cacheKey))
		cacheKey = string(hash.Sum([]byte("")))
		// 将哈希后的结果转换为16进制字符串
		cacheKey = fmt.Sprintf("%x", cacheKey)
	}

	// 添加缓存前缀
	cacheKey = CacheApiPrefix + cacheKey
	return cacheKey
}

// serialize 将给定的responseCache对象序列化为字符串形式
//
// 参数:
//     cache - 需要序列化的responseCache对象
//
// 返回值:
//     如果序列化成功，返回序列化后的字符串；如果序列化失败，返回空字符串
func serialize(cache responseCache) string {
	// 创建一个新的bytes.Buffer实例
	buf := new(bytes.Buffer)
	// 创建一个新的gob编码器，将buf作为输出
	enc := gob.NewEncoder(buf)
	// 使用gob编码器将cache对象编码到buf中
	if err := enc.Encode(cache); err != nil {
		// 如果编码过程中发生错误，返回空字符串
		return ""
	} else {
		// 如果编码成功，将buf中的字符串返回
		return buf.String()
	}
}

// unserialize 反序列化给定的字符串数据为responseCache实例
// 参数：
//     data: 待反序列化的字符串数据
// 返回值：
//     *responseCache: 反序列化后的responseCache实例指针，如果反序列化失败则返回nil
func unserialize(data string) *responseCache {
	// 初始化一个空的responseCache实例
	var g1 = responseCache{}
	// 创建一个gob解码器，用于解码字节数据
	dec := gob.NewDecoder(bytes.NewBuffer([]byte(data)))
	// 使用gob解码器解码data数据到g1实例中
	if err := dec.Decode(&g1); err != nil {
		// 如果解码过程中出现错误，则返回nil
		return nil
	} else {
		// 如果解码成功，则返回指向g1的指针
		return &g1
	}
}
