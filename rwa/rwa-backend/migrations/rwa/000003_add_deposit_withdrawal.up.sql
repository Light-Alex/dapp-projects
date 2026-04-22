-- deposit_operations table
CREATE TABLE IF NOT EXISTS deposit_operations (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL,
    operation_id VARCHAR(66),
    usdc_amount NUMERIC(38,18) NOT NULL,
    usdm_amount NUMERIC(38,18) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    tx_hash VARCHAR(66),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_deposit_ops_account ON deposit_operations(account_id);
CREATE INDEX idx_deposit_ops_status ON deposit_operations(status);

-- withdrawal_operations table
CREATE TABLE IF NOT EXISTS withdrawal_operations (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL,
    operation_id VARCHAR(66),
    usdm_amount NUMERIC(38,18) NOT NULL,
    usdc_amount NUMERIC(38,18) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    tx_hash VARCHAR(66),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_withdrawal_ops_account ON withdrawal_operations(account_id);
CREATE INDEX idx_withdrawal_ops_status ON withdrawal_operations(status);

-- contract_transactions table (tracks on-chain tx sent by backend)
CREATE TABLE IF NOT EXISTS contract_transactions (
    id BIGSERIAL PRIMARY KEY,
    tx_hash VARCHAR(66) NOT NULL UNIQUE,
    method_name VARCHAR(100) NOT NULL,
    contract_address VARCHAR(42) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    gas_used BIGINT,
    error_message TEXT,
    related_order_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ
);
CREATE INDEX idx_contract_tx_status ON contract_transactions(status);
CREATE INDEX idx_contract_tx_order ON contract_transactions(related_order_id);
