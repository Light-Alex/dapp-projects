// 定义包名为 nftchainservice，该包主要提供与 NFT 链服务相关的功能
package nftchainservice

import (
	"context"
	"time"

	// 导入以太坊账户 ABI 相关的包，用于处理合约的 ABI 信息
	"github.com/ethereum/go-ethereum/accounts/abi"
	// 导入以太坊通用地址类型的包，用于处理以太坊地址
	"github.com/ethereum/go-ethereum/common"
	// 导入错误处理包，用于更方便地包装和处理错误
	"github.com/pkg/errors"

	// 导入自定义的链客户端包，用于与区块链节点进行交互
	"github.com/ProjectsTask/EasySwapBase/chain/chainclient"
	// 导入自定义的 HTTP 客户端包，用于进行 HTTP 请求
	"github.com/ProjectsTask/EasySwapBase/xhttp"
)

// 定义默认的超时时间，单位为秒
const defaultTimeout = 10 //uint s

// NodeService 是一个接口，定义了与 NFT 链服务相关的操作方法
type NodeService interface {
	// FetchOnChainMetadata 方法用于从链上获取指定链 ID、NFT 集合地址和代币 ID 的元数据
	FetchOnChainMetadata(chainID int64, collectionAddr string, tokenID string) (*JsonMetadata, error)
	// FetchNftOwner 方法用于获取指定链 ID、NFT 集合地址和代币 ID 的 NFT 所有者地址
	FetchNftOwner(chainID int64, collectionAddr string, tokenID string) (common.Address, error)
	// GetNFTTransferEvent 方法用于获取指定链名称、起始块和结束块范围内的 NFT 转移事件
	GetNFTTransferEvent(chainName string, fromBlock, toBlock uint64) ([]*TransferLog, error)
	// GetNFTTransferEventGoroutine 方法用于使用 goroutine 并发获取指定链名称、起始块、结束块、块大小和通道大小的 NFT 转移事件
	GetNFTTransferEventGoroutine(chainName string, fromBlock, toBlock, blockSize, channelSize uint64) ([]*TransferLog, error)
}

// Service 结构体表示一个服务实例，包含了与 NFT 链服务相关的各种配置和客户端
type Service struct {
	// 上下文，用于控制服务的生命周期和传递请求范围的数据
	ctx context.Context

	// 合约的 ABI 信息，用于与智能合约进行交互
	Abi *abi.ABI
	// HTTP 客户端，用于与外部服务进行 HTTP 请求
	HttpClient *xhttp.Client
	// 链客户端，用于与区块链节点进行交互
	NodeClient chainclient.ChainClient
	// 链的名称，例如 "Ethereum"、"BSC" 等
	ChainName string
	// 节点的名称
	NodeName string
	// 名称标签列表，用于在元数据中查找名称相关的信息
	NameTags []string
	// 图像标签列表，用于在元数据中查找图像相关的信息
	ImageTags []string
	// 属性标签列表，用于在元数据中查找属性相关的信息
	AttributesTags []string
	// 特征名称标签列表，用于在元数据中查找特征名称相关的信息
	TraitNameTags []string
	// 特征值标签列表，用于在元数据中查找特征值相关的信息
	TraitValueTags []string
}

// New 函数用于创建一个新的 Service 实例
// 参数说明：
// - ctx: 上下文，用于控制服务的生命周期
// - endpoint: 区块链节点的端点地址，用于与区块链节点建立连接
// - chainName: 链的名称
// - chainID: 链的 ID
// - nameTags: 名称标签列表
// - imageTags: 图像标签列表
// - attributesTags: 属性标签列表
// - traitNameTags: 特征名称标签列表
// - traitValueTags: 特征值标签列表
// 返回值说明：
// - *Service: 新创建的 Service 实例
// - error: 如果创建过程中出现错误，返回相应的错误信息
func New(ctx context.Context, endpoint, chainName string, chainID int, nameTags, imageTags, attributesTags,
	traitNameTags, traitValueTags []string) (*Service, error) {
	// 获取默认的 HTTP 客户端配置
	conf := xhttp.GetDefaultConfig()
	// 禁用强制尝试使用 HTTP2 协议
	conf.ForceAttemptHTTP2 = false
	// 设置 HTTP 请求的超时时间
	conf.HTTPTimeout = time.Duration(defaultTimeout) * time.Second
	// 设置拨号超时时间
	conf.DialTimeout = time.Duration(defaultTimeout-5) * time.Second
	// 设置拨号保持活动的时间
	conf.DialKeepAlive = time.Duration(defaultTimeout+10) * time.Second

	// 创建链客户端实例
	nodeClient, err := chainclient.New(chainID, endpoint)
	if err != nil {
		// 如果创建链客户端失败，返回错误信息并包装错误原因
		return nil, errors.Wrap(err, "failed on create node client")
	}

	// 获取 NFT 合约的 ABI 信息
	abi, err := NftContractMetaData.GetAbi()
	if err != nil {
		// 如果获取合约 ABI 信息失败，返回错误信息并包装错误原因
		return nil, errors.Wrap(err, "failed on get contract abi")
	}

	// 返回新创建的 Service 实例
	return &Service{
		ctx:            ctx,
		Abi:            abi,
		HttpClient:     xhttp.NewClient(conf),
		NodeClient:     nodeClient,
		ChainName:      chainName,
		NameTags:       nameTags,
		ImageTags:      imageTags,
		AttributesTags: attributesTags,
		TraitNameTags:  traitNameTags,
		TraitValueTags: traitValueTags,
	}, nil
}
