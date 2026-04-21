package eip

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// TestToCheckSumAddress 测试 ToCheckSumAddress 函数的功能
func TestToCheckSumAddress(t *testing.T) {
	// 测试用例数组
	cases := []string{
		"0xe01511d7333A18e969758BBdC9C7f50CcF30160A",
		"0x62d17DE1fbDF36597F12F19717C39985A921426e",
		"0x6F702345360D6D8533d2362eC834bf5f1aB63910",
	}
	// 遍历测试用例数组
	for _, c := range cases {
		// 为每个测试用例运行子测试
		t.Run(c, func(t *testing.T) {
			// 调用 ToCheckSumAddress 函数，并将地址转换为小写
			res, err := ToCheckSumAddress(strings.ToLower(c))
			// 断言没有错误发生
			assert.Nil(t, err)
			// 断言返回的结果与原始地址相等
			assert.Equal(t, res, c)
		})
	}
}
