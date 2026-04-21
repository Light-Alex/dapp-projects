package middleware

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/ProjectsTask/EasySwapBase/errcode"
	"github.com/ProjectsTask/EasySwapBase/stores/xkv"
	"github.com/ProjectsTask/EasySwapBase/xhttp"
)

const CR_LOGIN_MSG_KEY string = "cache:es:login:msg"
const CR_LOGIN_KEY string = "cache:es:login:address:data"
const CR_LOGIN_SALT string = "es_login_salt&$%"

// 设置路由cookie
// AuthMiddleWare 是一个认证中间件函数，用于验证请求中的会话令牌。
// 它接收一个 xkv.Store 类型的上下文对象 ctx，用于与缓存交互。
// 该函数返回一个 gin.HandlerFunc 类型的处理函数，可用于 Gin 框架的路由中。
//
// 主要功能包括:
// 1. 从请求头获取 session_id，如果为空则跳过验证。
// 2. 支持多个 session_id，用逗号分隔。
// 3. 对每个 session_id 进行以下验证:
//   - 解码 session_id，如果失败则返回令牌验证错误。
//   - 解密解码后的 session_id，如果失败则返回令牌过期错误。
//   - 从缓存中读取解密后的 session_id 对应的数据，如果失败则返回令牌过期错误。
//
// 4. 如果验证失败则返回相应错误:
//   - 令牌格式错误返回 ErrTokenVerify
//   - 令牌过期返回 ErrTokenExpire
//
// 5. 验证通过则继续处理请求。
func AuthMiddleWare(ctx *xkv.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 session_id，格式为 cache:es:login:address:data:<用户地址>
		values := c.Request.Header.Get("session_id")
		// 如果 session_id 为空，跳过验证，继续处理请求
		if values == "" {
			c.Next()
			return
		}

		// 将多个 session_id 按逗号分隔成切片
		sessionIDs := strings.Split(values, ",")
		// 遍历每个 session_id 进行验证
		for _, sessionID := range sessionIDs {
			// 将 session_id 从十六进制字符串解码为字节切片
			encryptCode, err := hex.DecodeString(sessionID)
			// 如果解码失败，返回令牌验证错误，并终止请求处理
			if err != nil {
				xhttp.Error(c, errcode.ErrTokenVerify)
				c.Abort()
				return
			}

			// 对解码后的 session_id 进行解密操作
			decrptCode, err := AesDecryptOFB(encryptCode, []byte(CR_LOGIN_SALT))
			// 如果解密失败，返回令牌过期错误，并终止请求处理
			if err != nil {
				xhttp.Error(c, errcode.ErrTokenExpire)
				c.Abort()
				return
			}
			// 从缓存中读取解密后的 session_id 对应的数据
			result, err := ctx.Get(string(decrptCode))
			// 如果读取失败或数据为空，返回令牌过期错误，并终止请求处理
			if result == "" || err != nil {
				xhttp.Error(c, errcode.ErrTokenExpire)
				c.Abort()
				return
			}
		}

		// 所有 session_id 验证通过，继续处理请求
		c.Next()
	}
}

