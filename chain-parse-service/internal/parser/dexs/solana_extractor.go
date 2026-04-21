package dex

import (
	"encoding/base64"
	"encoding/binary"
	"math/big"
	"strconv"
	"strings"
	"time"

	"unified-tx-parser/internal/types"

	"github.com/mr-tron/base58"
)

// SolanaDexExtractor provides shared functionality for Solana-based DEX extractors
type SolanaDexExtractor struct {
	*BaseDexExtractor
}

// NewSolanaDexExtractor creates a Solana DEX extractor with the given configuration
func NewSolanaDexExtractor(cfg *BaseDexExtractorConfig) *SolanaDexExtractor {
	if cfg == nil {
		cfg = &BaseDexExtractorConfig{}
	}
	return &SolanaDexExtractor{
		BaseDexExtractor: NewBaseDexExtractor(cfg),
	}
}

// IsSolanaChainSupported checks if Solana is supported
func (s *SolanaDexExtractor) IsSolanaChainSupported(chainType types.ChainType) bool {
	return chainType == types.ChainTypeSolana
}

// ---------- Byte parsing utilities (offset-based, boundary-safe) ----------

// ParseU8 parses an unsigned 8-bit integer at the given offset.
// Returns the value and the new offset. Returns zero on out-of-bounds.
func ParseU8(data []byte, offset int) (uint8, int) {
	if offset < 0 || offset >= len(data) {
		return 0, offset
	}
	return data[offset], offset + 1
}

// ParseBool parses a boolean (1 byte) at the given offset.
// Returns the value and the new offset. Returns false on out-of-bounds.
func ParseBool(data []byte, offset int) (bool, int) {
	if offset < 0 || offset >= len(data) {
		return false, offset
	}
	return data[offset] != 0, offset + 1
}

// ParseU16LE parses an unsigned 16-bit little-endian integer at the given offset.
// Returns the value and the new offset. Returns zero on out-of-bounds.
func ParseU16LE(data []byte, offset int) (uint16, int) {
	if offset < 0 || offset+2 > len(data) {
		return 0, offset
	}
	return binary.LittleEndian.Uint16(data[offset : offset+2]), offset + 2
}

// ParseU64LE parses an unsigned 64-bit little-endian integer at the given offset.
// Returns the value and the new offset. Returns zero on out-of-bounds.
func ParseU64LE(data []byte, offset int) (uint64, int) {
	if offset < 0 || offset+8 > len(data) {
		return 0, offset
	}
	return binary.LittleEndian.Uint64(data[offset : offset+8]), offset + 8
}

// ParseI64LE parses a signed 64-bit little-endian integer at the given offset.
// Returns the value and the new offset. Returns zero on out-of-bounds.
func ParseI64LE(data []byte, offset int) (int64, int) {
	if offset < 0 || offset+8 > len(data) {
		return 0, offset
	}
	return int64(binary.LittleEndian.Uint64(data[offset : offset+8])), offset + 8
}

// ParseU128LE parses an unsigned 128-bit little-endian integer at the given offset.
// Returns the value as Uint128 and the new offset. Returns zero on out-of-bounds.
func ParseU128LE(data []byte, offset int) (Uint128, int) {
	if offset < 0 || offset+16 > len(data) {
		return Uint128{}, offset
	}
	return Uint128{
		Low:  binary.LittleEndian.Uint64(data[offset : offset+8]),
		High: binary.LittleEndian.Uint64(data[offset+8 : offset+16]),
	}, offset + 16
}

// ParsePubkey parses a 32-byte Solana public key at the given offset and returns
// its base58-encoded string representation. Returns empty string on out-of-bounds.
func ParsePubkey(data []byte, offset int) (string, int) {
	if offset < 0 || offset+32 > len(data) {
		return "", offset
	}
	pubkeyBytes := data[offset : offset+32]
	return base58.Encode(pubkeyBytes), offset + 32
}

// ParseString parses a Borsh-encoded string (4-byte LE length prefix + UTF-8 data)
// at the given offset. Returns the string and new offset. Returns empty string on error.
func ParseString(data []byte, offset int) (string, int) {
	if offset < 0 || offset+4 > len(data) {
		return "", offset
	}
	length := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if length < 0 || offset+length > len(data) {
		return "", offset
	}
	return string(data[offset : offset+length]), offset + length
}

// ---------- Uint128 type ----------

// Uint128 represents a 128-bit unsigned integer
type Uint128 struct {
	Low  uint64
	High uint64
}

// ToBigInt converts Uint128 to *big.Int
func (u Uint128) ToBigInt() *big.Int {
	result := new(big.Int).SetUint64(u.High)
	result.Lsh(result, 64)
	result.Or(result, new(big.Int).SetUint64(u.Low))
	return result
}

// String implements fmt.Stringer for Uint128
func (u Uint128) String() string {
	return u.ToBigInt().String()
}

// ---------- Solana conversion utilities ----------

// LamportsToSOL converts lamports (uint64) to SOL (float64).
// 1 SOL = 1_000_000_000 lamports.
func LamportsToSOL(lamports uint64) float64 {
	return float64(lamports) / 1e9
}

// ---------- Discriminator matching ----------

// ExtractDiscriminator extracts the event discriminator (first 8 bytes)
func (s *SolanaDexExtractor) ExtractDiscriminator(eventData []byte) ([]byte, error) {
	if len(eventData) < 8 {
		return nil, &InsufficientDataError{Needed: 8, Got: len(eventData), Field: "discriminator"}
	}
	return eventData[:8], nil
}

