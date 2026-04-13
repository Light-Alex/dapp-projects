package bsc

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"unified-tx-parser/internal/model"
	bscProcessor "unified-tx-parser/internal/parser/chains/bsc"
	dex "unified-tx-parser/internal/parser/dexs"
	"unified-tx-parser/internal/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	pancakeSwapV2FactoryAddr = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"
	pancakeSwapV3FactoryAddr = "0x0BFbCF9fa4f9C56B0F40a671Ad40E0805A091865"

	pancakeSwapV2EventSig      = "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"
	pancakeSwapV3EventSig      = "0x19b47279256b2a23a1665c810c8d55a1758940ee09377d4f8d26497a3577dc83"
	pancakeMintV2EventSig      = "0x4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f"
	pancakeBurnV2EventSig      = "0xdccd412f0b1252819cb1fd330b93224ca42612892bb3f4f789976e6d81936496"
	pancakeMintV3EventSig      = "0x7a53080ba414158be7ec69b987b5fb7d07dee101fe85488f0853ae16239d0bde"
	pancakeBurnV3EventSig      = "0x0c396cd989a39f4459b5fa1aed6a9a8dcdbc45908acfd67e028cd568da98982c"
	pancakePairCreatedEventSig = "0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9"
	pancakePoolCreatedEventSig = "0x783cca1c0412dd0d695e784568c96da2e9c22ff989357a2e8b1d9b2b4e6b7118"
)

// PancakeSwapExtractor parses PancakeSwap V2/V3 DEX events on BSC and Ethereum.
type PancakeSwapExtractor struct {
	pairCache sync.Map // 表示该pool合约是否由V2/V3 factory合约创建
	Client    *bscProcessor.BSCProcessor
	*dex.EVMDexExtractor
}

// NewPancakeSwapExtractor creates a PancakeSwap extractor with EVM base class.
func NewPancakeSwapExtractor() *PancakeSwapExtractor {
	cfg := &dex.BaseDexExtractorConfig{
		Protocols:        []string{"pancakeswap", "pancakeswap-v2", "pancakeswap-v3"},
		SupportedChains:  []types.ChainType{types.ChainTypeBSC},
		LoggerModuleName: "dex-pancakeswap",
	}
	return &PancakeSwapExtractor{
		EVMDexExtractor: dex.NewEVMDexExtractor(cfg),
	}
}

// SetBSCProcessor 设置EVM处理器（用于获取链上数据）
func (b *PancakeSwapExtractor) SetBSCProcessor(processor interface{}) {
	if bscProcessor, ok := processor.(*bscProcessor.BSCProcessor); ok {
		b.Client = bscProcessor
	}
}

func (p *PancakeSwapExtractor) ExtractDexData(ctx context.Context, blocks []types.UnifiedBlock) (*types.DexData, error) {
	dexData := &types.DexData{
		Pools:        make([]model.Pool, 0),
		Transactions: make([]model.Transaction, 0),
		Liquidities:  make([]model.Liquidity, 0),
		Reserves:     make([]model.Reserve, 0),
		Tokens:       make([]model.Token, 0),
	}

	for _, block := range blocks {
		if !p.IsChainSupported(block.ChainType) {
			continue
		}

		for _, tx := range block.Transactions {
			ethLogs := dex.ExtractEVMLogsFromTransaction(&tx)
			if len(ethLogs) == 0 {
				continue
			}

			swapIdx := int64(0)
			for _, log := range ethLogs {
				if !p.isPancakeSwapLog(log) {
					continue
				}

				logType := p.getLogType(log)
				eventIndex := dex.ExtractEventIndex(log)
				switch logType {
				case "swap_v2":
					if modelTx := p.parseV2Swap(log, &tx, eventIndex, swapIdx); modelTx != nil {
						dexData.Transactions = append(dexData.Transactions, *modelTx)
						swapIdx++
					}
				case "swap_v3":
					if modelTx := p.parseV3Swap(log, &tx, eventIndex, swapIdx); modelTx != nil {
						dexData.Transactions = append(dexData.Transactions, *modelTx)
						swapIdx++
					}
				case "mint":
					if liq := p.parseLiquidity(log, &tx, "add", eventIndex); liq != nil {
						dexData.Liquidities = append(dexData.Liquidities, *liq)
					}
				case "burn":
					if liq := p.parseLiquidity(log, &tx, "remove", eventIndex); liq != nil {
						dexData.Liquidities = append(dexData.Liquidities, *liq)
					}
				case "pair_created":
					if pool := p.parseV2PairCreated(log, &tx); pool != nil {
						dexData.Pools = append(dexData.Pools, *pool)
					}
				case "pool_created":
					if pool := p.parseV3PoolCreated(log, &tx); pool != nil {
						dexData.Pools = append(dexData.Pools, *pool)
					}
				}
			}
		}
	}

	return dexData, nil
}

