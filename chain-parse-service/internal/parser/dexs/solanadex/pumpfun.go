package solanadex

import (
	"context"
	"fmt"
	"math/big"

	"unified-tx-parser/internal/model"
	dex "unified-tx-parser/internal/parser/dexs"
	"unified-tx-parser/internal/types"
)

const (
	pumpFunProgramID = "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"
)

// PumpFun event discriminators (first 8 bytes of Anchor event data)
var (
	pumpFunCreateDiscriminator   = []byte{27, 114, 169, 77, 222, 235, 99, 118}
	pumpFunTradeDiscriminator    = []byte{189, 219, 127, 211, 78, 230, 97, 238}
	pumpFunCompleteDiscriminator = []byte{95, 114, 97, 156, 212, 46, 152, 8}
)

// PumpFunExtractor parses PumpFun DEX events on Solana.
type PumpFunExtractor struct {
	*dex.SolanaDexExtractor
}

// NewPumpFunExtractor creates a PumpFun extractor with the Solana base class.
func NewPumpFunExtractor() *PumpFunExtractor {
	cfg := &dex.BaseDexExtractorConfig{
		Protocols:        []string{"pumpfun"},
		SupportedChains:  []types.ChainType{types.ChainTypeSolana},
		LoggerModuleName: "dex-pumpfun",
	}
	return &PumpFunExtractor{
		SolanaDexExtractor: dex.NewSolanaDexExtractor(cfg),
	}
}

// ExtractDexData extracts PumpFun DEX data from unified blocks.
func (p *PumpFunExtractor) ExtractDexData(ctx context.Context, blocks []types.UnifiedBlock) (*types.DexData, error) {
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
			events := dex.ExtractSolanaEvents(&tx)
			if len(events) == 0 {
				continue
			}

			swapIdx := int64(0)
			for eventIdx, event := range events {
				// 校验事件来源程序是否为 PumpFun
				if event.ProgramID != pumpFunProgramID {
					continue
				}

				if len(event.Data) < 8 {
					continue
				}
				disc := event.Data[:8]

				switch {
				case dex.MatchDiscriminatorBytes(disc, pumpFunTradeDiscriminator):
					if modelTx := p.parseTradeEvent(event.Data[8:], &tx, int64(eventIdx), swapIdx); modelTx != nil {
						dexData.Transactions = append(dexData.Transactions, *modelTx)
						swapIdx++
					}

				case dex.MatchDiscriminatorBytes(disc, pumpFunCreateDiscriminator):
					pool, token := p.parseCreateEvent(event.Data[8:], &tx)
					if pool != nil {
						dexData.Pools = append(dexData.Pools, *pool)
					}
					if token != nil {
						dexData.Tokens = append(dexData.Tokens, *token)
					}

				case dex.MatchDiscriminatorBytes(disc, pumpFunCompleteDiscriminator):
					if liq := p.parseCompleteEvent(event.Data[8:], &tx, int64(eventIdx)); liq != nil {
						dexData.Liquidities = append(dexData.Liquidities, *liq)
					}
				}
			}
		}
	}

	return dexData, nil
}

// SupportsBlock checks if any transaction in the block contains PumpFun events.
func (p *PumpFunExtractor) SupportsBlock(block *types.UnifiedBlock) bool {
	if !p.IsChainSupported(block.ChainType) {
		return false
	}
	for _, tx := range block.Transactions {
		events := dex.ExtractSolanaEvents(&tx)
		for _, event := range events {
			// 校验事件来源程序是否为 PumpFun
			if event.ProgramID != pumpFunProgramID {
				continue
			}
			if len(event.Data) < 8 {
				continue
			}
			disc := event.Data[:8]
			if dex.MatchDiscriminatorBytes(disc, pumpFunTradeDiscriminator) ||
				dex.MatchDiscriminatorBytes(disc, pumpFunCreateDiscriminator) ||
				dex.MatchDiscriminatorBytes(disc, pumpFunCompleteDiscriminator) {
				return true
			}
		}
	}
	return false
}

