package bsc

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"unified-tx-parser/internal/model"
	bscProcessor "unified-tx-parser/internal/parser/chains/bsc"
	dex "unified-tx-parser/internal/parser/dexs"
	"unified-tx-parser/internal/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

/*
 * Four.meme 是 BSC 链上的 Meme 币发行和交易平台，采用 bonding curve（联合曲线）定价机制。
 * 合约分为 V1 和 V2 两个版本。
 *
 * 事件生命周期：
 *   TokenCreate → TokenPurchase (多次买入) / TokenSale (可随时卖出)
 *                                       ↓
 *                           (bonding curve 达到目标市值)
 *                                       ↓
 *                           LiquidityAdded (V2: 流动性迁移到 DEX)
 *
 * V1 事件（合约地址: 0xEC4549...3bFbC，2024-09-05 之前）：
 *   - TokenCreate:   创建新代币，包含创建者、代币地址、名称、符号、总量、上线时间
 *   - TokenPurchase: 买入代币（bonding curve 阶段），包含代币地址、账户、代币数量、花费的 BNB
 *   - TokenSale:     卖出代币（bonding curve 阶段），包含代币地址、账户、代币数量、获得的 BNB
 *
 * V2 事件（合约地址: 0x5c9520...0762b，2024-09-05 之后）：
 *   - TokenCreate:   创建新代币，相比 V1 多了 launchFee 字段
 *   - TokenPurchase: 买入代币，相比 V1 多了 price、cost、fee、offers、funds 字段
 *   - TokenSale:     卖出代币，参数结构同 V2 TokenPurchase
 *   - LiquidityAdded: bonding curve 结束后流动性添加到 DEX（如 PancakeSwap），包含 base/quote 地址和数量
 *
 * 核心区别：V1 只有 bonding curve 阶段的买卖，V2 增加了 LiquidityAdded 事件，
 * 意味着代币从 Four.meme 的联合曲线定价过渡到去中心化交易所的 AMM 定价。
 */

const (
	// TokenManager V1 contract address (tokens created before 2024-09-05)
	fourMemeV1Addr = "0xEC4549caDcE5DA21Df6E6422d448034B5233bFbC"
	// TokenManager V2 contract address (tokens created after 2024-09-05)
	fourMemeV2Addr = "0x5c952063c7fc8610FFDB798152D69F0B9550762b"

	// V2 event signatures
	fourMemeV2TokenCreateSig    = "0x396d5e902b675b032348d3d2e9517ee8f0c4a926603fbc075d3d282ff00cad20"
	fourMemeV2TokenPurchaseSig  = "0x7db52723a3b2cdd6164364b3b766e65e540d7be48ffa89582956d8eaebe62942"
	fourMemeV2TokenSaleSig      = "0x0a5575b3648bae2210cee56bf33254cc1ddfbc7bf637c0af2ac18b14fb1bae19"
	fourMemeV2LiquidityAddedSig = "0xc18aa71171b358b706fe3dd345299685ba21a5316c66ffa9e319268b033c44b0"

	// V1 event signatures
	fourMemeV1TokenCreateSig   = "0xc60523754e4c8d044ae75f841c3a7f27fefeed24c086155510c2ae0edf538fa0"
	fourMemeV1TokenPurchaseSig = "0x623b3804fa71d67900d064613da8f94b9617215ee90799290593e1745087ad18"
	fourMemeV1TokenSaleSig     = "0x3aa3f154f6bf5e3490d1a7205aa8d1412e76d26f9d186830de86fb9309224040"
)

var (
	fourMemeV1AddrLower = strings.ToLower(fourMemeV1Addr)
	fourMemeV2AddrLower = strings.ToLower(fourMemeV2Addr)
)

// FourMemeExtractor parses FourMeme V1/V2 DEX events on BSC.
type FourMemeExtractor struct {
	Client *bscProcessor.BSCProcessor
	*dex.EVMDexExtractor
}

// NewFourMemeExtractor creates a FourMeme extractor with EVM base class.
func NewFourMemeExtractor() *FourMemeExtractor {
	cfg := &dex.BaseDexExtractorConfig{
		Protocols:        []string{"fourmeme"},
		SupportedChains:  []types.ChainType{types.ChainTypeBSC},
		LoggerModuleName: "dex-fourmeme",
	}
	return &FourMemeExtractor{
		EVMDexExtractor: dex.NewEVMDexExtractor(cfg),
	}
}

