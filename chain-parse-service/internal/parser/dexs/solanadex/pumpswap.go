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
	pumpSwapProgramID = "pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA"
)

// PumpSwap event discriminators (first 8 bytes of Anchor event data)
var (
	pumpSwapBuyDiscriminator        = []byte{103, 244, 82, 31, 44, 245, 119, 119}
	pumpSwapSellDiscriminator       = []byte{62, 47, 55, 10, 165, 3, 220, 42}
	pumpSwapCreatePoolDiscriminator = []byte{177, 49, 12, 210, 160, 118, 167, 116}
	pumpSwapDepositDiscriminator    = []byte{120, 248, 61, 83, 31, 142, 107, 144}
	pumpSwapWithdrawDiscriminator   = []byte{22, 9, 133, 26, 160, 44, 71, 192}
)

// PumpSwapExtractor parses PumpSwap AMM events on Solana.
type PumpSwapExtractor struct {
	*dex.SolanaDexExtractor
}

// NewPumpSwapExtractor creates a PumpSwap extractor with the Solana base class.
func NewPumpSwapExtractor() *PumpSwapExtractor {
	cfg := &dex.BaseDexExtractorConfig{
		Protocols:        []string{"pumpswap"},
		SupportedChains:  []types.ChainType{types.ChainTypeSolana},
		LoggerModuleName: "dex-pumpswap",
	}
	return &PumpSwapExtractor{
		SolanaDexExtractor: dex.NewSolanaDexExtractor(cfg),
	}
}

// ExtractDexData extracts PumpSwap DEX data from unified blocks.
func (ps *PumpSwapExtractor) ExtractDexData(ctx context.Context, blocks []types.UnifiedBlock) (*types.DexData, error) {
	dexData := &types.DexData{
		Pools:        make([]model.Pool, 0),
		Transactions: make([]model.Transaction, 0),
		Liquidities:  make([]model.Liquidity, 0),
		Reserves:     make([]model.Reserve, 0),
		Tokens:       make([]model.Token, 0),
	}

	for _, block := range blocks {
		if !ps.IsChainSupported(block.ChainType) {
			continue
		}

		for _, tx := range block.Transactions {
			events := dex.ExtractSolanaEvents(&tx)
			if len(events) == 0 {
				continue
			}

			swapIdx := int64(0)
			for eventIdx, event := range events {
				// 校验事件来源程序是否为 PumpSwap
				if event.ProgramID != pumpSwapProgramID {
					continue
				}

				if len(event.Data) < 8 {
					continue
				}
				disc := event.Data[:8]

				switch {
				case dex.MatchDiscriminatorBytes(disc, pumpSwapBuyDiscriminator):
					if modelTx := ps.parseBuyEvent(event.Data[8:], &tx, int64(eventIdx), swapIdx); modelTx != nil {
						dexData.Transactions = append(dexData.Transactions, *modelTx)
						swapIdx++
					}

				case dex.MatchDiscriminatorBytes(disc, pumpSwapSellDiscriminator):
					if modelTx := ps.parseSellEvent(event.Data[8:], &tx, int64(eventIdx), swapIdx); modelTx != nil {
						dexData.Transactions = append(dexData.Transactions, *modelTx)
						swapIdx++
					}

				case dex.MatchDiscriminatorBytes(disc, pumpSwapCreatePoolDiscriminator):
					if pool := ps.parseCreatePoolEvent(event.Data[8:], &tx); pool != nil {
						dexData.Pools = append(dexData.Pools, *pool)
					}

				case dex.MatchDiscriminatorBytes(disc, pumpSwapDepositDiscriminator):
					if liq := ps.parseDepositEvent(event.Data[8:], &tx, int64(eventIdx)); liq != nil {
						dexData.Liquidities = append(dexData.Liquidities, *liq)
					}

				case dex.MatchDiscriminatorBytes(disc, pumpSwapWithdrawDiscriminator):
					if liq := ps.parseWithdrawEvent(event.Data[8:], &tx, int64(eventIdx)); liq != nil {
						dexData.Liquidities = append(dexData.Liquidities, *liq)
					}
				}
			}
		}
	}

	return dexData, nil
}

