package eip

import (
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/sha3"
	"strconv"
	"strings"
)

// ToCheckSumAddress 将给定的以太坊地址转换为校验和地址格式
//
// 参数:
//     address: 字符串类型，以太坊地址
//
// 返回值:
//     string: 转换后的校验和地址
//     error: 如果有错误发生则返回错误信息，否则返回nil
func ToCheckSumAddress(address string) (string, error) {
	// 检查地址是否为空
	if address == "" {
		return "", errors.New("empty address")
	}

	// 如果地址以"0x"开头，则去除"0x"
	if strings.HasPrefix(address, "0x") {
		address = address[2:]
	}
	
	// 将地址从十六进制字符串解码为字节数组
	bytes, err := hex.DecodeString(address)
	if err != nil {
		return "", err
	}

	// 计算地址的Keccak256哈希值
	hash := calculateKeccak256([]byte(strings.ToLower(address)))
	result := "0x"
	for i, b := range bytes {
		// 计算并拼接校验和
		result += checksumByte(b>>4, hash[i]>>4)
		result += checksumByte(b&0xF, hash[i]&0xF)
	}

	return result, nil
}

// checksumByte 函数用于将给定的字节addr转换为16进制字符串，并根据hash的值决定是否将结果转换为大写。
//
// 参数:
//     addr byte: 要转换的字节。
//     hash byte: 用于决定是否将结果转换为大写的阈值。
//
// 返回值:
//     string: 转换后的16进制字符串，可能为大写或小写。
func checksumByte(addr byte, hash byte) string {
	// 将addr转换为16进制字符串
	result := strconv.FormatUint(uint64(addr), 16)

	// 如果hash大于等于8，则将result转换为大写
	if hash >= 8 {
		// 将result转换为大写
		return strings.ToUpper(result)
	} else {
		// 直接返回result
		return result
	}
}

// calculateKeccak256 计算给定地址的Keccak256哈希值
//
// 参数:
//     addr []byte: 输入地址字节切片
//
// 返回值:
//     []byte: 计算得到的Keccak256哈希值字节切片
func calculateKeccak256(addr []byte) []byte {
	// 创建一个新的Keccak256哈希实例
	hash := sha3.NewLegacyKeccak256()
	// 将输入地址写入哈希实例
	hash.Write(addr)
	// 计算哈希值并返回结果
	return hash.Sum(nil)
}