// SetBSCProcessor 设置EVM处理器（用于获取链上数据）
func (b *FourMemeExtractor) SetBSCProcessor(processor interface{}) {
	if bscProcessor, ok := processor.(*bscProcessor.BSCProcessor); ok {
		b.Client = bscProcessor
	}
}

func (f *FourMemeExtractor) SupportsBlock(block *types.UnifiedBlock) bool {
	if block.ChainType != types.ChainTypeBSC {
		return false
	}
	for _, tx := range block.Transactions {
		ethLogs := dex.ExtractEVMLogsFromTransaction(&tx)
		for _, log := range ethLogs {
			if f.isFourMemeLog(log) {
				return true
			}
		}
	}
	return false
}

func (f *FourMemeExtractor) ExtractDexData(ctx context.Context, blocks []types.UnifiedBlock) (*types.DexData, error) {
	dexData := &types.DexData{
		Pools:        make([]model.Pool, 0),
		Transactions: make([]model.Transaction, 0),
		Liquidities:  make([]model.Liquidity, 0),
		Reserves:     make([]model.Reserve, 0),
		Tokens:       make([]model.Token, 0),
	}

	for _, block := range blocks {
		if block.ChainType != types.ChainTypeBSC {
			continue
		}

		for _, tx := range block.Transactions {
			ethLogs := dex.ExtractEVMLogsFromTransaction(&tx)
			if len(ethLogs) == 0 {
				continue
			}

			// swapIdx是同一笔交易内 swap 事件的顺序编号
			swapIdx := int64(0)
			for logIdx, log := range ethLogs {
				if !f.isFourMemeLog(log) {
					continue
				}

				topic0 := log.Topics[0].Hex()
				contractVersion := f.getContractVersion(log)

				switch {
				case topic0 == fourMemeV2TokenPurchaseSig && contractVersion == 2:
					if modelTx := f.parseV2Purchase(log, &tx, int64(logIdx), swapIdx); modelTx != nil {
						dexData.Transactions = append(dexData.Transactions, *modelTx)
						swapIdx++
					}

				case topic0 == fourMemeV2TokenSaleSig && contractVersion == 2:
					if modelTx := f.parseV2Sale(log, &tx, int64(logIdx), swapIdx); modelTx != nil {
						dexData.Transactions = append(dexData.Transactions, *modelTx)
						swapIdx++
					}

				case topic0 == fourMemeV1TokenPurchaseSig && contractVersion == 1:
					if modelTx := f.parseV1Purchase(log, &tx, int64(logIdx), swapIdx); modelTx != nil {
						dexData.Transactions = append(dexData.Transactions, *modelTx)
						swapIdx++
					}

				case topic0 == fourMemeV1TokenSaleSig && contractVersion == 1:
					if modelTx := f.parseV1Sale(log, &tx, int64(logIdx), swapIdx); modelTx != nil {
						dexData.Transactions = append(dexData.Transactions, *modelTx)
						swapIdx++
					}

				case topic0 == fourMemeV2TokenCreateSig && contractVersion == 2:
					if pool := f.parseV2TokenCreate(log, &tx); pool != nil {
						dexData.Pools = append(dexData.Pools, *pool)
					}

				case topic0 == fourMemeV1TokenCreateSig && contractVersion == 1:
					if pool := f.parseV1TokenCreate(log, &tx); pool != nil {
						dexData.Pools = append(dexData.Pools, *pool)
					}

				case topic0 == fourMemeV2LiquidityAddedSig && contractVersion == 2:
					if liq := f.parseV2LiquidityAdded(log, &tx, int64(logIdx)); liq != nil {
						dexData.Liquidities = append(dexData.Liquidities, *liq)
					}
				}
			}
		}
	}

	return dexData, nil
}

func (f *FourMemeExtractor) isFourMemeLog(log *ethtypes.Log) bool {
	if len(log.Topics) == 0 {
		return false
	}
	addr := strings.ToLower(log.Address.Hex())
	if addr != fourMemeV1AddrLower && addr != fourMemeV2AddrLower {
		return false
	}

	topic0 := log.Topics[0].Hex()
	return topic0 == fourMemeV2TokenCreateSig ||
		topic0 == fourMemeV2TokenPurchaseSig ||
		topic0 == fourMemeV2TokenSaleSig ||
		topic0 == fourMemeV2LiquidityAddedSig ||
		topic0 == fourMemeV1TokenCreateSig ||
		topic0 == fourMemeV1TokenPurchaseSig ||
		topic0 == fourMemeV1TokenSaleSig
}

