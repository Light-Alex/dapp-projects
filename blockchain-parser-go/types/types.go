package types

import (
	"math/big"
	"time"
)

// ==================== Core Data Models ====================

// Account represents a user account in the system
type Account struct {
	ID          int       `json:"id"`
	Email       string    `json:"email"`
	Address     string    `json:"address"`
	BNBAmount   *big.Int  `json:"bnb_amount"`
	USDTAmount  *big.Int  `json:"usdt_amount"`
	CreatedTime time.Time `json:"created_time"`
	UpdatedTime time.Time `json:"updated_time"`
}

// Transaction represents a blockchain transaction record
type Transaction struct {
	ID           int       `json:"id"`
	TxHash       string    `json:"tx_hash"`
	BlockNumber  int64     `json:"block_number"`
	FromAddress  string    `json:"from_address"`
	ToAddress    string    `json:"to_address"`
	Value        *big.Int  `json:"value"`
	TokenDecimal uint8     `json:"token_decimals"`
	TokenAddress string    `json:"token_address"`
	TokenSymbol  string    `json:"token_symbol"`
	Status       int       `json:"status"`
	CreatedTime  time.Time `json:"created_time"`
}

// Withdrawal represents a withdrawal request
type Withdrawal struct {
	ID           int       `json:"id"`
	AccountID    int       `json:"account_id"`
	Amount       *big.Int  `json:"amount"`
	TokenDecimal uint8     `json:"token_decimals"`
	TokenSymbol  string    `json:"token_symbol"`
	ToAddress    string    `json:"to_address"`
	TxHash       string    `json:"tx_hash"`
	Status       string    `json:"status"`
	CreatedTime  time.Time `json:"created_time"`
	UpdatedTime  time.Time `json:"updated_time"`
}

// ==================== Database Models ====================

// AccountDB represents the database-specific account structure
// Uses string types for big.Int values to handle database storage
type AccountDB struct {
	ID          int       `db:"id"`
	Email       string    `db:"email"`
	Address     string    `db:"address"`
	BNBAmount   string    `db:"bnb_amount"`
	USDTAmount  string    `db:"usdt_amount"`
	CreatedTime time.Time `db:"created_time"`
	UpdatedTime time.Time `db:"updated_time"`
}

// TransactionDB represents the database-specific transaction structure
type TransactionDB struct {
	ID            int       `db:"id"`
	TxHash        string    `db:"tx_hash"`
	BlockNumber   int64     `db:"block_number"`
	FromAddress   string    `db:"from_address"`
	ToAddress     string    `db:"to_address"`
	Value         string    `db:"value"`
	TokenDecimals uint8     `db:"token_decimals"`
	TokenAddress  string    `db:"token_address"`
	TokenSymbol   string    `db:"token_symbol"`
	Status        int       `db:"status"`
	CreatedTime   time.Time `db:"created_time"`
}

// WithdrawalDB represents the database-specific withdrawal structure
type WithdrawalDB struct {
	ID            int       `db:"id"`
	AccountID     int       `db:"account_id"`
	Amount        string    `db:"amount"`
	TokenDecimals uint8     `db:"token_decimals"`
	TokenSymbol   string    `db:"token_symbol"`
	ToAddress     string    `db:"to_address"`
	TxHash        string    `db:"tx_hash"`
	Status        string    `db:"status"`
	CreatedTime   time.Time `db:"created_time"`
	UpdatedTime   time.Time `db:"updated_time"`
}

// ==================== Database Schema Constants ====================

// CreateAccountTable is the SQL statement to create the account table
const CreateAccountTable = `
CREATE TABLE IF NOT EXISTS account (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    address VARCHAR(42) UNIQUE NOT NULL,
    bnb_amount VARCHAR(100) DEFAULT '0',
    usdt_amount VARCHAR(100) DEFAULT '0',
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`

// CreateTransactionTable is the SQL statement to create the transaction table
const CreateTransactionTable = `
CREATE TABLE IF NOT EXISTS transaction (
    id SERIAL PRIMARY KEY,
    tx_hash VARCHAR(66) UNIQUE NOT NULL,
    block_number BIGINT NOT NULL,
    from_address VARCHAR(42) NOT NULL,
    to_address VARCHAR(42) NOT NULL,
    value VARCHAR(100) NOT NULL,
    token_decimals INT DEFAULT 18,
    token_address VARCHAR(42),
    token_symbol VARCHAR(32),
    status INT DEFAULT 0,
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`

