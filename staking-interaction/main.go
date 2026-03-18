package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"staking-interaction/contracts"
	"staking-interaction/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// 定义配置结构体
type StakeConfig struct {
	MyToken       *contracts.MyToken
	MtkContracts  *contracts.MtkContracts
	Auth          *bind.TransactOpts
	Client        *ethclient.Client
	TokenContract common.Address
	StakeContract common.Address
}

func approveStaking(config *StakeConfig, amount *big.Int) error {
	// 调用 MyToken 合约的 approve 方法
	tx, err := config.MyToken.Approve(config.Auth, config.StakeContract, amount)
	if err != nil {
		return fmt.Errorf("failed to approve staking: %w", err)
	}

	log.Printf("Approve transaction sent: %s", tx.Hash().Hex())

	// 等待交易确认
	receipt, err := bind.WaitMined(context.Background(), config.Client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction confirmation: %w", err)
	}

	log.Printf("Approve transaction confirmed. Block: %d, Gas Used: %d", receipt.BlockNumber.Uint64(), receipt.GasUsed)
	return nil
}

func transfer(config *StakeConfig, to common.Address, amount *big.Int) error {
	// 调用 MyToken 合约的 transfer 方法
	tx, err := config.MyToken.Transfer(config.Auth, to, amount)
	if err != nil {
		return fmt.Errorf("failed to transfer tokens: %w", err)
	}

	log.Printf("Transfer transaction sent: %s", tx.Hash().Hex())

	// 等待交易确认
	receipt, err := bind.WaitMined(context.Background(), config.Client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction confirmation: %w", err)
	}

	log.Printf("Transfer transaction confirmed. Block: %d, Gas Used: %d", receipt.BlockNumber.Uint64(), receipt.GasUsed)
	return nil
}

func balanceOf(config *StakeConfig, user common.Address) (*big.Int, error) {
	// 调用 MyToken 合约的 balanceOf 方法
	balance, err := config.MyToken.BalanceOf(&bind.CallOpts{}, user)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

func decimals(config *StakeConfig) (uint8, error) {
	// 调用 MyToken 合约的 decimals 方法
	decimals, err := config.MyToken.Decimals(&bind.CallOpts{})
	if err != nil {
		return 0, fmt.Errorf("failed to get decimals: %w", err)
	}

	return decimals, nil
}

func getUserActiveStakes(config *StakeConfig, user common.Address) ([]contracts.MtkContractsStake, error) {
	// 调用 MtkContracts 合约的 getUserActiveStakes 方法
	activeStakes, err := config.MtkContracts.GetUserActiveStakes(&bind.CallOpts{}, user)
	if err != nil {
		return nil, fmt.Errorf("failed to get user active stakes: %w", err)
	}

	return activeStakes, nil
}

func calculateReward(config *StakeConfig, user common.Address, stakeId *big.Int) (*big.Int, error) {
	// 调用 MtkContracts 合约的 calculateReward 方法
	reward, err := config.MtkContracts.CalculateReward(&bind.CallOpts{}, user, stakeId)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate reward: %w", err)
	}

	return reward, nil
}

func isStakeExpired(config *StakeConfig, user common.Address, stakeId *big.Int) (bool, error) {
	// 调用 MtkContracts 合约的 isStakeExpired 方法
	expired, err := config.MtkContracts.IsStakeExpired(&bind.CallOpts{}, user, stakeId)
	if err != nil {
		return false, fmt.Errorf("failed to check if stake is expired: %w", err)
	}

	return expired, nil
}

func stake(config *StakeConfig, amount *big.Int, stakingType uint8) error {
	// 调用 MtkContracts 合约的 stake 方法
	tx, err := config.MtkContracts.Stake(config.Auth, amount, stakingType)
	if err != nil {
		return fmt.Errorf("failed to stake tokens: %w", err)
	}

	log.Printf("Stake transaction sent: %s", tx.Hash().Hex())

	// 等待交易确认
	receipt, err := bind.WaitMined(context.Background(), config.Client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction confirmation: %w", err)
	}

	log.Printf("Stake transaction confirmed. Block: %d, Gas Used: %d", receipt.BlockNumber.Uint64(), receipt.GasUsed)
	return nil
}

func withdraw(config *StakeConfig, stakeId *big.Int) error {
	// 调用 MtkContracts 合约的 withdraw 方法
	tx, err := config.MtkContracts.Withdraw(config.Auth, stakeId)
	if err != nil {
		return fmt.Errorf("failed to withdraw tokens: %w", err)
	}

	log.Printf("Withdraw transaction sent: %s", tx.Hash().Hex())

	// 等待交易确认
	receipt, err := bind.WaitMined(context.Background(), config.Client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction confirmation: %w", err)
	}

	log.Printf("Withdraw transaction confirmed. Block: %d, Gas Used: %d", receipt.BlockNumber.Uint64(), receipt.GasUsed)
	return nil
}

