package eth

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// makeMinimalEthLog creates a minimal *ethtypes.Log with a single topic for testing
// internal methods like getLogType, isUniswapLog, etc.
func makeMinimalEthLog(topic0 string) *ethtypes.Log {
	log := &ethtypes.Log{}
	if topic0 != "" {
		log.Topics = []common.Hash{common.HexToHash(topic0)}
	}
	return log
}

// makeMinimalEthLogWithAddr creates a minimal *ethtypes.Log with address and topic.
func makeMinimalEthLogWithAddr(addr, topic0 string) *ethtypes.Log {
	log := &ethtypes.Log{
		Address: common.HexToAddress(addr),
	}
	if topic0 != "" {
		log.Topics = []common.Hash{common.HexToHash(topic0)}
	}
	return log
}

// prepopulatePairCache 预填充 pairCache，使 isUniswapLog 在无 Client 时也能识别 pair 地址
func prepopulatePairCache(ext *UniswapExtractor, addrs ...string) {
	for _, addr := range addrs {
		key := strings.ToLower(common.HexToAddress(addr).Hex())
		ext.pairCache.Store(key, true)
	}
}
