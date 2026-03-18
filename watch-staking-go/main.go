package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// 配置常量
const (
	StakingContractAddress = "0x6287A4e265CfEA1B9C87C1dC692363d69f58378c"
	RPCURL                 = "wss://bnb-testnet.g.alchemy.com/v2/Z7oyiL7IXfGadWb_-cwpe"
	ReconnectInterval      = 5 * time.Second
	MaxReconnectAttempts   = 10
)

// 事件类型
type EventType int

const (
	AllEvents EventType = iota
	StakedEvent
	WithdrawnEvent
)

// 监听配置
type WatchConfig struct {
	UserAddress common.Address
	EventType   EventType
}

// 事件监听器
type StakingEventListener struct {
	client            *ethclient.Client
	config            WatchConfig
	contractAddress   common.Address
	stakedFilter      ethereum.FilterQuery
	withdrawnFilter   ethereum.FilterQuery
	isRunning         bool
	reconnectAttempts int
	mu                sync.RWMutex
	cancel            context.CancelFunc
}

// 合约事件结构定义（对应 MtkContracts ABI）
type MtkContracts struct {
	Abi []byte
}

// StakedEventData 事件数据
type StakedEventData struct {
	User      common.Address
	StakeId   *big.Int
	Amount    *big.Int
	Period    uint8
	Timestamp *big.Int
	Raw       types.Log
}

// WithdrawnEventData 事件数据
type WithdrawnEventData struct {
	User        common.Address
	StakeId     *big.Int
	Principal   *big.Int
	Reward      *big.Int
	TotalAmount *big.Int
	Raw         types.Log
}

// ABI 定义（从 JSON 文件加载）
var MtkContractsABI []byte

func init() {
	// 从 abis 目录加载 ABI
	abiBytes, err := os.ReadFile("abis/MtkContracts.json")
	if err != nil {
		log.Fatalf("Failed to load ABI: %v", err)
	}
	MtkContractsABI = abiBytes
}

