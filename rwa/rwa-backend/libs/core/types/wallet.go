package types

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
)

type Wallet struct {
	Address    common.Address    `json:"address"`
	PrivateKey *ecdsa.PrivateKey `json:"-"`
}
