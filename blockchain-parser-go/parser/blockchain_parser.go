package parser

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"blockchain-parser-go/config"
	"blockchain-parser-go/database"
	"blockchain-parser-go/redis"
	blockchainTypes "blockchain-parser-go/types"
	"blockchain-parser-go/utils"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const ERC20_ABI_JSON = `[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"}]`

type BlockchainParser struct {
	client             *ethclient.Client
	usdtABI            abi.ABI
	projectAddress     common.Address
	usdtAddress        common.Address
	db                 *database.DB
	redis              *redis.RedisClient
	scanInterval       time.Duration
	confirmationBlocks int64
	isProcessing       bool
}

func NewBlockchainParser(cfg *config.Config, db *database.DB, redis *redis.RedisClient) (*BlockchainParser, error) {
	client, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return nil, err
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}
	log.Printf("ChainID: %d", chainID)

	myERC20ABIData, err := os.ReadFile("abis/MyERC20.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read ABI: %w", err)
	}

	usdtABI, err := abi.JSON(bytes.NewReader(myERC20ABIData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return &BlockchainParser{
		client:             client,
		usdtABI:            usdtABI,
		projectAddress:     common.HexToAddress(cfg.ProjectAddress),
		usdtAddress:        common.HexToAddress(cfg.USDTAddress),
		db:                 db,
		redis:              redis,
		scanInterval:       time.Duration(cfg.ScanInterval) * time.Millisecond,
		confirmationBlocks: int64(cfg.ConfirmationBlocks),
	}, nil
}

func (p *BlockchainParser) InitData() error {
	log.Println("Initializing data...")

	user1Address := common.HexToAddress("0x6687e46C68C00bd1C10F8cc3Eb000B1752737e94")
	user2Address := common.HexToAddress("0x3DA0E0Aa4Db801ca5161412513558E9096E288bA")
	user3Address := common.HexToAddress("0x24ca20D955cc2B6f722eCfdE9A7053435885dDe9")

	// 初始化user1
	user1Balance, err := p.client.BalanceAt(context.Background(), user1Address, nil)
	if err != nil {
		return err
	}
	user1USDTBalance, err := utils.ReadContract(
		p.client,
		p.usdtABI,
		p.usdtAddress,
		"balanceOf",
		user1Address,
	)
	if err != nil {
		return err
	}
	user1USDTBalanceInt, ok := user1USDTBalance[0].(*big.Int)
	if !ok {
		return fmt.Errorf("failed to convert user1USDTBalance to *big.Int")
	}
	p.updateOrCreateUser(user1Address.Hex(), "user1@example.com", user1Balance, user1USDTBalanceInt)

	// 初始化user2
	user2Balance, err := p.client.BalanceAt(context.Background(), user2Address, nil)
	if err != nil {
		return err
	}
	user2USDTBalance, err := utils.ReadContract(
		p.client,
		p.usdtABI,
		p.usdtAddress,
		"balanceOf",
		user2Address,
	)
	if err != nil {
		return err
	}
	user2USDTBalanceInt, ok := user2USDTBalance[0].(*big.Int)
	if !ok {
		return fmt.Errorf("failed to convert user2USDTBalance to *big.Int")
	}
	p.updateOrCreateUser(user2Address.Hex(), "user2@example.com", user2Balance, user2USDTBalanceInt)

	// 初始化user3
	user3Balance, err := p.client.BalanceAt(context.Background(), user3Address, nil)
	if err != nil {
		return err
	}
	user3USDTBalance, err := utils.ReadContract(
		p.client,
		p.usdtABI,
		p.usdtAddress,
		"balanceOf",
		user3Address,
	)
	if err != nil {
		return err
	}
	user3USDTBalanceInt, ok := user3USDTBalance[0].(*big.Int)
	if !ok {
		return fmt.Errorf("failed to convert user3USDTBalance to *big.Int")
	}
	p.updateOrCreateUser(user3Address.Hex(), "user3@example.com", user3Balance, user3USDTBalanceInt)

	log.Println("Initialized 3 account records")

	// 初始化withdrawal表
	// 查询Withdrawal表是否为空
	withdrawalCount, err := p.db.WithdrawalCounts()
	if err != nil {
		return err
	}
	if withdrawalCount == 0 {
		log.Println("Withdrawal table is empty, initializing...")

		user1, err := p.db.GetAccountByAddress(user1Address.Hex())
		if err != nil {
			return err
		}
		user2, err := p.db.GetAccountByAddress(user2Address.Hex())
		if err != nil {
			return err
		}

		if user1 != nil && user2 != nil {
			// 初始化user1的withdrawal记录
			bnbAmount, err := utils.StringToWei("0.0001", 18) // 0.0001 BNBAmount
			if err != nil {
				return err
			}
			withdrawal := &blockchainTypes.Withdrawal{
				AccountID:    user1.ID,
				Amount:       bnbAmount,
				TokenDecimal: 18,
				TokenSymbol:  "BNB",
				ToAddress:    user3Address.Hex(),
			}
			if err := p.db.CreateWithdrawal(withdrawal); err != nil {
				return err
			}

			// 初始化user2的withdrawal记录
			erc20Amount, err := utils.StringToWei("10", 18) // 10 ERC20
			if err != nil {
				return err
			}

			withdrawal = &blockchainTypes.Withdrawal{
				AccountID:    user2.ID,
				Amount:       erc20Amount,
				TokenDecimal: 18,
				TokenSymbol:  "USDT",
				ToAddress:    user3Address.Hex(),
			}
			if err := p.db.CreateWithdrawal(withdrawal); err != nil {
				return err
			}
		}
	} else {
		// 将withdrawal表所有数据的status设置为init
		if err := p.db.UpdateWithdrawalAllStatus(); err != nil {
			return err
		}
	}

	log.Println("Initialized withdrawal records")
	return nil
}

func (p *BlockchainParser) Start() {
	// 初始化最后处理的区块
	lastBlock, err := p.redis.GetLastProcessedBlock()
	if err != nil {
		log.Fatalf("Error getting last processed block: %v", err)
	}

	if lastBlock == 0 {
		currentBlockNumber, err := p.client.BlockNumber(context.Background())
		lastBlock = int64(currentBlockNumber)
		if err != nil {
			log.Fatalf("Error getting current block number: %v", err)
		}
		lastBlock = int64(lastBlock) - 100
		if err := p.redis.SetLastProcessedBlock(lastBlock); err != nil {
			log.Fatalf("Error setting last processed block: %v", err)
		}
	}

	go p.processBlocks()
	log.Println("Blockchain parser started")
}

func (p *BlockchainParser) processBlocks() {
	ticker := time.NewTicker(p.scanInterval)
	defer ticker.Stop()

	for range ticker.C {
		if p.isProcessing {
			continue
		}

		p.isProcessing = true
		if err := p.processNewBlocks(); err != nil {
			log.Printf("Error processing blocks: %v", err)
		}
		p.isProcessing = false
	}
}

func (p *BlockchainParser) processNewBlocks() error {
	lastBlock, err := p.redis.GetLastProcessedBlock()
	if err != nil {
		return err
	}

	header, err := p.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return err
	}

	currentBlock := header.Number.Int64()
	confirmedBlock := currentBlock - p.confirmationBlocks

	if lastBlock >= confirmedBlock {
		return nil
	}

	log.Printf("Processing blocks from %d to %d", lastBlock+1, confirmedBlock)

	for blockNumber := lastBlock + 1; blockNumber <= confirmedBlock; blockNumber++ {
		if err := p.processBlock(big.NewInt(blockNumber)); err != nil {
			log.Printf("Error processing block %d: %+v", blockNumber, err)
			continue
		}
		if err := p.redis.SetLastProcessedBlock(blockNumber); err != nil {
			log.Printf("Error setting last processed block: %v", err)
		}
	}

	return nil
}

