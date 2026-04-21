// 定义包名为 v1
package v1

// 定义常量 CursorDelimiter，用于表示游标分隔符
// 在某些需要分割字符串以表示游标的场景中使用
const (
	CursorDelimiter = "_"
)

// 定义一个自定义类型 chainIDMap，它是一个从 int 类型到 string 类型的映射
// 用于存储链 ID 到链名称的映射关系
type chainIDMap map[int]string

// 初始化一个 chainIDMap 类型的变量 chainIDToChain
// 该变量存储了一些常见链 ID 对应的链名称
// 键为链 ID，值为对应的链名称字符串
var chainIDToChain = chainIDMap{
	1:        "eth",      // 链 ID 为 1 对应以太坊主网
	10:       "optimism", // 链 ID 为 10 对应 Optimism 网络
	11155111: "sepolia",  // 链 ID 为 11155111 对应 Sepolia 测试网
}
