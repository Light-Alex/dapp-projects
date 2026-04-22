package evm_helper

import (
	"strconv"

	"github.com/ethereum/go-ethereum/ethclient"
)

type EvmClient struct {
	rpcInfo       map[uint64]*EvmChain
	httpClientMap map[uint64]*ethclient.Client
	wssClientMap  map[uint64]*ethclient.Client
}

func NewEvmClient(rpcInfo RpcInfoMap) (*EvmClient, error) {
	httpClientMap := make(map[uint64]*ethclient.Client)
	wssClientMap := make(map[uint64]*ethclient.Client)
	for chainID, chain := range rpcInfo {
		httpClient, err := ethclient.Dial(chain.RpcUrl)
		if err != nil {
			return nil, err
		}
		httpClientMap[chainID] = httpClient

		wssClient, err := ethclient.Dial(chain.WssUrl)
		if err != nil {
			return nil, err
		}
		wssClientMap[chainID] = wssClient
	}
	return &EvmClient{
		httpClientMap: httpClientMap,
		wssClientMap:  wssClientMap,
		rpcInfo:       rpcInfo,
	}, nil
}

func (e *EvmClient) MustGetRpcInfo(chainID uint64) *EvmChain {
	chain, ok := e.rpcInfo[chainID]
	if !ok {
		panic("no rpc info for chainID: " + strconv.FormatUint(chainID, 10))
	}
	return chain
}

func (e *EvmClient) MustGetHttpClient(chainID uint64) *ethclient.Client {
	client, ok := e.httpClientMap[chainID]
	if !ok {
		panic("no http client for chainID: " + strconv.FormatUint(chainID, 10))
	}
	return client
}

func (e *EvmClient) MustGetWssClient(chainID uint64) *ethclient.Client {
	client, ok := e.wssClientMap[chainID]
	if !ok {
		panic("no wss client for chainID: " + strconv.FormatUint(chainID, 10))
	}
	return client
}