func (p *BlockchainParser) processBlock(blockNumber *big.Int) error {
	block, err := p.client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		return fmt.Errorf("failed to get block %d: %w", blockNumber.Int64(), err)
	}

	for _, tx := range block.Transactions() {
		// 检查是否是与相关方相关的交易
		from, err := getTransactionFrom(tx)
		if err != nil || from == "" {
			log.Printf("No from user in transaction %s, err: %v", tx.Hash().Hex(), err)
			return err
		}

		isFromProject := from == p.projectAddress.Hex()

		isToProject := false
		isToContract := false
		if tx.To() != nil {
			isToProject = tx.To().Hex() == p.projectAddress.Hex()
			isToContract = tx.To().Hex() == p.usdtAddress.Hex()
		}

		if isFromProject || isToProject || isToContract {
			if err := p.processTransaction(tx, blockNumber); err != nil {
				log.Printf("Error processing transaction %s: %v", tx.Hash().Hex(), err)
				return err
			}
		}
	}

	return nil
}

func (p *BlockchainParser) processTransaction(tx *types.Transaction, blockNumber *big.Int) error {
	txHash := tx.Hash().Hex()
	// Check if transaction already exists
	if exists, err := p.db.TransactionExists(txHash); err != nil || exists {
		return err
	}

	if processing, err := p.redis.IsTxProcessing(txHash); err != nil || processing {
		return err
	}

	if err := p.redis.SetTxProcessing(txHash, 5*time.Minute); err != nil {
		return err
	}
	defer p.redis.RemoveTxProcessing(txHash)

	receipt, err := p.client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		log.Printf("Error getting transaction receipt for %s: %v", txHash, err)
		return err
	}

	gasFee := new(big.Int).Mul(receipt.EffectiveGasPrice, new(big.Int).SetUint64(receipt.GasUsed))

	fromUserAddress, err := getTransactionFrom(tx)
	if err != nil || fromUserAddress == "" {
		return err
	}

	fromUser, _ := p.findUserByAddress(fromUserAddress)
	if fromUser != nil {
		if err = p.db.UpdateAccountBalance(fromUser.ID, "BNB", new(big.Int).Neg(gasFee)); err != nil {
			return fmt.Errorf("Failed to update BNB balance(gas fee) for user %d: %w", fromUser.ID, err)
		}

		log.Printf("Updated BNB balance for user %d %s: -%s BNB (gas fee)", fromUser.ID, fromUser.Address, utils.FormatWeiToEther(gasFee).String())
	}

	// 跳过合约创建交易
	if tx.To() == nil {
		return nil
	}

	// Process BNB transfer
	if len(tx.Data()) == 0 && tx.Value().Sign() > 0 {
		return p.processBNBTransfer(tx, receipt, blockNumber)
	}

	// Process ERC20 transfer
	if len(tx.Data()) > 0 && tx.To().Hex() == p.usdtAddress.Hex() {
		return p.processERC20Transfer(tx, receipt, blockNumber)
	}

	return nil
}

