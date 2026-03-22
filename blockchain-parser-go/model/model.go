package types

import (
	"math/big"
	"time"
)

type Account struct {
	ID          int       `json:"id"`
	Email       string    `json:"email"`
	Address     string    `json:"address"`
	BNBAmount   *big.Int  `json:"bnb_amount"`
	USDTAmount  *big.Int  `json:"usdt_amount"`
	CreatedTime time.Time `json:"created_time"`
	UpdatedTime time.Time `json:"updated_time"`
}

type Transaction struct {
	ID           int       `json:"id"`
	TxHash       string    `json:"tx_hash"`
	BlockNumber  int64     `json:"block_number"`
	FromAddress  string    `json:"from_address"`
	ToAddress    string    `json:"to_address"`
	Value        *big.Int  `json:"value"`
	TokenAddress string    `json:"token_address"`
	TokenSymbol  string    `json:"token_symbol"`
	Status       int       `json:"status"`
	CreatedTime  time.Time `json:"created_time"`
}

type Withdrawal struct {
	ID          int       `json:"id"`
	AccountID   int       `json:"account_id"`
	Amount      *big.Int  `json:"amount"`
	TokenSymbol string    `json:"token_symbol"`
	ToAddress   string    `json:"to_address"`
	TxHash      string    `json:"tx_hash"`
	Status      string    `json:"status"`
	CreatedTime time.Time `json:"created_time"`
	UpdatedTime time.Time `json:"updated_time"`
}

type AccountDB struct {
	ID          int       `db:"id"`
	Email       string    `db:"email"`
	Address     string    `db:"address"`
	BNBAmount   string    `db:"bnb_amount"`
	USDTAmount  string    `db:"usdt_amount"`
	CreatedTime time.Time `db:"created_time"`
	UpdatedTime time.Time `db:"updated_time"`
}

type TransactionDB struct {
	ID           int       `db:"id"`
	TxHash       string    `db:"tx_hash"`
	BlockNumber  int64     `db:"block_number"`
	FromAddress  string    `db:"from_address"`
	ToAddress    string    `db:"to_address"`
	Value        string    `db:"value"`
	TokenAddress string    `db:"token_address"`
	TokenSymbol  string    `db:"token_symbol"`
	Status       int       `db:"status"`
	CreatedTime  time.Time `db:"created_time"`
}
