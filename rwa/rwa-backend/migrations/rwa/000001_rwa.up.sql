-- Create account table
CREATE TABLE IF NOT EXISTS account (
    id BIGSERIAL PRIMARY KEY,
    address VARCHAR(255) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_account_address ON account (address);

-- Create stocks table
CREATE TABLE IF NOT EXISTS stocks (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    exchange VARCHAR(255) NOT NULL,
    about TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    contract VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_stocks_symbol ON stocks (symbol);

-- Create trading_accounts table
CREATE TABLE IF NOT EXISTS trading_accounts (
    id VARCHAR(255) PRIMARY KEY,
    external_account_id VARCHAR(255),
    provider VARCHAR(255),
    account_type VARCHAR(255),
    status VARCHAR(255),
    is_active BOOLEAN NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create positions table
CREATE TABLE IF NOT EXISTS positions (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL,
    symbol VARCHAR(255) NOT NULL,
    asset_type VARCHAR(255),
    quantity NUMERIC NOT NULL,
    average_price NUMERIC NOT NULL,
    market_value NUMERIC NOT NULL,
    unrealized_pnl NUMERIC NOT NULL,
    realized_pnl NUMERIC NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_positions_account_id ON positions (account_id);

CREATE INDEX IF NOT EXISTS idx_positions_symbol ON positions (symbol);

-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    client_order_id VARCHAR(255) NOT NULL,
    account_id BIGINT NOT NULL,
    symbol VARCHAR(255) NOT NULL,
    asset_type VARCHAR(255) NOT NULL,
    side VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL,
    quantity NUMERIC NOT NULL,
    price NUMERIC NOT NULL,
    stop_price NUMERIC NOT NULL,
    status VARCHAR(50) NOT NULL,
    filled_quantity NUMERIC NOT NULL,
    filled_price NUMERIC NOT NULL,
    remaining_quantity NUMERIC NOT NULL,
    contract_tx_hash VARCHAR(255),
    event_log_id BIGINT,
    external_order_id VARCHAR(255),
    provider VARCHAR(255),
    commission VARCHAR(255),
    commission_asset VARCHAR(255),
    metadata TEXT,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    submitted_at TIMESTAMP,
    filled_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    expired_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_orders_client_order_id ON orders (client_order_id);

CREATE INDEX IF NOT EXISTS idx_orders_account_id ON orders (account_id);

CREATE INDEX IF NOT EXISTS idx_orders_symbol ON orders (symbol);

CREATE INDEX IF NOT EXISTS idx_orders_contract_tx_hash ON orders (contract_tx_hash);

CREATE INDEX IF NOT EXISTS idx_orders_event_log_id ON orders (event_log_id);

-- Create order_executions table
CREATE TABLE IF NOT EXISTS order_executions (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    execution_id VARCHAR(255),
    quantity VARCHAR(255),
    price VARCHAR(255),
    commission VARCHAR(255),
    commission_asset VARCHAR(255),
    provider VARCHAR(255),
    external_id VARCHAR(255),
    executed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_order_executions_order_id ON order_executions (order_id);

-- Create event_logs table
CREATE TABLE IF NOT EXISTS event_logs (
    id BIGSERIAL PRIMARY KEY,
    tx_hash VARCHAR(255) NOT NULL,
    block_number BIGINT NOT NULL,
    block_hash VARCHAR(255),
    log_index INTEGER NOT NULL,
    contract_address VARCHAR(255) NOT NULL,
    event_name VARCHAR(255),
    event_data TEXT,
    account_id BIGINT,
    order_id BIGINT,
    processed_at TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_event_logs_tx_hash ON event_logs (tx_hash);

CREATE INDEX IF NOT EXISTS idx_event_logs_block_number ON event_logs (block_number);

CREATE INDEX IF NOT EXISTS idx_event_logs_contract_address ON event_logs (contract_address);

CREATE INDEX IF NOT EXISTS idx_event_logs_account_id ON event_logs (account_id);

CREATE INDEX IF NOT EXISTS idx_event_logs_order_id ON event_logs (order_id);