package orderbookindexer

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/ProjectsTask/EasySwapBase/chain/chainclient"
	"github.com/ProjectsTask/EasySwapBase/chain/types"
	"github.com/ProjectsTask/EasySwapBase/logger/xzap"
	"github.com/ProjectsTask/EasySwapBase/ordermanager"
	"github.com/ProjectsTask/EasySwapBase/stores/gdb"
	"github.com/ProjectsTask/EasySwapBase/stores/gdb/orderbookmodel/base"
	"github.com/ProjectsTask/EasySwapBase/stores/gdb/orderbookmodel/multi"
	"github.com/ProjectsTask/EasySwapBase/stores/xkv"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/threading"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ProjectsTask/EasySwapSync/service/comm"
	"github.com/ProjectsTask/EasySwapSync/service/config"
)

const (
	EventIndexType   = 6
	SleepInterval    = 10 // in seconds
	SyncBlockPeriod  = 10
	LogMakeTopic     = "0xfc37f2ff950f95913eb7182357ba3c14df60ef354bc7d6ab1ba2815f249fffe6"
	LogCancelTopic   = "0x0ac8bb53fac566d7afc05d8b4df11d7690a7b27bdc40b54e4060f9b21fb849bd"
	LogMatchTopic    = "0xf629aecab94607bc43ce4aebd564bf6e61c7327226a797b002de724b9944b20e"
	contractAbi      = `[{"inputs":[],"name":"CannotFindNextEmptyKey","type":"error"},{"inputs":[],"name":"CannotFindPrevEmptyKey","type":"error"},{"inputs":[{"internalType":"OrderKey","name":"orderKey","type":"bytes32"}],"name":"CannotInsertDuplicateOrder","type":"error"},{"inputs":[],"name":"CannotInsertEmptyKey","type":"error"},{"inputs":[],"name":"CannotInsertExistingKey","type":"error"},{"inputs":[],"name":"CannotRemoveEmptyKey","type":"error"},{"inputs":[],"name":"CannotRemoveMissingKey","type":"error"},{"inputs":[],"name":"EnforcedPause","type":"error"},{"inputs":[],"name":"ExpectedPause","type":"error"},{"inputs":[],"name":"InvalidInitialization","type":"error"},{"inputs":[],"name":"NotInitializing","type":"error"},{"inputs":[{"internalType":"address","name":"owner","type":"address"}],"name":"OwnableInvalidOwner","type":"error"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"OwnableUnauthorizedAccount","type":"error"},{"inputs":[],"name":"ReentrancyGuardReentrantCall","type":"error"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"offset","type":"uint256"},{"indexed":false,"internalType":"bytes","name":"msg","type":"bytes"}],"name":"BatchMatchInnerError","type":"event"},{"anonymous":false,"inputs":[],"name":"EIP712DomainChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint64","name":"version","type":"uint64"}],"name":"Initialized","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"OrderKey","name":"orderKey","type":"bytes32"},{"indexed":true,"internalType":"address","name":"maker","type":"address"}],"name":"LogCancel","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"OrderKey","name":"orderKey","type":"bytes32"},{"indexed":true,"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"indexed":true,"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"indexed":true,"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"indexed":false,"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"indexed":false,"internalType":"Price","name":"price","type":"uint128"},{"indexed":false,"internalType":"uint64","name":"expiry","type":"uint64"},{"indexed":false,"internalType":"uint64","name":"salt","type":"uint64"}],"name":"LogMake","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"OrderKey","name":"makeOrderKey","type":"bytes32"},{"indexed":true,"internalType":"OrderKey","name":"takeOrderKey","type":"bytes32"},{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"indexed":false,"internalType":"structLibOrder.Order","name":"makeOrder","type":"tuple"},{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"indexed":false,"internalType":"structLibOrder.Order","name":"takeOrder","type":"tuple"},{"indexed":false,"internalType":"uint128","name":"fillPrice","type":"uint128"}],"name":"LogMatch","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"OrderKey","name":"orderKey","type":"bytes32"},{"indexed":false,"internalType":"uint64","name":"salt","type":"uint64"}],"name":"LogSkipOrder","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint128","name":"newProtocolShare","type":"uint128"}],"name":"LogUpdatedProtocolShare","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"recipient","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"LogWithdrawETH","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Paused","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Unpaused","type":"event"},{"inputs":[{"internalType":"OrderKey[]","name":"orderKeys","type":"bytes32[]"}],"name":"cancelOrders","outputs":[{"internalType":"bool[]","name":"successes","type":"bool[]"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"components":[{"internalType":"OrderKey","name":"oldOrderKey","type":"bytes32"},{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"newOrder","type":"tuple"}],"internalType":"structLibOrder.EditDetail[]","name":"editDetails","type":"tuple[]"}],"name":"editOrders","outputs":[{"internalType":"OrderKey[]","name":"newOrderKeys","type":"bytes32[]"}],"stateMutability":"payable","type":"function"},{"inputs":[],"name":"eip712Domain","outputs":[{"internalType":"bytes1","name":"fields","type":"bytes1"},{"internalType":"string","name":"name","type":"string"},{"internalType":"string","name":"version","type":"string"},{"internalType":"uint256","name":"chainId","type":"uint256"},{"internalType":"address","name":"verifyingContract","type":"address"},{"internalType":"bytes32","name":"salt","type":"bytes32"},{"internalType":"uint256[]","name":"extensions","type":"uint256[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"OrderKey","name":"","type":"bytes32"}],"name":"filledAmount","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"}],"name":"getBestOrder","outputs":[{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"orderResult","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"collection","type":"address"},{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"}],"name":"getBestPrice","outputs":[{"internalType":"Price","name":"price","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"collection","type":"address"},{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"Price","name":"price","type":"uint128"}],"name":"getNextBestPrice","outputs":[{"internalType":"Price","name":"nextBestPrice","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"uint256","name":"count","type":"uint256"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"OrderKey","name":"firstOrderKey","type":"bytes32"}],"name":"getOrders","outputs":[{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order[]","name":"resultOrders","type":"tuple[]"},{"internalType":"OrderKey","name":"nextOrderKey","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint128","name":"newProtocolShare","type":"uint128"},{"internalType":"address","name":"newVault","type":"address"},{"internalType":"string","name":"EIP712Name","type":"string"},{"internalType":"string","name":"EIP712Version","type":"string"}],"name":"initialize","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order[]","name":"newOrders","type":"tuple[]"}],"name":"makeOrders","outputs":[{"internalType":"OrderKey[]","name":"newOrderKeys","type":"bytes32[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"sellOrder","type":"tuple"},{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"buyOrder","type":"tuple"}],"name":"matchOrder","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"sellOrder","type":"tuple"},{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"buyOrder","type":"tuple"},{"internalType":"uint256","name":"msgValue","type":"uint256"}],"name":"matchOrderWithoutPayback","outputs":[{"internalType":"uint128","name":"costValue","type":"uint128"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"sellOrder","type":"tuple"},{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"buyOrder","type":"tuple"}],"internalType":"structLibOrder.MatchDetail[]","name":"matchDetails","type":"tuple[]"}],"name":"matchOrders","outputs":[{"internalType":"bool[]","name":"successes","type":"bool[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"enumLibOrder.Side","name":"","type":"uint8"},{"internalType":"Price","name":"","type":"uint128"}],"name":"orderQueues","outputs":[{"internalType":"OrderKey","name":"head","type":"bytes32"},{"internalType":"OrderKey","name":"tail","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"OrderKey","name":"","type":"bytes32"}],"name":"orders","outputs":[{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"order","type":"tuple"},{"internalType":"OrderKey","name":"next","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"pause","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"paused","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"enumLibOrder.Side","name":"","type":"uint8"}],"name":"priceTrees","outputs":[{"internalType":"Price","name":"root","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"protocolShare","outputs":[{"internalType":"uint128","name":"","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"renounceOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint128","name":"newProtocolShare","type":"uint128"}],"name":"setProtocolShare","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newVault","type":"address"}],"name":"setVault","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"unpause","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"withdrawETH","outputs":[],"stateMutability":"nonpayable","type":"function"},{"stateMutability":"payable","type":"receive"}]`
	FixForCollection = 0
	FixForItem       = 1
	List             = 0
	Bid              = 1

	HexPrefix   = "0x"
	ZeroAddress = "0x0000000000000000000000000000000000000000"
)

