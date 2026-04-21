package nftchainservice

import (
	"context"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	evmTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"github.com/ProjectsTask/EasySwapBase/chain"
	logTypes "github.com/ProjectsTask/EasySwapBase/chain/types"
)

const hex = 16

var EVMTransferTopic = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
var TokenIdExp = new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil)

var BlockTimeGap = map[string]int{
	chain.Eth:      12,
	chain.Optimism: 2,
	chain.Sepolia:  12,
}

type TransferLog struct {
	Address         string        `json:"address" gencodec:"required"`
	TransactionHash string        `json:"transactionHash" gencodec:"required"`
	BlockNumber     uint64        `json:"blockNumber"`
	BlockTime       uint64        `json:"blockTime"`
	BlockHash       string        `json:"blockHash"`
	Data            []byte        `json:"data" gencodec:"required"`
	Topics          []common.Hash `json:"topics" gencodec:"required"`
	Topic0          string        `json:"topic0"`
	From            string        `json:"topic1"`
	To              string        `json:"topic2"`
	TokenID         string        `json:"topic3"`
	TxIndex         uint          `json:"transactionIndex"`
	Index           uint          `json:"logIndex"`
	Removed         bool          `json:"removed"`
}

// GetNFTTransferEvent 方法用于获取指定区块范围内的 NFT 转移事件日志。
// 参数 fromBlock 表示起始区块编号，toBlock 表示结束区块编号。
// 返回值为 TransferLog 结构体指针切片和错误信息。
func (s *Service) GetNFTTransferEvent(fromBlock, toBlock uint64) ([]*TransferLog, error) {
	// 获取起始区块的时间戳
	var startBlockTime uint64
	var err error
	// 根据不同的链名称获取起始区块的时间戳
	switch s.ChainName {
	case chain.Eth, chain.Optimism, chain.Sepolia:
		// 调用 NodeClient 的 BlockTimeByNumber 方法获取起始区块的时间戳
		blockTimestamp, err := s.NodeClient.BlockTimeByNumber(context.Background(), big.NewInt(int64(fromBlock)))
		if err != nil {
			// 如果获取失败，返回错误信息
			return nil, errors.Wrap(err, "failed on get block time")
		}
		startBlockTime = blockTimestamp
	}

	// 定义转移事件的主题
	var transferTopic string
	// 根据不同的链名称设置转移事件的主题
	switch s.ChainName {
	case chain.Eth, chain.Optimism, chain.Sepolia:
		// 将 EVMTransferTopic 转换为字符串
		transferTopic = EVMTransferTopic.String()
	default:
		// 如果链名称不支持，返回错误信息
		return nil, errors.Wrap(err, "unsupported chain")
	}

	// 定义日志过滤查询条件
	logFilter := logTypes.FilterQuery{
		// 设置起始区块编号
		FromBlock: new(big.Int).SetUint64(fromBlock),
		// 设置结束区块编号
		ToBlock: new(big.Int).SetUint64(toBlock),
		// 设置主题过滤条件
		Topics: [][]string{
			{transferTopic},
		},
	}

	// 使用 NodeClient 的 FilterLogs 方法过滤日志
	logs, err := s.NodeClient.FilterLogs(context.Background(), logFilter)
	if err != nil {
		// 如果过滤失败，返回错误信息
		return nil, errors.Wrap(err, "failed on filter logs")
	}

	// 定义转移日志切片
	var transferLogs []*TransferLog
	// 遍历过滤后的日志
	for _, log := range logs {
		var evmLog evmTypes.Log
		var ok bool
		// 将日志转换为 evmTypes.Log 类型
		evmLog, ok = log.(evmTypes.Log)
		if ok {
			var topics [4]string
			// 将 evmLog 的主题转换为十六进制字符串
			for i := range evmLog.Topics {
				topics[i] = evmLog.Topics[i].Hex()
			}
			// 如果主题 3 为空，跳过当前日志
			if topics[3] == "" {
				continue
			}

			// 将主题 3 的字节数组转换为 big.Int 类型的 tokenId
			tokenId := new(big.Int).SetBytes(evmLog.Topics[3][:])
			// 创建 TransferLog 结构体实例
			transferLog := &TransferLog{
				// 设置合约地址
				Address: evmLog.Address.String(),
				// 设置交易哈希
				TransactionHash: evmLog.TxHash.String(),
				// 设置区块编号
				BlockNumber: evmLog.BlockNumber,
				// 设置区块时间，根据起始区块时间和区块间隔计算
				BlockTime: startBlockTime + (evmLog.BlockNumber-fromBlock)*uint64(BlockTimeGap[s.ChainName]),
				// 设置区块哈希
				BlockHash: evmLog.BlockHash.String(),
				// 设置日志数据
				Data: evmLog.Data,
				// 设置日志主题
				Topics: evmLog.Topics,
				// 设置主题 0
				Topic0: topics[0],
				// 设置转移的发送方地址
				From: common.HexToAddress(topics[1]).String(),
				// 设置转移的接收方地址
				To: common.HexToAddress(topics[2]).String(),
				// 设置 token ID
				TokenID: tokenId.String(),
				// 设置交易索引
				TxIndex: evmLog.TxIndex,
				// 设置日志索引
				Index: evmLog.Index,
				// 设置日志是否被移除
				Removed: evmLog.Removed,
			}
			// 将转移日志添加到切片中
			transferLogs = append(transferLogs, transferLog)
			continue
		}
	}

	// 按区块编号对转移日志进行排序
	sort.Slice(transferLogs, func(i, j int) bool {
		return transferLogs[i].BlockNumber < transferLogs[j].BlockNumber
	})

	// 返回转移日志切片和错误信息
	return transferLogs, nil
}

// isInSlice 函数用于检查给定的字符串是否在指定的字符串切片中。
//
// 参数：
//   str - 需要检查的字符串
//   slice - 需要检查的字符串切片
//
// 返回值：
//   如果字符串在切片中，则返回 true；否则返回 false
func (s *Service) isInSlice(str string, slice []string) bool {
	// 将输入的字符串转换为统一的地址格式
	addr, err := chain.UniformAddress(s.ChainName, str)
	if err != nil {
		// 如果转换失败，则返回false
		return false
	}

	// 遍历输入的字符串切片
	for _, item := range slice {
		// 将切片中的每个字符串转换为统一的地址格式
		itemAddr, _ := chain.UniformAddress(s.ChainName, item)
		// 如果转换后的地址与输入的字符串地址相同，则返回true
		if itemAddr == addr {
			return true
		}
	}
	// 遍历结束后，如果没有找到匹配的地址，则返回false
	return false
}