func (f *FourMemeExtractor) getContractVersion(log *ethtypes.Log) int {
	addr := strings.ToLower(log.Address.Hex())
	switch addr {
	case fourMemeV1AddrLower:
		return 1
	case fourMemeV2AddrLower:
		return 2
	default:
		return 0
	}
}

/* parseV2Purchase 解析 V2 TokenPurchase 事件，表示用户在 bonding curve 阶段买入代币。
 *
 * 事件签名: TokenPurchase(address token, address account, uint256 price, uint256 amount, uint256 cost, uint256 fee, uint256 offers, uint256 funds)
 *
 * 参数说明:
 *   - token:   交易的代币合约地址
 *   - account: 买入者的钱包地址
 *   - price:   代币单价（wei）
 *   - amount:  买入的代币数量
 *   - cost:    花费的总成本（BNB，wei）
 *   - fee:     平台手续费（BNB，wei）
 *   - offers:  当前 bonding curve 上的挂单量
 *   - funds:   实际投入的资金（BNB，wei）
 *
 * 用户支付的总 BNB (cost) = 实际购币资金 (funds) + 平台手续费 (fee)
 *
 * ABI 编码: 8 个参数全部为 non-indexed，data = 8 * 32 = 256 字节
 */
func (f *FourMemeExtractor) parseV2Purchase(log *ethtypes.Log, tx *types.UnifiedTransaction, logIdx, swapIdx int64) *model.Transaction {
	if len(log.Data) < 256 {
		f.GetLogger().WithField("tx_hash", tx.TxHash).Warnf("V2 TokenPurchase data too short: %d bytes", len(log.Data))
		return nil
	}

	tokenAddr := common.BytesToAddress(log.Data[0:32]).Hex()
	account := common.BytesToAddress(log.Data[32:64]).Hex()
	price := new(big.Int).SetBytes(log.Data[64:96])
	amount := new(big.Int).SetBytes(log.Data[96:128])
	cost := new(big.Int).SetBytes(log.Data[128:160])
	fee := new(big.Int).SetBytes(log.Data[160:192])

	priceFloat := weiToFloat(price)
	costFloat := weiToFloat(cost)
	feeFloat := weiToFloat(fee)

	return &model.Transaction{
		Addr:        tokenAddr,
		Router:      fourMemeV2Addr,
		Factory:     fourMemeV2Addr,
		Pool:        tokenAddr,
		Hash:        tx.TxHash,
		From:        account,
		Side:        "buy",
		Amount:      amount,
		Price:       priceFloat,
		Value:       costFloat,
		Time:        uint64(tx.Timestamp.Unix()),
		EventIndex:  logIdx,
		TxIndex:     int64(tx.TxIndex),
		SwapIndex:   swapIdx,
		BlockNumber: dex.GetBlockNumber(tx),
		Extra: &model.TransactionExtra{
			QuotePrice:    fmt.Sprintf("%.18f", priceFloat),
			Type:          "buy",
			TokenDecimals: 18,
			QuoteAddr:     fmt.Sprintf("fee:%.18f", feeFloat),
		},
		ChainType: string(types.ChainTypeBSC),
		Protocol:  "fourmeme-v2",
	}
}

/* parseV2Sale 解析 V2 TokenSale 事件，表示用户在 bonding curve 阶段卖出代币。
 *
 * 事件签名: TokenSale(address token, address account, uint256 price, uint256 amount, uint256 cost, uint256 fee, uint256 offers, uint256 funds)
 *
 * 参数说明:
 *   - token:   交易的代币合约地址
 *   - account: 卖出者的钱包地址
 *   - price:   代币单价（wei）
 *   - amount:  卖出的代币数量
 *   - cost:    获得的总金额（BNB，wei）
 *   - fee:     平台手续费（BNB，wei）
 *   - offers:  当前 bonding curve 上的挂单量
 *   - funds:   实际收到的资金（BNB，wei）
 *
 * ABI 编码: 8 个参数全部为 non-indexed，data = 8 * 32 = 256 字节
 */