type Order struct {
	Side     uint8
	SaleKind uint8
	Maker    common.Address
	Nft      struct {
		TokenId        *big.Int
		CollectionAddr common.Address
		Amount         *big.Int
	}
	Price  *big.Int
	Expiry uint64
	Salt   uint64
}

type Service struct {
	ctx          context.Context
	cfg          *config.Config
	db           *gorm.DB
	kv           *xkv.Store
	orderManager *ordermanager.OrderManager
	chainClient  chainclient.ChainClient
	chainId      int64
	chain        string
	parsedAbi    abi.ABI
}

var MultiChainMaxBlockDifference = map[string]uint64{
	"eth":        1,
	"optimism":   2,
	"starknet":   1,
	"arbitrum":   2,
	"base":       2,
	"zksync-era": 2,
}

// New 是一个构造函数，用于创建一个新的 Service 实例。
// 该函数接收多个参数，包括上下文、配置、数据库连接、键值存储、链客户端、链ID、链名称和订单管理器，
// 并返回一个初始化后的 Service 指针。
// 参数:
// - ctx: 上下文环境，用于控制操作的生命周期。
// - cfg: 配置信息，包含服务运行所需的各种参数。
// - db: 数据库连接，用于与数据库交互。
// - xkv: 键值存储，用于存储和检索数据。
// - chainClient: 链客户端，用于与区块链进行交互。
// - chainId: 链的唯一标识符。
// - chain: 链的名称。
// - orderManager: 订单管理器，用于管理订单相关操作。
// 返回值:
// - *Service: 初始化后的 Service 指针。
func New(ctx context.Context, cfg *config.Config, db *gorm.DB, xkv *xkv.Store, chainClient chainclient.ChainClient, chainId int64, chain string, orderManager *ordermanager.OrderManager) *Service {
	// 通过ABI实例化parsedAbi，这里忽略了可能出现的错误
	parsedAbi, _ := abi.JSON(strings.NewReader(contractAbi))
	// 返回一个新的Service实例，包含传入的参数和解析后的ABI
	return &Service{
		// 上下文环境
		ctx: ctx,
		// 配置信息
		cfg: cfg,
		// 数据库连接
		db: db,
		// 键值存储
		kv: xkv,
		// 链客户端
		chainClient: chainClient,
		// 订单管理器
		orderManager: orderManager,
		// 链名称
		chain: chain,
		// 链ID
		chainId: chainId,
		// 解析后的ABI
		parsedAbi: parsedAbi,
	}
}