func (p *PancakeSwapExtractor) SupportsBlock(block *types.UnifiedBlock) bool {
	if !p.IsChainSupported(block.ChainType) {
		return false
	}
	for _, tx := range block.Transactions {
		ethLogs := dex.ExtractEVMLogsFromTransaction(&tx)
		for _, log := range ethLogs {
			if p.isPancakeSwapLog(log) {
				return true
			}
		}
	}
	return false
}

// 通过 Client 调用 Pair 合约的 factory() 方法
func (p *PancakeSwapExtractor) getPairFactory(pairAddr common.Address) (common.Address, error) {
	if p.Client == nil {
		return common.Address{}, fmt.Errorf("client is nil")
	}

	// 构造调用数据：只有 4 字节的函数选择器，无参数
	// callData := []byte{0xc45a0155}
	factorySelector := crypto.Keccak256([]byte("factory()"))[:4]

	// 调用合约
	result, err := p.Client.ReadContract(pairAddr, factorySelector)
	if err != nil {
		return common.Address{}, err
	}

	// 解析返回值：32 字节，address 类型左对齐
	if len(result) < 32 {
		return common.Address{}, fmt.Errorf("invalid factory return data")
	}

	return common.BytesToAddress(result[12:32]), nil
}

// isPancakeSwapLog 检查日志是否为 PancakeSwap V2/V3 事件
func (p *PancakeSwapExtractor) isPancakeSwapLog(log *ethtypes.Log) bool {
	if len(log.Topics) == 0 {
		return false
	}

	topic0 := log.Topics[0].Hex()

	createPoolEvents := []string{pancakePairCreatedEventSig, pancakePoolCreatedEventSig}
	poolEvents := []string{pancakeSwapV2EventSig, pancakeMintV2EventSig, pancakeBurnV2EventSig, pancakeSwapV3EventSig, pancakeMintV3EventSig, pancakeBurnV3EventSig}
	pancakeFactoryAddrs := []string{strings.ToLower(pancakeSwapV2FactoryAddr), strings.ToLower(pancakeSwapV3FactoryAddr)}

	logAddr := log.Address
	logAddrStr := strings.ToLower(log.Address.Hex())
	// PairCreated/PoolCreated 由 Factory 发出，检查 log.Address
	if dex.IsElementInSlice(topic0, createPoolEvents) {
		return dex.IsElementInSlice(logAddrStr, pancakeFactoryAddrs)
	}

	// Swap/Mint/Burn 由 Pair/Pool 发出
	if dex.IsElementInSlice(topic0, poolEvents) {
		if value, ok := p.pairCache.Load(logAddrStr); ok {
			return value.(bool)
		}

		// 获取factory合约地址
		factoryAddr, err := p.getPairFactory(logAddr)
		if err != nil {
			p.pairCache.Store(logAddrStr, false)
			return false
		}

		factoryAddrStr := strings.ToLower(factoryAddr.Hex())
		if dex.IsElementInSlice(factoryAddrStr, pancakeFactoryAddrs) {
			p.pairCache.Store(logAddrStr, true)
			return true
		}

		p.pairCache.Store(logAddrStr, false)
	}

	return false
}

func (p *PancakeSwapExtractor) getLogType(log *ethtypes.Log) string {
	if len(log.Topics) == 0 {
		return ""
	}
	topic0 := log.Topics[0].Hex()
	switch topic0 {
	case pancakeSwapV2EventSig:
		return "swap_v2"
	case pancakeSwapV3EventSig:
		return "swap_v3"
	case pancakeMintV2EventSig, pancakeMintV3EventSig:
		return "mint"
	case pancakeBurnV2EventSig, pancakeBurnV3EventSig:
		return "burn"
	case pancakePairCreatedEventSig:
		return "pair_created"
	case pancakePoolCreatedEventSig:
		return "pool_created"
	default:
		return ""
	}
}