func (f *FourMemeExtractor) parseV2Sale(log *ethtypes.Log, tx *types.UnifiedTransaction, logIdx, swapIdx int64) *model.Transaction {
	if len(log.Data) < 256 {
		f.GetLogger().WithField("tx_hash", tx.TxHash).Warnf("V2 TokenSale data too short: %d bytes", len(log.Data))
		return nil
	}

	tokenAddr := common.BytesToAddress(log.Data[0:32]).Hex()
	account := common.BytesToAddress(log.Data[32:64]).Hex()
	price := new(big.Int).SetBytes(log.Data[64:96])
	amount := new(big.Int).SetBytes(log.Data[96:128])
	cost := new(big.Int).SetBytes(log.Data[128:160])
	fee := new(big.Int).SetBytes(log.Data[160:192])

	priceFloat := weiToFloat(price)
	costFloat := weiToFloat(cost)
	feeFloat := weiToFloat(fee)

	return &model.Transaction{
		Addr:        tokenAddr,
		Router:      fourMemeV2Addr,
		Factory:     fourMemeV2Addr,
		Pool:        tokenAddr,
		Hash:        tx.TxHash,
		From:        account,
		Side:        "sell",
		Amount:      amount,
		Price:       priceFloat,
		Value:       costFloat,
		Time:        uint64(tx.Timestamp.Unix()),
		EventIndex:  logIdx,
		TxIndex:     int64(tx.TxIndex),
		SwapIndex:   swapIdx,
		BlockNumber: dex.GetBlockNumber(tx),
		Extra: &model.TransactionExtra{
			QuotePrice:    fmt.Sprintf("%.18f", priceFloat),
			Type:          "sell",
			TokenDecimals: 18,
			QuoteAddr:     fmt.Sprintf("fee:%.18f", feeFloat),
		},
		ChainType: string(types.ChainTypeBSC),
		Protocol:  "fourmeme-v2",
	}
}

/* parseV1Purchase 解析 V1 TokenPurchase 事件，表示用户在 bonding curve 阶段买入代币。
 *
 * 事件签名: TokenPurchase(address token, address account, uint256 tokenAmount, uint256 etherAmount)
 *
 * 参数说明:
 *   - token:       交易的代币合约地址
 *   - account:     买入者的钱包地址
 *   - tokenAmount: 买入的代币数量
 *   - etherAmount: 花费的 BNB 数量（wei）
 *
 * ABI 编码: 4 个参数全部为 non-indexed，data = 4 * 32 = 128 字节
 */
func (f *FourMemeExtractor) parseV1Purchase(log *ethtypes.Log, tx *types.UnifiedTransaction, logIdx, swapIdx int64) *model.Transaction {
	if len(log.Data) < 128 {
		f.GetLogger().WithField("tx_hash", tx.TxHash).Warnf("V1 TokenPurchase data too short: %d bytes", len(log.Data))
		return nil
	}

	tokenAddr := common.BytesToAddress(log.Data[0:32]).Hex()
	account := common.BytesToAddress(log.Data[32:64]).Hex()
	tokenAmount := new(big.Int).SetBytes(log.Data[64:96])
	etherAmount := new(big.Int).SetBytes(log.Data[96:128])

	price := dex.CalcPrice(etherAmount, tokenAmount)
	costFloat := weiToFloat(etherAmount)

	return &model.Transaction{
		Addr:        tokenAddr,
		Router:      fourMemeV1Addr,
		Factory:     fourMemeV1Addr,
		Pool:        tokenAddr,
		Hash:        tx.TxHash,
		From:        account,
		Side:        "buy",
		Amount:      tokenAmount,
		Price:       price,
		Value:       costFloat,
		Time:        uint64(tx.Timestamp.Unix()),
		EventIndex:  logIdx,
		TxIndex:     int64(tx.TxIndex),
		SwapIndex:   swapIdx,
		BlockNumber: dex.GetBlockNumber(tx),
		Extra: &model.TransactionExtra{
			QuotePrice:    fmt.Sprintf("%.18f", price),
			Type:          "buy",
			TokenDecimals: 18,
		},
		ChainType: string(types.ChainTypeBSC),
		Protocol:  "fourmeme-v1",
	}
}

/* parseV1Sale 解析 V1 TokenSale 事件，表示用户在 bonding curve 阶段卖出代币。
 *
 * 事件签名: TokenSale(address token, address account, uint256 tokenAmount, uint256 etherAmount)
 *
 * 参数说明:
 *   - token:       交易的代币合约地址
 *   - account:     卖出者的钱包地址
 *   - tokenAmount: 卖出的代币数量
 *   - etherAmount: 获得的 BNB 数量（wei）
 *
 * ABI 编码: 4 个参数全部为 non-indexed，data = 4 * 32 = 128 字节
 */