// Start 方法用于启动服务的主要功能，它会并发地启动两个关键的循环任务。
// 这两个任务分别是订单簿事件同步循环和集合底价变化的维护循环。
func (s *Service) Start() {
	// 启动一个安全的 goroutine 来执行订单簿事件同步循环。
	// SyncOrderBookEventLoop 方法会持续监听区块链上的订单簿事件，
	// 并根据事件类型更新数据库和订单管理队列。
	threading.GoSafe(s.SyncOrderBookEventLoop)
	// 启动另一个安全的 goroutine 来执行集合底价变化的维护循环。
	// UpKeepingCollectionFloorChangeLoop 方法会定期清理过期的集合底价变化数据，
	// 并更新集合的底价信息。
	threading.GoSafe(s.UpKeepingCollectionFloorChangeLoop)
}

// SyncOrderBookEventLoop 是一个服务方法，用于持续同步订单簿事件。
// 它从数据库中获取最后同步的区块高度，并与当前区块链的最新区块高度进行比较。
// 然后，它会从最后同步的区块开始，每次同步一定数量的区块（SyncBlockPeriod），直到达到当前区块高度。
// 对于每个同步的区块，它会获取该区块内的所有日志，并根据日志的主题（topic）调用相应的事件处理函数。
// 最后，它会更新数据库中最后同步的区块高度。
func (s *Service) SyncOrderBookEventLoop() {
	// 定义一个变量来存储索引状态
	var indexedStatus base.IndexedStatus
	// 从数据库中查询最后同步的区块高度
	if err := s.db.WithContext(s.ctx).Table(base.IndexedStatusTableName()).
		Where("chain_id = ? and index_type = ?", s.chainId, EventIndexType).
		First(&indexedStatus).Error; err != nil {
		// 如果查询失败，记录错误日志并返回
		xzap.WithContext(s.ctx).Error("failed on get listing index status",
			zap.Error(err))
		return
	}

	// 获取最后同步的区块高度
	lastSyncBlock := uint64(indexedStatus.LastIndexedBlock)
	// 进入无限循环，持续同步订单簿事件
	for {
		// 检查上下文是否被取消
		select {
		case <-s.ctx.Done():
			// 如果上下文被取消，记录日志并退出循环
			xzap.WithContext(s.ctx).Info("SyncOrderBookEventLoop stopped due to context cancellation")
			return
		default:
		}

		// 获取当前区块链的最新区块高度
		currentBlockNum, err := s.chainClient.BlockNumber()
		if err != nil {
			// 如果获取失败，记录错误日志并等待一段时间后重试
			xzap.WithContext(s.ctx).Error("failed on get current block number", zap.Error(err))
			time.Sleep(SleepInterval * time.Second)
			continue
		}

		// 检查最后同步的区块高度是否接近当前区块高度
		if lastSyncBlock > currentBlockNum-MultiChainMaxBlockDifference[s.chain] {
			// 如果接近，等待一段时间后重试
			time.Sleep(SleepInterval * time.Second)
			continue
		}

		// 计算本次同步的起始区块高度
		startBlock := lastSyncBlock
		// 计算本次同步的结束区块高度
		endBlock := startBlock + SyncBlockPeriod
		// 使用内置的 math.Min 函数更新结束区块高度，确保不超过当前区块高度减去最大允许的区块差异
		endBlock = uint64(math.Min(float64(endBlock), float64(currentBlockNum-MultiChainMaxBlockDifference[s.chain])))

		// 构建日志过滤查询条件
		query := types.FilterQuery{
			FromBlock: new(big.Int).SetUint64(startBlock),
			ToBlock:   new(big.Int).SetUint64(endBlock),
			Addresses: []string{s.cfg.ContractCfg.DexAddress},
		}

		// 根据查询条件获取日志
		logs, err := s.chainClient.FilterLogs(s.ctx, query)
		if err != nil {
			// 如果获取日志失败，记录错误日志并等待一段时间后重试
			xzap.WithContext(s.ctx).Error("failed on get log", zap.Error(err))
			time.Sleep(SleepInterval * time.Second)
			continue
		}

		// 遍历获取到的日志
		for _, log := range logs {
			// 将日志转换为以太坊日志类型
			ethLog := log.(ethereumTypes.Log)
			// 根据日志的主题（topic）调用相应的事件处理函数
			switch ethLog.Topics[0].String() {
			case LogMakeTopic:
				// 处理挂单事件
				s.handleMakeEvent(ethLog)
			case LogCancelTopic:
				// 处理取消订单事件
				s.handleCancelEvent(ethLog)
			case LogMatchTopic:
				// 处理订单匹配事件
				s.handleMatchEvent(ethLog)
			default:
			}
		}

		// 更新最后同步的区块高度
		lastSyncBlock = endBlock + 1
		// 将最后同步的区块高度更新到数据库中
		if err := s.db.WithContext(s.ctx).Table(base.IndexedStatusTableName()).
			Where("chain_id = ? and index_type = ?", s.chainId, EventIndexType).
			Update("last_indexed_block", lastSyncBlock).Error; err != nil {
			// 如果更新失败，记录错误日志并返回
			xzap.WithContext(s.ctx).Error("failed on update orderbook event sync block number",
				zap.Error(err))
			return
		}

		// 记录同步信息
		xzap.WithContext(s.ctx).Info("sync orderbook event ...",
			zap.Uint64("start_block", startBlock),
			zap.Uint64("end_block", endBlock))
	}
}

