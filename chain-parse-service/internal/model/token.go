// Package model 定义了链上数据解析的核心数据模型
package model

// Token 代币数据结构
// 用于记录ERC20代币的基本信息和元数据
type Token struct {
	Addr         string  `json:"addr"`                   // 代币合约地址
	Name         string  `json:"name"`                   // 代币名称
	Symbol       string  `json:"symbol"`                 // 代币符号
	Decimals     int     `json:"decimals"`               // 代币精度(小数位数)
	ChainType    string  `json:"chain_type"`             // 链类型(bsc/ethereum/solana/sui) / chain type
	TRLThreshold int     `json:"trlThreshold,omitempty"` // TRL阈值 / TRL threshold for this token
	IsStable     bool    `json:"is_stable"`              // 是否为稳定币
	CreatedAt    string  `json:"created_at,omitempty"`   // 创建时间(ISO 8601格式) / ISO 8601 format
	UsdPrice     float64 `json:"price_usd,omitempty"`    // 代币的USD价格 / USD price of the token
}