/*
 * parseTradeEvent 从 Borsh 编码的数据中解析 PumpFun TradeEvent（交易事件）。
 *
 * PumpFun 是 Solana 上的 memecoin launchpad，采用 Bonding Curve（联合曲线）机制定价：
 *   - 代币在 Bonding Curve 阶段通过买卖自动定价，无需传统流动性池
 *   - 当曲线市值达到阈值（约 69 SOL / ~$12k）后代币"毕业"，迁移至 Raydium DEX
 *   - 买入时 SOL 注入曲线，代币增发；卖出时代币销毁，SOL 返还
 *
 * 事件参数：
 *   - mint (Pubkey 32字节):                代币的 Mint 地址，即代币的唯一标识（也是 Bonding Curve 的标识，类比 EVM：相当于 ERC20 的合约地址）
 *   - sol_amount (u64):                    本次交易的 SOL 数量（单位：lamports，1 SOL = 10^9 lamports）
 *   - token_amount (u64):                  本次交易的代币数量（单位：最小精度，PumpFun 代币精度为 6）
 *   - is_buy (bool 1字节):                 交易方向，true=买入（SOL→Token），false=卖出（Token→SOL）
 *   - user (Pubkey 32字节):                交易发起者的钱包地址
 *   - timestamp (i64):                     交易的 Unix 时间戳（秒）
 *   - virtual_sol_reserves (u64):          虚拟 SOL 储备量（Bonding Curve 定价用的虚拟值，包含初始虚拟储备）
 *   - virtual_token_reserves (u64):        虚拟代币储备量（Bonding Curve 定价用的虚拟值）
 *   - real_sol_reserves (u64):             真实 SOL 储备量（Curve 中实际锁定的 SOL 数量）
 *   - real_token_reserves (u64):           真实代币储备量（Curve 中实际剩余的代币数量）
 *
 * 定价公式：price = virtual_sol_reserves / virtual_token_reserves
 *
 * 可选字段（新版本 PumpFun 程序包含）：
 *   - fee_recipient (Pubkey 32字节):      平台手续费接收地址（PumpFun 官方多签钱包）
 *   - fee_basis_points (u64):             平台手续费费率（单位：基点，100 = 1%）
 *   - fee (u64):                          本次交易实际扣除的平台手续费（单位：lamports）
 *   - creator (Pubkey 32字节):            代币创建者钱包地址
 *   - creator_fee_basis_points (u64):     创建者交易手续费费率（基点）
 *   - creator_fee (u64):                  本次交易实际分配给创建者的手续费（单位：lamports）
 */
