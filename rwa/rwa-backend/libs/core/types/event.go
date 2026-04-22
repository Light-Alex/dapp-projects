package types

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type ContractType string

const (
	ContractTypeGate       ContractType = "gate"
	ContractTypeInstrument ContractType = "instrument"
	ContractTypeOrder      ContractType = "order"
)

type EventLogWithId struct {
	Address        common.Address `json:"address"`
	BlockHash      string         `json:"blockHash"`
	BlockNumber    string         `json:"blockNumber"`
	BlockTimestamp string         `json:"blockTimestamp"`
	Data           string         `json:"data"`
	Index          string         `json:"logIndex"`
	Topics         []string       `json:"topics"`
	TxHash         string         `json:"transactionHash"`
	TxIndex        string         `json:"transactionIndex"`
	Removed        bool           `json:"removed"`
	EventId        uint64         `json:"eventId"`
	ContractType   ContractType   `json:"contractType"`
}

func (i *EventLogWithId) ConvertToEthLog() (*types.Log, error) {
	topics := make([]common.Hash, len(i.Topics))
	for idx, topic := range i.Topics {
		topics[idx] = common.HexToHash(topic)
	}

	blockNumber, err := hexutil.DecodeUint64(i.BlockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to decode BlockNumber %q: %w", i.BlockNumber, err)
	}
	txIndex, err := hexutil.DecodeUint64(i.TxIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode TxIndex %q: %w", i.TxIndex, err)
	}
	logIndex, err := hexutil.DecodeUint64(i.Index)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Index %q: %w", i.Index, err)
	}
	blockTimestamp, err := hexutil.DecodeUint64(i.BlockTimestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode BlockTimestamp %q: %w", i.BlockTimestamp, err)
	}

	return &types.Log{
		Address:        i.Address,
		Topics:         topics,
		Data:           common.FromHex(i.Data),
		BlockNumber:    blockNumber,
		TxHash:         common.HexToHash(i.TxHash),
		TxIndex:        uint(txIndex),
		BlockHash:      common.HexToHash(i.BlockHash),
		Index:          uint(logIndex),
		Removed:        i.Removed,
		BlockTimestamp: blockTimestamp,
	}, nil
}

type BlockRpcResult struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Number string `json:"number"`
	} `json:"result"`
	Id    int `json:"id"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type EventLogRpcResult struct {
	Jsonrpc string            `json:"jsonrpc"`
	Result  []*EventLogWithId `json:"result"`
	Id      int               `json:"id"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}
