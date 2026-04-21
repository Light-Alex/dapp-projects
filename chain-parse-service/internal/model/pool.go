// Package model 定义了链上数据解析的核心数据模型
package model

// Pool 流动性池数据结构
// 用于记录DEX流动性池的详细信息
type Pool struct {
	Addr      string                 `json:"addr"`            // 流动性池地址/创建的代币合约地址 / pool address
	Factory   string                 `json:"factory"`         // 工厂合约地址(如bluefin) / bluefin factory address
	Protocol  string                 `json:"protocol"`        // 协议名称(如bluefin) / bluefin
	ChainType string                 `json:"chain_type"`      // 链类型(bsc/ethereum/solana/sui) / chain type
	Tokens    map[int]string         `json:"tokens"`          // 池中的代币地址映射，key为代币索引
	Args      map[string]interface{} `json:"args,omitempty"`  // 额外的池参数
	Extra     *PoolExtra             `json:"extra,omitempty"` // 池的额外扩展信息
	Fee       int                    `json:"fee"`             // 手续费(bps, fee_basis_points, 基点) / 1 基点 = 0.01% = 1/10000 / 1 hundredths of a bip = 1/1e6 / (EVM,SUI 用 hundredths of a bip，Solana 用 bps)
}

// PoolExtra 池的额外扩展信息
type PoolExtra struct {
	Hash   string `json:"tx_hash"`           // 创建池的交易哈希
	From   string `json:"tx_from"`           // 创建交易的发起地址
	Time   uint64 `json:"tx_time,omitempty"` // 创建交易的时间戳
	Stable bool   `json:"stable,omitempty"`  // 是否为稳定币池
}