/*
 * parseV2Swap 解析 PancakeSwap V2 的 Swap 事件
 *
 * 事件签名：
 *   Swap(address indexed sender, uint256 amount0In, uint256 amount1In, uint256 amount0Out, uint256 amount1Out, address indexed to)
 *
 * 事件含义：
 *   这是 AMM（自动做市商）流动性池中的核心交换事件，当用户通过流动性池进行代币兑换时触发。
 *   恒定乘积公式：x * y = k（token0 数量 × token1 数量 = 常数）
 *
 * 事件参数说明：
 *   sender (indexed)  - 发起交易的地址（调用 Router 合约的用户），indexed 表示该参数可作为事件过滤器
 *   amount0In         - 输入的 token0 数量（wei），如果没有输入则为 0
 *   amount1In         - 输入的 token1 数量（wei），如果没有输入则为 0
 *   amount0Out        - 输出的 token0 数量（wei），如果没有输出则为 0
 *   amount1Out        - 输出的 token1 数量（wei），如果没有输出则为 0
 *   to (indexed)      - 接收输出代币的目标地址
 *
 * 交换方向判定：
 *   - 如果 amount0In > 0：用户用 token0 换取 token1（token0 → token1）
 *   - 如果 amount1In > 0：用户用 token1 换取 token0（token1 → token0）
 *   - amountXIn 和 amountXOut 不会同时为非零值（同一代币不会同时输入和输出）
 *
 * 实际案例：
 *   用户用 1 BNB 换取 USDT：
 *   - sender = 0x1234...abcd（用户地址）
 *   - amount0In = 1000000000000000000（1 BNB，假设 token0 是 WBNB）
 *   - amount1In = 0
 *   - amount0Out = 0
 *   - amount1Out = 500000000（500 USDT，假设 token1 是 USDT）
 *   - to = 0x1234...abcd（接收地址，通常是用户自己）
 */
func (p *PancakeSwapExtractor) parseV2Swap(log *ethtypes.Log, tx *types.UnifiedTransaction, logIdx, swapIdx int64) *model.Transaction {
	if len(log.Data) < 128 {
		p.GetLogger().WithField("tx_hash", tx.TxHash).Warn("V2 swap log data too short")
		return nil
	}

	amount0In := new(big.Int).SetBytes(log.Data[0:32])
	amount1In := new(big.Int).SetBytes(log.Data[32:64])
	amount0Out := new(big.Int).SetBytes(log.Data[64:96])
	amount1Out := new(big.Int).SetBytes(log.Data[96:128])

	// Determine swap direction: nonzero amountIn side is what the user paid
	var amountIn, amountOut *big.Int
	var direction int // 0 = token0->token1, 1 = token1->token0
	if amount0In.Sign() > 0 {
		amountIn = amount0In
		amountOut = amount1Out
		direction = 0
	} else {
		amountIn = amount1In
		amountOut = amount0Out
		direction = 1
	}

	price := dex.CalcPrice(amountIn, amountOut)
	value := p.estimateV2SwapValue(amount0In, amount1In, amount0Out, amount1Out, direction)
	poolAddr := log.Address.Hex()

	return &model.Transaction{
		Addr:        poolAddr,
		Router:      tx.ToAddress,
		Factory:     pancakeSwapV2FactoryAddr,
		Pool:        poolAddr,
		Hash:        tx.TxHash,
		From:        tx.FromAddress,
		Side:        "swap",
		Amount:      amountIn,
		Price:       price,
		Value:       value,
		Time:        uint64(tx.Timestamp.Unix()),
		EventIndex:  logIdx,
		TxIndex:     int64(tx.TxIndex),
		SwapIndex:   swapIdx,
		BlockNumber: dex.GetBlockNumber(tx),
		Protocol:    "pancakeswap-v2",
		ChainType:   string(types.ChainTypeBSC),
		Extra: &model.TransactionExtra{
			QuotePrice: fmt.Sprintf("%.18f", price),
			Type:       "swap",
		},
	}
}

