package main

import (
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

/* calculateEventSigForEVM 计算 EVM 事件的签名哈希。
 *
 * 参数：
 *   - eventName: 事件签名，格式如 "Transfer(address,address,uint256)"
 *
 * 返回值：
 *   - keccak256 哈希的十六进制字符串，如 "0x..."
 */
func calculateEventSigForEVM(eventName string) string {
	eventName = strings.Join(strings.Fields(eventName), "")
	eventHash := crypto.Keccak256Hash([]byte(eventName))
	return eventHash.Hex()
}

/* calculateEventSigForSolana 计算 Solana Anchor 事件的 discriminator。
 *
 * 参数：
 *   - eventName: 事件名称，如 "CreateEvent"
 *
 * 返回值：
 *   - sha256("event:事件名") 的前 8 个字节
 */
func calculateEventSigForSolana(eventName string) []byte {
	eventName = strings.Join(strings.Fields(eventName), "")
	// Anchor 事件 discriminator 前缀是 "event:"
	eventNameWithPrefix := "event:" + eventName

	// 计算 sha256 哈希
	hash := sha256.Sum256([]byte(eventNameWithPrefix))

	// 取前 8 个字节作为 discriminator
	discriminator := hash[:8]

	return discriminator
}

/* calculateEventSigForSui 计算 Sui 事件的标识符。
 *
 * 参数：
 *   - packageAddr: 包地址
 *   - moduleName: 模块名称
 *   - eventName: 事件名称
 *
 * 返回值：
 *   - 格式为 "package地址::模块名::事件名" 的字符串
 */
func calculateEventSigForSui(packageAddr, moduleName, eventName string) string {
	packageAddr = strings.Join(strings.Fields(packageAddr), "")
	moduleName = strings.Join(strings.Fields(moduleName), "")
	eventName = strings.Join(strings.Fields(eventName), "")
	return packageAddr + "::" + moduleName + "::" + eventName
}

func TestEventSig_EVM(t *testing.T) {
	// Four.meme
	fourMemeV1Events := []string{}
	fourMemeV1TokenCreateEvent := "TokenCreate(address,address,uint256,string,string,uint256,uint256)"
	fourMemeV1TokenPurchaseEvent := "TokenPurchase(address,address,uint256,uint256)"
	fourMemeV1TokenSaleEvent := "TokenSale(address,address,uint256,uint256)"
	fourMemeV1Events = append(fourMemeV1Events, fourMemeV1TokenCreateEvent, fourMemeV1TokenPurchaseEvent, fourMemeV1TokenSaleEvent)

	fourMemeV2Events := []string{}
	fourMemeV2TokenCreateEvent := "TokenCreate(address,address,uint256,string,string,uint256,uint256,uint256)"
	fourMemeV2TokenPurchaseEvent := "TokenPurchase(address,address,uint256,uint256,uint256,uint256,uint256,uint256)"
	fourMemeV2TokenSaleEvent := "TokenSale(address,address,uint256,uint256,uint256,uint256,uint256,uint256)"
	fourMemeV2TradeStopEvent := "TradeStop(address)"
	fourMemeV2LiquidityAddedEvent := "LiquidityAdded(address,uint256,address,uint256)"
	fourMemeV2Events = append(fourMemeV2Events, fourMemeV2TokenCreateEvent, fourMemeV2TokenPurchaseEvent, fourMemeV2TokenSaleEvent, fourMemeV2TradeStopEvent, fourMemeV2LiquidityAddedEvent)

	t.Logf("fourMemeV1Events: %v", fourMemeV1Events)
	for _, event := range fourMemeV1Events {
		t.Logf("%s: %v", event, calculateEventSigForEVM(event))
	}
	t.Logf("fourMemeV2Events: %v", fourMemeV2Events)
	for _, event := range fourMemeV2Events {
		t.Logf("%s: %v", event, calculateEventSigForEVM(event))
	}
	t.Log("------------------------------------------------------------------------------------------")

	// PancakeSwap
	pancakeSwapV2Events := []string{
		"Swap(address,uint256,uint256,uint256,uint256,address)",
		"Mint(address,uint256,uint256)",
		"Burn(address,uint256,uint256,address)",
		"PairCreated(address,address,address,uint256)",
	}

	pancakeSwapV3Events := []string{
		"Swap(address,address,int256,int256,uint160,uint128,int24,uint128,uint128)",
		"Mint(address,address,int24,int24,uint128,uint256,uint256)",
		"Burn(address,int24,int24,uint128,uint256,uint256)",
		"PoolCreated(address,address,uint24,int24,address)",
	}

	t.Logf("pancakeSwapV2Events: %v", pancakeSwapV2Events)
	for _, event := range pancakeSwapV2Events {
		t.Logf("%s: %v", event, calculateEventSigForEVM(event))
	}
	t.Logf("pancakeSwapV3Events: %v", pancakeSwapV3Events)
	for _, event := range pancakeSwapV3Events {
		t.Logf("%s: %v", event, calculateEventSigForEVM(event))
	}
	t.Log("------------------------------------------------------------------------------------------")

	// Uniswap
	uniswapV2Events := []string{
		"Swap(address,uint256,uint256,uint256,uint256,address)",
		"Mint(address,uint256,uint256)",
		"Burn(address,uint256,uint256,address)",
		"PairCreated(address,address,address,uint256)",
	}

	uniswapV3Events := []string{
		"Swap(address,address,int256,int256,uint160,uint128,int24)",
		"Mint(address,address,int24,int24,uint128,uint256,uint256)",
		"Burn(address,int24,int24,uint128,uint256,uint256)",
		"PoolCreated(address,address,uint24,int24,address)",
	}

	t.Logf("uniswapV2Events: %v", uniswapV2Events)
	for _, event := range uniswapV2Events {
		t.Logf("event:%s -> %v", event, calculateEventSigForEVM(event))
	}
	t.Logf("uniswapV3Events: %v", uniswapV3Events)
	for _, event := range uniswapV3Events {
		t.Logf("event:%s -> %v", event, calculateEventSigForEVM(event))
	}
}

func TestEventSig_Solana(t *testing.T) {
	// PumpFun
	pumpFunEvents := []string{
		"CreateEvent",
		"TradeEvent",
		"CompleteEvent",
	}

	t.Logf("PumpFun Event Discriminators:")
	for _, event := range pumpFunEvents {
		discriminator := calculateEventSigForSolana(event)
		t.Logf("event:%s -> %v", event, discriminator)
	}
	t.Log("------------------------------------------------------------------------------------------")

	// PumpSwap
	pumSwapEvents := []string{
		"BuyEvent",
		"SellEvent",
		"CreatePoolEvent",
		"DepositEvent",
		"WithdrawEvent",
	}

	t.Logf("PumpSwap Event Discriminators:")
	for _, event := range pumSwapEvents {
		discriminator := calculateEventSigForSolana(event)
		t.Logf("event:%s -> %v", event, discriminator)
	}
}

func TestEventSig_Sui(t *testing.T) {
	// bluefin
	bluefinAmmAddr := "0x3492c874c1e3b3e2984e8c41b589e642d4d0a5d6459e5a9cfc2d52fd7c89c267"
	eventsModuleName := "events"
	bluefinEvents := []string{
		"PoolCreated",
		"LiquidityProvided",
		"LiquidityRemoved",
		"AssetSwap",
		"FlashSwap",
	}
	t.Logf("bluefinEvents: %v", bluefinEvents)
	for _, event := range bluefinEvents {
		t.Logf("event:%s -> %v", event, calculateEventSigForSui(bluefinAmmAddr, eventsModuleName, event))
	}
	t.Log("------------------------------------------------------------------------------------------")

	// cetus

	cetusClmmPoolAddr := "0x1eabed72c53feb3805120a081dc15963c204dc8d091542592abaf7a35689b2fb"
	cetusLiquidityModuleAddr := "0xdb5cd62a06c79695bfc9982eb08534706d3752fe123b48e0144f480209b3117f"
	cetusDlmmAddr := "0x5664f9d3fd82c84023870cfbda8ea84e14c8dd56ce557ad2116e0668581a682b"
	factoryModuleName := "factory"
	poolModuleName := "pool"

	cetusClmmfactoryEvents := []string{
		"CreatePoolEvent",
	}
	t.Logf("Cetus clmm factory events: ")
	for _, event := range cetusClmmfactoryEvents {
		t.Logf("event:%s -> %v", event, calculateEventSigForSui(cetusClmmPoolAddr, factoryModuleName, event))
	}

	cetusClmmPoolEvents := []string{
		"AddLiquidityEvent",
		"RemoveLiquidityEvent",
		"SwapEvent",
		"CollectFeeEvent",
		"CollectRewardEvent",
	}
	t.Logf("Cetus clmm pool events: ")
	for _, event := range cetusClmmPoolEvents {
		t.Logf("event:%s -> %v", event, calculateEventSigForSui(cetusClmmPoolAddr, poolModuleName, event))
	}

	cetusV2ClmmPoolEvents := []string{
		"AddLiquidityV2Event",
		"RemoveLiquidityV2Event",
	}
	t.Logf("Cetus v2 clmm pool events: ")
	for _, event := range cetusV2ClmmPoolEvents {
		t.Logf("event:%s -> %v", event, calculateEventSigForSui(cetusLiquidityModuleAddr, poolModuleName, event))
	}

	cetusDlmmPoolEvents := []string{
		"SwapEvent",
		"AddLiquidityEvent",
		"RemoveLiquidityEvent",
	}
	t.Logf("Cetus dlmm pool events: ")
	for _, event := range cetusDlmmPoolEvents {
		t.Logf("event:%s -> %v", event, calculateEventSigForSui(cetusDlmmAddr, poolModuleName, event))
	}
}