func (p *PumpFunExtractor) parseTradeEvent(data []byte, tx *types.UnifiedTransaction, eventIdx, swapIdx int64) *model.Transaction {
	// Minimum required: mint(32) + sol_amount(8) + token_amount(8) + is_buy(1) + user(32) + timestamp(8) = 89 bytes
	if len(data) < 89 {
		p.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpFun trade: data too short")
		return nil
	}

	off := 0

	var mint string
	mint, off = dex.ParsePubkey(data, off)

	var solAmount, tokenAmount uint64
	solAmount, off = dex.ParseU64LE(data, off)
	tokenAmount, off = dex.ParseU64LE(data, off)

	var isBuy bool
	isBuy, off = dex.ParseBool(data, off)

	var user string
	user, off = dex.ParsePubkey(data, off)

	var timestamp int64
	timestamp, off = dex.ParseI64LE(data, off)
	_ = off // remaining optional fields not needed for core transaction

	if mint == "" || user == "" {
		p.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpFun trade: failed to parse mint or user")
		return nil
	}

	side := "sell"
	if isBuy {
		side = "buy"
	}

	// Price: SOL per token (normalize decimals: SOL has 9, PumpFun tokens have 6)
	// price = (solAmount / 1e9) / (tokenAmount / 1e6) = solAmount / tokenAmount / 1e3
	var price float64
	if tokenAmount > 0 {
		price = float64(solAmount) / float64(tokenAmount) / 1e3
	}

	solAmountBig := new(big.Int).SetUint64(solAmount)
	value := dex.LamportsToSOL(solAmount)

	txTime := uint64(tx.Timestamp.Unix())
	if timestamp > 0 {
		txTime = uint64(timestamp)
	}

	return &model.Transaction{
		Addr:        mint,
		Router:      pumpFunProgramID,
		Factory:     pumpFunProgramID,
		Pool:        mint, // PumpFun uses mint as the bonding curve identifier
		Hash:        tx.TxHash,
		From:        user,
		Side:        side,
		Amount:      solAmountBig,
		Price:       price,
		Value:       value,
		Time:        txTime,
		EventIndex:  eventIdx,
		TxIndex:     int64(tx.TxIndex),
		SwapIndex:   swapIdx,
		BlockNumber: dex.GetBlockNumber(tx),
		Protocol:    "pumpfun",
		ChainType:   string(types.ChainTypeSolana),
		Extra: &model.TransactionExtra{
			QuotePrice:    fmt.Sprintf("%.18f", price),
			Type:          "swap",
			TokenDecimals: 6, // PumpFun token decimals
		},
	}
}

/*
 * parseCreateEvent 从 Borsh 编码的数据中解析 PumpFun CreateEvent（代币创建事件）。
 *
 * 当用户在 PumpFun 上部署新代币时触发此事件，同时会创建一个 Bonding Curve 账户
 * 来管理该代币的价格曲线。代币初始通过 Bonding Curve 定价，达到市值阈值后迁移至 Raydium。
 *
 * 事件参数：
 *   - name (String 4字节长度前缀+UTF-8):   代币名称，如 "Pepe Coin"
 *   - symbol (String 4字节长度前缀+UTF-8):  代币符号/简称，如 "PEPE"
 *   - uri (String 4字节长度前缀+UTF-8):     代币元数据 URI，指向 JSON 格式的链下元数据（logo、描述、社交链接等）
 *   - mint (Pubkey 32字节):                 代币的 Mint 账户地址，全局唯一标识该代币
 *   - bonding_curve (Pubkey 32字节):        Bonding Curve 账户地址，管理该代币的定价曲线和储备资产（类比 EVM DEX 的 Pair/Pool 合约地址）
 *   - user (Pubkey 32字节):                 代币创建者的钱包地址
 *
 * 可选字段（新版本 PumpFun 程序包含）：
 *   - creator (Pubkey 32字节):             代币创建者钱包地址（与 user 相同，用于区分手续费归属）
 *   - timestamp (i64):                     创建时间的 Unix 时间戳（秒）
 *   - virtual_token_reserves (u64):        初始虚拟代币储备量（Bonding Curve 定价起始值）
 *   - virtual_sol_reserves (u64):          初始虚拟 SOL 储备量（Bonding Curve 定价起始值）
 *   - real_token_reserves (u64):           初始真实代币储备量
 *   - token_total_supply (u64):            代币总供应量（PumpFun 默认 10 亿，精度 6，即 1_000_000_000_000_000）
 */