func (p *BlockchainParser) processBNBTransfer(tx *types.Transaction, receipt *types.Receipt, blockNumber *big.Int) error {
	from, err := getTransactionFrom(tx)
	if err != nil || from == "" {
		return err
	}

	fromUser, _ := p.findUserByAddress(from)

	to := tx.To().Hex()
	toUser, _ := p.findUserByAddress(to)

	// Save transaction
	transaction := &blockchainTypes.Transaction{
		TxHash:       tx.Hash().Hex(),
		BlockNumber:  blockNumber.Int64(),
		FromAddress:  from,
		ToAddress:    to,
		Value:        tx.Value(),
		TokenDecimal: 18,
		TokenSymbol:  "BNB",
		Status:       int(receipt.Status),
		CreatedTime:  time.Now(),
	}

	if err := p.db.SaveTransaction(transaction); err != nil {
		return err
	}

	// Update user balance if transaction succeeded
	if receipt.Status == 1 {
		amountInEther := utils.FormatWeiToEther(tx.Value()).String()

		if fromUser != nil {
			if err := p.db.UpdateAccountBalance(fromUser.ID, "BNB", new(big.Int).Neg(tx.Value())); err != nil {
				return fmt.Errorf("Failed to update BNB balance for user %d %s: %w", fromUser.ID, fromUser.Address, err)
			}
			log.Printf("Updated BNB balance for user %d %s: -%s BNB", fromUser.ID, fromUser.Address, amountInEther)
		}

		if toUser != nil {
			if err := p.db.UpdateAccountBalance(toUser.ID, "BNB", tx.Value()); err != nil {
				return fmt.Errorf("Failed to update BNB balance for user %d %s: %w", toUser.ID, toUser.Address, err)
			}
			log.Printf("Updated BNB balance for user %d %s: +%s BNB", toUser.ID, toUser.Address, amountInEther)
		}
	}

	return nil
}