// 处理挂单事件
// handleMakeEvent 处理LogMake事件，当有新的订单挂单时触发。
// 该函数会解析事件数据，将订单信息存入数据库和活动表，并将订单添加到订单管理队列。
func (s *Service) handleMakeEvent(log ethereumTypes.Log) {
	// 定义一个结构体来存储解析后的事件数据
	var event struct {
		OrderKey [32]byte
		Nft      struct {
			TokenId        *big.Int
			CollectionAddr common.Address
			Amount         *big.Int
		}
		Price  *big.Int
		Expiry uint64
		Salt   uint64
	}

	// Unpack data
	// 通过ABI解析日志数据，将其存入event结构体
	err := s.parsedAbi.UnpackIntoInterface(&event, "LogMake", log.Data)
	if err != nil {
		// 如果解析失败，记录错误日志并返回
		xzap.WithContext(s.ctx).Error("Error unpacking LogMake event:", zap.Error(err))
		return
	}
	// Extract indexed fields from topics
	// 从日志的topics中提取索引字段
	side := uint8(new(big.Int).SetBytes(log.Topics[1].Bytes()).Uint64())
	saleKind := uint8(new(big.Int).SetBytes(log.Topics[2].Bytes()).Uint64())
	maker := common.BytesToAddress(log.Topics[3].Bytes())

	// 根据side和saleKind确定订单类型
	var orderType int64
	if side == Bid { // 买单
		if saleKind == FixForCollection { // 针对集合的买单
			orderType = multi.CollectionBidOrder
		} else { // 针对某个具体NFT的买单
			orderType = multi.ItemBidOrder
		}
	} else { // 卖单
		orderType = multi.ListingOrder
	}
	// 创建一个新的订单结构体
	newOrder := multi.Order{
		CollectionAddress: event.Nft.CollectionAddr.String(),
		MarketplaceId:     multi.MarketOrderBook,
		TokenId:           event.Nft.TokenId.String(),
		OrderID:           HexPrefix + hex.EncodeToString(event.OrderKey[:]),
		OrderStatus:       multi.OrderStatusActive,
		EventTime:         time.Now().Unix(),
		ExpireTime:        int64(event.Expiry),
		CurrencyAddress:   s.cfg.ContractCfg.EthAddress,
		Price:             decimal.NewFromBigInt(event.Price, 0),
		Maker:             maker.String(),
		Taker:             ZeroAddress,
		QuantityRemaining: event.Nft.Amount.Int64(),
		Size:              event.Nft.Amount.Int64(),
		OrderType:         orderType,
		Salt:              int64(event.Salt),
	}
	// 将订单信息存入数据库，如果订单已存在则不做处理
	if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&newOrder).Error; err != nil {
		// 如果存储失败，记录错误日志
		xzap.WithContext(s.ctx).Error("failed on create order",
			zap.Error(err))
	}
	// 获取该日志所在区块的时间
	blockTime, err := s.chainClient.BlockTimeByNumber(s.ctx, big.NewInt(int64(log.BlockNumber)))
	if err != nil {
		// 如果获取失败，记录错误日志并返回
		xzap.WithContext(s.ctx).Error("failed to get block time", zap.Error(err))
		return
	}
	// 根据side和saleKind确定活动类型
	var activityType int
	if side == Bid {
		if saleKind == FixForCollection {
			activityType = multi.CollectionBid
		} else {
			activityType = multi.ItemBid
		}
	} else {
		activityType = multi.Listing
	}
	// 创建一个新的活动结构体
	newActivity := multi.Activity{ // 将订单信息存入活动表
		ActivityType:      activityType,
		Maker:             maker.String(),
		Taker:             ZeroAddress,
		MarketplaceID:     multi.MarketOrderBook,
		CollectionAddress: event.Nft.CollectionAddr.String(),
		TokenId:           event.Nft.TokenId.String(),
		CurrencyAddress:   s.cfg.ContractCfg.EthAddress,
		Price:             decimal.NewFromBigInt(event.Price, 0),
		BlockNumber:       int64(log.BlockNumber),
		TxHash:            log.TxHash.String(),
		EventTime:         int64(blockTime),
	}
	// 将活动信息存入数据库，如果活动已存在则不做处理
	if err := s.db.WithContext(s.ctx).Table(multi.ActivityTableName(s.chain)).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&newActivity).Error; err != nil {
		// 如果存储失败，记录警告日志
		xzap.WithContext(s.ctx).Warn("failed on create activity",
			zap.Error(err))
	}

	// 将订单信息存入订单管理队列
	if err := s.orderManager.AddToOrderManagerQueue(&multi.Order{
		ExpireTime:        newOrder.ExpireTime,
		OrderID:           newOrder.OrderID,
		CollectionAddress: newOrder.CollectionAddress,
		TokenId:           newOrder.TokenId,
		Price:             newOrder.Price,
		Maker:             newOrder.Maker,
	}); err != nil {
		// 如果添加失败，记录错误日志
		xzap.WithContext(s.ctx).Error("failed on add order to manager queue",
			zap.Error(err),
			zap.String("order_id", newOrder.OrderID))
	}
}