// estimateV2SwapValue estimates the USD value of a V2 swap using QuoteAssets.
// For V2 swaps we don't have token addresses in the event, so we use the raw
// amount * price calculation as a fallback. This provides a relative value that
// can be refined when pool token addresses are known from PairCreated events.
func (p *PancakeSwapExtractor) estimateV2SwapValue(amount0In, amount1In, amount0Out, amount1Out *big.Int, direction int) float64 {
	var amountIn, amountOut *big.Int
	if direction == 0 {
		amountIn = amount0In
		amountOut = amount1Out
	} else {
		amountIn = amount1In
		amountOut = amount0Out
	}

	price := dex.CalcPrice(amountIn, amountOut)
	return dex.CalcValue(amountIn, price)
}

/*
 * parseV3Swap 解析 PancakeSwap V3 的 Swap 事件
 *
 * 事件签名：
 *   Swap(address indexed sender, address indexed recipient, int256 amount0, int256 amount1, uint160 sqrtPriceX96, uint128 liquidity, int24 tick)
 *
 * 事件含义：
 *   PancakeSwap V3 采用了集中流动性的设计，流动性只在特定价格区间内提供。
 *   与 V2 的恒定乘积公式 x * y = k 不同，V3 使用更复杂的价格和流动性管理。
 *
 * 事件参数说明：
 *   sender (indexed)     - 发起交易的地址（调用 Router 合约的用户）
 *   recipient (indexed)  - 接收输出代币的目标地址
 *   amount0              - token0 的变动量（有符号整数）
 *   amount1              - token1 的变动量（有符号整数）
 *   sqrtPriceX96         - 交易后的平方根价格，格式为 sqrt(price) * 2^96
 *   liquidity            - 交易后的流动性池活跃流动性数量
 *   tick                 - 交易后的价格刻度（对应 sqrtPriceX96）
 *
 * 有符号整数含义（V3 与 V2 的关键区别）：
 *   amount0/amount1 使用 int256（有符号）：
 *   - 正值（> 0）：代币流入池子（用户支付）
 *   - 负值（< 0）：代币流出池子（用户接收）
 *
 *   示例：用户用 token0 换 token1
 *   - amount0 = 1000000000（正值，用户支付 1000 token0）
 *   - amount1 = -500000000（负值，用户收到 500 token1）
 *
 * sqrtPriceX96 价格计算：
 *   - 实际价格 = (sqrtPriceX96 / 2^96)²
 *   - 使用 Q96.96 定点数格式（96位小数）
 *   - 示例：sqrtPriceX96 = 1000000... → 价格 = (1000000 / 2^96)²
 *
 * tick 价格刻度：
 *   - tick 与 sqrtPriceX96 的关系：sqrtPriceX96 = 2^96 * 1.0001^(tick/2)
 *   - tick 必须是整数，价格只能在离散的 tick 值之间跳变
 *   - 每个 tick 代表 0.01% 的价格变化（1.0001 的倍数）
 *
 * liquidity 流动性：
 *   - 表示当前价格点位的活跃流动性
 *   - 随着价格移动，流动性会在不同价格区间激活/停用
 *
 * Topics 数组结构：
 *   Topics[0]: 事件签名 keccak256("Swap(address,address,int256,int256,uint160,uint128,int24)")
 *   Topics[1]: sender (indexed)
 *   Topics[2]: recipient (indexed)
 *   Data: amount0(32) + amount1(32) + sqrtPriceX96(32) + liquidity(32) + tick(32) = 160字节
	 *   注意：所有 non-indexed 参数在 ABI 编码中都会被 padding 到 32 字节，无论原始类型大小
*/
func (p *PancakeSwapExtractor) parseV3Swap(log *ethtypes.Log, tx *types.UnifiedTransaction, logIdx, swapIdx int64) *model.Transaction {
	if len(log.Data) < 160 {
		p.GetLogger().WithField("tx_hash", tx.TxHash).Warn("V3 swap log data too short")
		return nil
	}

	// V3 Swap amounts are signed int256:
	//   positive = flows into pool (user pays)
	//   negative = flows out of pool (user receives)
	amount0 := dex.ToSignedInt256(log.Data[0:32])
	amount1 := dex.ToSignedInt256(log.Data[32:64])
	sqrtPriceX96 := new(big.Int).SetBytes(log.Data[64:96])

	// Pick the positive amount as amountIn (what user paid)
	amountIn := new(big.Int).Abs(amount0)
	amountOut := new(big.Int).Abs(amount1)
	if amount0.Sign() < 0 {
		amountIn, amountOut = amountOut, amountIn
	}

	price := dex.CalcV3Price(sqrtPriceX96)
	value := dex.CalcValue(amountIn, price)
	poolAddr := log.Address.Hex()

	return &model.Transaction{
		Addr:        poolAddr,
		Router:      tx.ToAddress,
		Factory:     pancakeSwapV3FactoryAddr,
		Pool:        poolAddr,
		Hash:        tx.TxHash,
		From:        tx.FromAddress,
		Side:        "swap",
		Amount:      amountIn,
		Price:       price,
		Value:       value,
		Time:        uint64(tx.Timestamp.Unix()),
		EventIndex:  logIdx,
		TxIndex:     int64(tx.TxIndex),
		SwapIndex:   swapIdx,
		BlockNumber: dex.GetBlockNumber(tx),
		Protocol:    "pancakeswap-v3",
		ChainType:   string(types.ChainTypeBSC),
		Extra: &model.TransactionExtra{
			QuotePrice: fmt.Sprintf("%.18f", price),
			Type:       "swap",
		},
	}
}

