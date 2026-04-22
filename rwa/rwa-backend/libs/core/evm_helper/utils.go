package evm_helper

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/go-resty/resty/v2"
)

// GetTransactionReceipts retrieves multiple transaction receipts by hashes using resty client
func GetTransactionReceipts(ctx context.Context, rpcURL string, txHashes []string) ([]*TransactionReceipt, error) {
	if len(txHashes) == 0 {
		return []*TransactionReceipt{}, nil
	}
	client := resty.New()
	// Build batch JSON-RPC requests
	requests := make([]JSONRPCRequest, len(txHashes))
	for i, txHash := range txHashes {
		requests[i] = JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "eth_getTransactionReceipt",
			Params:  []interface{}{txHash},
			ID:      i + 1,
		}
	}

	var responses []JSONRPCResponse

	resp, err := client.R().
		SetContext(ctx).
		SetBody(requests).
		SetResult(&responses).
		Post(rpcURL)

	if err != nil {
		return nil, fmt.Errorf("failed to send batch request: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode())
	}

	// Process responses
	receipts := make([]*TransactionReceipt, len(txHashes))
	for _, response := range responses {
		if response.Error != nil {
			return nil, fmt.Errorf("JSON-RPC error for request %d: %d - %s", response.ID, response.Error.Code, response.Error.Message)
		}

		// response.ID is 1-based, convert to 0-based index
		index := response.ID - 1
		if index >= 0 && index < len(receipts) {
			receipts[index] = response.Result
		}
	}
	return receipts, nil
}

// GetSingleTransactionReceipt retrieves a single transaction receipt by hash using resty client
func GetSingleTransactionReceipt(ctx context.Context, rpcURL string, txHash string) (*TransactionReceipt, error) {
	client := resty.New()

	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_getTransactionReceipt",
		Params:  []interface{}{txHash},
		ID:      1,
	}

	var response JSONRPCResponse

	resp, err := client.R().
		SetContext(ctx).
		SetBody(request).
		SetResult(&response).
		Post(rpcURL)

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode())
	}

	if response.Error != nil {
		return nil, fmt.Errorf("JSON-RPC error: %d - %s", response.Error.Code, response.Error.Message)
	}

	if response.Result == nil {
		return nil, fmt.Errorf("transaction receipt not found")
	}

	return response.Result, nil
}

// hexToBigInt converts a hex string to big.Int
func hexToBigInt(hexStr string) (*big.Int, error) {
	if hexStr == "" {
		return big.NewInt(0), nil
	}

	// Remove 0x prefix if present
	if strings.HasPrefix(hexStr, "0x") {
		hexStr = hexStr[2:]
	}

	if hexStr == "" {
		return big.NewInt(0), nil
	}

	result := new(big.Int)
	result, ok := result.SetString(hexStr, 16)
	if !ok {
		return nil, fmt.Errorf("invalid hex string: %s", hexStr)
	}

	return result, nil
}

// ParseTransactionReceipt converts a TransactionReceipt to ParsedTransactionReceipt with big.Int values
func ParseTransactionReceipt(receipt *TransactionReceipt) (*ParsedTransactionReceipt, error) {
	if receipt == nil {
		return nil, fmt.Errorf("receipt is nil")
	}

	parsed := &ParsedTransactionReceipt{
		BlockHash:       receipt.BlockHash,
		ContractAddress: receipt.ContractAddress,
		From:            receipt.From,
		LogsBloom:       receipt.LogsBloom,
		To:              receipt.To,
		TransactionHash: receipt.TransactionHash,
	}

	// Parse hex strings to big.Int
	var err error

	if parsed.BlockNumber, err = hexToBigInt(receipt.BlockNumber); err != nil {
		return nil, fmt.Errorf("failed to parse block number: %w", err)
	}

	if parsed.CumulativeGasUsed, err = hexToBigInt(receipt.CumulativeGasUsed); err != nil {
		return nil, fmt.Errorf("failed to parse cumulative gas used: %w", err)
	}

	if parsed.EffectiveGasPrice, err = hexToBigInt(receipt.EffectiveGasPrice); err != nil {
		return nil, fmt.Errorf("failed to parse effective gas price: %w", err)
	}

	if parsed.GasUsed, err = hexToBigInt(receipt.GasUsed); err != nil {
		return nil, fmt.Errorf("failed to parse gas used: %w", err)
	}

	if parsed.Status, err = hexToBigInt(receipt.Status); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	if parsed.TransactionIndex, err = hexToBigInt(receipt.TransactionIndex); err != nil {
		return nil, fmt.Errorf("failed to parse transaction index: %w", err)
	}

	if parsed.Type, err = hexToBigInt(receipt.Type); err != nil {
		return nil, fmt.Errorf("failed to parse type: %w", err)
	}

	// Parse logs
	parsed.Logs = make([]*ParsedLog, len(receipt.Logs))
	for i, log := range receipt.Logs {
		parsedLog, err := parseLog(&log)
		if err != nil {
			return nil, fmt.Errorf("failed to parse log %d: %w", i, err)
		}
		parsed.Logs[i] = parsedLog
	}

	return parsed, nil
}

// ParseTransactionReceipts converts multiple TransactionReceipts to ParsedTransactionReceipts with big.Int values
func ParseTransactionReceipts(receipts []*TransactionReceipt) ([]*ParsedTransactionReceipt, error) {
	if len(receipts) == 0 {
		return []*ParsedTransactionReceipt{}, nil
	}

	parsed := make([]*ParsedTransactionReceipt, len(receipts))
	for i, receipt := range receipts {
		if receipt == nil {
			parsed[i] = nil
			continue
		}

		parsedReceipt, err := ParseTransactionReceipt(receipt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse receipt %d: %w", i, err)
		}
		parsed[i] = parsedReceipt
	}

	return parsed, nil
}

// parseLog converts a Log to ParsedLog with big.Int values
func parseLog(log *Log) (*ParsedLog, error) {
	parsed := &ParsedLog{
		Address:         log.Address,
		Topics:          log.Topics,
		Data:            log.Data,
		TransactionHash: log.TransactionHash,
		BlockHash:       log.BlockHash,
		Removed:         log.Removed,
	}

	var err error

	if parsed.BlockNumber, err = hexToBigInt(log.BlockNumber); err != nil {
		return nil, fmt.Errorf("failed to parse block number: %w", err)
	}

	if parsed.TransactionIndex, err = hexToBigInt(log.TransactionIndex); err != nil {
		return nil, fmt.Errorf("failed to parse transaction index: %w", err)
	}

	if parsed.LogIndex, err = hexToBigInt(log.LogIndex); err != nil {
		return nil, fmt.Errorf("failed to parse log index: %w", err)
	}

	return parsed, nil
}

// IsTransactionSuccessful checks if a transaction was successful based on status
func IsTransactionSuccessful(receipt *TransactionReceipt) bool {
	return receipt != nil && receipt.Status == "0x1"
}
