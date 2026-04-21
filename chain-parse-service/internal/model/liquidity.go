// Package model 定义了链上数据解析的核心数据模型
package model

import (
	"math/big"
)

// Liquidity 流动性事件数据结构
// 用于记录添加/移除流动性的事件信息
type Liquidity struct {
	Addr      string          `json:"addr"`            // 主代币（Base Token）的合约地址 / pool地址 / 添加流动性涉及的其中一个token的地址
	Router    string          `json:"router"`          // 路由合约地址
	Factory   string          `json:"factory"`         // 工厂合约地址
	Pool      string          `json:"pool"`            // 流动性池地址
	Hash      string          `json:"hash"`            // 交易哈希
	From      string          `json:"from"`            // 交易发起地址
	Pos       string          `json:"pos"`             // 头寸标识符 / Position identifier
	Side      string          `json:"side"`            // 操作类型: add/remove(添加流动性/移除流动性)
	Amount    *big.Int        `json:"amount"`          // 添加的主代币数量 / 添加/移除的总代币数量
	Value     float64         `json:"value"`           // 添加的计价代币数量(BNB, USDT, ...)
	Time      uint64          `json:"time"`            // 区块时间戳
	Key       string          `json:"key"`             // 唯一键值，用于标识流动性事件 / Unique key for the liquidity event
	Extra     *LiquidityExtra `json:"extra,omitempty"` // 额外信息
	Protocol  string          `json:"protocol"`        // DEX协议名称(pancakeswap/uniswap/bluefin等) / DEX protocol name
	ChainType string          `json:"chain_type"`      // 链类型(bsc/ethereum/solana/sui) / chain type
}

// LiquidityExtra 流动性事件的额外扩展信息
type LiquidityExtra struct {
	Key     string    `json:"key"`     // 唯一键值，用于标识流动性事件 / Unique key for the liquidity event
	Amounts *big.Int  `json:"amounts"` // 添加的计价代币数量（BNB 或其他 BEP20 代币）/ 涉及的代币数量 / Amounts of tokens involved in the liquidity event
	Values  []float64 `json:"values"`  // [主代币数量, 计价代币数量] / [token0数量, token1数量] / 代币价值(USD或其他法币) / Values of tokens in USD or other fiat currency
	Time    uint64    `json:"time"`    // 交易所在区块时间戳 / Timestamp of the liquidity event
}
