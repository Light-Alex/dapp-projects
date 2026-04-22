package evm_helper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockRPCServer creates a test HTTP server that responds to JSON-RPC requests
func mockRPCServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestBody interface{}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			t.Errorf("Failed to decode request body: %v", err)
			return
		}

		// Mock response for single request
		if req, ok := requestBody.(map[string]interface{}); ok {
			response := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      int(req["id"].(float64)),
				Result: &TransactionReceipt{
					BlockHash:         "0x443092b9d3ad62c707f098ca9e9f6f86ac25084b969039127cdadfc60780bf20",
					BlockNumber:       "0x167d458",
					ContractAddress:   "",
					CumulativeGasUsed: "0x1618ebd",
					EffectiveGasPrice: "0xe2c8823",
					From:              "0x4838b106fce9647bdf1e7877bf73ce8b0bad5f97",
					GasUsed:           "0x565f",
					LogsBloom:         "0x00000000000000000000000000000000000100004000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000",
					Status:            "0x1",
					To:                "0x388c818ca8b9251b393131c08a736a67ccb19297",
					TransactionHash:   "0x850a4188dc787bb795c082a4236a97e78e9c1c321bb16856e75b3cf9dcde8e85",
					TransactionIndex:  "0xee",
					Type:              "0x2",
					Logs: []Log{
						{
							Address:          "0x388c818ca8b9251b393131c08a736a67ccb19297",
							Topics:           []string{"0x27f12abfe35860a9a927b465bb3d4a9c23c8428174b83f278fe45ed7b4da2662"},
							Data:             "0x0000000000000000000000000000000000000000000000000028165a20dd318a",
							BlockNumber:      "0x167d458",
							TransactionHash:  "0x850a4188dc787bb795c082a4236a97e78e9c1c321bb16856e75b3cf9dcde8e85",
							TransactionIndex: "0xee",
							BlockHash:        "0x443092b9d3ad62c707f098ca9e9f6f86ac25084b969039127cdadfc60780bf20",
							LogIndex:         "0x264",
							Removed:          false,
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Mock response for batch requests
		if reqs, ok := requestBody.([]interface{}); ok {
			responses := make([]JSONRPCResponse, len(reqs))
			for i, reqInterface := range reqs {
				req := reqInterface.(map[string]interface{})
				responses[i] = JSONRPCResponse{
					JSONRPC: "2.0",
					ID:      int(req["id"].(float64)),
					Result: &TransactionReceipt{
						BlockHash:         "0x443092b9d3ad62c707f098ca9e9f6f86ac25084b969039127cdadfc60780bf20",
						BlockNumber:       "0x167d458",
						ContractAddress:   "",
						CumulativeGasUsed: "0x1618eb",
						EffectiveGasPrice: "0xe2c8823",
						From:              "0x4838b106fce9647bdf1e7877bf73ce8b0bad5f97",
						GasUsed:           "0x565f",
						LogsBloom:         "0x00000000000000000000000000000000000100004000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000",
						Status:            "0x1",
						To:                "0x388c818ca8b9251b393131c08a736a67ccb19297",
						TransactionHash:   req["params"].([]interface{})[0].(string), // Use the actual tx hash from request
						TransactionIndex:  "0xee",
						Type:              "0x2",
						Logs: []Log{
							{
								Address:          "0x388c818ca8b9251b393131c08a736a67ccb19297",
								Topics:           []string{"0x27f12abfe35860a9a927b465bb3d4a9c23c8428174b83f278fe45ed7b4da2662"},
								Data:             "0x0000000000000000000000000000000000000000000000000028165a20dd318a",
								BlockNumber:      "0x167d458",
								TransactionHash:  req["params"].([]interface{})[0].(string),
								TransactionIndex: "0xee",
								BlockHash:        "0x443092b9d3ad62c707f098ca9e9f6f86ac25084b969039127cdadfc60780bf20",
								LogIndex:         "0x264",
								Removed:          false,
							},
						},
					},
				}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(responses)
			return
		}

		http.Error(w, "Invalid request", http.StatusBadRequest)
	}))
}

func TestGetSingleTransactionReceipt(t *testing.T) {
	server := mockRPCServer(t)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	txHash := "0x850a4188dc787bb795c082a4236a97e78e9c1c321bb16856e75b3cf9dcde8e85"

	receipt, err := GetSingleTransactionReceipt(ctx, server.URL, txHash)
	if err != nil {
		t.Fatalf("GetSingleTransactionReceipt failed: %v", err)
	}
	if receipt == nil {
		t.Fatal("Receipt is nil")
	}

	// Verify receipt fields
	if receipt.BlockHash != "0x443092b9d3ad62c707f098ca9e9f6f86ac25084b969039127cdadfc60780bf20" {
		t.Errorf("Expected BlockHash %s, got %s", "0x443092b9d3ad62c707f098ca9e9f6f86ac25084b969039127cdadfc60780bf20", receipt.BlockHash)
	}
	if receipt.BlockNumber != "0x167d458" {
		t.Errorf("Expected BlockNumber %s, got %s", "0x167d458", receipt.BlockNumber)
	}
	if receipt.Status != "0x1" {
		t.Errorf("Expected Status %s, got %s", "0x1", receipt.Status)
	}
	if receipt.TransactionHash != txHash {
		t.Errorf("Expected TransactionHash %s, got %s", txHash, receipt.TransactionHash)
	}
	if len(receipt.Logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(receipt.Logs))
	}

	t.Logf("✅ GetSingleTransactionReceipt test passed")
}