func (p *BlockchainParser) processERC20Transfer(tx *types.Transaction, receipt *types.Receipt, blockNumber *big.Int) error {
	_, err := getTransactionFrom(tx)
	if err != nil {
		return err
	}

	decimalsData, err := utils.ReadContract(p.client, p.usdtABI, p.usdtAddress, "decimals")
	if err != nil {
		return err
	}
	decimals, ok := decimalsData[0].(uint8)
	if !ok {
		return fmt.Errorf("decimals is not uint8 type")
	}

	symbolData, err := utils.ReadContract(p.client, p.usdtABI, p.usdtAddress, "symbol")
	if err != nil {
		return err
	}
	symbol, ok := symbolData[0].(string)
	if !ok {
		return fmt.Errorf("symbol is not string type")
	}

	// Parse Transfer events
	for _, eventLog := range receipt.Logs {
		if eventLog.Address.Hex() != p.usdtAddress.Hex() {
			log.Printf("Skipping log with non-USDT address: %s", eventLog.Address.Hex())
			continue
		}

		if len(eventLog.Topics) < 3 {
			log.Printf("Skipping log with less than 3 topics: %s", eventLog.Topics)
			continue
		}

		event, err := p.usdtABI.EventByID(eventLog.Topics[0])
		if err != nil || event.Name != "Transfer" {
			log.Printf("Skipping log with non-Transfer event: %s, err: %v", event.Name, err)
			continue
		}

		// transferEvent := struct {
		// 	From  common.Address
		// 	To    common.Address
		// 	Value *big.Int
		// }{}

		from := common.HexToAddress(eventLog.Topics[1].Hex()).Hex()
		to := common.HexToAddress(eventLog.Topics[2].Hex()).Hex()

		// 解析 value（非 indexed 参数，在 Data 中）
		// 必须使用结构体包装，即使只有一个字段
		transferValue := struct {
			Value *big.Int
		}{}

		if err := p.usdtABI.UnpackIntoInterface(&transferValue, "Transfer", eventLog.Data); err != nil {
			log.Printf("Failed to unpack Transfer event: %v", err)
			continue
		}

		// Check if transfer is from or to project address
		if from != p.projectAddress.Hex() && to != p.projectAddress.Hex() {
			log.Printf("Skipping log with non-project address: %s, %s", from, to)
			continue
		}

		fromUser, _ := p.findUserByAddress(from)

		toUser, _ := p.findUserByAddress(to)

		// Save transaction
		transaction := &blockchainTypes.Transaction{
			TxHash:       tx.Hash().Hex(),
			BlockNumber:  blockNumber.Int64(),
			FromAddress:  from,
			ToAddress:    to,
			Value:        transferValue.Value,
			TokenDecimal: decimals,
			TokenAddress: p.usdtAddress.Hex(),
			TokenSymbol:  symbol,
			Status:       int(receipt.Status),
			CreatedTime:  time.Now(),
		}

		if err := p.db.SaveTransaction(transaction); err != nil {
			return err
		}

		// Update user balance if transaction succeeded
		if receipt.Status == 1 {
			amountInEther := utils.FormatWeiToEther(transferValue.Value).String()

			if fromUser != nil {
				if err := p.db.UpdateAccountBalance(fromUser.ID, "USDT", new(big.Int).Neg(transferValue.Value)); err != nil {
					return fmt.Errorf("Failed to update %s balance for user %d %s: %w", symbol, fromUser.ID, fromUser.Address, err)
				}
				log.Printf("Updated %s balance for user %d %s: -%s %s", symbol, fromUser.ID, fromUser.Address, amountInEther, symbol)
			}

			if toUser != nil {
				if err := p.db.UpdateAccountBalance(toUser.ID, "USDT", transferValue.Value); err != nil {
					return fmt.Errorf("Failed to update %s balance for user %d %s: %w", symbol, toUser.ID, toUser.Address, err)
				}
				log.Printf("Updated %s balance for user %d %s: +%s %s", symbol, toUser.ID, toUser.Address, amountInEther, symbol)
			}
		}
	}

	return nil
}

