package utils

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func IsZeroAddress(address string) bool {
	return address == "0x0000000000000000000000000000000000000000" ||
		address == "0x" ||
		strings.Trim(address, "0") == "x"
}

func FormatWeiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18))
}

func FormatWeiToUSDT(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e6))
}

func IsValidAddress(address string) bool {
	return common.IsHexAddress(address) && !IsZeroAddress(address)
}

// StringToWei 将带小数点的代币金额转换为 wei
// value: 代币金额字符串，支持小数点，如 "1.5"
// decimals: 代币精度，如 USDT 是 6
// 返回: wei 单位的大整数
func StringToWei(value string, decimals int) (*big.Int, error) {
	// 使用 big.Float 来支持小数
	bigFloat, ok := new(big.Float).SetString(value)
	if !ok {
		return nil, fmt.Errorf("invalid value format: %s", value)
	}

	// 计算乘数: 10^decimals
	multiplier := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))

	// 相乘得到 wei 值
	result := new(big.Float).Mul(bigFloat, multiplier)

	// 转换为 big.Int（截断小数部分）
	wei := new(big.Int)
	result.Int(wei) // Int() 方法会截断小数部分

	return wei, nil
}

// ReadContract 调用合约的只读方法（view/pure）
// client: 以太坊客户端连接
// contractABI: 合约ABI对象
// contractAddress: 合约地址
// methodName: 要调用的方法名
// params: 方法参数（可选）
// 返回: 返回值切片和错误，调用者需要根据合约方法定义进行类型断言
func ReadContract(client *ethclient.Client, contractABI abi.ABI, contractAddress common.Address, methodName string, params ...interface{}) ([]any, error) {
	// 1. 打包方法调用数据
	data, err := contractABI.Pack(methodName, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack method call: %w", err)
	}

	// 2. 调用合约
	callMsg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}

	result, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	// 3. 解包返回结果
	unpacked, err := contractABI.Unpack(methodName, result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack result: %w", err)
	}

	return unpacked, nil
}