// handleMatchEvent 处理LogMatch事件，当订单匹配成功时触发。
// 该函数会解析事件数据，更新订单状态、数量和NFT所有者信息，
// 并将交易信息存入活动表和价格更新队列。
func (s *Service) handleMatchEvent(log ethereumTypes.Log) {
	// 定义一个结构体，用于存储解包后的日志事件数据
	var event struct {
		MakeOrder Order
		TakeOrder Order
		FillPrice *big.Int
	}

	// 使用ABI解包日志数据
	err := s.parsedAbi.UnpackIntoInterface(&event, "LogMatch", log.Data)
	if err != nil {
		// 如果解包失败，记录错误日志并返回
		xzap.WithContext(s.ctx).Error("Error unpacking LogMatch event:", zap.Error(err))
		return
	}

	// 通过topic获取订单ID
	makeOrderId := HexPrefix + hex.EncodeToString(log.Topics[1].Bytes())
	takeOrderId := HexPrefix + hex.EncodeToString(log.Topics[2].Bytes())
	var owner string
	var collection string
	var tokenId string
	var from string
	var to string
	var sellOrderId string
	var buyOrder multi.Order

	// 判断订单是买单还是卖单
	if event.MakeOrder.Side == Bid { // 买单， 由卖方发起交易撮合
		// 设置NFT所有者
		owner = strings.ToLower(event.MakeOrder.Maker.String())
		// 设置NFT集合地址
		collection = event.TakeOrder.Nft.CollectionAddr.String()
		// 设置NFT的Token ID
		tokenId = event.TakeOrder.Nft.TokenId.String()
		// 设置交易发起方
		from = event.TakeOrder.Maker.String()
		// 设置交易接收方
		to = event.MakeOrder.Maker.String()
		// 设置卖方订单ID
		sellOrderId = takeOrderId

		// 更新卖方订单状态
		if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
			Where("order_id = ?", takeOrderId).
			Updates(map[string]interface{}{
				"order_status":       multi.OrderStatusFilled,
				"quantity_remaining": 0,
				"taker":              to,
			}).Error; err != nil {
			// 如果更新失败，记录错误日志并返回
			xzap.WithContext(s.ctx).Error("failed on update order status",
				zap.String("order_id", takeOrderId))
			return
		}

		// 查询买方订单信息，不存在则无需更新，说明不是从平台前端发起的交易
		if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
			Where("order_id = ?", makeOrderId).
			First(&buyOrder).Error; err != nil {
			// 如果查询失败，记录错误日志并返回
			xzap.WithContext(s.ctx).Error("failed on get buy order",
				zap.Error(err))
			return
		}
		// 更新买方订单的剩余数量
		if buyOrder.QuantityRemaining > 1 {
			if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
				Where("order_id = ?", makeOrderId).
				Update("quantity_remaining", buyOrder.QuantityRemaining-1).Error; err != nil {
				// 如果更新失败，记录错误日志并返回
				xzap.WithContext(s.ctx).Error("failed on update order quantity_remaining",
					zap.String("order_id", makeOrderId))
				return
			}
		} else {
			if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
				Where("order_id = ?", makeOrderId).
				Updates(map[string]interface{}{
					"order_status":       multi.OrderStatusFilled,
					"quantity_remaining": 0,
				}).Error; err != nil {
				// 如果更新失败，记录错误日志并返回
				xzap.WithContext(s.ctx).Error("failed on update order status",
					zap.String("order_id", makeOrderId))
				return
			}
		}
	} else { // 卖单， 由买方发起交易撮合， 同理
		// 设置NFT所有者
		owner = strings.ToLower(event.TakeOrder.Maker.String())
		// 设置NFT集合地址
		collection = event.MakeOrder.Nft.CollectionAddr.String()
		// 设置NFT的Token ID
		tokenId = event.MakeOrder.Nft.TokenId.String()
		// 设置交易发起方
		from = event.MakeOrder.Maker.String()
		// 设置交易接收方
		to = event.TakeOrder.Maker.String()
		// 设置卖方订单ID
		sellOrderId = makeOrderId

		// 更新卖方订单状态
		if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
			Where("order_id = ?", makeOrderId).
			Updates(map[string]interface{}{
				"order_status":       multi.OrderStatusFilled,
				"quantity_remaining": 0,
				"taker":              to,
			}).Error; err != nil {
			// 如果更新失败，记录错误日志并返回
			xzap.WithContext(s.ctx).Error("failed on update order status",
				zap.String("order_id", makeOrderId))
			return
		}

		// 查询买方订单信息，不存在则无需更新，说明不是从平台前端发起的交易
		if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
			Where("order_id = ?", takeOrderId).
			First(&buyOrder).Error; err != nil {
			// 如果查询失败，记录错误日志并返回
			xzap.WithContext(s.ctx).Error("failed on get buy order",
				zap.Error(err))
			return
		}
		// 更新买方订单的剩余数量
		if buyOrder.QuantityRemaining > 1 {
			if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
				Where("order_id = ?", takeOrderId).
				Update("quantity_remaining", buyOrder.QuantityRemaining-1).Error; err != nil {
				// 如果更新失败，记录错误日志并返回
				xzap.WithContext(s.ctx).Error("failed on update order quantity_remaining",
					zap.String("order_id", takeOrderId))
				return
			}
		} else {
			if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
				Where("order_id = ?", takeOrderId).
				Updates(map[string]interface{}{
					"order_status":       multi.OrderStatusFilled,
					"quantity_remaining": 0,
				}).Error; err != nil {
				// 如果更新失败，记录错误日志并返回
				xzap.WithContext(s.ctx).Error("failed on update order status",
					zap.String("order_id", takeOrderId))
				return
			}
		}
	}

	// 获取该日志所在区块的时间
	blockTime, err := s.chainClient.BlockTimeByNumber(s.ctx, big.NewInt(int64(log.BlockNumber)))
	if err != nil {
		// 如果获取失败，记录错误日志并返回
		xzap.WithContext(s.ctx).Error("failed to get block time", zap.Error(err))
		return
	}
	// 创建一个新的活动结构体
	newActivity := multi.Activity{
		ActivityType:      multi.Sale,
		Maker:             event.MakeOrder.Maker.String(),
		Taker:             event.TakeOrder.Maker.String(),
		MarketplaceID:     multi.MarketOrderBook,
		CollectionAddress: collection,
		TokenId:           tokenId,
		CurrencyAddress:   s.cfg.ContractCfg.EthAddress,
		Price:             decimal.NewFromBigInt(event.FillPrice, 0),
		BlockNumber:       int64(log.BlockNumber),
		TxHash:            log.TxHash.String(),
		EventTime:         int64(blockTime),
	}
	// 将活动信息存入数据库，如果活动已存在则不做处理
	if err := s.db.WithContext(s.ctx).Table(multi.ActivityTableName(s.chain)).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&newActivity).Error; err != nil {
		// 如果存储失败，记录警告日志
		xzap.WithContext(s.ctx).Warn("failed on create activity",
			zap.Error(err))
	}

	// 更新NFT的所有者
	if err := s.db.WithContext(s.ctx).Table(multi.ItemTableName(s.chain)).
		Where("collection_address = ? and token_id = ?", strings.ToLower(collection), tokenId).
		Update("owner", owner).Error; err != nil {
		// 如果更新失败，记录错误日志并返回
		xzap.WithContext(s.ctx).Error("failed to update item owner",
			zap.Error(err))
		return
	}

	// 将交易信息存入价格更新队列
	if err := ordermanager.AddUpdatePriceEvent(s.kv, &ordermanager.TradeEvent{
		OrderId:        sellOrderId,
		CollectionAddr: collection,
		EventType:      ordermanager.Buy,
		TokenID:        tokenId,
		From:           from,
		To:             to,
	}, s.chain); err != nil {
		// 如果添加失败，记录错误日志
		xzap.WithContext(s.ctx).Error("failed on add update price event",
			zap.Error(err),
			zap.String("type", "sale"),
			zap.String("order_id", sellOrderId))
	}
}