// SupportsBlock checks if any transaction in the block contains PumpSwap events.
func (ps *PumpSwapExtractor) SupportsBlock(block *types.UnifiedBlock) bool {
	if !ps.IsChainSupported(block.ChainType) {
		return false
	}
	for _, tx := range block.Transactions {
		events := dex.ExtractSolanaEvents(&tx)
		for _, event := range events {
			// 校验事件来源程序是否为 PumpSwap
			if event.ProgramID != pumpSwapProgramID {
				continue
			}
			if len(event.Data) < 8 {
				continue
			}
			disc := event.Data[:8]
			if dex.MatchDiscriminatorBytes(disc, pumpSwapBuyDiscriminator) ||
				dex.MatchDiscriminatorBytes(disc, pumpSwapSellDiscriminator) ||
				dex.MatchDiscriminatorBytes(disc, pumpSwapCreatePoolDiscriminator) ||
				dex.MatchDiscriminatorBytes(disc, pumpSwapDepositDiscriminator) ||
				dex.MatchDiscriminatorBytes(disc, pumpSwapWithdrawDiscriminator) {
				return true
			}
		}
	}
	return false
}

/*
 * parseBuyEvent 从 Borsh 编码的数据中解析 PumpSwap BuyEvent（买入事件）。
 *
 * 当用户在 PumpSwap 上购买代币时触发此事件。
 *
 * 事件参数：
 *   - base_amount_out (u64):    买家收到的代币数量（基础代币，即被购买的代币）
 *   - quote_amount_in (u64):    买家支付的 SOL 数量（报价代币，单位 lamports）
 *   - lp_fee (u64):             流动性提供者手续费（分配给流动性提供者（LP）的手续费）
 *   - protocol_fee (u64):       协议手续费（分配给协议方（PumpSwap 团队/国库）的手续费）
 *   - pool (Pubkey 32字节):     交易池账户地址
 *   - user (Pubkey 32字节):     买家钱包地址
 *   - base_mint (Pubkey 32字节): 基础代币的 Mint 地址（被购买的代币）
 *   - quote_mint (Pubkey 32字节): 报价代币的 Mint 地址（通常为 SOL）
 */
func (ps *PumpSwapExtractor) parseBuyEvent(data []byte, tx *types.UnifiedTransaction, eventIdx, swapIdx int64) *model.Transaction {
	// Minimum: 4*u64(32) + 4*Pubkey(128) = 160 bytes
	if len(data) < 160 {
		ps.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpSwap buy: data too short")
		return nil
	}

	off := 0

	var baseAmountOut, quoteAmountIn, lpFee, protocolFee uint64
	baseAmountOut, off = dex.ParseU64LE(data, off)
	quoteAmountIn, off = dex.ParseU64LE(data, off)
	lpFee, off = dex.ParseU64LE(data, off)
	protocolFee, off = dex.ParseU64LE(data, off)
	_ = lpFee
	_ = protocolFee

	var pool, user, baseMint, quoteMint string
	pool, off = dex.ParsePubkey(data, off)
	user, off = dex.ParsePubkey(data, off)
	baseMint, off = dex.ParsePubkey(data, off)
	quoteMint, off = dex.ParsePubkey(data, off)
	_ = off

	if pool == "" || baseMint == "" {
		ps.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpSwap buy: failed to parse pool or base_mint")
		return nil
	}

	// Price: SOL per base token (normalize: quote=SOL 9 decimals, base=token 6 decimals)
	// price = (quoteAmountIn / 1e9) / (baseAmountOut / 1e6) = quoteAmountIn / baseAmountOut / 1e3
	var price float64
	if baseAmountOut > 0 {
		price = float64(quoteAmountIn) / float64(baseAmountOut) / 1e3
	}

	quoteAmountBig := new(big.Int).SetUint64(quoteAmountIn)
	value := dex.LamportsToSOL(quoteAmountIn)

	return &model.Transaction{
		Addr:        baseMint,
		Router:      pumpSwapProgramID,
		Factory:     pumpSwapProgramID,
		Pool:        pool,
		Hash:        tx.TxHash,
		From:        user,
		Side:        "buy",
		Amount:      quoteAmountBig,
		Price:       price,
		Value:       value,
		Time:        uint64(tx.Timestamp.Unix()),
		EventIndex:  eventIdx,
		TxIndex:     int64(tx.TxIndex),
		SwapIndex:   swapIdx,
		BlockNumber: dex.GetBlockNumber(tx),
		Protocol:    "pumpswap",
		ChainType:   string(types.ChainTypeSolana),
		Extra: &model.TransactionExtra{
			QuoteAddr:     quoteMint,
			QuotePrice:    fmt.Sprintf("%.18f", price),
			Type:          "swap",
			TokenDecimals: 6,
		},
	}
}