// CreateWithdrawalTable is the SQL statement to create the withdrawal table
const CreateWithdrawalTable = `
CREATE TABLE IF NOT EXISTS withdrawal (
    id SERIAL PRIMARY KEY,
    account_id INT NOT NULL REFERENCES account(id),
    amount VARCHAR(100) NOT NULL,
    token_decimals INT DEFAULT 18,
    token_symbol VARCHAR(32) NOT NULL,
    to_address VARCHAR(42) NOT NULL,
    tx_hash VARCHAR(66) DEFAULT '',
    status VARCHAR(32) DEFAULT 'init',
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`

// CreateIndexes creates database indexes for better query performance
const CreateIndexes = `
CREATE INDEX IF NOT EXISTS idx_account_address ON account(address);
CREATE INDEX IF NOT EXISTS idx_tx_hash ON transaction(tx_hash);
CREATE INDEX IF NOT EXISTS idx_tx_from ON transaction(from_address);
CREATE INDEX IF NOT EXISTS idx_tx_to ON transaction(to_address);
CREATE INDEX IF NOT EXISTS idx_tx_block ON transaction(block_number);
CREATE INDEX IF NOT EXISTS idx_withdrawal_status ON withdrawal(status);
CREATE INDEX IF NOT EXISTS idx_withdrawal_account ON withdrawal(account_id);
`

// ==================== Transaction Status Constants ====================

// TransactionStatus represents the status of a transaction
type TransactionStatus int

const (
	// TransactionStatusPending indicates the transaction is pending
	TransactionStatusPending TransactionStatus = 0
	// TransactionStatusSuccess indicates the transaction succeeded
	TransactionStatusSuccess TransactionStatus = 1
	// TransactionStatusFailed indicates the transaction failed
	TransactionStatusFailed TransactionStatus = 2
)