/*
 * parseLiquidity 解析 PancakeSwap V2/V3 的 Mint（添加流动性）和 Burn（移除流动性）事件
 *
 * ==================== V2 事件 ====================
 *
 * V2 Mint 事件签名：
 *   Mint(address indexed sender, uint256 amount0, uint256 amount1)
 *
 *   含义：流动性提供者向 V2 池子存入两种代币，获得 LP 代币作为凭证
 *   参数说明：
 *     sender (indexed) - 触发 Mint 的地址（通常是 Router 合约）
 *     amount0          - 存入的 token0 数量（wei）
 *     amount1          - 存入的 token1 数量（wei）
 *
 *   Topics: [0]=事件签名, [1]=sender
 *   Data:   amount0(32字节) + amount1(32字节) = 64字节
 *
 *   示例：用户添加 1 BNB + 2000 USDT 流动性
 *     sender  = 0x10ED43C7...（Router）
 *     amount0 = 1000000000000000000  （1 BNB）
 *     amount1 = 2000000000000000000  （2000 USDT）
 *
 * V2 Burn 事件签名：
 *   Burn(address indexed sender, uint256 amount0, uint256 amount1, address indexed to)
 *
 *   含义：流动性提供者销毁 LP 代币，取回两种代币
 *   参数说明：
 *     sender (indexed) - 触发 Burn 的地址（通常是 Router 合约）
 *     amount0          - 取回的 token0 数量（wei）
 *     amount1          - 取回的 token1 数量（wei）
 *     to (indexed)     - 接收代币的目标地址（流动性提供者）
 *
 *   Topics: [0]=事件签名, [1]=sender, [2]=to
 *   Data:   amount0(32字节) + amount1(32字节) = 64字节
 *
 * ==================== V3 事件 ====================
 *
 * V3 Mint 事件签名：
 *   Mint(address sender, address indexed owner, int24 indexed tickLower, int24 indexed tickUpper, uint128 amount, uint256 amount0, uint256 amount1)
 *
 *   含义：流动性提供者在 V3 指定价格区间 [tickLower, tickUpper] 内添加流动性
 *   参数说明：
 *     sender             - 触发 Mint 的地址（Router 合约，non-indexed）
 *     owner (indexed)    - 流动性头寸的所有者地址
 *     tickLower (indexed)- 价格区间下界对应的 tick 值
 *     tickUpper (indexed)- 价格区间上界对应的 tick 值
 *     amount             - 添加的流动性数量（L 值，非代币数量）
 *     amount0            - 存入的 token0 数量（wei）
 *     amount1            - 存入的 token1 数量（wei）
 *
 *   Topics: [0]=事件签名, [1]=owner, [2]=tickLower, [3]=tickUpper
 *   Data:   sender(32字节) + amount(32字节) + amount0(32字节) + amount1(32字节) = 128字节
 *
 *   示例：用户在 tick [88200, 89000] 区间添加流动性
 *     sender    = 0x10ED43C7...（Router）
 *     owner     = 0x1234...abcd（用户地址）
 *     tickLower = 88200
 *     tickUpper = 89000
 *     amount    = 5000000000000000000（流动性 L 值）
 *     amount0   = 1000000000000000000  （1 token0）
 *     amount1   = 2000000000000000000  （2 token1）
 *
 * V3 Burn 事件签名：
 *   Burn(address indexed owner, int24 indexed tickLower, int24 indexed tickUpper, uint128 amount, uint256 amount0, uint256 amount1)
 *
 *   含义：流动性提供者从指定价格区间 [tickLower, tickUpper] 移除流动性
 *   参数说明：
 *     owner (indexed)    - 流动性头寸的所有者地址
 *     tickLower (indexed)- 价格区间下界对应的 tick 值
 *     tickUpper (indexed)- 价格区间上界对应的 tick 值
 *     amount             - 移除的流动性数量（L 值）
 *     amount0            - 取回的 token0 数量（wei）
 *     amount1            - 取回的 token1 数量（wei）
 *
 *   Topics: [0]=事件签名, [1]=owner, [2]=tickLower, [3]=tickUpper
 *   Data:   amount(32字节) + amount0(32字节) + amount1(32字节) = 96字节
 *
 * ==================== V2 vs V3 对比 ====================
 *
 * | 特性         | V2                        | V3                                |
 * |-------------|---------------------------|-----------------------------------|
 * | 流动性范围   | 全价格区间 (0, ∞)         | 自定义区间 [tickLower, tickUpper] |
 * | LP 凭证     | ERC20 LP 代币（可转让）     | NFT 头寸（不可转让，不可分割）     |
 * | amount 含义  | 等同于代币数量              | 流动性 L 值（非代币数量）          |
 * | Mint sender | indexed                   | non-indexed                       |
 */