// handleCancelEvent 处理LogCancel事件，当订单被取消时触发。
// 该函数会更新订单状态为已取消，将取消信息存入活动表，并将交易信息存入价格更新队列。
func (s *Service) handleCancelEvent(log ethereumTypes.Log) {
	// 从日志的topics中提取订单ID
	orderId := HexPrefix + hex.EncodeToString(log.Topics[1].Bytes())
	// 注释掉的代码，原计划从日志的topics中提取订单创建者地址
	//maker := common.BytesToAddress(log.Topics[2].Bytes())
	// 更新数据库中订单的状态为已取消
	if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
		Where("order_id = ?", orderId).
		Update("order_status", multi.OrderStatusCancelled).Error; err != nil {
		// 如果更新失败，记录错误日志并返回
		xzap.WithContext(s.ctx).Error("failed on update order status",
			zap.String("order_id", orderId))
		return
	}

	// 定义一个变量来存储取消的订单信息
	var cancelOrder multi.Order
	// 从数据库中查询取消的订单信息
	if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
		Where("order_id = ?", orderId).
		First(&cancelOrder).Error; err != nil {
		// 如果查询失败，记录错误日志并返回
		xzap.WithContext(s.ctx).Error("failed on get cancel order",
			zap.Error(err))
		return
	}

	// 获取该日志所在区块的时间
	blockTime, err := s.chainClient.BlockTimeByNumber(s.ctx, big.NewInt(int64(log.BlockNumber)))
	if err != nil {
		// 如果获取失败，记录错误日志并返回
		xzap.WithContext(s.ctx).Error("failed to get block time", zap.Error(err))
		return
	}
	// 根据订单类型确定活动类型
	var activityType int
	if cancelOrder.OrderType == multi.ListingOrder {
		// 挂单取消
		activityType = multi.CancelListing
	} else if cancelOrder.OrderType == multi.CollectionBidOrder {
		// 集合买单取消
		activityType = multi.CancelCollectionBid
	} else {
		// 单个NFT买单取消
		activityType = multi.CancelItemBid
	}
	// 创建一个新的活动结构体，用于记录订单取消信息
	newActivity := multi.Activity{
		ActivityType:      activityType,
		Maker:             cancelOrder.Maker,
		Taker:             ZeroAddress,
		MarketplaceID:     multi.MarketOrderBook,
		CollectionAddress: cancelOrder.CollectionAddress,
		TokenId:           cancelOrder.TokenId,
		CurrencyAddress:   s.cfg.ContractCfg.EthAddress,
		Price:             cancelOrder.Price,
		BlockNumber:       int64(log.BlockNumber),
		TxHash:            log.TxHash.String(),
		EventTime:         int64(blockTime),
	}
	// 将活动信息存入数据库，如果活动已存在则不做处理
	if err := s.db.WithContext(s.ctx).Table(multi.ActivityTableName(s.chain)).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&newActivity).Error; err != nil {
		// 如果存储失败，记录警告日志
		xzap.WithContext(s.ctx).Warn("failed on create activity",
			zap.Error(err))
	}

	// 将交易信息存入价格更新队列，通知价格更新
	if err := ordermanager.AddUpdatePriceEvent(s.kv, &ordermanager.TradeEvent{
		OrderId:        cancelOrder.OrderID,
		CollectionAddr: cancelOrder.CollectionAddress,
		TokenID:        cancelOrder.TokenId,
		EventType:      ordermanager.Cancel,
	}, s.chain); err != nil {
		// 如果添加失败，记录错误日志
		xzap.WithContext(s.ctx).Error("failed on add update price event",
			zap.Error(err),
			zap.String("type", "cancel"),
			zap.String("order_id", cancelOrder.OrderID))
	}
}

