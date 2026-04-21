package xkv

import (
	"encoding/json"
	"log"
	"reflect"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/kv"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/ProjectsTask/EasySwapBase/kit/convert"
)

const (
	// getAndDelScript 获取并删除key所关联的值lua脚本
	getAndDelScript = `local current = redis.call('GET', KEYS[1]);
if (current) then
    redis.call('DEL', KEYS[1]);
end
return current;`
)

// Store 键值存取器结构详情
type Store struct {
	kv.Store
	Redis *redis.Redis
}

// NewStore 新建键值存取器
func NewStore(c kv.KvConf) *Store {
	// 如果配置为空或缓存总权重小于等于0
	if len(c) == 0 || cache.TotalWeights(c) <= 0 {
		// 记录日志并退出程序
		log.Fatal("no cache nodes")
	}

	// 根据配置创建Redis客户端
	cn := redis.MustNewRedis(c[0].RedisConf)
	// 返回Store结构体指针
	return &Store{
		// 创建kv.Store实例
		Store: kv.NewStore(c),
		// 将创建的Redis客户端赋值给Redis字段
		Redis: cn,
	}
}

// GetInt 返回给定key所关联的int值
func (s *Store) GetInt(key string) (int, error) {
	// 从Store中获取指定key的值
	value, err := s.Get(key)
	if err != nil {
		// 如果发生错误，返回0和错误
		return 0, err
	}

	// 将获取到的值转换为int类型，并返回
	return convert.ToInt(value), nil
}

// SetInt 将int value关联到给定key，seconds为key的过期时间（秒）
func (s *Store) SetInt(key string, value int, seconds ...int) error {
	return s.SetString(key, convert.ToString(value), seconds...)
}

// GetInt64 返回给定key所关联的int64值
func (s *Store) GetInt64(key string) (int64, error) {
	value, err := s.Get(key)
	if err != nil {
		return 0, err
	}

	return convert.ToInt64(value), nil
}

// SetInt64 将int64 value关联到给定key，seconds为key的过期时间（秒）
func (s *Store) SetInt64(key string, value int64, seconds ...int) error {
	return s.SetString(key, convert.ToString(value), seconds...)
}

// GetBytes 返回给定key所关联的[]byte值
func (s *Store) GetBytes(key string) ([]byte, error) {
	value, err := s.Get(key)
	if err != nil {
		return nil, err
	}

	return []byte(value), nil
}

// GetDel 返回并删除给定key所关联的string值
func (s *Store) GetDel(key string) (string, error) {
	resp, err := s.Eval(getAndDelScript, key)
	if err != nil {
		return "", errors.Wrap(err, "eval script err")
	}

	return convert.ToString(resp), nil
}

// SetString 将string value关联到给定key，seconds为key的过期时间（秒）
func (s *Store) SetString(key, value string, seconds ...int) error {
	if len(seconds) != 0 {
		return errors.Wrapf(s.Setex(key, value, seconds[0]), "setex by seconds = %v err", seconds[0])
	}

	return errors.Wrap(s.Set(key, value), "set err")
}


// Read 从存储中读取指定 key 对应的数据，并将数据反序列化为 obj。
//
// 参数：
//     key: 存储中的键名
//     obj: 用于存储反序列化后的数据的目标对象
//
// 返回值：
//     bool: 如果成功读取并反序列化数据，则返回 true；否则返回 false
//     error: 如果读取或反序列化过程中出现错误，则返回相应的错误信息；否则返回 nil
func (s *Store) Read(key string, obj interface{}) (bool, error) {
	// 检查传入的对象是否有效
	if !isValid(obj) {
		return false, errors.New("obj is invalid")
	}

	// 从存储中获取指定键的字节数据
	value, err := s.GetBytes(key)
	if err != nil {
		return false, errors.Wrap(err, "get bytes err")
	}
	// 如果获取到的字节数据长度为0，则返回错误
	if len(value) == 0 {
		return false, nil
	}

	// 将字节数据反序列化为传入的对象
	err = json.Unmarshal(value, obj)
	if err != nil {
		return false, errors.Wrap(err, "json unmarshal value to obj err")
	}

	// 返回成功标识和nil错误
	return true, nil
}

// Write 将对象obj序列化后关联到给定key，seconds为key的过期时间（秒）
func (s *Store) Write(key string, obj interface{}, seconds ...int) error {
	value, err := json.Marshal(obj)
	if err != nil {
		return errors.Wrap(err, "json marshal obj err")
	}

	return s.SetString(key, string(value), seconds...)
}

// GetFunc 给定key不存在时调用的数据获取函数
type GetFunc func() (interface{}, error)

// ReadOrGet 将给定key所关联的值反序列化到obj对象
// 若给定key不存在则调用数据获取函数，调用成功时赋值至obj对象
// 并将其序列化后关联到给定key，seconds为key的过期时间（秒）
func (s *Store) ReadOrGet(key string, obj interface{}, gf GetFunc, seconds ...int) error {
	// 尝试从缓存中读取数据
	isExist, err := s.Read(key, obj)
	if err != nil {
		return errors.Wrap(err, "read obj by err") // 读取对象时发生错误
	}

	// 如果缓存中不存在数据
	if !isExist {
		data, err := gf()
		if err != nil {
			return err // 获取数据时发生错误
		}

		// 检查获取的数据是否有效
		if !isValid(data) {
			return errors.New("get data is invalid") // 获取的数据无效
		}

		// 将获取的数据赋值给传入的obj参数
		ov, dv := reflect.ValueOf(obj).Elem(), reflect.ValueOf(data).Elem()
		if ov.Type() != dv.Type() {
			return errors.New("obj type and get data type are not equal") // obj类型和获取的数据类型不匹配
		}
		ov.Set(dv)

		// 将数据写入缓存
		_ = s.Write(key, data, seconds...)
	}

	return nil
}

// isValid 判断对象是否合法
func isValid(obj interface{}) bool {
	if obj == nil {
		return false
	}

	if reflect.ValueOf(obj).Kind() != reflect.Ptr {
		return false
	}

	return true
}