func (p *PancakeSwapExtractor) parseLiquidity(log *ethtypes.Log, tx *types.UnifiedTransaction, side string, logIdx int64) *model.Liquidity {
	if len(log.Data) < 64 {
		return nil
	}

	topic0 := log.Topics[0].Hex()

	var amount0, amount1 *big.Int
	switch topic0 {
	case pancakeMintV3EventSig:
		// V3 Mint data: [address sender, uint128 amount, uint256 amount0, uint256 amount1]
		if len(log.Data) >= 128 {
			amount0 = new(big.Int).SetBytes(log.Data[64:96])
			amount1 = new(big.Int).SetBytes(log.Data[96:128])
		}
	case pancakeBurnV3EventSig:
		// V3 Burn data: [uint128 amount, uint256 amount0, uint256 amount1]
		if len(log.Data) >= 96 {
			amount0 = new(big.Int).SetBytes(log.Data[32:64])
			amount1 = new(big.Int).SetBytes(log.Data[64:96])
		}
	default:
		// V2 Mint/Burn data: [uint256 amount0, uint256 amount1]
		amount0 = new(big.Int).SetBytes(log.Data[0:32])
		amount1 = new(big.Int).SetBytes(log.Data[32:64])
	}

	if amount0 == nil {
		amount0 = big.NewInt(0)
	}
	if amount1 == nil {
		amount1 = big.NewInt(0)
	}

	totalAmount := new(big.Int).Add(amount0, amount1)
	val0, _ := new(big.Float).SetInt(amount0).Float64()
	val1, _ := new(big.Float).SetInt(amount1).Float64()

	poolAddr := log.Address.Hex()
	key := fmt.Sprintf("%s_%s_%d", tx.TxHash, side, logIdx)

	return &model.Liquidity{
		Addr:      poolAddr,
		Router:    tx.ToAddress,
		Factory:   p.getFactoryAddress(log),
		Pool:      poolAddr,
		Hash:      tx.TxHash,
		From:      tx.FromAddress,
		Side:      side,
		Amount:    totalAmount,
		Value:     val0 + val1,
		Time:      uint64(tx.Timestamp.Unix()),
		Key:       key,
		Protocol:  "pancakeswap",
		ChainType: string(types.ChainTypeBSC),
		Extra: &model.LiquidityExtra{
			Key:     key,
			Amounts: amount1,
			Values:  []float64{val0, val1},
			Time:    uint64(tx.Timestamp.Unix()),
		},
	}
}