// GetAuthUserAddress 从请求头的 session_id 中获取认证用户的地址。
// 该函数接收一个 gin.Context 类型的上下文对象 c 和一个 xkv.Store 类型的上下文对象 ctx，用于获取请求头信息和与缓存交互。
// 函数返回一个包含用户地址的字符串切片和一个错误对象。
//
// 主要功能包括:
// 1. 从请求头获取 session_id，如果为空则返回错误。
// 2. 支持多个 session_id，用逗号分隔。
// 3. 对每个 session_id 进行以下操作:
//   - 解码 session_id，如果失败则返回解码错误。
//   - 解密解码后的 session_id，如果失败则返回无效 cookie 错误。
//   - 从缓存中读取解密后的 session_id 对应的数据，如果失败则返回读取缓存错误。
//   - 解析数据获取用户地址，如果格式错误或地址为空则返回相应错误。
//
// 4. 将所有有效的用户地址添加到切片中并返回。
func GetAuthUserAddress(c *gin.Context, ctx *xkv.Store) ([]string, error) {
	// 从请求头中获取 session_id
	values := c.Request.Header.Get("session_id")
	// 如果 session_id 为空，返回获取令牌失败的错误
	if values == "" {
		return nil, errors.New("failed on get token")
	}

	// 将多个 session_id 按逗号分隔成切片
	sessionIDs := strings.Split(values, ",")
	// 用于存储用户地址的切片
	var addrs []string
	// 遍历每个 session_id 进行处理
	for _, sessionID := range sessionIDs {
		// 将 session_id 从十六进制字符串解码为字节切片
		encryptCode, err := hex.DecodeString(sessionID)
		// 如果解码失败，返回解码 cookie 失败的错误
		if err != nil {
			return nil, errors.Wrap(err, "failed on decode cookie")
		}

		// 对解码后的 session_id 进行解密操作
		decrptCode, err := AesDecryptOFB(encryptCode, []byte(CR_LOGIN_SALT))
		// 如果解密失败，返回无效 cookie 的错误
		if err != nil {
			return nil, errors.Wrap(err, "invalid cookie")
		}
		// 从缓存中读取解密后的 session_id 对应的数据
		result, err := ctx.Get(string(decrptCode))
		// 如果读取失败或数据为空，返回从缓存读取 cookie 失败的错误
		if result == "" || err != nil {
			return nil, errors.Wrap(err, "failed on read cookie from cache")
		}
		// 按特定格式分割解密后的字符串，获取用户地址部分
		arr := strings.Split(string(decrptCode), CR_LOGIN_KEY+":")
		// 如果分割结果不符合预期，返回用户缓存信息格式错误
		if len(arr) != 2 {
			return nil, errors.New("user cache info format err")
		}

		// 如果用户地址为空，返回无效用户地址错误
		if arr[1] == "" {
			return nil, errors.New("invalid user address")
		}
		// 将有效的用户地址添加到切片中
		addrs = append(addrs, arr[1])
	}

	// 返回包含所有有效用户地址的切片和 nil 错误
	return addrs, nil
}

// AesDecryptOFB 使用OFB（Output Feedback）模式对AES加密的数据进行解密。
// 此函数接收加密数据和密钥作为输入，返回解密后的数据和可能出现的错误。
//
// 参数:
//   - data: 要解密的加密数据，格式为字节切片。
//   - key: 用于解密的AES密钥，格式为字节切片。
//
// 返回值:
//   - 解密后的数据，格式为字节切片。
//   - 若解密过程中出现错误，返回包含错误信息的错误对象；若解密成功，返回 nil。
func AesDecryptOFB(data []byte, key []byte) ([]byte, error) {
	// 创建一个新的AES加密块
	block, _ := aes.NewCipher([]byte(key))
	// 从加密数据中提取初始化向量（IV）
	iv := data[:aes.BlockSize]
	// 去除加密数据中的IV部分
	data = data[aes.BlockSize:]
	// 检查数据长度是否为块大小的倍数
	if len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("data is not a multiple of the block size")
	}

	// 创建一个字节切片用于存储解密后的数据
	out := make([]byte, len(data))
	// 使用OFB模式创建一个流加密器
	mode := cipher.NewOFB(block, iv)
	// 对数据进行解密
	mode.XORKeyStream(out, data)

	// 去除填充字节
	out = PKCS7UnPadding(out)
	// 返回解密后的数据和 nil 错误
	return out, nil
}

// 去码
// PKCS7UnPadding 用于去除 PKCS#7 填充的数据。
// 在 PKCS#7 填充方案中，填充字节的值等于填充字节的数量。
// 该函数接收一个经过 PKCS#7 填充的字节切片，返回去除填充后的原始数据。
//
// 参数:
//   - origData: 经过 PKCS#7 填充的字节切片。
//
// 返回值:
//   - 返回去除填充后的原始数据字节切片。
func PKCS7UnPadding(origData []byte) []byte {
	// 获取填充数据的总长度
	length := len(origData)
	// 获取填充字节的值，该值表示填充字节的数量
	unpadding := int(origData[length-1])
	// 截取去除填充后的原始数据
	return origData[:(length - unpadding)]
}