// String returns the string representation of the transaction status
func (s TransactionStatus) String() string {
	switch s {
	case TransactionStatusPending:
		return "pending"
	case TransactionStatusSuccess:
		return "success"
	case TransactionStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// ==================== Withdrawal Status Constants ====================

// WithdrawalStatus represents the status of a withdrawal request
type WithdrawalStatus string

const (
	// WithdrawalStatusInit indicates the withdrawal is initialized
	WithdrawalStatusInit WithdrawalStatus = "init"
	// WithdrawalStatusProcessing indicates the withdrawal is being processed
	WithdrawalStatusProcessing WithdrawalStatus = "processing"
	// WithdrawalStatusSuccess indicates the withdrawal succeeded
	WithdrawalStatusSuccess WithdrawalStatus = "success"
	// WithdrawalStatusFailed indicates the withdrawal failed
	WithdrawalStatusFailed WithdrawalStatus = "failed"
)

// IsValidWithdrawalStatus checks if the given status is valid
func IsValidWithdrawalStatus(status string) bool {
	switch WithdrawalStatus(status) {
	case WithdrawalStatusInit, WithdrawalStatusProcessing, WithdrawalStatusSuccess, WithdrawalStatusFailed:
		return true
	default:
		return false
	}
}

// ==================== API Response Types ====================

// APIResponse represents a standard API response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	ErrorCode string      `json:"error_code,omitempty"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// PaginationResponse represents a paginated response
type PaginationResponse struct {
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// ==================== Custom Error Types ====================

// DatabaseError represents a database operation error
type DatabaseError struct {
	Query string
	Err   error
}

// Error returns the error message
func (e *DatabaseError) Error() string {
	return "database error in query '" + e.Query + "': " + e.Err.Error()
}

// Unwrap returns the underlying error
func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error returns the error message
func (e *ValidationError) Error() string {
	return "validation error on field '" + e.Field + "': " + e.Message
}

// BlockchainError represents a blockchain-related error
type BlockchainError struct {
	Operation string
	TxHash    string
	Err       error
}

// Error returns the error message
func (e *BlockchainError) Error() string {
	if e.TxHash != "" {
		return "blockchain error in operation '" + e.Operation + "' (tx: " + e.TxHash + "): " + e.Err.Error()
	}
	return "blockchain error in operation '" + e.Operation + "': " + e.Err.Error()
}

// Unwrap returns the underlying error
func (e *BlockchainError) Unwrap() error {
	return e.Err
}

// ==================== Token Type Constants ====================

// TokenSymbol represents supported token symbols
type TokenSymbol string

const (
	// TokenSymbolBNB represents BNB token
	TokenSymbolBNB TokenSymbol = "BNB"
	// TokenSymbolUSDT represents USDT token
	TokenSymbolUSDT TokenSymbol = "USDT"
)

// IsValidTokenSymbol checks if the token symbol is supported
func IsValidTokenSymbol(symbol string) bool {
	switch TokenSymbol(symbol) {
	case TokenSymbolBNB, TokenSymbolUSDT:
		return true
	default:
		return false
	}
}

// ==================== Conversion Helper Functions ====================

// AccountDBToAccount converts AccountDB to Account
func AccountDBToAccount(db *AccountDB) (*Account, error) {
	bnbAmount := new(big.Int)
	usdtAmount := new(big.Int)

	if _, ok := bnbAmount.SetString(db.BNBAmount, 10); !ok {
		return nil, &ValidationError{
			Field:   "bnb_amount",
			Message: "invalid bnb_amount format",
		}
	}

	if _, ok := usdtAmount.SetString(db.USDTAmount, 10); !ok {
		return nil, &ValidationError{
			Field:   "usdt_amount",
			Message: "invalid usdt_amount format",
		}
	}

	return &Account{
		ID:          db.ID,
		Email:       db.Email,
		Address:     db.Address,
		BNBAmount:   bnbAmount,
		USDTAmount:  usdtAmount,
		CreatedTime: db.CreatedTime,
		UpdatedTime: db.UpdatedTime,
	}, nil
}

// TransactionDBToTransaction converts TransactionDB to Transaction
func TransactionDBToTransaction(db *TransactionDB) (*Transaction, error) {
	value := new(big.Int)

	if _, ok := value.SetString(db.Value, 10); !ok {
		return nil, &ValidationError{
			Field:   "value",
			Message: "invalid value format",
		}
	}

	return &Transaction{
		ID:           db.ID,
		TxHash:       db.TxHash,
		BlockNumber:  db.BlockNumber,
		FromAddress:  db.FromAddress,
		ToAddress:    db.ToAddress,
		Value:        value,
		TokenDecimal: db.TokenDecimals,
		TokenAddress: db.TokenAddress,
		TokenSymbol:  db.TokenSymbol,
		Status:       db.Status,
		CreatedTime:  db.CreatedTime,
	}, nil
}

// WithdrawalDBToWithdrawal converts WithdrawalDB to Withdrawal
func WithdrawalDBToWithdrawal(db *WithdrawalDB) (*Withdrawal, error) {
	amount := new(big.Int)

	if _, ok := amount.SetString(db.Amount, 10); !ok {
		return nil, &ValidationError{
			Field:   "amount",
			Message: "invalid amount format",
		}
	}

	return &Withdrawal{
		ID:           db.ID,
		AccountID:    db.AccountID,
		Amount:       amount,
		TokenDecimal: db.TokenDecimals,
		TokenSymbol:  db.TokenSymbol,
		ToAddress:    db.ToAddress,
		TxHash:       db.TxHash,
		Status:       db.Status,
		CreatedTime:  db.CreatedTime,
		UpdatedTime:  db.UpdatedTime,
	}, nil
}

// AccountToAccountDB converts Account to AccountDB
func AccountToAccountDB(account *Account) *AccountDB {
	return &AccountDB{
		ID:          account.ID,
		Email:       account.Email,
		Address:     account.Address,
		BNBAmount:   account.BNBAmount.String(),
		USDTAmount:  account.USDTAmount.String(),
		CreatedTime: account.CreatedTime,
		UpdatedTime: account.UpdatedTime,
	}
}

// TransactionToTransactionDB converts Transaction to TransactionDB
func TransactionToTransactionDB(tx *Transaction) *TransactionDB {
	return &TransactionDB{
		ID:            tx.ID,
		TxHash:        tx.TxHash,
		BlockNumber:   tx.BlockNumber,
		FromAddress:   tx.FromAddress,
		ToAddress:     tx.ToAddress,
		Value:         tx.Value.String(),
		TokenDecimals: tx.TokenDecimal,
		TokenAddress:  tx.TokenAddress,
		TokenSymbol:   tx.TokenSymbol,
		Status:        tx.Status,
		CreatedTime:   tx.CreatedTime,
	}
}

// WithdrawalToWithdrawalDB converts Withdrawal to WithdrawalDB
func WithdrawalToWithdrawalDB(withdrawal *Withdrawal) *WithdrawalDB {
	return &WithdrawalDB{
		ID:            withdrawal.ID,
		AccountID:     withdrawal.AccountID,
		Amount:        withdrawal.Amount.String(),
		TokenDecimals: withdrawal.TokenDecimal,
		TokenSymbol:   withdrawal.TokenSymbol,
		ToAddress:     withdrawal.ToAddress,
		TxHash:        withdrawal.TxHash,
		Status:        withdrawal.Status,
		CreatedTime:   withdrawal.CreatedTime,
		UpdatedTime:   withdrawal.UpdatedTime,
	}
}