// 创建新的监听器
func NewStakingEventListener(config WatchConfig) (*StakingEventListener, error) {
	client, err := ethclient.DialContext(context.Background(), RPCURL)
	if err != nil {
		log.Fatalf("Failed to connect to the BSC network: %v", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := client.ChainID(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to network: %w", err)
	}
	log.Println("✅ 连接到 BSC 测试网成功")

	contractAddress := common.HexToAddress(StakingContractAddress)
	listener := &StakingEventListener{
		client:          client,
		config:          config,
		contractAddress: contractAddress,
	}

	// 创建过滤器
	if err := listener.createFilters(); err != nil {
		return nil, fmt.Errorf("failed to create filters: %w", err)
	}

	return listener, nil
}

// 创建事件过滤器
func (l *StakingEventListener) createFilters() error {
	// 计算 Staked 事件签名
	stakedSig := []byte("Staked(address,uint256,uint256,uint8,uint256)")
	stakedHash := crypto.Keccak256Hash(stakedSig)

	// 计算 Withdrawn 事件签名
	withdrawnSig := []byte("Withdrawn(address,uint256,uint256,uint256,uint256)")
	withdrawnHash := crypto.Keccak256Hash(withdrawnSig)

	// Staked 事件过滤器
	stakedTopics := [][]common.Hash{
		{stakedHash}, // 事件签名
	}
	if l.config.UserAddress != (common.Address{}) {
		// 添加用户地址过滤（第一个参数是 indexed）
		stakedTopics = append(stakedTopics, []common.Hash{common.BytesToHash(l.config.UserAddress[:])})
	}
	l.stakedFilter = ethereum.FilterQuery{
		Addresses: []common.Address{l.contractAddress},
		Topics:    stakedTopics,
	}

	// Withdrawn 事件过滤器
	withdrawnTopics := [][]common.Hash{
		{withdrawnHash}, // 事件签名
	}
	if l.config.UserAddress != (common.Address{}) {
		// 添加用户地址过滤（第一个参数是 indexed）
		withdrawnTopics = append(withdrawnTopics, []common.Hash{common.BytesToHash(l.config.UserAddress[:])})
	}
	l.withdrawnFilter = ethereum.FilterQuery{
		Addresses: []common.Address{l.contractAddress},
		Topics:    withdrawnTopics,
	}

	return nil
}

// 处理 Staked 事件
func (l *StakingEventListener) handleStaked(log types.Log) error {
	if l.config.EventType == WithdrawnEvent {
		return nil
	}

	// 解析事件数据
	event := struct {
		User      common.Address
		StakeId   *big.Int
		Amount    *big.Int
		Period    uint8
		Timestamp *big.Int
	}{}

	if len(log.Data) > 0 {
		// 解析非 indexed 参数（data 字段）
		// ABI 编码规则：每个参数占据 32 字节槽
		// stakeId: bytes 0-32 (uint256)
		// amount: bytes 32-64 (uint256)
		// period: bytes 64-96 (uint8，但填充到32字节)
		// timestamp: bytes 96-128 (uint256)

		if len(log.Data) < 128 {
			return fmt.Errorf("invalid data length: %d, expected 128", len(log.Data))
		}

		stakeId := new(big.Int).SetBytes(log.Data[0:32])
		amount := new(big.Int).SetBytes(log.Data[32:64])
		period := uint8(log.Data[64]) // uint8 只占第一个字节
		timestamp := new(big.Int).SetBytes(log.Data[96:128])

		event.StakeId = stakeId
		event.Amount = amount
		event.Period = period
		event.Timestamp = timestamp
	}

	// indexed 参数从 Topics 获取
	if len(log.Topics) > 1 {
		event.User = common.HexToAddress(log.Topics[1].Hex())
	}

	fmt.Println("\n📈 ===== Staked 事件 =====")
	fmt.Printf("用户: %s\n", event.User.Hex())
	fmt.Printf("质押ID: %s\n", event.StakeId.String())
	fmt.Printf("质押金额: %s Tokens\n", formatWei(event.Amount))
	fmt.Printf("期限类型: %d (0=30天, 1=90天, 2=180天, 3=1年)\n", event.Period)
	fmt.Printf("时间戳: %s\n", time.Unix(event.Timestamp.Int64(), 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("交易哈希: %s\n", log.TxHash.Hex())
	fmt.Printf("区块号: %d\n", log.BlockNumber)
	fmt.Println("========================")

	return nil
}

// 处理 Withdrawn 事件
func (l *StakingEventListener) handleWithdrawn(log types.Log) error {
	if l.config.EventType == StakedEvent {
		return nil
	}

	// 解析事件数据
	event := struct {
		User        common.Address
		StakeId     *big.Int
		Principal   *big.Int
		Reward      *big.Int
		TotalAmount *big.Int
	}{}

	if len(log.Data) > 0 {
		// 解析非 indexed 参数（data 字段）
		// stakeId (32 bytes) + principal (32 bytes) + reward (32 bytes) + totalAmount (32 bytes)
		stakeId := new(big.Int).SetBytes(log.Data[0:32])
		principal := new(big.Int).SetBytes(log.Data[32:64])
		reward := new(big.Int).SetBytes(log.Data[64:96])
		totalAmount := new(big.Int).SetBytes(log.Data[96:128])

		event.StakeId = stakeId
		event.Principal = principal
		event.Reward = reward
		event.TotalAmount = totalAmount
	}

	// indexed 参数从 Topics 获取
	if len(log.Topics) > 1 {
		event.User = common.HexToAddress(log.Topics[1].Hex())
	}

	fmt.Println("\n💰 ===== Withdrawn 事件 =====")
	fmt.Printf("用户: %s\n", event.User.Hex())
	fmt.Printf("质押ID: %s\n", event.StakeId.String())
	fmt.Printf("本金: %s Tokens\n", formatWei(event.Principal))
	fmt.Printf("奖励: %s Tokens\n", formatWei(event.Reward))
	fmt.Printf("总金额: %s Tokens\n", formatWei(event.TotalAmount))
	fmt.Printf("交易哈希: %s\n", log.TxHash.Hex())
	fmt.Printf("区块号: %d\n", log.BlockNumber)
	fmt.Println("============================")

	return nil
}

// 格式化 Wei 为 Token
func formatWei(wei *big.Int) string {
	if wei == nil {
		return "0"
	}

	// 假设 18 位小数
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	before := new(big.Int).Div(wei, divisor)
	after := new(big.Int).Mod(wei, divisor)

	if after.Sign() == 0 {
		return before.String()
	}

	// 格式化为最多 4 位小数
	afterStr := after.String()
	for len(afterStr) < 18 {
		afterStr = "0" + afterStr
	}
	afterStr = afterStr[:4] // 取前 4 位
	afterStr = trimRight(afterStr, "0")

	if afterStr == "" {
		return before.String()
	}
	return fmt.Sprintf("%s.%s", before.String(), afterStr)
}

func trimRight(s, cutset string) string {
	for len(s) > 0 && s[len(s)-1:] == cutset {
		s = s[:len(s)-1]
	}
	return s
}

// 开始监听事件
func (l *StakingEventListener) Start(ctx context.Context) error {
	l.mu.Lock()
	l.isRunning = true
	l.mu.Unlock()

	fmt.Printf("\n🎯 开始监听质押合约事件...\n")
	fmt.Printf("合约地址: %s\n", l.contractAddress.Hex())
	if l.config.UserAddress != (common.Address{}) {
		fmt.Printf("过滤用户: %s\n", l.config.UserAddress.Hex())
	}
	if l.config.EventType != AllEvents {
		eventType := map[EventType]string{
			StakedEvent:    "Staked",
			WithdrawnEvent: "Withdrawn",
		}
		fmt.Printf("事件类型: %s\n", eventType[l.config.EventType])
	}
	fmt.Println("等待事件...")

	// 创建订阅
	ctx, cancel := context.WithCancel(ctx)
	l.cancel = cancel

	// 启动监听协程
	go l.watchLoop(ctx)

	return nil
}

// 监听循环（带重连）
func (l *StakingEventListener) watchLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("⏹️  监听已停止")
			return
		default:
			if err := l.watchEvents(ctx); err != nil {
				log.Printf("❌ 监听错误: %v", err)

				if l.reconnectAttempts >= MaxReconnectAttempts {
					log.Printf("❌ 达到最大重连次数 (%d)，停止尝试", MaxReconnectAttempts)
					return
				}

				l.reconnectAttempts++
				log.Printf("🔄 %d 秒后尝试重连... (%d/%d)",
					ReconnectInterval/time.Second, l.reconnectAttempts, MaxReconnectAttempts)

				select {
				case <-ctx.Done():
					return
				case <-time.After(ReconnectInterval):
					continue
				}
			}
		}
	}
}

// 监听事件
func (l *StakingEventListener) watchEvents(ctx context.Context) error {
	// 创建订阅
	var subStaked ethereum.Subscription
	var subWithdrawn ethereum.Subscription
	var err error

	// 订阅 Staked 事件
	if l.config.EventType != WithdrawnEvent {
		logs := make(chan types.Log)
		subStaked, err = l.client.SubscribeFilterLogs(ctx, l.stakedFilter, logs)
		if err != nil {
			return fmt.Errorf("failed to subscribe Staked events: %w", err)
		}
		defer subStaked.Unsubscribe()

		// 处理 Staked 事件
		go func() {
			for {
				select {
				case err := <-subStaked.Err():
					if err != nil {
						log.Printf("Staked 订阅错误: %v", err)
					}
					return
				case vLog := <-logs:
					if err := l.handleStaked(vLog); err != nil {
						log.Printf("处理 Staked 事件错误: %v", err)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// 订阅 Withdrawn 事件
	if l.config.EventType != StakedEvent {
		logs := make(chan types.Log)
		subWithdrawn, err = l.client.SubscribeFilterLogs(ctx, l.withdrawnFilter, logs)
		if err != nil {
			return fmt.Errorf("failed to subscribe Withdrawn events: %w", err)
		}
		defer subWithdrawn.Unsubscribe()

		// 处理 Withdrawn 事件
		go func() {
			for {
				select {
				case err := <-subWithdrawn.Err():
					if err != nil {
						log.Printf("Withdrawn 订阅错误: %v", err)
					}
					return
				case vLog := <-logs:
					if err := l.handleWithdrawn(vLog); err != nil {
						log.Printf("处理 Withdrawn 事件错误: %v", err)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	l.reconnectAttempts = 0 // 重置重连计数

	// 等待上下文取消或订阅错误
	<-ctx.Done()
	return nil
}

// 停止监听
func (l *StakingEventListener) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.cancel != nil {
		l.cancel()
	}

	l.isRunning = false
	if l.client != nil {
		l.client.Close()
	}
	log.Println("⏹️  监听已停止")
}

func main() {
	// 从环境变量读取配置
	config := WatchConfig{
		EventType: AllEvents,
	}

	// 可选：过滤特定用户地址
	if userAddr := os.Getenv("USER_ADDRESS"); userAddr != "" {
		config.UserAddress = common.HexToAddress(userAddr)
	}

	// 可选：过滤事件类型
	if eventType := os.Getenv("EVENT_TYPE"); eventType != "" {
		switch eventType {
		case "Staked":
			config.EventType = StakedEvent
		case "Withdrawn":
			config.EventType = WithdrawnEvent
		}
	}

	// 创建监听器
	listener, err := NewStakingEventListener(config)
	if err != nil {
		log.Fatalf("创建监听器失败: %v", err)
	}

	// 设置信号处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n⚠️  收到退出信号，正在关闭...")
		cancel()
		listener.Stop()
		os.Exit(0)
	}()

	// 开始监听
	if err := listener.Start(ctx); err != nil {
		log.Fatalf("启动监听失败: %v", err)
	}

	// 保持程序运行
	<-ctx.Done()
}
