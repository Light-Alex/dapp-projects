package database

import (
	"database/sql"
	"fmt"
	"log"
	"math/big"

	"blockchain-parser-go/types"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func NewDB(host string, port int, user, password, dbname string) (*DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Create tables
	if _, err := db.Exec(types.CreateAccountTable); err != nil {
		return nil, fmt.Errorf("error creating account table: %v", err)
	}

	if _, err := db.Exec(types.CreateTransactionTable); err != nil {
		return nil, fmt.Errorf("error creating transaction table: %v", err)
	}

	if _, err := db.Exec(types.CreateWithdrawalTable); err != nil {
		return nil, fmt.Errorf("error creating withdrawal table: %v", err)
	}

	if _, err := db.Exec(types.CreateIndexes); err != nil {
		return nil, fmt.Errorf("error creating indexes: %v", err)
	}

	log.Println("Database connected and tables initialized")
	return &DB{db}, nil
}

func (db *DB) GetAccountByAddress(address string) (*types.Account, error) {
	var accountDB types.AccountDB
	query := `SELECT id, email, address, bnb_amount, usdt_amount, created_time, updated_time 
                  FROM account WHERE address = $1`

	err := db.QueryRow(query, address).Scan(
		&accountDB.ID, &accountDB.Email, &accountDB.Address,
		&accountDB.BNBAmount, &accountDB.USDTAmount,
		&accountDB.CreatedTime, &accountDB.UpdatedTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	bnbAmount := new(big.Int)
	usdtAmount := new(big.Int)

	if _, ok := bnbAmount.SetString(accountDB.BNBAmount, 10); !ok {
		return nil, fmt.Errorf("invalid bnb_amount format")
	}
	if _, ok := usdtAmount.SetString(accountDB.USDTAmount, 10); !ok {
		return nil, fmt.Errorf("invalid usdt_amount format")
	}

	return &types.Account{
		ID:          accountDB.ID,
		Email:       accountDB.Email,
		Address:     accountDB.Address,
		BNBAmount:   bnbAmount,
		USDTAmount:  usdtAmount,
		CreatedTime: accountDB.CreatedTime,
		UpdatedTime: accountDB.UpdatedTime,
	}, nil
}

func (db *DB) TransactionExists(txHash string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM transaction WHERE tx_hash = $1`
	err := db.QueryRow(query, txHash).Scan(&count)
	return count > 0, err
}

func (db *DB) SaveTransaction(tx *types.Transaction) error {
	query := `INSERT INTO transaction 
                  (tx_hash, block_number, from_address, to_address, value, token_address, token_symbol, status) 
                  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := db.Exec(query,
		tx.TxHash, tx.BlockNumber, tx.FromAddress, tx.ToAddress,
		tx.Value.String(), tx.TokenAddress, tx.TokenSymbol, tx.Status,
	)
	return err
}

func (db *DB) CreateAccount(account *types.Account) error {
	query := `INSERT INTO account (email, address, bnb_amount, usdt_amount) 
                  VALUES ($1, $2, $3, $4)`

	_, err := db.Exec(query,
		account.Email, account.Address,
		account.BNBAmount.String(), account.USDTAmount.String(),
	)
	return err
}

func (db *DB) UpdateAccount(account *types.Account) error {
	query := `UPDATE account SET email = $1, bnb_amount = $2, usdt_amount = $3, updated_time = CURRENT_TIMESTAMP WHERE id = $4`

	_, err := db.Exec(query,
		account.Email, account.BNBAmount.String(), account.USDTAmount.String(), account.ID,
	)
	return err
}

func (db *DB) UpdateAccountBalance(accountID int, tokenSymbol string, amount *big.Int) error {
	var query string
	if tokenSymbol == "BNB" {
		// 将 VARCHAR 转换为 NUMERIC 进行计算，然后再转回 TEXT
		query = `UPDATE account SET bnb_amount = ((CAST(bnb_amount AS NUMERIC) + CAST($1 AS NUMERIC))::TEXT), updated_time = CURRENT_TIMESTAMP WHERE id = $2`
	} else if tokenSymbol == "USDT" {
		query = `UPDATE account SET usdt_amount = ((CAST(usdt_amount AS NUMERIC) + CAST($1 AS NUMERIC))::TEXT), updated_time = CURRENT_TIMESTAMP WHERE id = $2`
	} else {
		return fmt.Errorf("unsupported token symbol: %s", tokenSymbol)
	}

	_, err := db.Exec(query, amount.String(), accountID)
	return err
}

func (db *DB) GetPendingWithdrawals() ([]*types.Withdrawal, error) {
	query := `SELECT id, account_id, amount, token_decimals, token_symbol, to_address, tx_hash, status, created_time, updated_time 
                  FROM withdrawal WHERE status = 'init'`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []*types.Withdrawal
	for rows.Next() {
		var w types.Withdrawal
		var amountStr string

		err := rows.Scan(
			&w.ID, &w.AccountID, &amountStr, &w.TokenDecimal, &w.TokenSymbol, &w.ToAddress,
			&w.TxHash, &w.Status, &w.CreatedTime, &w.UpdatedTime,
		)
		if err != nil {
			return nil, err
		}

		w.Amount = new(big.Int)
		if _, ok := w.Amount.SetString(amountStr, 10); !ok {
			return nil, fmt.Errorf("invalid amount format")
		}

		withdrawals = append(withdrawals, &w)
	}

	return withdrawals, nil
}

func (db *DB) WithdrawalCounts() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM withdrawal`
	err := db.QueryRow(query).Scan(&count)
	return count, err
}

func (db *DB) CreateWithdrawal(withdrawal *types.Withdrawal) error {
	query := `INSERT INTO withdrawal (account_id, amount, token_decimals, token_symbol, to_address) 
                  VALUES ($1, $2, $3, $4, $5)`

	_, err := db.Exec(query,
		withdrawal.AccountID, withdrawal.Amount.String(),
		withdrawal.TokenDecimal, withdrawal.TokenSymbol, withdrawal.ToAddress,
	)
	return err
}

func (db *DB) UpdateWithdrawalAllStatus() error {
	query := `UPDATE withdrawal SET status = 'init'`
	_, err := db.Exec(query)
	return err
}

func (db *DB) UpdateWithdrawalStatus(id int, status, txHash string) error {
	query := `UPDATE withdrawal SET status = $1, tx_hash = $2, updated_time = CURRENT_TIMESTAMP WHERE id = $3`
	_, err := db.Exec(query, status, txHash, id)
	return err
}