/*
 * parseSellEvent 从 Borsh 编码的数据中解析 PumpSwap SellEvent（卖出事件）。
 *
 * 当用户在 PumpSwap 上卖出代币时触发此事件。
 *
 * 事件参数：
 *   - base_amount_in (u64):      卖出方支付的代币数量（基础代币，即被卖出的代币）
 *   - quote_amount_out (u64):    卖出方收到的 SOL 数量（报价代币，单位 lamports）
 *   - lp_fee (u64):              流动性提供者手续费
 *   - protocol_fee (u64):        协议手续费
 *   - pool (Pubkey 32字节):      交易池账户地址
 *   - user (Pubkey 32字节):      卖出方钱包地址
 *   - base_mint (Pubkey 32字节):  基础代币的 Mint 地址（被卖出的代币）
 *   - quote_mint (Pubkey 32字节): 报价代币的 Mint 地址（通常为 SOL）
 */
func (ps *PumpSwapExtractor) parseSellEvent(data []byte, tx *types.UnifiedTransaction, eventIdx, swapIdx int64) *model.Transaction {
	// Minimum: 4*u64(32) + 4*Pubkey(128) = 160 bytes
	if len(data) < 160 {
		ps.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpSwap sell: data too short")
		return nil
	}

	off := 0

	var baseAmountIn, quoteAmountOut, lpFee, protocolFee uint64
	baseAmountIn, off = dex.ParseU64LE(data, off)
	quoteAmountOut, off = dex.ParseU64LE(data, off)
	lpFee, off = dex.ParseU64LE(data, off)
	protocolFee, off = dex.ParseU64LE(data, off)
	_ = lpFee
	_ = protocolFee

	var pool, user, baseMint, quoteMint string
	pool, off = dex.ParsePubkey(data, off)
	user, off = dex.ParsePubkey(data, off)
	baseMint, off = dex.ParsePubkey(data, off)
	quoteMint, off = dex.ParsePubkey(data, off)
	_ = off

	if pool == "" || baseMint == "" {
		ps.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpSwap sell: failed to parse pool or base_mint")
		return nil
	}

	// Price: SOL per base token (normalize: quote=SOL 9 decimals, base=token 6 decimals)
	// price = (quoteAmountOut / 1e9) / (baseAmountIn / 1e6) = quoteAmountOut / baseAmountIn / 1e3
	var price float64
	if baseAmountIn > 0 {
		price = float64(quoteAmountOut) / float64(baseAmountIn) / 1e3
	}

	quoteAmountBig := new(big.Int).SetUint64(quoteAmountOut)
	value := dex.LamportsToSOL(quoteAmountOut)

	return &model.Transaction{
		Addr:        baseMint,
		Router:      pumpSwapProgramID,
		Factory:     pumpSwapProgramID,
		Pool:        pool,
		Hash:        tx.TxHash,
		From:        user,
		Side:        "sell",
		Amount:      quoteAmountBig,
		Price:       price,
		Value:       value,
		Time:        uint64(tx.Timestamp.Unix()),
		EventIndex:  eventIdx,
		TxIndex:     int64(tx.TxIndex),
		SwapIndex:   swapIdx,
		BlockNumber: dex.GetBlockNumber(tx),
		Protocol:    "pumpswap",
		ChainType:   string(types.ChainTypeSolana),
		Extra: &model.TransactionExtra{
			QuoteAddr:     quoteMint,
			QuotePrice:    fmt.Sprintf("%.18f", price),
			Type:          "swap",
			TokenDecimals: 6,
		},
	}
}