func (f *FourMemeExtractor) parseV1Sale(log *ethtypes.Log, tx *types.UnifiedTransaction, logIdx, swapIdx int64) *model.Transaction {
	if len(log.Data) < 128 {
		f.GetLogger().WithField("tx_hash", tx.TxHash).Warnf("V1 TokenSale data too short: %d bytes", len(log.Data))
		return nil
	}

	tokenAddr := common.BytesToAddress(log.Data[0:32]).Hex()
	account := common.BytesToAddress(log.Data[32:64]).Hex()
	tokenAmount := new(big.Int).SetBytes(log.Data[64:96])
	etherAmount := new(big.Int).SetBytes(log.Data[96:128])

	price := dex.CalcPrice(etherAmount, tokenAmount)
	costFloat := weiToFloat(etherAmount)

	return &model.Transaction{
		Addr:        tokenAddr,
		Router:      fourMemeV1Addr,
		Factory:     fourMemeV1Addr,
		Pool:        tokenAddr,
		Hash:        tx.TxHash,
		From:        account,
		Side:        "sell",
		Amount:      tokenAmount,
		Price:       price,
		Value:       costFloat,
		Time:        uint64(tx.Timestamp.Unix()),
		EventIndex:  logIdx,
		TxIndex:     int64(tx.TxIndex),
		SwapIndex:   swapIdx,
		BlockNumber: dex.GetBlockNumber(tx),
		Extra: &model.TransactionExtra{
			QuotePrice:    fmt.Sprintf("%.18f", price),
			Type:          "sell",
			TokenDecimals: 18,
		},
		ChainType: string(types.ChainTypeBSC),
		Protocol:  "fourmeme-v1",
	}
}

/* parseV2TokenCreate 解析 V2 TokenCreate 事件，表示在 bonding curve 阶段创建新代币。
 *
 * 事件签名: TokenCreate(address creator, address token, uint256 requestId, string name, string symbol, uint256 totalSupply, uint256 launchTime, uint256 launchFee)
 *
 * 参数说明:
 *   - creator:     创建者钱包地址
 *   - token:       代币合约地址
 *   - requestId:   创建请求 ID
 *   - name:        代币名称（动态编码，占 64 字节）
 *   - symbol:      代币符号（动态编码，占 64 字节）
 *   - totalSupply: 代币总供应量
 *   - launchTime:  代币上线时间（Unix 时间戳）
 *   - launchFee:   上线手续费（BNB，wei，V2 新增字段）
 *
 * ABI 编码注意：
 *   - string 类型使用动态编码（offset + length + data）
 *   - name 和 symbol 各占 64 字节，所以 data >= 256 字节
 */
func (f *FourMemeExtractor) parseV2TokenCreate(log *ethtypes.Log, tx *types.UnifiedTransaction) *model.Pool {
	if len(log.Data) < 256 {
		f.GetLogger().WithField("tx_hash", tx.TxHash).Warnf("V2 TokenCreate data too short: %d bytes", len(log.Data))
		return nil
	}

	creator := common.BytesToAddress(log.Data[0:32]).Hex()
	tokenAddr := common.BytesToAddress(log.Data[32:64]).Hex()

	return &model.Pool{
		Addr:      tokenAddr,
		Factory:   fourMemeV2Addr,
		Protocol:  "fourmeme-v2",
		ChainType: string(types.ChainTypeBSC),
		Tokens:    map[int]string{0: tokenAddr},
		Fee:       0,
		Args: map[string]interface{}{
			"creator": creator,
			"version": 2,
		},
		Extra: &model.PoolExtra{
			Hash: tx.TxHash,
			From: creator,
			Time: uint64(tx.Timestamp.Unix()),
		},
	}
}

