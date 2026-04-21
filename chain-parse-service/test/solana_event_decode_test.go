package main

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"

	"github.com/mr-tron/base58"
)

// 已知的事件 discriminator 映射表
var knownDiscriminators = map[string]string{
	// PumpFun
	"1b72ad4ddeeb6376": "PumpFun:CreateEvent",
	"bddb7fd34ee661ee": "PumpFun:TradeEvent",
	"5f72619cd42e9808": "PumpFun:CompleteEvent",
	// 其他 DEX (待补充)
	// "xxxxxxxx": "Raydium:SwapEvent",
	// "xxxxxxxx": "Orca:SwapEvent",
	// "xxxxxxxx": "Jupiter:RouteEvent",
}

// printEventData 打印事件数据的可读格式
func printEventData(data []byte) string {
	var sb strings.Builder

	if len(data) < 8 {
		return fmt.Sprintf("[数据过短: %d 字节]", len(data))
	}

	// 提取 discriminator
	disc := data[:8]
	discHex := fmt.Sprintf("%x", disc)
	sb.WriteString(fmt.Sprintf("Discriminator: %s", discHex))

	// 尝试识别事件类型
	if name, ok := knownDiscriminators[discHex]; ok {
		sb.WriteString(fmt.Sprintf(" [%s]", name))
	} else {
		sb.WriteString(" [未知事件类型]")
	}
	sb.WriteString("\n")

	// 分段打印数据（每 32 字节一行，Solana pubkey 长度）
	for i := 8; i < len(data); i += 32 {
		end := i + 32
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]

		// 打印偏移和 hex
		sb.WriteString(fmt.Sprintf("Offset %04d: %x", i, chunk))

		// 如果是 32 字节，尝试解析为 pubkey
		if len(chunk) == 32 {
			pubkey := base58.Encode(chunk)
			sb.WriteString(fmt.Sprintf(" (可能为 Pubkey: %s)", pubkey))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// parseBorshFields 尝试解析 Borsh 字段
func parseBorshFields(data []byte) map[string]interface{} {
	result := make(map[string]interface{})
	offset := 8 // 跳过 discriminator

	// 尝试解析 Pubkey (32 字节)
	if len(data) >= offset+32 {
		pubkey := base58.Encode(data[offset : offset+32])
		result["field_0_pubkey"] = pubkey
		offset += 32
	}

	// 尝试解析 u64 (8 字节)
	if len(data) >= offset+8 {
		u64Val := binary.LittleEndian.Uint64(data[offset : offset+8])
		result["field_1_u64"] = u64Val
		offset += 8
	}

	// 尝试解析第二个 u64
	if len(data) >= offset+8 {
		u64Val := binary.LittleEndian.Uint64(data[offset : offset+8])
		result["field_2_u64"] = u64Val
		offset += 8
	}

	// 尝试解析 bool (1 字节)
	if len(data) >= offset+1 {
		boolVal := data[offset] != 0
		result["field_3_bool"] = boolVal
		offset += 1
	}

	// 尝试解析第二个 Pubkey
	if len(data) >= offset+32 {
		pubkey := base58.Encode(data[offset : offset+32])
		result["field_4_pubkey"] = pubkey
		offset += 32
	}

	// 尝试解析 i64 timestamp (8 字节)
	if len(data) >= offset+8 {
		i64Val := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
		result["field_5_i64_timestamp"] = i64Val
		offset += 8
	}

	result["remaining_bytes"] = len(data) - offset

	return result
}

func Test_SolanaEventDecode(t *testing.T) {
	// 测试数据：PumpFun TradeEvent 的 base64 编码
	b64Data := "vdt/007mYe4dGrPt6OFkOXJLs/aOoEOX7RLs6KKEjOubEFcitskwT+UZAQEAAAAAr5oxVHcAAAABxc/vyd/tlaoaxur5ahQ+fo0mxDVHgOGRHcky79uRpUENbddpAAAAANnjv0AHAAAAU1SYlk9dAwAC9AzdAAAAAFO8hUq+XgIA6JMUH7GOnxV02BDheOGeMGBOMXWqLkoy38hgByfRBwlfAAAAAAAAAEZxAgAAAAAAvNCU9woZvHz5LwHMkYwhcCpyDrOBWVZnh4V12TB6VMAeAAAAAAAAAHXFAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAGJ1eV9leGFjdF9zb2xfaW4BAAAAAAAAAAAAAAAAAAAAAA=="

	decoded, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		t.Errorf("Failed to decode base64 string: %v", err)
		return
	}

	t.Logf("\n========== 事件解析结果 ==========")
	t.Logf("总长度: %d 字节", len(decoded))
	t.Logf("\n--- 数据分段 ---")
	t.Log(printEventData(decoded))

	t.Logf("\n--- Borsh 字段解析尝试 ---")
	fields := parseBorshFields(decoded)
	for key, val := range fields {
		t.Logf("  %s: %v", key, val)
	}

	t.Logf("\n--- 原始 Hex (完整) ---")
	t.Logf("%x", decoded)
}

// Test_UnknownDiscriminator 测试识别未知 discriminator
// 用于发现新事件类型
func Test_UnknownDiscriminator(t *testing.T) {
	// 如果遇到未知事件，可以提取其 discriminator
	unknownEventB64 := "" // 在这里填入未知事件的 base64

	if unknownEventB64 == "" {
		t.Skip("没有未知事件数据")
	}

	decoded, err := base64.StdEncoding.DecodeString(unknownEventB64)
	if err != nil {
		t.Errorf("Failed to decode: %v", err)
		return
	}

	if len(decoded) < 8 {
		t.Errorf("数据过短")
		return
	}

	disc := decoded[:8]
	discHex := fmt.Sprintf("%x", disc)

	t.Logf("\n========== 未知事件 Discriminator ==========")
	t.Logf("Hex: %s", discHex)
	t.Logf("建议添加到 knownDiscriminators 映射表中")

	if name, ok := knownDiscriminators[discHex]; ok {
		t.Logf("实际上这是一个已知事件: %s", name)
	} else {
		t.Logf("这是新的事件类型，请确认属于哪个 DEX 协议")
	}
}
