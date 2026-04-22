package evm_helper

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// VerifyMessage verifies an Ethereum signed message signature
// This function is equivalent to ethers.js verifyMessage
func VerifyMessage(message string, signatureHex string, expectedAddr string) (string, error) {
	// Add EIP-191 prefix
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))
	prefixedMsg := prefix + message
	hash := crypto.Keccak256Hash([]byte(prefixedMsg))

	// Decode signature
	sig, err := hexutil.Decode(signatureHex)
	if err != nil {
		return "", err
	}
	if len(sig) != 65 {
		return "", fmt.Errorf("invalid signature length")
	}

	// Adjust recovery ID (v - 27)
	if sig[64] >= 27 {
		sig[64] -= 27
	}

	// Recover public key
	pubKey, err := crypto.Ecrecover(hash.Bytes(), sig)
	if err != nil {
		return "", err
	}

	// Convert public key to address
	pubKeyECDSA, err := crypto.UnmarshalPubkey(pubKey)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal public key: %w", err)
	}
	recoveredAddr := crypto.PubkeyToAddress(*pubKeyECDSA).Hex()

	// Verify address matches expected
	if recoveredAddr != expectedAddr {
		return "", fmt.Errorf("signature mismatch: expected %s, got %s", expectedAddr, recoveredAddr)
	}
	return recoveredAddr, nil
}
