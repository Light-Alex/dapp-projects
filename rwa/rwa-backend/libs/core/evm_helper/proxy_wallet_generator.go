package evm_helper

import (
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/AnchoredLabs/rwa-backend/libs/core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

// ProxyWalletGenerator proxy wallet generator
type ProxyWalletGenerator struct {
	mnemonic       string
	derivationPath string
	cachedWallets  map[string]*types.Wallet
	mutex          sync.RWMutex
}

// NewProxyWalletGenerator creates a new proxy wallet generator
func NewProxyWalletGenerator(mnemonic string, derivationPath string) *ProxyWalletGenerator {
	return &ProxyWalletGenerator{
		mnemonic:       mnemonic,
		derivationPath: derivationPath,
		cachedWallets:  make(map[string]*types.Wallet),
	}
}

// GetOrCreateWallet gets or creates a proxy wallet
func (pwg *ProxyWalletGenerator) GetOrCreateWallet(userAddress string) (*types.Wallet, error) {
	// Normalize user address
	normalizedUser := common.HexToAddress(userAddress).Hex()

	// Read lock to check cache
	pwg.mutex.RLock()
	if wallet, exists := pwg.cachedWallets[normalizedUser]; exists {
		pwg.mutex.RUnlock()
		return wallet, nil
	}
	pwg.mutex.RUnlock()

	// Write lock to create new wallet
	pwg.mutex.Lock()
	defer pwg.mutex.Unlock()

	// Double check
	if wallet, exists := pwg.cachedWallets[normalizedUser]; exists {
		return wallet, nil
	}

	// Generate proxy wallet
	wallet, err := pwg.generateProxyWallet(userAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to generate proxy wallet: %w", err)
	}

	// Cache wallet
	pwg.cachedWallets[normalizedUser] = wallet

	return wallet, nil
}

// generateProxyWallet generates a proxy wallet
func (pwg *ProxyWalletGenerator) generateProxyWallet(userAddress string) (*types.Wallet, error) {
	// Calculate proxy index from user address
	proxyIndex := pwg.calculateProxyIndex(userAddress)

	// Generate proxy address and private key
	proxyAddress, privateKey, err := pwg.generateProxyAddressAndKey(proxyIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to generate proxy address and key: %w", err)
	}

	// Create Wallet object
	wallet := &types.Wallet{
		Address:    proxyAddress,
		PrivateKey: privateKey,
	}

	return wallet, nil
}

// TODO
// calculateProxyIndex calculates proxy index
func (pwg *ProxyWalletGenerator) calculateProxyIndex(userAddress string) uint32 {
	// Use first 4 bytes of the parsed address to form a uint32
	addrBytes := common.HexToAddress(userAddress).Bytes() // 20 bytes
	index := binary.BigEndian.Uint32(addrBytes[0:4])
	// Ensure index is within reasonable range for BIP32 derivation
	// Use a large but safe range to minimize collisions
	return index % 0x7FFFFFFF // Max safe value for BIP32 non-hardened derivation
}

// generateProxyAddressAndKey generates proxy address and private key
func (pwg *ProxyWalletGenerator) generateProxyAddressAndKey(proxyIndex uint32) (common.Address, *ecdsa.PrivateKey, error) {
	// Generate seed from mnemonic
	seed, err := bip39.NewSeedWithErrorChecking(pwg.mnemonic, "")
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("failed to generate seed from mnemonic: %w", err)
	}

	// Create master key
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("failed to create master key: %w", err)
	}

	// Derivation path: m/44'/60'/0'/0/{proxyIndex}
	derivedKey, err := pwg.deriveKey(masterKey, proxyIndex)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("failed to derive key: %w", err)
	}

	// Generate Ethereum address and private key
	address, privateKey, err := pwg.generateEthereumAddressAndKey(derivedKey)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("failed to generate Ethereum address and key: %w", err)
	}

	return address, privateKey, nil
}

// deriveKey derives key from master key
func (pwg *ProxyWalletGenerator) deriveKey(masterKey *bip32.Key, proxyIndex uint32) (*bip32.Key, error) {
	// Simplified derivation path implementation
	// Assume path format is m/44'/60'/0'/0/{index}
	derivedKey, err := masterKey.NewChildKey(0x8000002C) // 44'
	if err != nil {
		return nil, err
	}

	derivedKey, err = derivedKey.NewChildKey(0x8000003C) // 60'
	if err != nil {
		return nil, err
	}

	derivedKey, err = derivedKey.NewChildKey(0x80000000) // 0'
	if err != nil {
		return nil, err
	}

	derivedKey, err = derivedKey.NewChildKey(0) // 0
	if err != nil {
		return nil, err
	}

	derivedKey, err = derivedKey.NewChildKey(proxyIndex) // {proxyIndex}
	if err != nil {
		return nil, err
	}

	return derivedKey, nil
}

// generateEthereumAddressAndKey generates Ethereum address and private key
func (pwg *ProxyWalletGenerator) generateEthereumAddressAndKey(key *bip32.Key) (common.Address, *ecdsa.PrivateKey, error) {
	// Generate ECDSA private key from BIP32 key
	privateKey, err := crypto.ToECDSA(key.Key)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("failed to convert to ECDSA private key: %w", err)
	}

	// Generate address from private key
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	return address, privateKey, nil
}