// UpKeepingCollectionFloorChangeLoop 是一个服务方法，用于维护集合底价的变化。
// 它会定期清理过期的集合底价变化数据，并在满足条件时更新集合的底价信息。
func (s *Service) UpKeepingCollectionFloorChangeLoop() {
	// 创建一个定时器，每天触发一次，用于清理过期的集合底价变化数据
	timer := time.NewTicker(comm.DaySeconds * time.Second)
	// 确保在函数结束时停止定时器，避免资源泄漏
	defer timer.Stop()
	// 创建另一个定时器，每隔 comm.MaxCollectionFloorTimeDifference 10秒触发一次，用于更新集合的底价信息
	updateFloorPriceTimer := time.NewTicker(comm.MaxCollectionFloorTimeDifference * time.Second)
	// 确保在函数结束时停止定时器，避免资源泄漏
	defer updateFloorPriceTimer.Stop()

	// 定义一个变量来存储索引状态
	var indexedStatus base.IndexedStatus
	// 从数据库中查询最后同步的时间
	if err := s.db.WithContext(s.ctx).Table(base.IndexedStatusTableName()).
		// 仅选择 last_indexed_time 字段
		Select("last_indexed_time").
		// 根据链 ID 和索引类型过滤数据
		Where("chain_id = ? and index_type = ?", s.chainId, comm.CollectionFloorChangeIndexType).
		// 将查询结果存储到 indexedStatus 变量中
		First(&indexedStatus).Error; err != nil {
		// 如果查询失败，记录错误日志并返回
		xzap.WithContext(s.ctx).Error("failed on get collection floor change index status",
			zap.Error(err))
		return
	}

	// 进入无限循环，持续维护集合底价的变化
	for {
		// 使用 select 语句监听多个通道事件
		select {
		// 监听上下文取消事件
		case <-s.ctx.Done():
			// 如果上下文被取消，记录日志并退出循环
			xzap.WithContext(s.ctx).Info("UpKeepingCollectionFloorChangeLoop stopped due to context cancellation")
			return
		// 监听定时器事件，每天触发一次
		case <-timer.C:
			// 调用 deleteExpireCollectionFloorChangeFromDatabase 方法清理过期的集合底价变化数据
			if err := s.deleteExpireCollectionFloorChangeFromDatabase(); err != nil {
				// 如果清理失败，记录错误日志
				xzap.WithContext(s.ctx).Error("failed on delete expire collection floor change",
					zap.Error(err))
			}
		// 监听更新底价定时器事件，每隔 comm.MaxCollectionFloorTimeDifference 10秒触发一次
		case <-updateFloorPriceTimer.C:
			// 检查项目配置名称是否为指定的项目名称
			if s.cfg.ProjectCfg.Name == gdb.OrderBookDexProject {
				// 调用 QueryCollectionsFloorPrice 方法查询集合的底价信息
				floorPrices, err := s.QueryCollectionsFloorPrice()
				if err != nil {
					// 如果查询失败，记录错误日志并继续下一次循环
					xzap.WithContext(s.ctx).Error("failed on query collections floor change",
						zap.Error(err))
					continue
				}

				// 调用 persistCollectionsFloorChange 方法将查询到的底价信息持久化到数据库中
				if err := s.persistCollectionsFloorChange(floorPrices); err != nil {
					// 如果持久化失败，记录错误日志并继续下一次循环
					xzap.WithContext(s.ctx).Error("failed on persist collections floor price",
						zap.Error(err))
					continue
				}
			}
		// 默认分支，当前没有匹配的通道事件时执行
		default:
		}
	}
}

// deleteExpireCollectionFloorChangeFromDatabase 方法用于从数据库中删除过期的集合底价变化数据。
// 过期的判断依据是 event_time 字段的值是否小于当前时间减去指定的时间范围。
// 返回值：如果删除过程中出现错误，返回包装后的错误信息；否则返回 nil。
func (s *Service) deleteExpireCollectionFloorChangeFromDatabase() error {
	// 构建 SQL 删除语句，使用 fmt.Sprintf 函数动态生成表名和时间范围
	// gdb.GetMultiProjectCollectionFloorPriceTableName 函数用于获取集合底价变化数据的表名
	// comm.CollectionFloorTimeRange 表示集合底价变化数据的有效时间范围
	stmt := fmt.Sprintf(`DELETE FROM %s where event_time < UNIX_TIMESTAMP() - %d`, gdb.GetMultiProjectCollectionFloorPriceTableName(s.cfg.ProjectCfg.Name, s.chain), comm.CollectionFloorTimeRange)

	// 执行 SQL 删除语句
	// s.db.Exec 方法用于执行 SQL 语句，返回一个 *gorm.DB 对象
	// 如果执行过程中出现错误，将错误信息存储在 err 变量中
	if err := s.db.Exec(stmt).Error; err != nil {
		// 如果出现错误，使用 errors.Wrap 函数包装错误信息，便于调试和定位问题
		return errors.Wrap(err, "failed on delete expire collection floor price")
	}

	// 如果删除过程中没有出现错误，返回 nil
	return nil
}