/*
 * parseCreatePoolEvent 从 Borsh 编码的数据中解析 PumpSwap CreatePoolEvent（创建池子事件）。
 *
 * 当用户在 PumpSwap 上创建新的交易池时触发此事件，同时会铸造 LP 代币作为流动性凭证。
 *
 * 事件参数：
 *   - creator (Pubkey 32字节):       池子创建者的钱包地址
 *   - base_mint (Pubkey 32字节):     基础代币的 Mint 地址
 *   - quote_mint (Pubkey 32字节):    报价代币的 Mint 地址（通常为 SOL）
 *   - lp_token_amount_out (u64):     铸造给创建者的 LP 代币数量
 *   - pool (Pubkey 32字节):          新创建的交易池账户地址
 *   - lp_mint (Pubkey 32字节):       LP 代币的 Mint 地址
 *   - base_amount_in (u64):          初始注入的基础代币数量
 *   - quote_amount_in (u64):         初始注入的报价代币数量（单位 lamports）
 */
func (ps *PumpSwapExtractor) parseCreatePoolEvent(data []byte, tx *types.UnifiedTransaction) *model.Pool {
	// Minimum: 3*Pubkey(96) + u64(8) + 2*Pubkey(64) + 2*u64(16) = 184 bytes
	if len(data) < 184 {
		ps.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpSwap create pool: data too short")
		return nil
	}

	off := 0

	var creator, baseMint, quoteMint string
	creator, off = dex.ParsePubkey(data, off)
	baseMint, off = dex.ParsePubkey(data, off)
	quoteMint, off = dex.ParsePubkey(data, off)

	var lpTokenAmountOut uint64
	lpTokenAmountOut, off = dex.ParseU64LE(data, off)
	_ = lpTokenAmountOut

	var pool, lpMint string
	pool, off = dex.ParsePubkey(data, off)
	lpMint, off = dex.ParsePubkey(data, off)
	_ = lpMint

	var baseAmountIn, quoteAmountIn uint64
	baseAmountIn, off = dex.ParseU64LE(data, off)
	quoteAmountIn, off = dex.ParseU64LE(data, off)

	// 消除编译器"未使用变量"警告的惯用写法
	_ = baseAmountIn
	_ = quoteAmountIn
	_ = off

	if pool == "" || baseMint == "" || quoteMint == "" {
		ps.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpSwap create pool: failed to parse required fields")
		return nil
	}

	return &model.Pool{
		Addr:      pool,
		Factory:   pumpSwapProgramID,
		Protocol:  "pumpswap",
		ChainType: string(types.ChainTypeSolana),
		Tokens:    map[int]string{0: baseMint, 1: quoteMint},
		Fee:       30, // PumpSwap default 0.3% (30 bps)
		Extra: &model.PoolExtra{
			Hash: tx.TxHash,
			From: creator,
			Time: uint64(tx.Timestamp.Unix()),
		},
	}
}

/*
 * parseDepositEvent 从 Borsh 编码的数据中解析 PumpSwap DepositEvent（添加流动性事件）。
 *
 * 当用户向 PumpSwap 交易池添加流动性时触发此事件，用户存入代币对并获得 LP 代币作为凭证。
 *
 * 事件参数：
 *   - base_amount_in (u64):       存入的基础代币数量
 *   - quote_amount_in (u64):      存入的报价代币数量（单位 lamports）
 *   - lp_token_amount_out (u64):  用户收到的 LP 代币数量
 *   - pool (Pubkey 32字节):       目标交易池账户地址
 *   - user (Pubkey 32字节):       添加流动性的用户钱包地址
 */