func (p *BlockchainParser) findUserByAddress(address string) (*blockchainTypes.Account, error) {
	// Check cache first
	if userID, err := p.redis.GetCachedUserByAddress(address); err == nil && userID > 0 {
		// We only have ID from cache, need to get full account info from DB
		// For simplicity, we'll just query DB again
	}

	user, err := p.db.GetAccountByAddress(address)
	if err != nil {
		return nil, err
	}

	if user != nil {
		if err := p.redis.CacheUserAddress(address, user.ID); err != nil {
			log.Printf("Error caching user address: %v", err)
		}
	}

	return user, nil
}

func (p *BlockchainParser) updateOrCreateUser(address string, email string, bnbAmount *big.Int, usdtAmount *big.Int) error {
	user, err := p.findUserByAddress(address)
	if err != nil {
		return err
	}
	if user == nil {
		user = &blockchainTypes.Account{
			Address:    address,
			Email:      email,
			BNBAmount:  bnbAmount,
			USDTAmount: usdtAmount,
		}
		if err := p.db.CreateAccount(user); err != nil {
			return err
		}
	} else {
		user.BNBAmount = bnbAmount
		user.USDTAmount = usdtAmount
		if err := p.db.UpdateAccount(user); err != nil {
			return err
		}
	}
	return nil
}

func getTransactionFrom(tx *types.Transaction) (string, error) {
	// 显示交易类型
	txType := tx.Type()

	// // 根据交易类型选择合适的 Signer
	// var signer types.Signer

	// switch txType {
	// case types.LegacyTxType:
	// 	// Legacy 交易（type 0）
	// 	signer = types.NewEIP155Signer(tx.ChainId())
	// case types.AccessListTxType:
	// 	// EIP-2930 交易（type 1）- 带访问列表
	// 	signer = types.NewEIP2930Signer(tx.ChainId())
	// case types.DynamicFeeTxType:
	// 	// EIP-1559 交易（type 2）- 动态手续费
	// 	signer = types.NewLondonSigner(tx.ChainId())
	// case 3:
	// 	// EIP-4844 Blob 交易（type 3）- Cancun 升级引入
	// 	// 注意：需要 go-ethereum v1.13.0+ 才支持 CancunSigner
	// 	signer = types.NewCancunSigner(tx.ChainId())
	// case 4:
	// 	// EIP 7702 交易（type 4）
	// 	// EIP-7702旨在通过允许EOAs将其执行委托给智能合约来解决这些用户体验问题，有效地赋予它们可编程能力，而无需用户迁移到全新的钱包
	// 	signer = types.NewPragueSigner(tx.ChainId())
	// default:
	// 	// 未知类型，尝试使用 London Signer（向后兼容）
	// 	signer = types.NewLondonSigner(tx.ChainId())
	// }

	cfg := config.LoadConfig()
	signer := types.NewPragueSigner(big.NewInt(cfg.ChainID))

	from, err := types.Sender(signer, tx)
	if err != nil {
		log.Printf("[DEBUG] Failed to get sender for transaction %s, type: %d, err: %v", tx.Hash().Hex(), txType, err)
		return "", err
	}
	return from.Hex(), nil
}
