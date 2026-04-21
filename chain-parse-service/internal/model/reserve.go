// Package model 定义了链上数据解析的核心数据模型
package model

import (
	"math/big"
)

// Reserve 储备金数据结构
// 用于记录流动性池中各代币的储备量信息
type Reserve struct {
	Addr      string           `json:"addr"`            // 流动性池地址 / pool address
	Amounts   map[int]*big.Int `json:"amounts"`         // 各代币的储备数量，key为代币索引
	Time      uint64           `json:"time"`            // 区块时间戳 / block time
	Value     map[int]float64  `json:"value,omitempty"` // 各代币的储备价值(USD或其他法币) / Value in USD or other fiat currency
	Protocol  string           `json:"protocol"`        // DEX协议名称(pancakeswap/uniswap/bluefin等) / DEX protocol name
	ChainType string           `json:"chain_type"`      // 链类型(bsc/ethereum/solana/sui) / chain type
}
