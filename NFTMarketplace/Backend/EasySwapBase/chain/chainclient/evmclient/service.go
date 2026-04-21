// 包 evmclient 提供了与以太坊虚拟机（EVM）兼容的区块链客户端服务。
package evmclient

import (
	"context"
	"math/big"

	// 导入以太坊官方 Go 客户端库
	"github.com/ethereum/go-ethereum"
	// 导入以太坊通用工具包
	"github.com/ethereum/go-ethereum/common"
	// 导入以太坊客户端包
	"github.com/ethereum/go-ethereum/ethclient"
	// 导入错误处理包
	"github.com/pkg/errors"

	// 导入自定义的日志类型包
	logTypes "github.com/ProjectsTask/EasySwapBase/chain/types"
)

// Service 结构体表示一个 EVM 客户端服务实例。
type Service struct {
	// client 是以太坊客户端实例，用于与 EVM 兼容的区块链进行交互。
	client *ethclient.Client
}

// New 创建一个新的 EVM 客户端服务实例。
// 参数 nodeUrl 是以太坊节点的 URL。
// 返回一个指向 Service 结构体的指针和可能出现的错误。
func New(nodeUrl string) (*Service, error) {
	// 尝试连接到指定的以太坊节点
	client, err := ethclient.Dial(nodeUrl)
	if err != nil {
		// 如果连接失败，返回错误信息
		return nil, errors.Wrap(err, "failed on create client")
	}

	// 返回一个新的 Service 实例
	return &Service{
		client: client,
	}, nil
}

// Client 返回底层的以太坊客户端实例。
// 返回值为一个空接口，允许调用者将其转换为所需的类型。
func (s *Service) Client() interface{} {
	return s.client
}

// FilterLogs 根据给定的查询条件过滤日志。
// 参数 ctx 是上下文，用于控制操作的生命周期。
// 参数 q 是过滤查询条件，包含地址、主题和块范围等信息。
// 返回一个包含过滤后日志的切片和可能出现的错误。
func (s *Service) FilterLogs(ctx context.Context, q logTypes.FilterQuery) ([]interface{}, error) {
	// 将查询中的地址字符串转换为以太坊地址类型
	var addresses []common.Address
	for _, addr := range q.Addresses {
		addresses = append(addresses, common.HexToAddress(addr))
	}

	// 将查询中的主题字符串转换为以太坊哈希类型
	var topicsHash [][]common.Hash
	for _, topics := range q.Topics {
		var topicHash []common.Hash
		for _, topic := range topics {
			topicHash = append(topicHash, common.HexToHash(topic))
		}
		topicsHash = append(topicsHash, topicHash)
	}

	// 构建以太坊过滤查询参数
	queryParam := ethereum.FilterQuery{
		FromBlock: q.FromBlock,
		ToBlock:   q.ToBlock,
		Addresses: addresses,
		Topics:    topicsHash,
	}

	// 调用以太坊客户端的 FilterLogs 方法过滤日志
	logs, err := s.client.FilterLogs(ctx, queryParam)
	if err != nil {
		// 如果过滤失败，返回错误信息
		return nil, errors.Wrap(err, "failed on get events")
	}

	// 将过滤后的日志转换为接口切片
	var logEvents []interface{}
	for _, log := range logs {
		logEvents = append(logEvents, log)
	}

	// 返回过滤后的日志和 nil 错误
	return logEvents, nil
}

// BlockTimeByNumber 根据给定的块号获取块的时间戳。
// 参数 ctx 是上下文，用于控制操作的生命周期。
// 参数 blockNum 是要查询的块号。
// 返回块的时间戳和可能出现的错误。
func (s *Service) BlockTimeByNumber(ctx context.Context, blockNum *big.Int) (uint64, error) {
	// 调用以太坊客户端的 HeaderByNumber 方法获取块头
	header, err := s.client.HeaderByNumber(ctx, blockNum)
	if err != nil {
		// 如果获取块头失败，返回错误信息
		return 0, errors.Wrap(err, "failed on get block header")
	}

	// 返回块头中的时间戳
	return header.Time, nil
}

// CallContractByChain 根据给定的参数调用智能合约。
// 参数 ctx 是上下文，用于控制操作的生命周期。
// 参数 param 是调用参数，包含 EVM 参数和块号等信息。
// 返回调用结果和可能出现的错误。
func (s *Service) CallContractByChain(ctx context.Context, param logTypes.CallParam) (interface{}, error) {
	// 调用 CallContract 方法执行合约调用
	return s.CallContract(ctx, param.EVMParam, param.BlockNumber)
}

// CallContract 根据给定的消息和块号调用智能合约。
// 参数 ctx 是上下文，用于控制操作的生命周期。
// 参数 msg 是以太坊调用消息，包含调用的目标合约地址、方法等信息。
// 参数 blockNumber 是调用时使用的块号。
// 返回调用结果的字节切片和可能出现的错误。
func (s *Service) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	// 调用以太坊客户端的 CallContract 方法执行合约调用
	return s.client.CallContract(ctx, msg, blockNumber)
}

// BlockNumber 获取当前区块链的最新块号。
// 返回最新块号和可能出现的错误。
func (s *Service) BlockNumber() (uint64, error) {
	var err error
	// 调用以太坊客户端的 BlockNumber 方法获取最新块号
	blockNum, err := s.client.BlockNumber(context.Background())
	if err != nil {
		// 如果获取块号失败，返回错误信息
		return 0, errors.Wrap(err, "failed on get evm block number")
	}

	// 返回最新块号和 nil 错误
	return blockNum, nil
}

// BlockWithTxs 根据给定的块号获取包含交易的块信息。
// 参数 ctx 是上下文，用于控制操作的生命周期。
// 参数 blockNumber 是要查询的块号。
// 返回包含交易的块信息和可能出现的错误。
func (s *Service) BlockWithTxs(ctx context.Context, blockNumber uint64) (interface{}, error) {
	// 调用以太坊客户端的 BlockByNumber 方法获取包含交易的块信息
	blockWithTxs, err := s.client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		// 如果获取块信息失败，返回错误信息
		return nil, errors.Wrap(err, "failed on get evm block")
	}
	// 返回包含交易的块信息和 nil 错误
	return blockWithTxs, nil
}