func (ps *PumpSwapExtractor) parseDepositEvent(data []byte, tx *types.UnifiedTransaction, eventIdx int64) *model.Liquidity {
	// Minimum: 3*u64(24) + 2*Pubkey(64) = 88 bytes
	if len(data) < 88 {
		ps.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpSwap deposit: data too short")
		return nil
	}

	off := 0

	var quoteAmountIn, lpTokenAmountOut uint64
	_, off = dex.ParseU64LE(data, off) // base_amount_in (skip, unknown decimals)
	quoteAmountIn, off = dex.ParseU64LE(data, off)
	lpTokenAmountOut, off = dex.ParseU64LE(data, off)
	_ = lpTokenAmountOut

	var pool, user string
	pool, off = dex.ParsePubkey(data, off)
	user, off = dex.ParsePubkey(data, off)
	_ = off

	if pool == "" {
		ps.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpSwap deposit: failed to parse pool")
		return nil
	}

	// Use quote (SOL) amount for Value since base token value is unknown without price oracle
	quoteAmountBig := new(big.Int).SetUint64(quoteAmountIn)
	quoteValue := dex.LamportsToSOL(quoteAmountIn)

	key := fmt.Sprintf("%s_add_%d", tx.TxHash, eventIdx)

	return &model.Liquidity{
		Addr:      pool,
		Router:    pumpSwapProgramID,
		Factory:   pumpSwapProgramID,
		Pool:      pool,
		Hash:      tx.TxHash,
		From:      user,
		Side:      "add",
		Amount:    quoteAmountBig,
		Value:     quoteValue * 2, // Approximate: assume equal value on both sides
		Time:      uint64(tx.Timestamp.Unix()),
		Key:       key,
		Protocol:  "pumpswap",
		ChainType: string(types.ChainTypeSolana),
		Extra: &model.LiquidityExtra{
			Key:     key,
			Amounts: new(big.Int).SetUint64(quoteAmountIn),
			Values:  []float64{0, quoteValue}, // base value unknown, quote in SOL
			Time:    uint64(tx.Timestamp.Unix()),
		},
	}
}

/*
 * parseWithdrawEvent 从 Borsh 编码的数据中解析 PumpSwap WithdrawEvent（移除流动性事件）。
 *
 * 当用户从 PumpSwap 交易池移除流动性时触发此事件，用户销毁 LP 代币并取回对应的代币对。
 *
 * 事件参数：
 *   - lp_token_amount_in (u64):   用户销毁的 LP 代币数量
 *   - base_amount_out (u64):      用户取回的基础代币数量
 *   - quote_amount_out (u64):     用户取回的报价代币数量（单位 lamports）
 *   - pool (Pubkey 32字节):       目标交易池账户地址
 *   - user (Pubkey 32字节):       移除流动性的用户钱包地址
 */
func (ps *PumpSwapExtractor) parseWithdrawEvent(data []byte, tx *types.UnifiedTransaction, eventIdx int64) *model.Liquidity {
	// Minimum: 3*u64(24) + 2*Pubkey(64) = 88 bytes
	if len(data) < 88 {
		ps.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpSwap withdraw: data too short")
		return nil
	}

	off := 0

	var quoteAmountOut uint64
	_, off = dex.ParseU64LE(data, off) // lp_token_amount_in (skip)
	_, off = dex.ParseU64LE(data, off) // base_amount_out (skip, unknown decimals)
	quoteAmountOut, off = dex.ParseU64LE(data, off)

	var pool, user string
	pool, off = dex.ParsePubkey(data, off)
	user, off = dex.ParsePubkey(data, off)
	_ = off

	if pool == "" {
		ps.GetLogger().WithField("tx_hash", tx.TxHash).Debug("PumpSwap withdraw: failed to parse pool")
		return nil
	}

	// Use quote (SOL) amount for Value since base token value is unknown without price oracle
	quoteAmountBig := new(big.Int).SetUint64(quoteAmountOut)
	quoteValue := dex.LamportsToSOL(quoteAmountOut)

	key := fmt.Sprintf("%s_remove_%d", tx.TxHash, eventIdx)

	return &model.Liquidity{
		Addr:      pool,
		Router:    pumpSwapProgramID,
		Factory:   pumpSwapProgramID,
		Pool:      pool,
		Hash:      tx.TxHash,
		From:      user,
		Side:      "remove",
		Amount:    quoteAmountBig,
		Value:     quoteValue * 2, // Approximate: assume equal value on both sides
		Time:      uint64(tx.Timestamp.Unix()),
		Key:       key,
		Protocol:  "pumpswap",
		ChainType: string(types.ChainTypeSolana),
		Extra: &model.LiquidityExtra{
			Key:     key,
			Amounts: new(big.Int).SetUint64(quoteAmountOut),
			Values:  []float64{0, quoteValue}, // base value unknown, quote in SOL
			Time:    uint64(tx.Timestamp.Unix()),
		},
	}
}
