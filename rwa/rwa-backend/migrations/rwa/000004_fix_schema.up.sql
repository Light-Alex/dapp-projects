-- Add missing fields to orders table
ALTER TABLE orders ADD COLUMN IF NOT EXISTS time_in_force VARCHAR(10) DEFAULT 'DAY';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS escrow_amount NUMERIC(38,18);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS escrow_asset VARCHAR(42);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS refund_amount NUMERIC(38,18);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS execute_tx_hash VARCHAR(66);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS cancel_tx_hash VARCHAR(66);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS accepted_at TIMESTAMPTZ;

-- Fix order_executions: VARCHAR -> NUMERIC for price/quantity
ALTER TABLE order_executions ALTER COLUMN quantity TYPE NUMERIC(38,18) USING quantity::NUMERIC(38,18);
ALTER TABLE order_executions ALTER COLUMN price TYPE NUMERIC(38,18) USING price::NUMERIC(38,18);

-- Add unique constraints for idempotency
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uq_orders_client_order_id') THEN
        ALTER TABLE orders ADD CONSTRAINT uq_orders_client_order_id UNIQUE (client_order_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uq_order_executions_execution_id') THEN
        ALTER TABLE order_executions ADD CONSTRAINT uq_order_executions_execution_id UNIQUE (execution_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uq_event_logs_tx_log') THEN
        ALTER TABLE event_logs ADD CONSTRAINT uq_event_logs_tx_log UNIQUE (tx_hash, log_index);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uq_positions_account_symbol') THEN
        ALTER TABLE positions ADD CONSTRAINT uq_positions_account_symbol UNIQUE (account_id, symbol);
    END IF;
END $$;

-- Use TIMESTAMPTZ for timestamps
ALTER TABLE orders ALTER COLUMN created_at TYPE TIMESTAMPTZ;
ALTER TABLE orders ALTER COLUMN updated_at TYPE TIMESTAMPTZ;
ALTER TABLE orders ALTER COLUMN submitted_at TYPE TIMESTAMPTZ;
ALTER TABLE orders ALTER COLUMN filled_at TYPE TIMESTAMPTZ;
ALTER TABLE orders ALTER COLUMN cancelled_at TYPE TIMESTAMPTZ;
ALTER TABLE orders ALTER COLUMN expired_at TYPE TIMESTAMPTZ;
