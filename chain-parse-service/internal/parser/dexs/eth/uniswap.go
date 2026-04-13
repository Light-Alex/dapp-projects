package eth

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"unified-tx-parser/internal/model"
	ethereumProcessor "unified-tx-parser/internal/parser/chains/ethereum"
	dex "unified-tx-parser/internal/parser/dexs"
	"unified-tx-parser/internal/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// Uniswap V2 contract addresses
	uniswapV2FactoryAddr = "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"

	// Uniswap V3 contract addresses
	uniswapV3FactoryAddr = "0x1F98431c8aD98523631AE4a59f267346ea31F984"

	// Event signatures (shared with PancakeSwap)
	swapV2EventSig      = "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"
	swapV3EventSig      = "0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67"
	mintV2EventSig      = "0x4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f"
	burnV2EventSig      = "0xdccd412f0b1252819cb1fd330b93224ca42612892bb3f4f789976e6d81936496"
	mintV3EventSig      = "0x7a53080ba414158be7ec69b987b5fb7d07dee101fe85488f0853ae16239d0bde"
	burnV3EventSig      = "0x0c396cd989a39f4459b5fa1aed6a9a8dcdbc45908acfd67e028cd568da98982c"
	pairCreatedEventSig = "0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9"
	poolCreatedEventSig = "0x783cca1c0412dd0d695e784568c96da2e9c22ff989357a2e8b1d9b2b4e6b7118"
)

// UniswapExtractor parses Uniswap V2/V3 DEX events on Ethereum and BSC.
type UniswapExtractor struct {
	pairCache sync.Map // 表示该pool合约是否由V2/V3 factory合约创建
	Client    *ethereumProcessor.EthereumProcessor
	*dex.EVMDexExtractor
}

// NewUniswapExtractor creates a Uniswap extractor with EVM base class.
func NewUniswapExtractor() *UniswapExtractor {
	cfg := &dex.BaseDexExtractorConfig{
		Protocols:        []string{"uniswap", "uniswap-v2", "uniswap-v3"},
		SupportedChains:  []types.ChainType{types.ChainTypeEthereum},
		LoggerModuleName: "dex-uniswap",
	}
	return &UniswapExtractor{
		EVMDexExtractor: dex.NewEVMDexExtractor(cfg),
	}
}

// SetEthereumProcessor 设置EVM处理器（用于获取链上数据）
func (b *UniswapExtractor) SetEthereumProcessor(processor interface{}) {
	if ethereumProcessor, ok := processor.(*ethereumProcessor.EthereumProcessor); ok {
		b.Client = ethereumProcessor
	}
}

func (u *UniswapExtractor) ExtractDexData(ctx context.Context, blocks []types.UnifiedBlock) (*types.DexData, error) {
	dexData := &types.DexData{
		Pools:        make([]model.Pool, 0),
		Transactions: make([]model.Transaction, 0),
		Liquidities:  make([]model.Liquidity, 0),
		Reserves:     make([]model.Reserve, 0),
		Tokens:       make([]model.Token, 0),
	}

	for _, block := range blocks {
		if !u.IsChainSupported(block.ChainType) {
			continue
		}

		u.GetLogger().Debugf("processing block %s with %d transactions", block.BlockNumber.String(), len(block.Transactions))

		for _, tx := range block.Transactions {
			// FIX #4: Use shared ExtractEVMLogsFromTransaction instead of duplicate code
			ethLogs := dex.ExtractEVMLogsFromTransaction(&tx)
			if len(ethLogs) == 0 {
				continue
			}

			// FIX #2: Track swapIdx per transaction, pass logIdx as eventIndex
			swapIdx := int64(0)
			for _, log := range ethLogs {
				if !u.isUniswapLog(log) {
					continue
				}

				logType := u.getLogType(log)
				eventIndex := dex.ExtractEventIndex(log)
				u.GetLogger().Debugf("found uniswap log, type: %s, address: %s", logType, log.Address.Hex())

				switch logType {
				case "swap_v2":
					if modelTx := u.parseV2Swap(log, &tx, eventIndex, swapIdx); modelTx != nil {
						dexData.Transactions = append(dexData.Transactions, *modelTx)
						swapIdx++
					}
				case "swap_v3":
					if modelTx := u.parseV3Swap(log, &tx, eventIndex, swapIdx); modelTx != nil {
						dexData.Transactions = append(dexData.Transactions, *modelTx)
						swapIdx++
					}
				case "mint":
					if liq := u.parseLiquidity(log, &tx, "add", eventIndex); liq != nil {
						dexData.Liquidities = append(dexData.Liquidities, *liq)
					}
				case "burn":
					if liq := u.parseLiquidity(log, &tx, "remove", eventIndex); liq != nil {
						dexData.Liquidities = append(dexData.Liquidities, *liq)
					}
				case "pair_created":
					if pool := u.parseV2PairCreated(log, &tx); pool != nil {
						dexData.Pools = append(dexData.Pools, *pool)
					}
				case "pool_created":
					if pool := u.parseV3PoolCreated(log, &tx); pool != nil {
						dexData.Pools = append(dexData.Pools, *pool)
					}
				}
			}
		}
	}

	return dexData, nil
}