/* parseV1TokenCreate 解析 V1 TokenCreate 事件，表示在 bonding curve 阶段创建新代币。
 *
 * 事件签名: TokenCreate(address creator, address token, uint256 requestId, string name, string symbol, uint256 totalSupply, uint256 launchTime)
 *
 * 参数说明:
 *   - creator:      创建者钱包地址
 *   - token:        代币合约地址
 *   - requestId:   创建请求 ID
 *   - name:        代币名称（动态编码，占 64 字节）
 *   - symbol:      代币符号（动态编码，占 64 字节）
 *   - totalSupply: 代币总供应量
 *   - launchTime:  代币上线时间（Unix 时间戳）
 *
 * ABI 编码注意：
 *   - string 类型使用动态编码（offset + length + data）
 *   - name 和 symbol 各占 64 字节，所以 data >= 224 字节
 *   - V1 版本没有 launchFee 字段，这是与 V2 的主要区别
 */
func (f *FourMemeExtractor) parseV1TokenCreate(log *ethtypes.Log, tx *types.UnifiedTransaction) *model.Pool {
	if len(log.Data) < 224 {
		f.GetLogger().WithField("tx_hash", tx.TxHash).Warnf("V1 TokenCreate data too short: %d bytes", len(log.Data))
		return nil
	}

	creator := common.BytesToAddress(log.Data[0:32]).Hex()
	tokenAddr := common.BytesToAddress(log.Data[32:64]).Hex()

	return &model.Pool{
		Addr:      tokenAddr,
		Factory:   fourMemeV1Addr,
		Protocol:  "fourmeme-v1",
		ChainType: string(types.ChainTypeBSC),
		Tokens:    map[int]string{0: tokenAddr},
		Fee:       0,
		Args: map[string]interface{}{
			"creator": creator,
			"version": 1,
		},
		Extra: &model.PoolExtra{
			Hash: tx.TxHash,
			From: creator,
			Time: uint64(tx.Timestamp.Unix()),
		},
	}
}

/* parseV2LiquidityAdded 解析 V2 LiquidityAdded 事件，表示用户在 bonding curve 阶段结束后添加流动性到 DEX。
 *
 * 事件签名: LiquidityAdded(address base, uint256 offers, address quote, uint256 funds)
 *
 * 参数说明:
 *   - base:    主代币（Base Token）的合约地址
 *   - offers:  添加的主代币数量
 *   - quote:   计价代币（Quote Token）的合约地址，address(0) 表示 BNB
 *   - funds:   添加的计价代币数量（BNB 或其他 BEP20 代币）
 *
 * 生命周期阶段：
 *   - bonding curve 阶段：用户通过 Four.meme 的联合曲线购买代币
 *   - bonding curve 结束后：达到市值目标，迁移到 DEX（如 PancakeSwap）
 *   - 添加流动性：用户将代币和 BNB 添加到 DEX 的流动性池中
 *
 * ABI 编码: 4 个参数全部为 non-indexed，data = 4 * 32 = 128 字节
 */
func (f *FourMemeExtractor) parseV2LiquidityAdded(log *ethtypes.Log, tx *types.UnifiedTransaction, logIdx int64) *model.Liquidity {
	if len(log.Data) < 128 {
		f.GetLogger().WithField("tx_hash", tx.TxHash).Warnf("V2 LiquidityAdded data too short: %d bytes", len(log.Data))
		return nil
	}

	baseAddr := common.BytesToAddress(log.Data[0:32]).Hex()
	offers := new(big.Int).SetBytes(log.Data[32:64])
	_ = common.BytesToAddress(log.Data[64:96]) // quoteAddr: address(0) = BNB, otherwise BEP20
	funds := new(big.Int).SetBytes(log.Data[96:128])

	fundsFloat := weiToFloat(funds)
	key := fmt.Sprintf("%s_liquidity_%d", tx.TxHash, logIdx)

	return &model.Liquidity{
		Addr:      baseAddr,
		Router:    fourMemeV2Addr,
		Factory:   fourMemeV2Addr,
		Pool:      baseAddr,
		Hash:      tx.TxHash,
		From:      tx.FromAddress,
		Side:      "add",
		Amount:    offers,
		Value:     fundsFloat,
		Time:      uint64(tx.Timestamp.Unix()),
		Key:       key,
		Protocol:  "fourmeme-v2",
		ChainType: string(types.ChainTypeBSC),
		Extra: &model.LiquidityExtra{
			Key:     key,
			Amounts: funds,
			Values:  []float64{weiToFloat(offers), fundsFloat},
			Time:    uint64(tx.Timestamp.Unix()),
		},
	}
}

func weiToFloat(wei *big.Int) float64 {
	if wei == nil || wei.Sign() == 0 {
		return 0
	}
	return dex.ConvertDecimals(wei, 18)
}
