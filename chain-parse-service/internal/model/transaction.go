// Package model 定义了链上数据解析的核心数据模型
package model

import (
	"math/big"
)

// Transaction 交易数据结构
// 用于记录DEX上的交易(如swap)事件信息
type Transaction struct {
	Addr        string            `json:"addr"`        // 发行的代币地址/池子地址
	Router      string            `json:"router"`      // dex路由合约地址
	Factory     string            `json:"factory"`     // dex工厂合约地址
	Pool        string            `json:"pool"`        // 流动性池地址
	Hash        string            `json:"hash"`        // 交易哈希
	From        string            `json:"from"`        // 交易发起地址
	Side        string            `json:"side"`        // "swap" | "buy" | "sell" | "route"
	Amount      *big.Int          `json:"amount"`      // 换出代币数量(buy) / 换入代币的数量(sell)
	Price       float64           `json:"price"`       // 换出代币单价(buy) / 换入代币的单价(sell)
	Value       float64           `json:"value"`       // 换出代币的实际价值(buy) / 换入代币的实际价值(sell) (可以以稳定币计价)
	Time        uint64            `json:"time"`        // 区块时间戳
	Extra       *TransactionExtra `json:"extra"`       // 额外扩展信息
	EventIndex  int64             `json:"index"`       // 事件在区块中的索引 / Index of the event in the transaction
	TxIndex     int64             `json:"txIndex"`     // 交易在区块中的索引
	SwapIndex   int64             `json:"swapIndex"`   // 同一笔交易内swap操作的顺序编号
	BlockNumber int64             `json:"blockNumber"` // 交易所在区块号 / Index of the transaction in the block
	Protocol    string            `json:"protocol"`    // DEX协议名称(pancakeswap/uniswap/bluefin等) / DEX protocol name
	ChainType   string            `json:"chain_type"`  // 链类型(bsc/ethereum/solana/sui) / chain type
}

// TransactionExtra 交易的额外扩展信息
type TransactionExtra struct {
	QuoteAddr     string `json:"quote_addr"`     // 报价代币地址
	QuotePrice    string `json:"quote_price"`    // 报价代币单价
	Type          string `json:"type"`           // 交易类型，如"swap", "add_liquidity", "remove_liquidity" / e.g., "swap", "add_liquidity", "remove_liquidity"
	TokenSymbol   string `json:"token_symbol"`   // 交易代币的符号 / Symbol of the token being traded
	TokenDecimals int    `json:"token_decimals"` // 交易代币的精度 / Decimals of the token being traded
}
