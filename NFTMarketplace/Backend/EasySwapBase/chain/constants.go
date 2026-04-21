package chain

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/ProjectsTask/EasySwapBase/evm/eip"
)

const (
	Eth      = "eth"
	Optimism = "optimism"
	Sepolia  = "sepolia"
)

const (
	EthChainID      = 1
	OptimismChainID = 10
	SepoliaChainID  = 11155111
)

// UniformAddress 函数用于将给定的地址转换为统一的校验和地址。
//
// 参数:
// chainName string: 区块链名称，当前未使用。
// address string: 需要转换的地址。
//
// 返回值:
// string: 转换后的统一校验和地址。
// error: 如果转换过程中发生错误，则返回错误信息。
func UniformAddress(chainName string, address string) (string, error) {
	// 使用eip库将地址转换为校验和地址
	addr, err := eip.ToCheckSumAddress(address)
	if err != nil {
		// 如果转换失败，则返回错误信息
		return "", errors.Wrap(err, "failed on uniform evm chain address")
	}
	// 将地址转换为小写并返回
	return strings.ToLower(addr), nil

}