// MatchDiscriminator compares two discriminators
func (s *SolanaDexExtractor) MatchDiscriminator(actual, expected []byte) bool {
	if len(actual) != len(expected) {
		return false
	}
	for i := range actual {
		if actual[i] != expected[i] {
			return false
		}
	}
	return true
}

// MatchDiscriminatorBytes is a package-level helper for comparing discriminators without
// requiring a SolanaDexExtractor receiver.
func MatchDiscriminatorBytes(data []byte, expected []byte) bool {
	if len(data) < len(expected) {
		return false
	}
	for i := range expected {
		if data[i] != expected[i] {
			return false
		}
	}
	return true
}

// ---------- Solana log extraction ----------

/*
 * SolanaEvent 表示从 Solana 交易日志中提取的一个 Anchor 事件，
 * 包含事件数据和发出该事件的程序 ID。
 */
type SolanaEvent struct {
	Data      []byte // 解码后的事件数据（含 discriminator）
	ProgramID string // 发出该事件的程序 ID（base58）
}

/*
 * ExtractSolanaEvents 从交易日志中提取所有 Anchor 事件，同时追踪每个事件的来源程序 ID。
 *
 * Solana 日志格式示例：
 *   Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P invoke [1]
 *   Program log: Instruction: Buy
 *   Program data: <base64 事件数据>
 *   Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P success
 *
 * 通过跟踪 "Program <ID> invoke" 行来确定 "Program data:" 对应的程序 ID。
 */
func ExtractSolanaEvents(tx *types.UnifiedTransaction) []SolanaEvent {
	if tx == nil || tx.RawData == nil {
		return nil
	}

	var logMessages []string

	switch rawData := tx.RawData.(type) {
	case map[string]interface{}:
		if logs, ok := rawData["log_messages"]; ok {
			logMessages = toStringSlice(logs)
		} else if logs, ok := rawData["logMessages"]; ok {
			logMessages = toStringSlice(logs)
		} else if meta, ok := rawData["meta"]; ok {
			if metaMap, ok := meta.(map[string]interface{}); ok {
				if logs, ok := metaMap["logMessages"]; ok {
					logMessages = toStringSlice(logs)
				} else if logs, ok := metaMap["log_messages"]; ok {
					logMessages = toStringSlice(logs)
				}
			}
		}
	}

	if len(logMessages) == 0 {
		return nil
	}

	// 追踪当前执行的程序 ID，使用栈结构支持嵌套调用（如 CPI 调用）
	// CPI: Solana 区块链中的 跨程序调用（Cross-Program Invocation）
	var programStack []string
	var events []SolanaEvent

	const invokePrefix = "Program "
	const invokeSuffix = " invoke ["
	const dataPrefix = "Program data: "

	for _, line := range logMessages {
		// 检测程序调用：Program <ID> invoke [<depth>]
		if strings.HasPrefix(line, invokePrefix) && strings.Contains(line, invokeSuffix) {
			trimmed := strings.TrimPrefix(line, invokePrefix)
			idx := strings.Index(trimmed, invokeSuffix)
			if idx > 0 {
				programID := strings.TrimSpace(trimmed[:idx])
				programStack = append(programStack, programID)
			}
			continue
		}

		// 检测程序返回：Program <ID> success / failed
		if strings.HasPrefix(line, invokePrefix) &&
			(strings.HasSuffix(line, " success") || strings.HasSuffix(line, " failed")) {
			if len(programStack) > 0 {
				programStack = programStack[:len(programStack)-1]
			}
			continue
		}

		// 检测事件数据：Program data: <base64>
		if strings.HasPrefix(line, dataPrefix) {
			b64Data := strings.TrimPrefix(line, dataPrefix)
			decoded, err := base64.StdEncoding.DecodeString(b64Data)
			if err != nil {
				continue
			}
			if len(decoded) < 8 {
				continue
			}

			// 取栈顶程序 ID 作为事件来源
			programID := ""
			if len(programStack) > 0 {
				programID = programStack[len(programStack)-1]
			}

			events = append(events, SolanaEvent{
				Data:      decoded,
				ProgramID: programID,
			})
		}
	}

	return events
}

/*
 * ExtractSolanaEventData 从交易日志中提取事件数据（不含程序 ID 信息）。
 * 保留此函数以保持向后兼容，新代码建议使用 ExtractSolanaEvents。
 */
func ExtractSolanaEventData(tx *types.UnifiedTransaction) [][]byte {
	events := ExtractSolanaEvents(tx)
	if len(events) == 0 {
		return nil
	}
	result := make([][]byte, 0, len(events))
	for _, evt := range events {
		result = append(result, evt.Data)
	}
	return result
}

// toStringSlice converts an interface{} to []string, handling common types.
func toStringSlice(v interface{}) []string {
	switch val := v.(type) {
	case []string:
		return val
	case []interface{}:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

// ---------- Error helpers ----------

// InsufficientDataError indicates that there is not enough data to parse a field
type InsufficientDataError struct {
	Needed int
	Got    int
	Field  string
}

func (e *InsufficientDataError) Error() string {
	return "insufficient data for " + e.Field + ": needed " + strconv.Itoa(e.Needed) + " bytes, got " + strconv.Itoa(e.Got)
}

// CacheTTL constants for Solana extractors
const (
	SolanaCacheDefaultTTL = 5 * time.Minute
	SolanaTokenCacheTTL   = 1 * time.Hour
)