func TestGetTransactionReceipts(t *testing.T) {
	server := mockRPCServer(t)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	txHashes := []string{
		"0x850a4188dc787bb795c082a4236a97e78e9c1c321bb16856e75b3cf9dcde8e85",
		"0x123456789abcdef123456789abcdef123456789abcdef123456789abcdef1234",
		"0xabcdef123456789abcdef123456789abcdef123456789abcdef123456789abc",
	}

	receipts, err := GetTransactionReceipts(ctx, server.URL, txHashes)
	if err != nil {
		t.Fatalf("GetTransactionReceipts failed: %v", err)
	}
	if len(receipts) != 3 {
		t.Fatalf("Expected 3 receipts, got %d", len(receipts))
	}

	// Verify each receipt
	for i, receipt := range receipts {
		if receipt == nil {
			t.Fatalf("Receipt %d is nil", i)
		}
		if receipt.Status != "0x1" {
			t.Errorf("Receipt %d: Expected Status %s, got %s", i, "0x1", receipt.Status)
		}
		if receipt.TransactionHash != txHashes[i] {
			t.Errorf("Receipt %d: Expected TransactionHash %s, got %s", i, txHashes[i], receipt.TransactionHash)
		}
	}

	t.Logf("✅ GetTransactionReceipts test passed")
}

func TestGetTransactionReceiptsEmpty(t *testing.T) {
	server := mockRPCServer(t)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	receipts, err := GetTransactionReceipts(ctx, server.URL, []string{})
	if err != nil {
		t.Fatalf("GetTransactionReceipts with empty array failed: %v", err)
	}
	if len(receipts) != 0 {
		t.Errorf("Expected empty receipts, got %d", len(receipts))
	}

	t.Logf("✅ GetTransactionReceiptsEmpty test passed")
}

func TestHexToBigInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "valid hex with 0x prefix",
			input:    "0x565f",
			expected: "22111",
		},
		{
			name:     "valid hex without prefix",
			input:    "565f",
			expected: "22111",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "0",
		},
		{
			name:     "only 0x",
			input:    "0x",
			expected: "0",
		},
		{
			name:     "invalid hex",
			input:    "0xgg",
			hasError: true,
		},
		{
			name:     "large number",
			input:    "0x167d458",
			expected: "23549016",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := hexToBigInt(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input %s, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %s: %v", tt.input, err)
				} else if result.String() != tt.expected {
					t.Errorf("Expected %s, got %s for input %s", tt.expected, result.String(), tt.input)
				}
			}
		})
	}

	t.Logf("✅ HexToBigInt test passed")
}