func main() {
	// 初始化客户端
	client, err := ethclient.DialContext(context.Background(), "https://bsc-testnet-dataseed.bnbchain.org")
	if err != nil {
		log.Fatalf("Failed to connect to the BSC network: %v", err)
	}

	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	// 从环境变量获取私钥
	privateKeyHex := utils.GetEnv("SEPOLIA_PRIVATE_KEY", "")
	if privateKeyHex == "" {
		log.Fatal("PRIVATE_KEY environment variable is not set")
	}

	// 加载私钥
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	log.Printf("User address: %s", fromAddress.Hex())

	// 获取链ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}
	log.Printf("Chain ID: %d", chainID.Int64())

	// 创建授权事务
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}

	// 这里需要加载合约ABI和地址
	tokenContractAddress := common.HexToAddress("0x93480Ce4b54baD6c60D8CDAEaeaF898fE00deBF2")
	stakeContractAddress := common.HexToAddress("0x6287A4e265CfEA1B9C87C1dC692363d69f58378c")

	// 假设你已经将ABI编译为Go绑定
	myToken, err := contracts.NewMyToken(tokenContractAddress, client)
	if err != nil {
		log.Fatalf("Failed to create MyToken contract: %v", err)
	}

	mtkContracts, err := contracts.NewMtkContracts(stakeContractAddress, client)
	if err != nil {
		log.Fatalf("Failed to create MtkContracts contract: %v", err)
	}

	// 创建配置结构体
	config := &StakeConfig{
		MyToken:       myToken,
		MtkContracts:  mtkContracts,
		Auth:          auth,
		Client:        client,
		TokenContract: tokenContractAddress,
		StakeContract: stakeContractAddress,
	}

	log.Printf("Go Ethereum SDK初始化完成")

	// 查询用户的token余额
	balance, err := balanceOf(config, fromAddress)
	if err != nil {
		log.Fatalf("Failed to get token balance: %v", err)
	}
	decimals, err := decimals(config)
	if err != nil {
		log.Fatalf("Failed to get decimals: %v", err)
	}
	// 转换为可读格式（除以 10^decimals）
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	readableBalance := new(big.Int).Div(balance, divisor)
	log.Printf("User token balance: %s", readableBalance.String())

	// 用户向质押合约授权100个token
	amount := new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	err = approveStaking(config, amount)
	if err != nil {
		log.Fatalf("Failed to approve tokens: %v", err)
	}

	// 查询用户的质押信息
	stakes, err := getUserActiveStakes(config, fromAddress)
	if err != nil {
		log.Fatalf("Failed to get user active stakes: %v", err)
	}
	log.Printf("User active stakes: %v", stakes)

	if len(stakes) == 0 {
		// 用户进行质押100个token
		err = stake(config, amount, 0)
		if err != nil {
			log.Fatalf("Failed to stake tokens: %v", err)
		}
		log.Printf("Successfully staked %s tokens", amount.String())
	}

	for _, stake := range stakes {
		// 查询用户的质押奖励
		reward, err := calculateReward(config, fromAddress, stake.StakeId)
		if err != nil {
			log.Fatalf("Failed to get user reward: %v", err)
		}
		log.Printf("User %s reward: %s, for stake ID: %s", fromAddress.Hex(), reward.String(), stake.StakeId.String())

		// 查询质押合约token余额
		contractBalance, err := balanceOf(config, config.StakeContract)
		if err != nil {
			log.Fatalf("Failed to get contract token balance: %v", err)
		}
		log.Printf("Contract token balance: %s", contractBalance.String())

		// 向质押合约转账
		if reward.Cmp(contractBalance) > 0 {
			transferAmount := new(big.Int).Sub(reward, contractBalance)
			err = transfer(config, config.StakeContract, transferAmount)
			if err != nil {
				log.Fatalf("Failed to transfer tokens to stake contract: %v", err)
			}
			log.Printf("Successfully transferred %s tokens to stake contract %s", transferAmount.String(), config.StakeContract.Hex())
		}

		// 用户进行奖励提现
		isExpired, err := isStakeExpired(config, fromAddress, stake.StakeId)
		if err != nil {
			log.Fatalf("Failed to check if stake is expired: %v", err)
		}
		if !isExpired {
			log.Printf("Stake ID %s is not expired", stake.StakeId.String())
			continue
		}

		err = withdraw(config, stake.StakeId)
		if err != nil {
			log.Fatalf("Failed to withdraw reward: %v", err)
		}
		log.Printf("Successfully withdrew reward for stake ID: %s", stake.StakeId.String())
	}

	// 查询用户的token余额
	balance, err = balanceOf(config, fromAddress)
	if err != nil {
		log.Fatalf("Failed to get token balance: %v", err)
	}
	readableBalance = new(big.Int).Div(balance, divisor)
	log.Printf("User token balance: %s", readableBalance.String())
}
