package evm_helper

import (
	"math/big"
)

type RpcInfoMap map[uint64]*EvmChain

type EvmChain struct {
	RpcUrl              string `json:"rpcUrl" yaml:"rpcUrl"`
	WssUrl              string `json:"wssUrl" yaml:"wssUrl"`
	NativeTokenSymbol   string `json:"nativeTokenSymbol" yaml:"nativeTokenSymbol"`
	NativeTokenDecimals uint8  `json:"nativeTokenDecimals" yaml:"nativeTokenDecimals"`
}

// Log represents an Ethereum log entry
type Log struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	BlockHash        string   `json:"blockHash"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
}

// TransactionReceipt represents an Ethereum transaction receipt
type TransactionReceipt struct {
	BlockHash         string `json:"blockHash"`
	BlockNumber       string `json:"blockNumber"`
	ContractAddress   string `json:"contractAddress"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	EffectiveGasPrice string `json:"effectiveGasPrice"`
	From              string `json:"from"`
	GasUsed           string `json:"gasUsed"`
	Logs              []Log  `json:"logs"`
	LogsBloom         string `json:"logsBloom"`
	Status            string `json:"status"`
	To                string `json:"to"`
	TransactionHash   string `json:"transactionHash"`
	TransactionIndex  string `json:"transactionIndex"`
	Type              string `json:"type"`
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string              `json:"jsonrpc"`
	ID      int                 `json:"id"`
	Result  *TransactionReceipt `json:"result,omitempty"`
	Error   *JSONRPCError       `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ParsedLog represents a parsed log with converted values
type ParsedLog struct {
	Address          string
	Topics           []string
	Data             string
	BlockNumber      *big.Int
	TransactionHash  string
	TransactionIndex *big.Int
	BlockHash        string
	LogIndex         *big.Int
	Removed          bool
}

// ParsedTransactionReceipt represents a parsed transaction receipt with converted values
type ParsedTransactionReceipt struct {
	BlockHash         string
	BlockNumber       *big.Int
	ContractAddress   string
	CumulativeGasUsed *big.Int
	EffectiveGasPrice *big.Int
	From              string
	GasUsed           *big.Int
	Logs              []*ParsedLog
	LogsBloom         string
	Status            *big.Int
	To                string
	TransactionHash   string
	TransactionIndex  *big.Int
	Type              *big.Int
}
