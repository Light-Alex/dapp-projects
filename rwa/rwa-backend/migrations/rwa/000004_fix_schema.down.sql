-- Revert TIMESTAMPTZ back to TIMESTAMP
ALTER TABLE orders ALTER COLUMN created_at TYPE TIMESTAMP;
ALTER TABLE orders ALTER COLUMN updated_at TYPE TIMESTAMP;
ALTER TABLE orders ALTER COLUMN submitted_at TYPE TIMESTAMP;
ALTER TABLE orders ALTER COLUMN filled_at TYPE TIMESTAMP;
ALTER TABLE orders ALTER COLUMN cancelled_at TYPE TIMESTAMP;
ALTER TABLE orders ALTER COLUMN expired_at TYPE TIMESTAMP;

-- Drop unique constraints
ALTER TABLE positions DROP CONSTRAINT IF EXISTS uq_positions_account_symbol;
ALTER TABLE event_logs DROP CONSTRAINT IF EXISTS uq_event_logs_tx_log;
ALTER TABLE order_executions DROP CONSTRAINT IF EXISTS uq_order_executions_execution_id;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS uq_orders_client_order_id;

-- Revert order_executions: NUMERIC -> VARCHAR for price/quantity
ALTER TABLE order_executions ALTER COLUMN price TYPE VARCHAR(255) USING price::VARCHAR(255);
ALTER TABLE order_executions ALTER COLUMN quantity TYPE VARCHAR(255) USING quantity::VARCHAR(255);

-- Drop added columns from orders table
ALTER TABLE orders DROP COLUMN IF EXISTS accepted_at;
ALTER TABLE orders DROP COLUMN IF EXISTS cancel_tx_hash;
ALTER TABLE orders DROP COLUMN IF EXISTS execute_tx_hash;
ALTER TABLE orders DROP COLUMN IF EXISTS refund_amount;
ALTER TABLE orders DROP COLUMN IF EXISTS escrow_asset;
ALTER TABLE orders DROP COLUMN IF EXISTS escrow_amount;
ALTER TABLE orders DROP COLUMN IF EXISTS time_in_force;