func TestIsTransactionSuccessful(t *testing.T) {
	tests := []struct {
		name     string
		receipt  *TransactionReceipt
		expected bool
	}{
		{
			name: "successful transaction",
			receipt: &TransactionReceipt{
				Status: "0x1",
			},
			expected: true,
		},
		{
			name: "failed transaction",
			receipt: &TransactionReceipt{
				Status: "0x0",
			},
			expected: false,
		},
		{
			name:     "nil receipt",
			receipt:  nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTransactionSuccessful(tt.receipt)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}

	t.Logf("✅ IsTransactionSuccessful test passed")
}

func TestParseTransactionReceipts(t *testing.T) {
	// Test data
	receipts := []*TransactionReceipt{
		{
			BlockHash:         "0x443092b9d3ad62c707f098ca9e9f6f86ac25084b969039127cdadfc60780bf20",
			BlockNumber:       "0x167d458",
			ContractAddress:   "",
			CumulativeGasUsed: "0x1618eb",
			EffectiveGasPrice: "0xe2c8823",
			From:              "0x4838b106fce9647bdf1e7877bf73ce8b0bad5f97",
			GasUsed:           "0x565f",
			Status:            "0x1",
			To:                "0x388c818ca8b9251b393131c08a736a67ccb19297",
			TransactionHash:   "0x850a4188dc787bb795c082a4236a97e78e9c1c321bb16856e75b3cf9dcde8e85",
			TransactionIndex:  "0xee",
			Type:              "0x2",
			Logs: []Log{
				{
					Address:          "0x388c818ca8b9251b393131c08a736a67ccb19297",
					Topics:           []string{"0x27f12abfe35860a9a927b465bb3d4a9c23c8428174b83f278fe45ed7b4da2662"},
					Data:             "0x0000000000000000000000000000000000000000000000000028165a20dd318a",
					BlockNumber:      "0x167d458",
					TransactionHash:  "0x850a4188dc787bb795c082a4236a97e78e9c1c321bb16856e75b3cf9dcde8e85",
					TransactionIndex: "0xee",
					BlockHash:        "0x443092b9d3ad62c707f098ca9e9f6f86ac25084b969039127cdadfc60780bf20",
					LogIndex:         "0x264",
					Removed:          false,
				},
			},
		},
		{
			BlockHash:         "0x443092b9d3ad62c707f098ca9e9f6f86ac25084b969039127cdadfc60780bf21",
			BlockNumber:       "0x167d459",
			ContractAddress:   "",
			CumulativeGasUsed: "0x1618ebe",
			EffectiveGasPrice: "0xe2c8824",
			From:              "0x4838b106fce9647bdf1e7877bf73ce8b0bad5f98",
			GasUsed:           "0x5660",
			Status:            "0x1",
			To:                "0x388c818ca8b9251b393131c08a736a67ccb19298",
			TransactionHash:   "0x850a4188dc787bb795c082a4236a97e78e9c1c321bb16856e75b3cf9dcde8e86",
			TransactionIndex:  "0xef",
			Type:              "0x2",
			Logs:              []Log{},
		},
		nil, // Test nil receipt handling
	}

	parsed, err := ParseTransactionReceipts(receipts)
	if err != nil {
		t.Fatalf("ParseTransactionReceipts failed: %v", err)
	}

	if len(parsed) != 3 {
		t.Fatalf("Expected 3 parsed receipts, got %d", len(parsed))
	}

	// Test first receipt
	if parsed[0] == nil {
		t.Fatal("First parsed receipt is nil")
	}
	if parsed[0].BlockNumber.String() != "23549016" { // 0x167d458
		t.Errorf("Expected BlockNumber 23549016, got %s", parsed[0].BlockNumber.String())
	}
	if parsed[0].GasUsed.String() != "22111" { // 0x565f
		t.Errorf("Expected GasUsed 22111, got %s", parsed[0].GasUsed.String())
	}
	if parsed[0].Status.String() != "1" { // 0x1
		t.Errorf("Expected Status 1, got %s", parsed[0].Status.String())
	}
	if len(parsed[0].Logs) != 1 {
		t.Errorf("Expected 1 log in first receipt, got %d", len(parsed[0].Logs))
	}

	// Test second receipt
	if parsed[1] == nil {
		t.Fatal("Second parsed receipt is nil")
	}
	if parsed[1].BlockNumber.String() != "23549017" { // 0x167d459
		t.Errorf("Expected BlockNumber 23549017, got %s", parsed[1].BlockNumber.String())
	}
	if parsed[1].GasUsed.String() != "22112" { // 0x5660
		t.Errorf("Expected GasUsed 22112, got %s", parsed[1].GasUsed.String())
	}
	if len(parsed[1].Logs) != 0 {
		t.Errorf("Expected 0 logs in second receipt, got %d", len(parsed[1].Logs))
	}

	// Test nil receipt
	if parsed[2] != nil {
		t.Error("Expected third parsed receipt to be nil")
	}

	t.Logf("✅ ParseTransactionReceipts test passed")
}

func TestParseTransactionReceiptsEmpty(t *testing.T) {
	parsed, err := ParseTransactionReceipts([]*TransactionReceipt{})
	if err != nil {
		t.Fatalf("ParseTransactionReceipts with empty array failed: %v", err)
	}
	if len(parsed) != 0 {
		t.Errorf("Expected empty array, got %d receipts", len(parsed))
	}

	t.Logf("✅ ParseTransactionReceiptsEmpty test passed")
}
