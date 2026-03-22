package service

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
	blockchainTypes "blockchain-parser-go/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type WithdrawalService struct {
	client       *ethclient.Client
	wallet       *bind.TransactOpts
	usdtAddress  common.Address
	usdtABI      abi.ABI
	db           *database.DB
	scanInterval time.Duration
}

func NewWithdrawalService(cfg *config.Config, db *database.DB) (*WithdrawalService, error) {
	client, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	chainID := big.NewInt(cfg.ChainID)
	wallet, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, err
	}

	myERC20ABIData, err := os.ReadFile("abis/MyERC20.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read ABI: %w", err)
	}

	usdtABI, err := abi.JSON(bytes.NewReader(myERC20ABIData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return &WithdrawalService{
		client:       client,
		wallet:       wallet,
		usdtAddress:  common.HexToAddress(cfg.USDTAddress),
		usdtABI:      usdtABI,
		db:           db,
		scanInterval: time.Duration(cfg.ScanInterval) * time.Millisecond,
	}, nil
}

func (s *WithdrawalService) Start() {
	go s.processWithdrawals()
	log.Println("Withdrawal service started")
}

func (s *WithdrawalService) processWithdrawals() {
	ticker := time.NewTicker(s.scanInterval)
	defer ticker.Stop()

	for range ticker.C {
		withdrawals, err := s.db.GetPendingWithdrawals()
		if err != nil {
			log.Printf("Error getting pending withdrawals: %v", err)
			continue
		}

		for _, withdrawal := range withdrawals {
			if err := s.processWithdrawal(withdrawal); err != nil {
				log.Printf("Error processing withdrawal %d: %v", withdrawal.ID, err)
			}
		}
	}
}

func (s *WithdrawalService) processWithdrawal(withdrawal *blockchainTypes.Withdrawal) error {
	// Update status to processing
	if err := s.db.UpdateWithdrawalStatus(withdrawal.ID, "processing", ""); err != nil {
		return err
	}

	var tx *types.Transaction
	var err error

	switch withdrawal.TokenSymbol {
	case "BNB":
		tx, err = s.withdrawBNB(withdrawal)
	case "USDT":
		tx, err = s.withdrawUSDT(withdrawal)
	default:
		err = fmt.Errorf("unsupported token symbol: %s", withdrawal.TokenSymbol)
	}

	if err != nil {
		s.db.UpdateWithdrawalStatus(withdrawal.ID, "failed", "")
		return err
	}

	// Update with transaction hash
	if err := s.db.UpdateWithdrawalStatus(withdrawal.ID, "processing", tx.Hash().Hex()); err != nil {
		return err
	}

	// Monitor transaction
	go s.monitorTransaction(tx, withdrawal)

	return nil
}

func (s *WithdrawalService) withdrawBNB(withdrawal *blockchainTypes.Withdrawal) (*types.Transaction, error) {
	nonce, err := s.client.PendingNonceAt(context.Background(), s.wallet.From)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// 获取当前 gas price
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(withdrawal.ToAddress),
		withdrawal.Amount,
		21000,
		gasPrice,
		nil,
	)

	signedTx, err := s.wallet.Signer(s.wallet.From, tx)
	if err != nil {
		return nil, err
	}

	if err := s.client.SendTransaction(context.Background(), signedTx); err != nil {
		log.Printf("[DEBUG] Failed to send transaction: %v", err)
		return nil, err
	}

	return signedTx, nil
}

func (s *WithdrawalService) withdrawUSDT(withdrawal *blockchainTypes.Withdrawal) (*types.Transaction, error) {
	// For USDT, we need to use the contract's transfer function
	// This is a simplified version - in practice you'd use the contract ABI
	// to properly encode the transfer function call
	nonce, err := s.client.PendingNonceAt(context.Background(), s.wallet.From)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// 获取当前 gas price
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	tx := types.NewTransaction(
		nonce,
		s.usdtAddress,
		big.NewInt(0), // ERC20 转账不需要发送原生代币（BNB），value 设为 0
		100000,
		gasPrice,
		s.encodeTransferCall(withdrawal.ToAddress, withdrawal.Amount),
	)

	signedTx, err := s.wallet.Signer(s.wallet.From, tx)
	if err != nil {
		return nil, err
	}

	if err := s.client.SendTransaction(context.Background(), signedTx); err != nil {
		log.Printf("[DEBUG] Failed to send transaction: %v", err)
		return nil, err
	}

	return signedTx, nil
}

func (s *WithdrawalService) monitorTransaction(tx *types.Transaction, withdrawal *blockchainTypes.Withdrawal) {
	// Wait for transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), s.client, tx)
	if err != nil {
		log.Printf("Error waiting for transaction %s: %v", tx.Hash().Hex(), err)
		s.db.UpdateWithdrawalStatus(withdrawal.ID, "failed", tx.Hash().Hex())
		return
	}

	if receipt.Status == 1 {
		s.db.UpdateWithdrawalStatus(withdrawal.ID, "success", tx.Hash().Hex())
		log.Printf("Withdrawal %d succeeded with tx %s", withdrawal.ID, tx.Hash().Hex())
	} else {
		s.db.UpdateWithdrawalStatus(withdrawal.ID, "failed", tx.Hash().Hex())
		switch withdrawal.TokenSymbol {
		case "BNB":
			s.db.UpdateAccountBalance(withdrawal.AccountID, "BNB", withdrawal.Amount)
		case "USDT":
			s.db.UpdateAccountBalance(withdrawal.AccountID, "USDT", withdrawal.Amount)
		}
		log.Printf("Withdrawal %d failed with tx %s, roll back account balance", withdrawal.ID, tx.Hash().Hex())
	}
}

func (s *WithdrawalService) encodeTransferCall(toAddress string, amount *big.Int) []byte {
	// This is a simplified version - in practice you'd use the ABI to properly encode
	// the function call: transfer(address,uint256)
	// For a real implementation, use the abi.JSON and abi.Pack methods
	data, err := s.usdtABI.Pack("transfer", common.HexToAddress(toAddress), amount)
	if err != nil {
		log.Printf("Failed to pack transfer call: %v", err)
		return []byte{}
	}
	return data
}