func (u *UniswapExtractor) SupportsBlock(block *types.UnifiedBlock) bool {
	if !u.IsChainSupported(block.ChainType) {
		return false
	}
	for _, tx := range block.Transactions {
		ethLogs := dex.ExtractEVMLogsFromTransaction(&tx)
		for _, log := range ethLogs {
			if u.isUniswapLog(log) {
				return true
			}
		}
	}
	return false
}

// 通过 Client 调用 Pair 合约的 factory() 方法
func (p *UniswapExtractor) getPairFactory(pairAddr common.Address) (common.Address, error) {
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

func (u *UniswapExtractor) isUniswapLog(log *ethtypes.Log) bool {
	if len(log.Topics) == 0 {
		return false
	}
	topic0 := log.Topics[0].Hex()

	createPoolEvents := []string{pairCreatedEventSig, poolCreatedEventSig}
	poolEvents := []string{swapV2EventSig, swapV3EventSig, mintV2EventSig, burnV2EventSig, mintV3EventSig, burnV3EventSig}
	uniswapFactoryAddrs := []string{strings.ToLower(uniswapV2FactoryAddr), strings.ToLower(uniswapV3FactoryAddr)}

	logAddr := log.Address
	logAddrStr := strings.ToLower(log.Address.Hex())
	// PairCreated/PoolCreated 由 Factory 发出，检查 log.Address
	if dex.IsElementInSlice(topic0, createPoolEvents) {
		return dex.IsElementInSlice(logAddrStr, uniswapFactoryAddrs)
	}

	// Swap/Mint/Burn 由 Pair/Pool 发出
	if dex.IsElementInSlice(topic0, poolEvents) {
		if value, ok := u.pairCache.Load(logAddrStr); ok {
			if boolVal, isBool := value.(bool); isBool {
				return boolVal
			}
		}

		// 获取factory合约地址
		factoryAddr, err := u.getPairFactory(logAddr)
		if err != nil {
			u.pairCache.Store(logAddrStr, false)
			return false
		}

		factoryAddrStr := strings.ToLower(factoryAddr.Hex())
		if dex.IsElementInSlice(factoryAddrStr, uniswapFactoryAddrs) {
			u.pairCache.Store(logAddrStr, true)
			return true
		}

		u.pairCache.Store(logAddrStr, false)
	}

	return false

}

func (u *UniswapExtractor) getLogType(log *ethtypes.Log) string {
	if len(log.Topics) == 0 {
		return ""
	}
	topic0 := log.Topics[0].Hex()
	switch topic0 {
	case swapV2EventSig:
		return "swap_v2"
	case swapV3EventSig:
		return "swap_v3"
	case mintV2EventSig, mintV3EventSig:
		return "mint"
	case burnV2EventSig, burnV3EventSig:
		return "burn"
	case pairCreatedEventSig:
		return "pair_created"
	case poolCreatedEventSig:
		return "pool_created"
	default:
		return ""
	}
}

/*
 * parseV2Swap 解析 Uniswap V2 的 Swap 事件
 *
 * 事件签名：
 *   Swap(address indexed sender, uint256 amount0In, uint256 amount1In, uint256 amount0Out, uint256 amount1Out, address indexed to)
 *
 * 事件含义：
 *   这是 Uniswap V2 AMM 流动性池中的核心交换事件，当用户通过恒定乘积公式 x * y = k 进行代币兑换时触发。
 *   Uniswap V2 是 PancakeSwap V2 的原始版本，两者事件结构完全相同，但费率不同：
 *   Uniswap V2 固定 0.30% 费率，PancakeSwap V2 固定 0.25% 费率。
 *
 * 事件参数说明：
 *   sender (indexed)  - 发起交易的地址（调用 Router 合约的用户），indexed 表示可作为事件过滤器
 *   amount0In         - 输入的 token0 数量（wei），如果没有输入则为 0
 *   amount1In         - 输入的 token1 数量（wei），如果没有输入则为 0
 *   amount0Out        - 输出的 token0 数量（wei），如果没有输出则为 0
 *   amount1Out        - 输出的 token1 数量（wei），如果没有输出则为 0
 *   to (indexed)      - 接收输出代币的目标地址
 *
 * Topics 数组结构：
 *   Topics[0]: 事件签名 keccak256("Swap(address,uint256,uint256,uint256,uint256,address)")
 *   Topics[1]: sender (indexed)
 *   Topics[2]: to (indexed)
 *   Data:      amount0In(32字节) + amount1In(32字节) + amount0Out(32字节) + amount1Out(32字节) = 128字节
 *
 * 交换方向判定：
 *   - amount0In > 0：用户用 token0 换取 token1（token0 → token1）
 *   - amount1In > 0：用户用 token1 换取 token0（token1 → token0）
 *   - 同一代币不会同时输入和输出（amountXIn 和 amountXOut 不同时为非零值）
 *
 * 实际案例：
 *   用户用 1 ETH 换取 USDC：
 *   - sender    = 0x7a25...（Router 合约地址）
 *   - amount0In = 1000000000000000000（1 WETH，假设 token0 是 WETH）
 *   - amount1In = 0
 *   - amount0Out = 0
 *   - amount1Out = 3000000000（3000 USDC，假设 token1 是 USDC）
 *   - to        = 0x1234...abcd（用户地址）
 */
func (u *UniswapExtractor) parseV2Swap(log *ethtypes.Log, tx *types.UnifiedTransaction, logIdx, swapIdx int64) *model.Transaction {
	if len(log.Data) < 128 {
		u.GetLogger().WithField("tx_hash", tx.TxHash).Warn("V2 swap log data too short")
		return nil
	}

	amount0In := new(big.Int).SetBytes(log.Data[0:32])
	amount1In := new(big.Int).SetBytes(log.Data[32:64])
	amount0Out := new(big.Int).SetBytes(log.Data[64:96])
	amount1Out := new(big.Int).SetBytes(log.Data[96:128])

	var amountIn, amountOut *big.Int
	if amount0In.Sign() > 0 {
		amountIn = amount0In
		amountOut = amount1Out
	} else {
		amountIn = amount1In
		amountOut = amount0Out
	}

	price := dex.CalcPrice(amountIn, amountOut)
	value := dex.CalcValue(amountIn, price)
	poolAddr := log.Address.Hex()

	return &model.Transaction{
		Addr:        poolAddr,
		Router:      tx.ToAddress,
		Factory:     uniswapV2FactoryAddr,
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
		Protocol:    "uniswap-v2",
		ChainType:   string(types.ChainTypeEthereum),
		Extra: &model.TransactionExtra{
			QuotePrice: fmt.Sprintf("%.18f", price),
			Type:       "swap",
		},
	}
}

/*
 * parseV3Swap 解析 Uniswap V3 的 Swap 事件
 *
 * 事件签名：
 *   Swap(address indexed sender, address indexed recipient, int256 amount0, int256 amount1, uint160 sqrtPriceX96, uint128 liquidity, int24 tick)
 *
 * 事件含义：
 *   Uniswap V3 的集中流动性 AMM 池中的核心交换事件。与 V2 的恒定乘积公式不同，
 *   V3 允许流动性提供者在指定价格区间内集中提供流动性，提高资金利用率。
 *   事件结构与 PancakeSwap V3 相同。
 *
 * 事件参数说明：
 *   sender (indexed)     - 发起交易的地址（调用 Router 合约的用户）
 *   recipient (indexed)  - 接收输出代币的目标地址
 *   amount0              - token0 的变动量（有符号整数，int256）
 *   amount1              - token1 的变动量（有符号整数，int256）
 *   sqrtPriceX96         - 交易后的平方根价格，格式为 sqrt(price) * 2^96
 *   liquidity            - 交易后的活跃流动性数量（当前 tick 区间的 L 值）
 *   tick                 - 交易后的价格刻度
 *
 * 有符号整数含义（V3 与 V2 的关键区别）：
 *   amount0/amount1 使用 int256（有符号），正值和负值代表不同方向：
 *   - 正值（> 0）：代币流入池子（用户支付的代币）
 *   - 负值（< 0）：代币流出池子（用户收到的代币）
 *
 *   示例：用户用 WETH 换 USDC
 *   - amount0 =  1000000000000000000（正值，用户支付 1 WETH）
 *   - amount1 = -3000000000         （负值，用户收到 3000 USDC）
 *
 * sqrtPriceX96 价格编码：
 *   - 使用 Q96.96 定点数格式，将浮点数编码为整数
 *   - 实际价格 = (sqrtPriceX96 / 2^96)^2
 *   - 代表 token0 相对于 token1 的价格
 *
 * tick 价格刻度系统：
 *   - tick 与 sqrtPriceX96 的数学关系：sqrtPriceX96 = 2^96 * 1.0001^(tick/2)
 *   - 每个 tick 代表 0.01% 的价格变化（即价格比为 1.0001）
 *   - tick 必须是 tickSpacing 的整数倍
 *
 * liquidity 活跃流动性：
 *   - 当前价格所在 tick 区间内的流动性数量
 *   - 当价格跨越区间边界时，活跃流动性会发生变化
 *
 * Topics 数组结构：
 *   Topics[0]: 事件签名 keccak256("Swap(address,address,int256,int256,uint160,uint128,int24)")
 *   Topics[1]: sender (indexed)
 *   Topics[2]: recipient (indexed)
 *   Data:      amount0(32) + amount1(32) + sqrtPriceX96(32) + liquidity(32) + tick(32) = 160字节
 *   注意：所有 non-indexed 参数在 ABI 编码中都会被 padding 到 32 字节，无论原始类型大小
 *
 * Uniswap V3 vs PancakeSwap V3：
 *   两者事件结构完全相同，差异仅在部署链和默认费率档位不同
 */
func (u *UniswapExtractor) parseV3Swap(log *ethtypes.Log, tx *types.UnifiedTransaction, logIdx, swapIdx int64) *model.Transaction {
	if len(log.Data) < 160 {
		u.GetLogger().WithField("tx_hash", tx.TxHash).Warn("V3 swap log data too short")
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
		Factory:     uniswapV3FactoryAddr,
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
		Protocol:    "uniswap-v3",
		ChainType:   string(types.ChainTypeEthereum),
		Extra: &model.TransactionExtra{
			QuotePrice: fmt.Sprintf("%.18f", price),
			Type:       "swap",
		},
	}
}

/*
 * parseLiquidity 解析 Uniswap V2/V3 的 Mint（添加流动性）和 Burn（移除流动性）事件
 *
 * ==================== V2 事件 ====================
 *
 * V2 Mint 事件签名：
 *   Mint(address indexed sender, uint256 amount0, uint256 amount1)
 *
 *   含义：流动性提供者向 V2 池子存入两种代币，获得 ERC20 LP 代币作为凭证
 *   LP 代币代表该流动性池中的份额比例，可随时转让
 *   参数说明：
 *     sender (indexed) - 触发 Mint 的地址（通常是 Router 合约）
 *     amount0          - 存入的 token0 数量（wei）
 *     amount1          - 存入的 token1 数量（wei）
 *
 *   Topics: [0]=事件签名, [1]=sender
 *   Data:   amount0(32字节) + amount1(32字节) = 64字节
 *
 *   示例：用户添加 10 ETH + 30000 USDC 流动性
 *     sender  = 0x68b3...（Router 合约）
 *     amount0 = 10000000000000000000  （10 WETH）
 *     amount1 = 30000000000000000000  （30000 USDC）
 *
 * V2 Burn 事件签名：
 *   Burn(address indexed sender, uint256 amount0, uint256 amount1, address indexed to)
 *
 *   含义：流动性提供者销毁 LP 代币，按比例取回池子中的两种代币
 *   参数说明：
 *     sender (indexed) - 触发 Burn 的地址（通常是 Pair 合约自身）
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
 *   含义：流动性提供者在 V3 指定价格区间 [tickLower, tickUpper] 内添加集中流动性
 *   获得 NFT 头寸（ERC721）而非 ERC20 LP 代币，代表特定价格区间的流动性份额
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
 *   示例：用户在 tick [−887220, 887220] 区间添加流动性（接近全范围）
 *     sender    = 0x68b3...（Router 合约）
 *     owner     = 0x1234...abcd（用户地址）
 *     tickLower = -887220
 *     tickUpper = 887220
 *     amount    = 5000000000000000000（流动性 L 值）
 *     amount0   = 10000000000000000000 （10 WETH）
 *     amount1   = 30000000000000000000 （30000 USDC）
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
 * | LP 凭证     | ERC20 LP 代币（可转让）     | NFT 头寸（ERC721，不可分割）       |
 * | amount 含义  | 等同于代币数量              | 流动性 L 值（非代币数量）          |
 * | Mint sender | indexed                   | non-indexed                       |
 * | 同币对池数   | 唯一                      | 可按不同费率创建多个 Pool          |
 */
func (u *UniswapExtractor) parseLiquidity(log *ethtypes.Log, tx *types.UnifiedTransaction, side string, logIdx int64) *model.Liquidity {
	if len(log.Data) < 64 {
		return nil
	}

	topic0 := log.Topics[0].Hex()

	var amount0, amount1 *big.Int
	switch topic0 {
	case mintV3EventSig:
		// V3 Mint data: [address sender, uint128 amount, uint256 amount0, uint256 amount1]
		if len(log.Data) >= 128 {
			amount0 = new(big.Int).SetBytes(log.Data[64:96])
			amount1 = new(big.Int).SetBytes(log.Data[96:128])
		}
	case burnV3EventSig:
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
		Factory:   u.getFactoryAddress(log),
		Pool:      poolAddr,
		Hash:      tx.TxHash,
		From:      tx.FromAddress,
		Side:      side,
		Amount:    totalAmount,
		Value:     val0 + val1,
		Time:      uint64(tx.Timestamp.Unix()),
		Key:       key,
		Protocol:  "uniswap",
		ChainType: string(types.ChainTypeEthereum),
		Extra: &model.LiquidityExtra{
			Key:     key,
			Amounts: amount1,
			Values:  []float64{val0, val1},
			Time:    uint64(tx.Timestamp.Unix()),
		},
	}
}

/*
 * parseV2PairCreated 解析 Uniswap V2 的 PairCreated（交易对创建）事件
 *
 * 事件签名：
 *   PairCreated(address indexed token0, address indexed token1, address pair, uint256)
 *
 * 事件含义：
 *   当通过 Uniswap V2 Factory 合约创建新的流动性池（交易对）时触发。
 *   这是 V2 生态的起点 —— 先创建 Pair，才能添加流动性和进行兑换。
 *   每个代币组合只会创建一个 Pair（与 V3 不同，V3 可创建多个不同费率的 Pool）。
 *
 * 事件参数说明：
 *   token0 (indexed) - 交易对中的第一种代币地址，按地址数值升序排列（地址值较小者为 token0）
 *   token1 (indexed) - 交易对中的第二种代币地址，按地址数值升序排列（地址值较大者为 token1）
 *   pair             - 新创建的 Pair 合约（流动性池）地址
 *   (未命名)          - 第 4 个参数在原合约中未使用，通常为 0
 *
 * Topics 数组结构：
 *   Topics[0]: 事件签名 keccak256("PairCreated(address,address,address,uint256)")
 *   Topics[1]: token0 (indexed)
 *   Topics[2]: token1 (indexed)
 *   Data:      pair(32字节) + 第4参数(32字节) = 64字节
 *
 * token0/token1 排序规则：
 *   按地址数值升序排列（地址值较小者为 token0）：
 *   示例：WETH(0xC02a...) > USDC(0xA0b8...)，所以 USDC 是 token0，WETH 是 token1
 *
 * Pair 地址来源：
 *   使用 Data[0:32] 而非 log.Address 获取 Pair 地址。
 *   log.Address 是发出事件的合约地址（即 Factory），Data 中的 pair 才是新创建的 Pair 地址。
 *
 * 与 Swap 事件的关联：
 *   Swap 事件只包含 amount0In/amount1In/amount0Out/amount1Out，不包含代币地址。
 *   需要通过 PairCreated 记录的 token0/token1 地址，才能确定 Swap 中各代币的具体类型。
 *
 * Uniswap V2 费率：固定 0.30%（3000 hundredths of a bip）
 *
 * 实际案例：
 *   创建 USDC/WETH 交易对：
 *   - token0 = 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48（USDC，地址值较小）
 *   - token1 = 0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2（WETH，地址值较大）
 *   - pair   = 0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc（新创建的 Pair 地址）
 */
func (u *UniswapExtractor) parseV2PairCreated(log *ethtypes.Log, tx *types.UnifiedTransaction) *model.Pool {
	if len(log.Topics) < 3 || len(log.Data) < 64 {
		return nil
	}

	token0 := common.BytesToAddress(log.Topics[1].Bytes()).Hex()
	token1 := common.BytesToAddress(log.Topics[2].Bytes()).Hex()
	pairAddr := common.BytesToAddress(log.Data[0:32]).Hex()

	return &model.Pool{
		Addr:      pairAddr,
		Factory:   uniswapV2FactoryAddr,
		Protocol:  "uniswap-v2",
		ChainType: string(types.ChainTypeEthereum),
		Tokens:    map[int]string{0: token0, 1: token1},
		Fee:       3000,
		Extra: &model.PoolExtra{
			Hash: tx.TxHash,
			From: tx.FromAddress,
			Time: uint64(tx.Timestamp.Unix()),
		},
	}
}

/*
 * parseV3PoolCreated 解析 Uniswap V3 的 PoolCreated（池子创建）事件
 *
 * 事件签名：
 *   PoolCreated(address indexed token0, address indexed token1, uint24 indexed fee, int24 tickSpacing, address pool)
 *
 * 事件含义：
 *   当通过 Uniswap V3 Factory 合约创建新的流动性池时触发。
 *   与 V2 的 PairCreated 相比，V3 增加了 fee（费率）和 tickSpacing（刻度间距）参数，
 *   因为同一个代币组合可以按不同费率创建多个池子。
 *
 * 事件参数说明：
 *   token0 (indexed)      - 交易对中的第一种代币地址（地址值较小者）
 *   token1 (indexed)      - 交易对中的第二种代币地址（地址值较大者）
 *   fee (indexed)         - 池子费率，单位为 hundredths of a bip（1/100 基点）
 *                          Uniswap V3 常见值：500=0.05%, 3000=0.30%, 10000=1.00%
 *   tickSpacing           - 刻度间距，决定可用的价格刻度密度
 *                          常见值：10(0.05%), 60(0.30%), 200(1.00%)
 *   pool                  - 新创建的 Pool 合约地址
 *
 * Topics 数组结构：
 *   Topics[0]: 事件签名 keccak256("PoolCreated(address,address,uint24,int24,address)")
 *   Topics[1]: token0 (indexed)
 *   Topics[2]: token1 (indexed)
 *   Topics[3]: fee (indexed, uint24 类型，高位补零)
 *   Data:      tickSpacing(32字节) + pool(32字节) = 64字节
 *
 * Pool 地址来源：
 *   使用 Data[32:64] 而非 log.Address 获取 Pool 地址。
 *   log.Address 是发出事件的 Factory 合约，Data 中的 pool 才是新创建的 Pool 地址。
 *   注意：与 V2 的 PairCreated 不同，V3 的 pool 在 Data 的第二个 32 字节（[32:64]），
 *   第一个 32 字节（[0:32]）是 tickSpacing。
 *
 * V2 vs V3 池子创建对比：
 *   V2: 同一 token0+token1 组合只能创建一个 Pair，费率固定 0.30%
 *   V3: 同一 token0+token1 组合可按不同 fee 创建多个 Pool
 *       例：USDC/WETH 可同时存在 0.05%、0.30%、1.00% 三个费率的池子
 *
 * fee 与 tickSpacing 的对应关系（Uniswap V3 默认）：
 *   fee=500   → tickSpacing=10  （低费率，高精度，适合稳定币对）
 *   fee=3000  → tickSpacing=60  （中等费率，最常用）
 *   fee=10000 → tickSpacing=200 （高费率，低精度，适合波动大的代币对）
 *
 * 实际案例：
 *   创建 USDC/WETH 0.30% 费率池子：
 *   - token0      = 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48（USDC）
 *   - token1      = 0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2（WETH）
 *   - fee         = 3000（0.30%）
 *   - tickSpacing = 60
 *   - pool        = 0x8ad599c3A0ff1De082011EFDDc58f1908eb6e6D8（新 Pool 地址）
 */
func (u *UniswapExtractor) parseV3PoolCreated(log *ethtypes.Log, tx *types.UnifiedTransaction) *model.Pool {
	if len(log.Topics) < 4 || len(log.Data) < 64 {
		return nil
	}

	token0 := common.BytesToAddress(log.Topics[1].Bytes()).Hex()
	token1 := common.BytesToAddress(log.Topics[2].Bytes()).Hex()
	fee := new(big.Int).SetBytes(log.Topics[3].Bytes())
	poolAddr := common.BytesToAddress(log.Data[32:64]).Hex()

	return &model.Pool{
		Addr:      poolAddr,
		Factory:   uniswapV3FactoryAddr,
		Protocol:  "uniswap-v3",
		ChainType: string(types.ChainTypeEthereum),
		Tokens:    map[int]string{0: token0, 1: token1},
		Fee:       int(fee.Int64()),
		Extra: &model.PoolExtra{
			Hash: tx.TxHash,
			From: tx.FromAddress,
			Time: uint64(tx.Timestamp.Unix()),
		},
	}
}

func (u *UniswapExtractor) getFactoryAddress(log *ethtypes.Log) string {
	if len(log.Topics) == 0 {
		return ""
	}
	topic0 := log.Topics[0].Hex()
	switch topic0 {
	case swapV2EventSig, mintV2EventSig, burnV2EventSig, pairCreatedEventSig:
		return uniswapV2FactoryAddr
	case swapV3EventSig, mintV3EventSig, burnV3EventSig, poolCreatedEventSig:
		return uniswapV3FactoryAddr
	default:
		return ""
	}
}