func (p *PumpFunExtractor) parseCreateEvent(data []byte, tx *types.UnifiedTransaction) (*model.Pool, *model.Token) {
	// Minimum: 3 strings (4+0 each minimum = 12) + 3 pubkeys (32 each = 96) = 108 bytes minimum
	if len(data) < 108 {
		p.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpFun create: data too short")
		return nil, nil
	}

	off := 0

	var name, symbol, uri string
	name, off = dex.ParseString(data, off)
	symbol, off = dex.ParseString(data, off)
	uri, off = dex.ParseString(data, off)

	var mint, bondingCurve, user string
	mint, off = dex.ParsePubkey(data, off)
	bondingCurve, off = dex.ParsePubkey(data, off)
	user, off = dex.ParsePubkey(data, off)
	_ = off // remaining optional fields parsed only if needed

	if mint == "" || bondingCurve == "" {
		p.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpFun create: failed to parse mint or bonding_curve")
		return nil, nil
	}

	pool := &model.Pool{
		Addr:      bondingCurve,
		Factory:   pumpFunProgramID,
		Protocol:  "pumpfun",
		ChainType: string(types.ChainTypeSolana),
		Tokens:    map[int]string{0: mint, 1: "So11111111111111111111111111111111"}, // SOL native mint
		Fee:       100,                                                              // PumpFun default 1%
		Extra: &model.PoolExtra{
			Hash: tx.TxHash,
			From: user,
			Time: uint64(tx.Timestamp.Unix()),
		},
	}

	token := &model.Token{
		Addr:      mint,
		Name:      name,
		Symbol:    symbol,
		Decimals:  6, // PumpFun tokens default to 6 decimals
		ChainType: string(types.ChainTypeSolana),
		CreatedAt: tx.Timestamp.Format("2006-01-02T15:04:05Z"),
	}

	// Store URI in pool Args for reference
	if uri != "" {
		pool.Args = map[string]any{"uri": uri}
	}

	return pool, token
}

/*
 * parseCompleteEvent 从 Borsh 编码的数据中解析 PumpFun CompleteEvent（毕业事件）。
 *
 * 当 Bonding Curve 中锁定的 SOL 达到目标阈值（约 69 SOL）时触发此事件，标志着代币
 * 从 Bonding Curve 阶段"毕业"。毕业后代币的流动性将自动迁移至 Raydium DEX，
 * Bonding Curve 中的 SOL 和代币全部转入 Raydium 的流动性池，开始自由市场交易。
 *
 * 事件参数：
 *   - user (Pubkey 32字节):          触发毕业的交易者钱包地址（即最后一笔买入使曲线达到阈值的用户）
 *   - mint (Pubkey 32字节):          代币的 Mint 账户地址，唯一标识该代币
 *   - bonding_curve (Pubkey 32字节): Bonding Curve 账户地址，毕业后的曲线储备将清空并转入 Raydium 池
 *   - timestamp (i64):               毕业时间的 Unix 时间戳（秒）
 */
func (p *PumpFunExtractor) parseCompleteEvent(data []byte, tx *types.UnifiedTransaction, eventIdx int64) *model.Liquidity {
	// Minimum: user(32) + mint(32) + bonding_curve(32) + timestamp(8) = 104 bytes
	if len(data) < 104 {
		p.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpFun complete: data too short")
		return nil
	}

	off := 0

	var user, mint, bondingCurve string
	user, off = dex.ParsePubkey(data, off)
	mint, off = dex.ParsePubkey(data, off)
	bondingCurve, off = dex.ParsePubkey(data, off)

	var timestamp int64
	timestamp, _ = dex.ParseI64LE(data, off)

	if mint == "" || bondingCurve == "" {
		p.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpFun complete: failed to parse mint or bonding_curve")
		return nil
	}

	txTime := uint64(tx.Timestamp.Unix())
	if timestamp > 0 {
		txTime = uint64(timestamp)
	}

	key := fmt.Sprintf("%s_graduate_%d", tx.TxHash, eventIdx)

	return &model.Liquidity{
		Addr:      bondingCurve,
		Router:    pumpFunProgramID,
		Factory:   pumpFunProgramID,
		Pool:      bondingCurve,
		Hash:      tx.TxHash,
		From:      user,
		Side:      "graduate",
		Amount:    big.NewInt(0),
		Value:     0,
		Time:      txTime,
		Key:       key,
		Protocol:  "pumpfun",
		ChainType: string(types.ChainTypeSolana),
		Extra: &model.LiquidityExtra{
			Key:  key,
			Time: txTime,
		},
	}
}
