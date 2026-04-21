// package chainclient 定义了与区块链交互的客户端相关功能
package chainclient

import (
	"context"
	"math/big"

	// 导入以太坊相关包
	"github.com/ethereum/go-ethereum"
	// 导入错误处理包
	"github.com/pkg/errors"

	// 导入自定义的链相关包
	"github.com/ProjectsTask/EasySwapBase/chain"
	// 导入自定义的 EVM 客户端包
	"github.com/ProjectsTask/EasySwapBase/chain/chainclient/evmclient"
	// 导入自定义的日志类型包
	logTypes "github.com/ProjectsTask/EasySwapBase/chain/types"
)

// ChainClient 定义了与区块链交互的客户端接口
type ChainClient interface {
	// FilterLogs 根据给定的查询条件过滤日志
	// 参数:
	//   - ctx: 上下文，用于控制操作的生命周期
	//   - q: 过滤查询条件
	// 返回值:
	//   - []interface{}: 过滤后的日志列表
	//   - error: 操作过程中可能出现的错误
	FilterLogs(ctx context.Context, q logTypes.FilterQuery) ([]interface{}, error)

	// BlockTimeByNumber 根据区块号获取区块的时间戳
	// 参数:
	//   - context.Context: 上下文，用于控制操作的生命周期
	//   - *big.Int: 区块号
	// 返回值:
	//   - uint64: 区块的时间戳
	//   - error: 操作过程中可能出现的错误
	BlockTimeByNumber(context.Context, *big.Int) (uint64, error)

	// Client 返回底层的区块链客户端实例
	// 返回值:
	//   - interface{}: 区块链客户端实例
	Client() interface{}

	// CallContract 调用智能合约的方法
	// 参数:
	//   - ctx: 上下文，用于控制操作的生命周期
	//   - msg: 调用消息，包含调用的目标合约地址、方法等信息
	//   - blockNumber: 调用时使用的区块号
	// 返回值:
	//   - []byte: 调用结果的字节切片
	//   - error: 操作过程中可能出现的错误
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)

	// CallContractByChain 根据链的参数调用智能合约
	// 参数:
	//   - ctx: 上下文，用于控制操作的生命周期
	//   - param: 调用参数
	// 返回值:
	//   - interface{}: 调用结果
	//   - error: 操作过程中可能出现的错误
	CallContractByChain(ctx context.Context, param logTypes.CallParam) (interface{}, error)

	// BlockNumber 获取当前区块链的最新区块号
	// 返回值:
	//   - uint64: 最新区块号
	//   - error: 操作过程中可能出现的错误
	BlockNumber() (uint64, error)

	// BlockWithTxs 根据区块号获取包含交易信息的区块
	// 参数:
	//   - ctx: 上下文，用于控制操作的生命周期
	//   - blockNumber: 区块号
	// 返回值:
	//   - interface{}: 包含交易信息的区块
	//   - error: 操作过程中可能出现的错误
	BlockWithTxs(ctx context.Context, blockNumber uint64) (interface{}, error)
}

// New 根据链 ID 和节点 URL 创建一个新的 ChainClient 实例
// 参数:
//   - chainID: 链的 ID
//   - nodeUrl: 节点的 URL
//
// 返回值:
//   - ChainClient: 新创建的 ChainClient 实例
//   - error: 操作过程中可能出现的错误
func New(chainID int, nodeUrl string) (ChainClient, error) {
	// 根据链 ID 选择不同的客户端实现
	switch chainID {
	// 支持以太坊、Optimism 和 Sepolia 链
	case chain.EthChainID, chain.OptimismChainID, chain.SepoliaChainID:
		return evmclient.New(nodeUrl)
	default:
		// 不支持的链 ID，返回错误
		return nil, errors.New("unsupported chain id")
	}
}