/*
 * parseV2PairCreated 解析 PancakeSwap V2 的 PairCreated（交易对创建）事件
 *
 * 事件签名：
 *   PairCreated(address indexed token0, address indexed token1, address pair, uint256)
 *
 * 事件含义：
 *   当通过 PancakeSwap V2 Factory 合约创建新的流动性池（交易对）时触发。
 *   这是整个 V2 生态的起点 —— 只有先创建 Pair，才能在其中添加流动性和进行兑换。
 *   每个代币组合（token0 + token1）只会创建一个 Pair 合约。
 *
 * 事件参数说明：
 *   token0 (indexed) - 交易对中的第一种代币地址，按地址数值大小排序（地址值较小者为 token0）
 *   token1 (indexed) - 交易对中的第二种代币地址，按地址数值大小排序（地址值较大者为 token1）
 *   pair             - 新创建的 Pair 合约（流动性池）地址
 *   (未命名)          - 第 4 个参数在原合约中为 totalSupply，通常未使用或为 0
 *
 * Topics 数组结构：
 *   Topics[0]: 事件签名 keccak256("PairCreated(address,address,address,uint256)")
 *   Topics[1]: token0 (indexed)
 *   Topics[2]: token1 (indexed)
 *   Data:      pair(32字节) + 第4参数(32字节) = 64字节
 *
 * token0/token1 排序规则：
 *   PancakeSwap/Uniswap V2 中，token0 和 token1 按地址数值升序排列：
 *   - token0 = 地址值较小的代币
 *   - token1 = 地址值较大的代币
 *   示例：WBNB(0x00...) < USDT(0x55...)，所以 WBNB 是 token0，USDT 是 token1
 *
 * 与 parseV2Swap 的关联：
 *   Swap 事件中只有 amount0In/amount1In/amount0Out/amount1Out，不包含代币地址。
 *   需要通过 PairCreated 事件记录 token0/token1 地址，才能知道 Swap 中每种代币具体是什么。
 *
 * 实际案例：
 *   创建 WBNB/USDT 交易对：
 *   - token0 = 0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c（WBNB，地址值较小）
 *   - token1 = 0x55d398326f99059fF775485246999027B3197955（USDT，地址值较大）
 *   - pair   = 0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE（新创建的 Pair 地址）
 */
func (p *PancakeSwapExtractor) parseV2PairCreated(log *ethtypes.Log, tx *types.UnifiedTransaction) *model.Pool {
	if len(log.Topics) < 3 || len(log.Data) < 64 {
		return nil
	}

	token0 := common.BytesToAddress(log.Topics[1].Bytes()).Hex()
	token1 := common.BytesToAddress(log.Topics[2].Bytes()).Hex()
	pairAddr := common.BytesToAddress(log.Data[0:32]).Hex()

	return &model.Pool{
		Addr:      pairAddr,
		Factory:   pancakeSwapV2FactoryAddr,
		Protocol:  "pancakeswap-v2",
		ChainType: string(types.ChainTypeBSC),
		Tokens:    map[int]string{0: token0, 1: token1},
		Fee:       2500,
		Extra: &model.PoolExtra{
			Hash: tx.TxHash,
			From: tx.FromAddress,
			Time: uint64(tx.Timestamp.Unix()),
		},
	}
}