// QueryCollectionsFloorPrice 查询集合的底价信息。
// 该函数通过 SQL 查询获取每个集合的最低价格，并将结果存储在 multi.CollectionFloorPrice 结构体切片中。
// 返回值：包含集合底价信息的切片和可能出现的错误。
func (s *Service) QueryCollectionsFloorPrice() ([]multi.CollectionFloorPrice, error) {
	// 获取当前时间戳（秒）
	timestamp := time.Now().Unix()
	// 获取当前时间戳（毫秒）
	timestampMilli := time.Now().UnixMilli()
	// 定义一个切片，用于存储查询到的集合底价信息
	var collectionFloorPrice []multi.CollectionFloorPrice
	// 构建 SQL 查询语句
	// 从项目的物品表和订单表中查询每个集合的最低价格
	// 使用左连接确保即使没有匹配的订单，也会包含集合信息
	// 过滤条件包括订单类型、订单状态、过期时间和订单创建者与物品所有者的匹配
	sql := fmt.Sprintf(`SELECT co.collection_address as collection_address,min(co.price) as price
FROM %s as ci
         left join %s co on co.collection_address = ci.collection_address and co.token_id = ci.token_id
WHERE (co.order_type = ? and
       co.order_status = ? and expire_time > ? and co.maker = ci.owner) group by co.collection_address`, gdb.GetMultiProjectItemTableName(s.cfg.ProjectCfg.Name, s.chain), gdb.GetMultiProjectOrderTableName(s.cfg.ProjectCfg.Name, s.chain))
	// 执行 SQL 查询
	// 使用 Raw 方法执行自定义 SQL 查询
	// 将查询结果扫描到 collectionFloorPrice 切片中
	if err := s.db.WithContext(s.ctx).Raw(
		sql,
		multi.ListingType,
		multi.OrderStatusActive,
		time.Now().Unix(),
	).Scan(&collectionFloorPrice).Error; err != nil {
		// 如果查询过程中出现错误，使用 errors.Wrap 包装错误信息并返回
		return nil, errors.Wrap(err, "failed on get collection floor price")
	}

	// 遍历查询结果，为每个集合的底价信息添加时间戳
	for i := 0; i < len(collectionFloorPrice); i++ {
		// 设置事件时间为当前时间戳（秒）
		collectionFloorPrice[i].EventTime = timestamp
		// 设置创建时间为当前时间戳（毫秒）
		collectionFloorPrice[i].CreateTime = timestampMilli
		// 设置更新时间为当前时间戳（毫秒）
		collectionFloorPrice[i].UpdateTime = timestampMilli
	}

	// 返回包含集合底价信息的切片和 nil 错误
	return collectionFloorPrice, nil
}

// persistCollectionsFloorChange 将集合底价变化信息持久化到数据库中。
// 该函数会将传入的 FloorPrices 切片按批次插入到数据库中，以避免一次性插入大量数据导致性能问题。
// 如果在插入过程中出现错误，函数会返回包装后的错误信息。
// 参数:
// - FloorPrices: 包含集合底价变化信息的切片。
// 返回值:
// - error: 如果插入过程中出现错误，返回包装后的错误信息；否则返回 nil。
func (s *Service) persistCollectionsFloorChange(FloorPrices []multi.CollectionFloorPrice) error {
	// 遍历 FloorPrices 切片，按批次处理数据
	for i := 0; i < len(FloorPrices); i += comm.DBBatchSizeLimit {
		// 计算当前批次的结束索引
		// 使用 math.Min 函数确保结束索引不超过切片的长度
		end := int(math.Min(float64(i+comm.DBBatchSizeLimit), float64(len(FloorPrices))))

		// 初始化存储 SQL 值占位符的切片
		valueStrings := make([]string, 0)
		// 初始化存储 SQL 值的切片，使用 any 类型替代 interface{}
		valueArgs := make([]any, 0)

		// 遍历当前批次的 FloorPrices 数据
		for _, t := range FloorPrices[i:end] {
			// 为每个数据项添加 SQL 值占位符
			valueStrings = append(valueStrings, "(?,?,?,?,?)")
			// 将数据项的具体值添加到 valueArgs 切片中
			valueArgs = append(valueArgs, t.CollectionAddress, t.Price, t.EventTime, t.CreateTime, t.UpdateTime)
		}

		// 构建 SQL 插入语句
		// 使用 fmt.Sprintf 函数动态生成表名和值占位符
		stmt := fmt.Sprintf(`INSERT INTO %s (collection_address,price,event_time,create_time,update_time)  VALUES %s
        ON DUPLICATE KEY UPDATE update_time=VALUES(update_time)`, gdb.GetMultiProjectCollectionFloorPriceTableName(s.cfg.ProjectCfg.Name, s.chain), strings.Join(valueStrings, ","))

		// 执行 SQL 插入语句
		// 如果执行过程中出现错误，将错误信息存储在 err 变量中
		if err := s.db.Exec(stmt, valueArgs...).Error; err != nil {
			// 如果出现错误，使用 errors.Wrap 函数包装错误信息，便于调试和定位问题
			return errors.Wrap(err, "failed on persist collection floor price info")
		}
	}
	// 如果插入过程中没有出现错误，返回 nil
	return nil
}
