package erc721

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

const (
	ERC721 = "erc721"
)

type NftErc721 struct {
	client   *ethclient.Client
	endpoint string
}

// NewNftErc721 根据给定的以太坊节点地址创建一个新的NftErc721实例
//
// 参数:
//     endpoint: string - 以太坊节点地址
//
// 返回值:
//     *NftErc721: 指向NftErc721结构体的指针
//     error: 错误信息，如果连接失败则返回错误信息，否则为nil
func NewNftErc721(endpoint string) (*NftErc721, error) {
	// 使用给定的endpoint连接以太坊客户端
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		// 如果连接失败，返回错误
		return nil, err
	}

	// 返回一个NftErc721结构体指针和一个nil错误
	return &NftErc721{
		client:   client,  // 以太坊客户端
		endpoint: endpoint,// 以太坊节点地址
	}, nil
}

// GetItemOwner 从给定的地址和tokenID中获取NFT的拥有者地址
//
// 参数:
//     address: NFT合约的地址字符串
//     tokenId: token的唯一标识符字符串
//
// 返回值:
//     string: 拥有者地址的十六进制字符串表示
//     error: 如果在获取拥有者地址过程中发生错误，则返回错误信息
func (n *NftErc721) GetItemOwner(address string, tokenId string) (string, error) {
	// 将传入的地址字符串转换为以太坊地址类型
	addr := common.HexToAddress(address)

	// 创建一个新的 Erc721Caller 实例
	instance, err := NewErc721Caller(addr, n.client)
	if err != nil {
		// 如果创建实例失败，则返回错误
		return "", err
	}

	// 将传入的 tokenId 字符串转换为大整数类型
	token := new(big.Int)
	token.SetString(tokenId, 10)

	// 调用 instance 的 OwnerOf 方法获取 token 的拥有者地址
	ownerOf, err := instance.OwnerOf(&bind.CallOpts{}, token)
	if err != nil {
		// 如果获取拥有者地址失败，则返回错误
		return "", err
	}

	// 返回拥有者地址的十六进制字符串表示
	return ownerOf.Hex(), nil
}