/*
 * parseV3PoolCreated 解析 PancakeSwap V3 的 PoolCreated（池子创建）事件
 *
 * 事件签名：
 *   PoolCreated(address indexed token0, address indexed token1, uint24 indexed fee, int24 tickSpacing, address pool)
 *
 * 事件含义：
 *   当通过 PancakeSwap V3 Factory 合约创建新的流动性池时触发。
 *   与 V2 的 PairCreated 相比，V3 增加了 fee（费率）和 tickSpacing（刻度间距）参数，
 *   因为同一个代币组合可以创建多个不同费率的池子。
 *
 * 事件参数说明：
 *   token0 (indexed)      - 交易对中的第一种代币地址（地址值较小者）
 *   token1 (indexed)      - 交易对中的第二种代币地址（地址值较大者）
 *   fee (indexed)         - 池子费率，单位为 hundredths of a bip（1/100 基点）
 *                          常见值：500=0.05%, 2500=0.25%, 3000=0.30%, 10000=1.00%
 *   tickSpacing           - 刻度间距，决定可用的价格刻度密度
 *                          常见值：10(0.05%), 50(0.25%), 60(0.30%), 200(1.00%)
 *   pool                  - 新创建的 Pool 合约地址
 *
 * Topics 数组结构：
 *   Topics[0]: 事件签名 keccak256("PoolCreated(address,address,uint24,int24,address)")
 *   Topics[1]: token0 (indexed)
 *   Topics[2]: token1 (indexed)
 *   Topics[3]: fee (indexed, uint24 类型，高位补零)
 *   Data:      tickSpacing(32字节) + pool(32字节) = 64字节
 *
 * V2 vs V3 池子创建对比：
 *   V2: 同一 token0+token1 组合只能创建一个 Pair，费率固定 0.25%
 *   V3: 同一 token0+token1 组合可按不同 fee 创建多个 Pool
 *       例：WBNB/USDT 可同时存在 0.05%、0.25%、1.00% 三个费率的池子
 *
 * fee 与 tickSpacing 的对应关系（PancakeSwap V3 默认）：
 *   fee=500   → tickSpacing=10  （低费率，高精度）
 *   fee=2500  → tickSpacing=50  （中等费率）
 *   fee=3000  → tickSpacing=60  （中等费率）
 *   fee=10000 → tickSpacing=200 （高费率，低精度）
 *
 * 实际案例：
 *   创建 WBNB/USDT 0.25% 费率池子：
 *   - token0      = 0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c（WBNB）
 *   - token1      = 0x55d398326f99059fF775485246999027B3197955（USDT）
 *   - fee         = 2500（0.25%）
 *   - tickSpacing = 50
 *   - pool        = 0x36696169c63e42cd08ce11f5deebbcebae652050（新 Pool 地址）
 */
func (p *PancakeSwapExtractor) parseV3PoolCreated(log *ethtypes.Log, tx *types.UnifiedTransaction) *model.Pool {
	if len(log.Topics) < 4 || len(log.Data) < 64 {
		return nil
	}

	token0 := common.BytesToAddress(log.Topics[1].Bytes()).Hex()
	token1 := common.BytesToAddress(log.Topics[2].Bytes()).Hex()
	fee := new(big.Int).SetBytes(log.Topics[3].Bytes())
	poolAddr := common.BytesToAddress(log.Data[32:64]).Hex()

	return &model.Pool{
		Addr:      poolAddr,
		Factory:   pancakeSwapV3FactoryAddr,
		Protocol:  "pancakeswap-v3",
		ChainType: string(types.ChainTypeBSC),
		Tokens:    map[int]string{0: token0, 1: token1},
		Fee:       int(fee.Int64()),
		Extra: &model.PoolExtra{
			Hash: tx.TxHash,
			From: tx.FromAddress,
			Time: uint64(tx.Timestamp.Unix()),
		},
	}
}

func (p *PancakeSwapExtractor) getFactoryAddress(log *ethtypes.Log) string {
	if len(log.Topics) == 0 {
		return ""
	}
	topic0 := log.Topics[0].Hex()
	switch topic0 {
	case pancakeSwapV2EventSig, pancakeMintV2EventSig, pancakeBurnV2EventSig, pancakePairCreatedEventSig:
		return pancakeSwapV2FactoryAddr
	case pancakeSwapV3EventSig, pancakeMintV3EventSig, pancakeBurnV3EventSig, pancakePoolCreatedEventSig:
		return pancakeSwapV3FactoryAddr
	default:
		return ""
	}
}

// isQuoteAsset checks if an address is a configured quote asset
func (p *PancakeSwapExtractor) isQuoteAsset(addr string) bool {
	if p.GetQuoteAssetRank(strings.ToLower(addr)) >= 0 {
		return true
	}
	return p.GetQuoteAssetRank(addr) >= 0
}

// getTokenSymbol returns a short symbol derived from the token address
func getTokenSymbol(tokenAddr string) string {
	if len(tokenAddr) >= 8 {
		return strings.ToUpper(tokenAddr[2:8])
	}
	return "UNKNOWN"
}
